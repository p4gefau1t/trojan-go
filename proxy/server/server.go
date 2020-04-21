package server

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"net"
	"reflect"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/direct"
	"github.com/p4gefau1t/trojan-go/protocol/simplesocks"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/p4gefau1t/trojan-go/proxy"
	"github.com/p4gefau1t/trojan-go/stat"
	"github.com/xtaci/smux"
)

type Server struct {
	common.Runnable
	proxy.Buildable

	listener net.Listener
	auth     stat.Authenticator
	meter    stat.TrafficMeter
	config   *conf.GlobalConfig
	ctx      context.Context
	cancel   context.CancelFunc
}

func (s *Server) handleMuxConn(stream *smux.Stream, passwordHash string) {
	inboundConn, err := simplesocks.NewInboundSimpleSocksConnSession(stream, passwordHash)
	if err != nil {
		stream.Close()
		log.Error(common.NewError("cannot start inbound session").Base(err))
		return
	}
	inboundConn.(protocol.NeedMeter).SetMeter(s.meter)
	defer inboundConn.Close()
	req := inboundConn.GetRequest()
	switch req.Command {
	case protocol.Connect:
		outboundConn, err := direct.NewOutboundConnSession(nil, req)
		if err != nil {
			log.Error(err)
			return
		}
		log.Info("user", passwordHash, "mux tunneling to", req.String())
		defer outboundConn.Close()
		proxy.ProxyConn(s.ctx, inboundConn, outboundConn)
	case protocol.Associate:
		outboundPacket, err := direct.NewOutboundPacketSession(s.ctx)
		common.Must(err)
		inboundPacket, err := trojan.NewPacketSession(inboundConn)
		proxy.ProxyPacket(s.ctx, inboundPacket, outboundPacket)
	default:
		log.Error(fmt.Sprintf("invalid command %d", req.Command))
		return
	}
}

func (s *Server) handleConn(conn net.Conn) {
	inboundConn, err := trojan.NewInboundConnSession(s.ctx, conn, s.config, s.auth)
	if err != nil {
		log.Error(common.NewError("failed to start inbound session, remote:" + conn.RemoteAddr().String()).Base(err))
		return
	}

	req := inboundConn.GetRequest()
	hash := inboundConn.(protocol.HasHash).GetHash()

	if req.Command == protocol.Mux {
		muxServer, err := smux.Server(inboundConn, nil)
		defer muxServer.Close()
		common.Must(err)
		for {
			stream, err := muxServer.AcceptStream()
			if err != nil {
				log.Debug("mux conn from", conn.RemoteAddr(), "closed:", err)
				return
			}
			go s.handleMuxConn(stream, hash)
		}
	}
	inboundConn.(protocol.NeedMeter).SetMeter(s.meter)

	if req.Command == protocol.Associate {
		inboundPacket, _ := trojan.NewPacketSession(inboundConn)
		defer inboundPacket.Close()

		outboundPacket, err := direct.NewOutboundPacketSession(s.ctx)
		if err != nil {
			log.Error(err)
			return
		}
		defer outboundPacket.Close()
		log.Info("UDP tunnel established")
		proxy.ProxyPacket(s.ctx, inboundPacket, outboundPacket)
		log.Debug("UDP tunnel closed")
		return
	}

	defer inboundConn.Close()
	outboundConn, err := direct.NewOutboundConnSession(nil, req)
	if err != nil {
		log.Error(err)
		return
	}
	defer outboundConn.Close()

	log.Info("conn from", conn.RemoteAddr(), "tunneling to", req.String())
	proxy.ProxyConn(s.ctx, inboundConn, outboundConn)
}

func (s *Server) handleInvalidConn(conn net.Conn, tlsConn *tls.Conn) {
	defer conn.Close()
	if len(s.config.TLS.HTTPResponse) > 0 {
		log.Warn("trying to response with a plain http response")
		conn.Write(s.config.TLS.HTTPResponse)
		return
	}

	if s.config.TLS.FallbackAddress != nil {
		defer func() {
			if r := recover(); r != nil {
				log.Error("recovered", r)
			}
		}()
		//HACK
		//obtain the bytes buffered by the tls conn
		v := reflect.ValueOf(*tlsConn)
		buf := v.FieldByName("rawInput").FieldByName("buf").Bytes()
		log.Debug("payload:" + string(buf))

		remote, err := net.Dial("tcp", s.config.TLS.FallbackAddress.String())
		if err != nil {
			log.Warn(common.NewError("failed to dial to tls fallback server").Base(err))
			return
		}
		log.Warn("proxying this invalid tls conn to the tls fallback server")
		remote.Write(buf)
		proxy.ProxyConn(s.ctx, conn, remote)
	} else {
		log.Warn("tls fallback port is unspecified, closing")
	}
}

func (s *Server) Run() error {
	var db *sql.DB
	var err error
	if s.config.MySQL.Enabled {
		db, err = common.ConnectDatabase(
			"mysql",
			s.config.MySQL.Username,
			s.config.MySQL.Password,
			s.config.MySQL.ServerHost,
			s.config.MySQL.ServerPort,
			s.config.MySQL.Database,
		)
		if err != nil {
			return common.NewError("failed to connect to database server").Base(err)
		}
	}
	if db == nil {
		s.auth = &stat.ConfigUserAuthenticator{
			Config: s.config,
		}
	} else {
		s.auth, err = stat.NewMixedAuthenticator(s.config, db)
		if err != nil {
			return common.NewError("failed to init auth").Base(err)
		}
		s.meter, err = stat.NewDBTrafficMeter(s.config, db)
		if err != nil {
			return common.NewError("failed to init traffic meter").Base(err)
		}
	}
	defer s.auth.Close()
	if s.meter != nil {
		defer s.meter.Close()
	}
	log.Info("server is running at", s.config.LocalAddress)

	var listener net.Listener
	if s.config.TCP.ReusePort || s.config.TCP.FastOpen || s.config.TCP.NoDelay {
		localIP, err := s.config.LocalAddress.ResolveIP(false)
		listener, err = ListenWithTCPOption(
			s.config.TCP.FastOpen,
			s.config.TCP.ReusePort,
			s.config.TCP.NoDelay,
			localIP,
			s.config.LocalAddress.String(),
		)
		if err != nil {
			return err
		}
	} else {
		listener, err = net.Listen("tcp", s.config.LocalAddress.String())
		if err != nil {
			return err
		}
	}
	s.listener = listener
	defer listener.Close()

	tlsConfig := &tls.Config{
		Certificates:             s.config.TLS.KeyPair,
		CipherSuites:             s.config.TLS.CipherSuites,
		PreferServerCipherSuites: s.config.TLS.PreferServerCipher,
		SessionTicketsDisabled:   !s.config.TLS.SessionTicket,
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return nil
			default:
			}
			return err
		}
		go func(conn net.Conn) {
			tlsConn := tls.Server(conn, tlsConfig)
			err = tlsConn.Handshake()
			if err != nil {
				log.Warn(common.NewError("failed to perform tls handshake, remote:" + conn.RemoteAddr().String()).Base(err))
				go s.handleInvalidConn(conn, tlsConn)
				return
			}
			go s.handleConn(tlsConn)
		}(conn)
	}
}

func (s *Server) Close() error {
	log.Info("shutting down server..")
	s.cancel()
	s.listener.Close()
	return nil
}

func (s *Server) Build(config *conf.GlobalConfig) (common.Runnable, error) {
	s.config = config
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s, nil
}

func init() {
	proxy.RegisterProxy(conf.Server, &Server{})
}
