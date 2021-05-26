package trojan

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/p4gefau1t/trojan-go/api"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/redirector"
	"github.com/p4gefau1t/trojan-go/statistic"
	"github.com/p4gefau1t/trojan-go/statistic/memory"
	"github.com/p4gefau1t/trojan-go/statistic/mysql"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/mux"
)

var auth statistic.Authenticator

// InboundConn is a trojan inbound connection
type InboundConn struct {
	net.Conn
	sent     uint64
	recv     uint64
	auth     statistic.Authenticator
	user     statistic.User
	hash     string
	metadata *tunnel.Metadata
	ip       string
}

func (c *InboundConn) Metadata() *tunnel.Metadata {
	return c.metadata
}

func (c *InboundConn) Write(p []byte) (int, error) {
	n, err := c.Conn.Write(p)
	c.sent += uint64(n)
	c.user.AddTraffic(n, 0)
	return n, err
}

func (c *InboundConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	c.recv += uint64(n)
	c.user.AddTraffic(0, n)
	return n, err
}

func (c *InboundConn) Close() error {
	log.Info("user", c.hash, "from", c.Conn.RemoteAddr(), "tunneling to", c.metadata.Address, "closed", "sent:", common.HumanFriendlyTraffic(c.sent), "recv:", common.HumanFriendlyTraffic(c.recv))
	c.user.DelIP(c.ip)
	return c.Conn.Close()
}

func (c *InboundConn) Auth() error {
	userHash := [56]byte{}
	n, err := c.Conn.Read(userHash[:])
	if err != nil || n != 56 {
		return common.NewError("failed to read hash").Base(err)
	}

	valid, user := c.auth.AuthUser(string(userHash[:]))
	if !valid {
		return common.NewError("invalid hash:" + string(userHash[:]))
	}
	c.hash = string(userHash[:])
	c.user = user

	ip, _, err := net.SplitHostPort(c.Conn.RemoteAddr().String())
	if err != nil {
		return common.NewError("failed to parse host:" + c.Conn.RemoteAddr().String()).Base(err)
	}

	c.ip = ip
	ok := user.AddIP(ip)
	if !ok {
		return common.NewError("ip limit reached")
	}

	crlf := [2]byte{}
	_, err = io.ReadFull(c.Conn, crlf[:])
	if err != nil {
		return err
	}

	c.metadata = &tunnel.Metadata{}
	if err := c.metadata.ReadFrom(c.Conn); err != nil {
		return err
	}

	_, err = io.ReadFull(c.Conn, crlf[:])
	if err != nil {
		return err
	}
	return nil
}

// Server is a trojan tunnel server
type Server struct {
	auth       statistic.Authenticator
	redir      *redirector.Redirector
	redirAddr  *tunnel.Address
	underlay   tunnel.Server
	connChan   chan tunnel.Conn
	muxChan    chan tunnel.Conn
	packetChan chan tunnel.PacketConn
	ctx        context.Context
	cancel     context.CancelFunc
}

func (s *Server) Close() error {
	s.cancel()
	return s.underlay.Close()
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.underlay.AcceptConn(&Tunnel{})
		if err != nil { // Closing
			log.Error(common.NewError("trojan failed to accept conn").Base(err))
			select {
			case <-s.ctx.Done():
				return
			default:
			}
			continue
		}
		go func(conn tunnel.Conn) {
			rewindConn := common.NewRewindConn(conn)
			rewindConn.SetBufferSize(128)
			defer rewindConn.StopBuffering()

			inboundConn := &InboundConn{
				Conn: rewindConn,
				auth: s.auth,
			}

			if err := inboundConn.Auth(); err != nil {
				rewindConn.Rewind()
				rewindConn.StopBuffering()
				log.Warn(common.NewError("connection with invalid trojan header from " + rewindConn.RemoteAddr().String()).Base(err))
				s.redir.Redirect(&redirector.Redirection{
					RedirectTo:  s.redirAddr,
					InboundConn: rewindConn,
				})
				return
			}

			rewindConn.StopBuffering()
			switch inboundConn.metadata.Command {
			case Connect:
				if inboundConn.metadata.DomainName == "MUX_CONN" {
					s.muxChan <- inboundConn
					log.Debug("mux(r) connection")
				} else {
					s.connChan <- inboundConn
					log.Debug("normal trojan connection")
				}

			case Associate:
				s.packetChan <- &PacketConn{
					Conn: inboundConn,
				}
				log.Debug("trojan udp connection")
			case Mux:
				s.muxChan <- inboundConn
				log.Debug("mux connection")
			default:
				log.Error(common.NewError(fmt.Sprintf("unknown trojan command %d", inboundConn.metadata.Command)))
			}
		}(conn)
	}
}

func (s *Server) AcceptConn(nextTunnel tunnel.Tunnel) (tunnel.Conn, error) {
	switch nextTunnel.(type) {
	case *mux.Tunnel:
		select {
		case t := <-s.muxChan:
			return t, nil
		case <-s.ctx.Done():
			return nil, common.NewError("trojan client closed")
		}
	default:
		select {
		case t := <-s.connChan:
			return t, nil
		case <-s.ctx.Done():
			return nil, common.NewError("trojan client closed")
		}
	}
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	select {
	case t := <-s.packetChan:
		return t, nil
	case <-s.ctx.Done():
		return nil, common.NewError("trojan client closed")
	}
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	ctx, cancel := context.WithCancel(ctx)

	if auth == nil {
		var err error
		if cfg.MySQL.Enabled {
			log.Debug("mysql enabled")
			auth, err = statistic.NewAuthenticator(ctx, mysql.Name)
		} else {
			log.Debug("auth by config file")
			auth, err = statistic.NewAuthenticator(ctx, memory.Name)
		}
		if err != nil {
			cancel()
			return nil, common.NewError("trojan failed to create authenticator")
		}
	}

	if cfg.API.Enabled {
		go api.RunService(ctx, Name+"_SERVER", auth)
	}

	redirAddr := tunnel.NewAddressFromHostPort("tcp", cfg.RemoteHost, cfg.RemotePort)
	s := &Server{
		underlay:   underlay,
		auth:       auth,
		redirAddr:  redirAddr,
		connChan:   make(chan tunnel.Conn, 32),
		muxChan:    make(chan tunnel.Conn, 32),
		packetChan: make(chan tunnel.PacketConn, 32),
		ctx:        ctx,
		cancel:     cancel,
		redir:      redirector.NewRedirector(ctx),
	}

	if !cfg.DisableHTTPCheck {
		redirConn, err := net.Dial("tcp", redirAddr.String())
		if err != nil {
			cancel()
			return nil, common.NewError("invalid redirect address. check your http server: " + redirAddr.String()).Base(err)
		}
		redirConn.Close()
	}

	go s.acceptLoop()
	log.Debug("trojan server created")
	return s, nil
}
