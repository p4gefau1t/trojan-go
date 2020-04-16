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
		log.Error(common.NewError("failed to start inbound session").Base(err))
		return
	}
	req := inbound.GetRequest()
	defer inbound.Close()
	if n.config.Mux.Enabled {
		stream, info, err := n.mux.OpenMuxConn()
		if err != nil {
			log.Error(common.NewError("failed to open mux stream").Base(err))
			return
		}
		outbound, err := mux.NewOutboundMuxConnSession(stream, req)
		if err != nil {
			stream.Close()
			log.Error(common.NewError("failed to start mux outbound session").Base(err))
			return
		}
		defer outbound.Close()
		log.Info("[transparent]conn from", conn.RemoteAddr(), "mux tunneling to", req, "mux id", info.id)
		proxy.ProxyConn(inbound, outbound)
		return
	}
	rwc, err := DialTLSToServer(n.config)
	if err != nil {
		log.Error(common.NewError("failed to dail to remote server").Base(err))
	}
	outbound, err := trojan.NewOutboundConnSession(req, rwc, n.config)
	if err != nil {
		log.Error("failed to start outbound session", err)
		return
	}
	defer outbound.Close()
	log.Info("[transparent]conn from", conn.RemoteAddr(), "tunneling to", req)
	proxy.ProxyConn(inbound, outbound)
}

func (n *NAT) listenUDP(errChan chan error) {
	inbound, err := nat.NewInboundPacketSession(n.config)
	if err != nil {
		log.Fatal(err)
	}
	n.packetInbound = inbound
	defer inbound.Close()
	req := &protocol.Request{
		Address: &common.Address{
			DomainName:  "UDP_CONN",
			AddressType: common.DomainName,
		},
		Command: protocol.Associate,
	}
	for {
		rwc, err := DialTLSToServer(n.config)
		if err != nil {
			log.Error(common.NewError("failed to dail to remote server").Base(err))
			continue
		}
		tunnel, err := trojan.NewOutboundConnSession(req, rwc, n.config)
		if err != nil {
			select {
			case <-n.ctx.Done():
				return
			default:
			}
			log.Error(err)
			continue
		}
		outbound, err := trojan.NewPacketSession(tunnel)
		common.Must(err)
		proxy.ProxyPacket(inbound, outbound)
		tunnel.Close()
	}
}

func (n *NAT) listenTCP(errChan chan error) {
	localIP, err := n.config.LocalAddress.ResolveIP(false)
	if err != nil {
		errChan <- err
	}
	listener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   localIP,
		Port: int(n.config.LocalAddress.Port),
	})
	if err != nil {
		errChan <- err
	}
	n.listener = listener
	defer listener.Close()
	for {
		conn, err := n.listener.Accept()
		if err != nil {
			select {
			case <-n.ctx.Done():
				return
			default:
			}
			errChan <- err
			break
		}
		go n.handleConn(conn)
	}
}

func (n *NAT) Run() error {
	log.Info("nat running at", n.config.LocalAddress)
	errChan := make(chan error, 2)
	go n.listenUDP(errChan)
	go n.listenTCP(errChan)
	select {
	case err := <-errChan:
		return err
	case <-n.ctx.Done():
		return nil
	}
}

func (n *NAT) Close() error {
	log.Info("shutting down nat...")
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
			log.Fatal(err)
		}
		n.mux = mux
	}
	return n, nil
}

func init() {
	proxy.RegisterProxy(conf.NAT, &NAT{})
}
