package nat

import (
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type NATInboundConnSession struct {
	protocol.ConnSession
	reqeust *protocol.Request
	conn    net.Conn
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
	addr, err := getOriginalTCPDest(i.conn.(*net.TCPConn))
	if err != nil {
		return common.NewError("failed to get original dst").Base(err)
	}
	req := &protocol.Request{
		IP:      addr.IP,
		Port:    uint16(addr.Port),
		Command: protocol.Connect,
	}
	if addr.IP.To4() != nil {
		req.AddressType = protocol.IPv4
	} else {
		req.AddressType = protocol.IPv6
	}
	i.reqeust = req
	return nil
}

func NewInboundConnSession(conn net.Conn) (protocol.ConnSession, error) {
	i := &NATInboundConnSession{
		conn: conn,
	}
	if err := i.parseRequest(); err != nil {
		return nil, common.NewError("failed to parse request").Base(err)
	}
	return i, nil
}
