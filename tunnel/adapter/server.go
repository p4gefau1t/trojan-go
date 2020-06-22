package adapter

import (
	"context"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/http"
	"github.com/p4gefau1t/trojan-go/tunnel/socks"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
)

type Server struct {
	underlay  tunnel.Server
	socksConn chan tunnel.Conn
	httpConn  chan tunnel.Conn
	nextSocks bool
	ctx       context.Context
	cancel    context.CancelFunc
}

func (s *Server) acceptConnLoop() {
	for {
		conn, err := s.underlay.AcceptConn(&Tunnel{})
		if err != nil {
			select {
			case <-s.ctx.Done():
				log.Debug("exiting")
				return
			default:
				continue
			}
		}
		rewindConn := common.NewRewindConn(conn)
		rewindConn.SetBufferSize(16)
		buf := [3]byte{}
		_, err = rewindConn.Read(buf[:])
		rewindConn.Rewind()
		rewindConn.StopBuffering()
		if err != nil {
			log.Error(common.NewError("failed to detect proxy protocol type").Base(err))
			continue
		}
		if buf[0] == 5 && s.nextSocks {
			log.Debug("socks5 connection")
			s.socksConn <- &transport.Conn{
				Conn: rewindConn,
			}
		} else {
			log.Debug("http connection")
			s.httpConn <- &transport.Conn{
				Conn: rewindConn,
			}
		}
	}
}

func (s *Server) AcceptConn(overlay tunnel.Tunnel) (tunnel.Conn, error) {
	if _, ok := overlay.(*http.Tunnel); ok {
		select {
		case conn := <-s.httpConn:
			return conn, nil
		case <-s.ctx.Done():
			return nil, common.NewError("adapter closed")
		}
	} else if _, ok := overlay.(*socks.Tunnel); ok {
		s.nextSocks = true
		select {
		case conn := <-s.socksConn:
			return conn, nil
		case <-s.ctx.Done():
			return nil, common.NewError("adapter closed")
		}
	} else {
		panic("invalid overlay")
	}
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	// no packet conn available, but it's ok to stuck here
	<-s.ctx.Done()
	return nil, common.NewError("adapter server closed")
}

func (s *Server) Close() error {
	s.cancel()
	return s.underlay.Close()
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
	ctx, cancel := context.WithCancel(ctx)
	server := &Server{
		underlay:  underlay,
		socksConn: make(chan tunnel.Conn, 32),
		httpConn:  make(chan tunnel.Conn, 32),
		ctx:       ctx,
		cancel:    cancel,
	}
	go server.acceptConnLoop()
	return server, nil
}
