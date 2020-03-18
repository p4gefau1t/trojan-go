package tproxy

import (
	"fmt"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type NATInboundConnSession struct {
	protocol.ConnSession
	reqeust *protocol.Request
	conn    *Conn
}

func (i *NATInboundConnSession) Read(p []byte) (int, error) {
	return i.conn.Read(p)
}

func (i *NATInboundConnSession) Write(p []byte) (int, error) {
	return i.conn.Write(p)
}

func (i *NATInboundConnSession) Close() error {
	return i.conn.Close()
}

func (i *NATInboundConnSession) GetRequest() *protocol.Request {
	return i.reqeust
}

func (i *NATInboundConnSession) parseRequest() error {
	remote := i.conn.RemoteAddr()
	hostStr, portStr, err := net.SplitHostPort(remote.String())
	if err != nil {
		return err
	}
	ip := net.ParseIP(hostStr)
	if ip == nil {
		return common.NewError("invalid host " + hostStr)
	}
	var port uint16
	fmt.Sscanf(portStr, "%d", &port)
	req := &protocol.Request{
		IP:   ip,
		Port: port,
	}
	if ip.To4() != nil {
		req.AddressType = protocol.IPv4
	} else {
		req.AddressType = protocol.IPv6
	}
	i.reqeust = req
	return nil
}

func NewInboundConnSession(conn net.Conn) (protocol.ConnSession, error) {
	i := &NATInboundConnSession{
		conn: conn.(*Conn),
	}
	if err := i.parseRequest(); err != nil {
		return nil, common.NewError("failed to parse request").Base(err)
	}
	return i, nil
}
