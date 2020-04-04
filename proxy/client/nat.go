// +build linux

package client

import (
	"context"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/mux"
	"github.com/p4gefau1t/trojan-go/protocol/nat"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/p4gefau1t/trojan-go/proxy"
)

type NAT struct {
	common.Runnable
	proxy.Buildable

	config        *conf.GlobalConfig
	ctx           context.Context
	cancel        context.CancelFunc
	packetInbound protocol.PacketSession
	listener      net.Listener
	mux           *muxPoolManager
}

func (n *NAT) handleConn(conn net.Conn) {
	inbound, err := nat.NewInboundConnSession(conn)
	if err != nil {
		log.DefaultLogger.Error(common.NewError("failed to start inbound session").Base(err))
		return
	}
	req := inbound.GetRequest()
	defer inbound.Close()
	if n.config.Mux.Enabled {
		stream, info, err := n.mux.OpenMuxConn()
		if err != nil {
			log.DefaultLogger.Error(common.NewError("failed to open mux stream").Base(err))
			return
		}
		outbound, err := mux.NewOutboundMuxConnSession(stream, req)
		if err != nil {
			stream.Close()
			log.DefaultLogger.Error(common.NewError("failed to start mux outbound session").Base(err))
			return
		}
		defer outbound.Close()
		log.DefaultLogger.Info("[transparent]conn from", conn.RemoteAddr(), "mux tunneling to", req, "mux id", info.id)
		proxy.ProxyConn(inbound, outbound)
		return
	}
	outbound, err := trojan.NewOutboundConnSession(req, nil, n.config)
	if err != nil {
		log.DefaultLogger.Error("failed to start outbound session", err)
		return
	}
	defer outbound.Close()
	log.DefaultLogger.Info("[transparent]conn from", conn.RemoteAddr(), "tunneling to", req)
	proxy.ProxyConn(inbound, outbound)
}

func (n *NAT) listenUDP() {
	inbound, err := nat.NewInboundPacketSession(n.config)
	if err != nil {
		log.DefaultLogger.Fatal(err)
	}
	n.packetInbound = inbound
	defer inbound.Close()
	req := protocol.Request{
		DomainName:  []byte("UDP_CONN"),
		AddressType: protocol.DomainName,
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
			log.DefaultLogger.Error(err)
			continue
		}
		outbound, err := trojan.NewPacketSession(tunnel)
		common.Must(err)
		proxy.ProxyPacket(inbound, outbound)
		tunnel.Close()
	}
}

func (n *NAT) Run() error {
	go n.listenUDP()
	log.DefaultLogger.Info("nat running at", n.config.LocalAddr)
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
			select {
			case <-n.ctx.Done():
				return nil
			default:
			}
			log.DefaultLogger.Error(err)
			continue
		}
		go n.handleConn(conn)
	}
}

func (n *NAT) Close() error {
	log.DefaultLogger.Info("shutting down nat...")
	n.cancel()
	n.listener.Close()
	n.packetInbound.Close()
	return nil
}

func (n *NAT) Build(config *conf.GlobalConfig) (common.Runnable, error) {
	n.ctx, n.cancel = context.WithCancel(context.Background())
	n.config = config
	if config.Mux.Enabled {
		mux, err := NewMuxPoolManager(n.ctx, config)
		if err != nil {
			log.DefaultLogger.Fatal(err)
		}
		n.mux = mux
	}
	return n, nil
}

func init() {
	proxy.RegisterProxy(conf.NAT, &NAT{})
}
