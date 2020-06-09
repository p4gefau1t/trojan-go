package mux

import (
	"context"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/xtaci/smux"
)

// Server is a smux server
type Server struct {
	underlay tunnel.Server
	connChan chan tunnel.Conn
	errChan  chan error
	ctx      context.Context
	cancel   context.CancelFunc
}

func (s *Server) acceptConnWorker() {
	for {
		conn, err := s.underlay.AcceptConn(&Tunnel{})
		if err != nil {
			log.Debug(err)
			select {
			case <-s.ctx.Done():
				return
			default:
			}
			continue
		}
		smuxConfig := smux.DefaultConfig()
		smuxConfig.KeepAliveDisabled = true
		smuxSession, err := smux.Server(conn, smuxConfig)
		if err != nil {
			s.errChan <- err
			continue
		}
		// TODO context
		go func(session *smux.Session, conn tunnel.Conn) {
			defer session.Close()
			defer conn.Close()
			for {
				stream, err := session.AcceptStream()
				if err != nil {
					s.errChan <- err
					return
				}
				s.connChan <- &Conn{
					rwc:  stream,
					Conn: conn,
				}
			}
		}(smuxSession, conn)
	}
}

func (s *Server) AcceptConn(tunnel.Tunnel) (tunnel.Conn, error) {
	select {
	case conn := <-s.connChan:
		return conn, nil
	case err := <-s.errChan:
		return nil, err
	case <-s.ctx.Done():
		return nil, common.NewError("mux client closed")
	}
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

func (s *Server) Close() error {
	s.cancel()
	return s.underlay.Close()
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
	ctx, cancel := context.WithCancel(ctx)
	server := &Server{
		underlay: underlay,
		ctx:      ctx,
		cancel:   cancel,
		connChan: make(chan tunnel.Conn, 32),
	}
	go server.acceptConnWorker()
	log.Debug("mux server created")
	return server, nil
}
