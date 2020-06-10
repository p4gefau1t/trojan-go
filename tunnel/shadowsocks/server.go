package shadowsocks

import (
	"context"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"github.com/shadowsocks/go-shadowsocks2/core"
)

type Server struct {
	underlay tunnel.Server
	core.Cipher
}

func (s *Server) AcceptConn(overlay tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := s.underlay.AcceptConn(overlay)
	if err != nil {
		return nil, common.NewError("shadowsocks failed to accept connection from underlying tunnel")
	}
	return &transport.Conn{
		Conn: s.Cipher.StreamConn(conn),
	}, nil
}

func (s *Server) AcceptPacket(t tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

func (s *Server) Close() error {
	return s.underlay.Close()
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	cipher, err := core.PickCipher(cfg.Shadowsocks.Method, nil, cfg.Shadowsocks.Password)
	if err != nil {
		return nil, common.NewError("invalid shadowsocks cipher").Base(err)
	}
	log.Info("shadowsocks client created")
	return &Server{
		underlay: underlay,
		Cipher:   cipher,
	}, nil
}
