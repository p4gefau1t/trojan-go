package direct

import (
	"io"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type DirectOutboundConnSession struct {
	protocol.ConnSession
	conn    io.ReadWriteCloser
	request *protocol.Request
}

func (o *DirectOutboundConnSession) Read(p []byte) (int, error) {
	return o.conn.Read(p)
}

func (o *DirectOutboundConnSession) Write(p []byte) (int, error) {
	return o.conn.Write(p)
}

func (o *DirectOutboundConnSession) Close() error {
	return o.conn.Close()
}

func NewOutboundConnSession(conn io.ReadWriteCloser, req *protocol.Request) (protocol.ConnSession, error) {
	o := &DirectOutboundConnSession{}
	o.request = req
	if conn == nil {
		newConn, err := net.Dial(req.Network(), req.String())
		if err != nil {
			return nil, err
		}
		o.conn = newConn
	} else {
		o.conn = conn
	}
	return o, nil
}

type DirectOutboundPacketSession struct {
	protocol.PacketSession
	conn    *net.UDPConn
	connSet chan int
}

func (o *DirectOutboundPacketSession) Close() error {
	o.connSet <- 0
	return o.conn.Close()
}

func (o *DirectOutboundPacketSession) ReadPacket() (*protocol.Request, []byte, error) {
	s := <-o.connSet
	if s == 0 {
		return nil, nil, common.NewError("closed")
	}
	buf := [protocol.MaxUDPPacketSize]byte{}
	n, remote, err := o.conn.ReadFromUDP(buf[:])
	if err != nil {
		return nil, nil, err
	}
	req := &protocol.Request{
		IP:          remote.IP,
		Port:        uint16(remote.Port),
		NetworkType: "udp",
		AddressType: protocol.IPv4,
	}
	if remote.IP.To16() != nil {
		req.AddressType = protocol.IPv6
	}
	return req, buf[0:n], nil
}

func (o *DirectOutboundPacketSession) WritePacket(req *protocol.Request, packet []byte) (int, error) {
	remoteAddr := &net.UDPAddr{
		IP:   req.IP,
		Port: int(req.Port),
	}
	if o.conn == nil {
		conn, err := net.DialUDP("udp", nil, remoteAddr)
		if err != nil {
			return 0, common.NewError("cannot dial to remote to init conn").Base(err)
		}
		o.conn = conn
		o.connSet <- 1
	}
	return o.conn.Write(packet)
}

func NewOutboundPacketSession() (protocol.PacketSession, error) {
	return &DirectOutboundPacketSession{
		connSet: make(chan int),
	}, nil
}
