// +build linux

package proxy

import (
	"context"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/nat"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
)

type NAT struct {
	common.Runnable

	config        *conf.GlobalConfig
	ctx           context.Context
	cancel        context.CancelFunc
	packetInbound protocol.PacketSession
	listener      net.Listener
}

func (n *NAT) handleConn(conn net.Conn) {
	inbound, err := nat.NewInboundConnSession(conn)
	if err != nil {
		logger.Error("failed to start inbound session", err)
		return
	}
	req := inbound.GetRequest()
	defer inbound.Close()
	outbound, err := trojan.NewOutboundConnSession(req, nil, n.config)
	if err != nil {
		logger.Error("failed to start outbound session", err)
		return
	}
	defer outbound.Close()
	logger.Info("transparent nat from", conn.RemoteAddr(), "tunneling to", req)
	proxyConn(inbound, outbound)
}

func (n *NAT) listenUDP() {
	inbound, err := nat.NewInboundPacketSession(n.config)
	n.packetInbound = inbound
	if err != nil {
		logger.Error(err)
		panic(err)
	}
	defer inbound.Close()
	req := protocol.Request{
		IP:          net.IPv4(233, 233, 233, 233),
		Port:        2333,
		AddressType: protocol.IPv4,
		Command:     protocol.Associate,
	}
	for {
		tunnel, err := trojan.NewOutboundConnSession(&req, nil, n.config)
		if err != nil {
			select {
			case <-n.ctx.Done():
				return
			default:
			}
			logger.Error(err)
			continue
		}
		outbound, err := trojan.NewPacketSession(tunnel)
		proxyPacket(inbound, outbound)
		tunnel.Close()
	}
}

func (n *NAT) Run() error {
	go n.listenUDP()
	logger.Info("nat running at", n.config.LocalAddr)
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   n.config.LocalIP,
		Port: int(n.config.LocalPort),
	})
	if err != nil {
		return err
	}
	n.listener = listener
	defer listener.Close()
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			logger.Error(err)
			continue
		}
		go n.handleConn(conn)
	}
}

func (n *NAT) Close() error {
	logger.Info("shutting down nat...")
	n.cancel()
	return nil
}
