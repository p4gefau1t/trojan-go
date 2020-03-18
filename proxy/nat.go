package proxy

import (
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol/nat"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
)

type NAT struct {
	common.Runnable
	config *conf.GlobalConfig
}

func (n *NAT) handleConn(conn net.Conn) {
	inbound, err := nat.NewInboundConnSession(conn)
	if err != nil {
		logger.Error("failed to start inbound session", err)
	}
	req := inbound.GetRequest()
	defer inbound.Close()
	outbound, err := trojan.NewOutboundConnSession(req, n.config)
	if err != nil {
		logger.Error("failed to start outbound session", err)
	}
	defer outbound.Close()
	logger.Info("transparent nat from", conn.RemoteAddr(), "tunneling to", req)
	proxyConn(inbound, outbound)
}

func (n *NAT) listenTCP(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			logger.Error(err)
		}
		go n.handleConn(conn)
	}
}

func (n *NAT) listenUDP() {

}

func (n *NAT) Run() error {
	logger.Info("nat running at", n.config.LocalAddr)
	tcpListener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   n.config.LocalIP,
		Port: int(n.config.LocalPort),
	})
	if err != nil {
		return err
	}
	n.listenTCP(tcpListener)
	return nil
}
