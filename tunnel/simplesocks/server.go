package simplesocks

import (
	"context"
	"fmt"
	"github.com/p4gefau1t/trojan-go/log"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/trojan"
)

// Server is a simplesocks server
type Server struct {
	underlay   tunnel.Server
	connChan   chan tunnel.Conn
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
		if err != nil {
			log.Error(common.NewError("simplesocks failed to accept connection from underlying tunnel").Base(err))
			select {
			case <-s.ctx.Done():
				return
			default:
			}
			continue
		}
		metadata := new(tunnel.Metadata)
		if err := metadata.ReadFrom(conn); err != nil {
			log.Error(common.NewError("simplesocks server faield to read header").Base(err))
			conn.Close()
			continue
		}
		switch metadata.Command {
		case Connect:
			s.connChan <- &Conn{
				metadata: metadata,
				Conn:     conn,
			}
		case Associate:
			s.packetChan <- &PacketConn{
				PacketConn: trojan.PacketConn{
					Conn: conn,
				},
			}
		default:
			log.Error(common.NewError(fmt.Sprintf("simplesocks unknown command %d", metadata.Command)))
			conn.Close()
		}
	}
}

func (s *Server) AcceptConn(tunnel.Tunnel) (tunnel.Conn, error) {
	select {
	case conn := <-s.connChan:
		return conn, nil
	case <-s.ctx.Done():
		return nil, common.NewError("simplesocks server closed")
	}
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	select {
	case packetConn := <-s.packetChan:
		return packetConn, nil
	case <-s.ctx.Done():
		return nil, common.NewError("simplesocks server closed")
	}
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
	ctx, cancel := context.WithCancel(ctx)
	server := &Server{
		underlay:   underlay,
		ctx:        ctx,
		connChan:   make(chan tunnel.Conn, 32),
		packetChan: make(chan tunnel.PacketConn, 32),
		cancel:     cancel,
	}
	go server.acceptLoop()
	log.Debug("simplesocks server created")
	return server, nil
}
