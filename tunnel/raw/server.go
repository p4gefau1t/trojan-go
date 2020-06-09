package raw

import (
	"net"

	"github.com/p4gefau1t/trojan-go/tunnel"
)

type Server struct {
	tcpListener net.Listener
	addr        *tunnel.Address
}

func (s *Server) AcceptConn(tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := s.tcpListener.Accept()
	if err != nil {
		return nil, err
	}
	return &Conn{
		TCPConn: conn.(*net.TCPConn),
	}, nil
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	packetConn, err := net.ListenPacket("udp", s.addr.String())
	if err != nil {
		return nil, err
	}
	return &PacketConn{
		UDPConn: packetConn.(*net.UDPConn),
	}, nil
}

func (s *Server) Close() error {
	return s.tcpListener.Close()
}

func NewServer(addr *tunnel.Address) (*Server, error) {
	l, err := net.Listen("tcp", addr.String())
	if err != nil {
		return nil, err
	}
	return &Server{
		addr:        addr,
		tcpListener: l,
	}, nil
}
