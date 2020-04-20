// +build linux

package client

import (
	"context"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/nat"
	"github.com/p4gefau1t/trojan-go/protocol/simplesocks"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/p4gefau1t/trojan-go/proxy"
)

type NAT struct {
	common.Runnable
	proxy.Buildable

	config        *conf.GlobalConfig
	ctx           context.Context
	cancel        context.CancelFunc
	inboundPacket protocol.PacketSession
	listener      net.Listener
	transport     TransportManager
}

func (n *NAT) openOutboundConn(req *protocol.Request) (protocol.ConnSession, error) {
	var outboundConn protocol.ConnSession
	//transport layer
	transport, err := n.transport.DialToServer()
	if err != nil {
		return nil, common.NewError("failed to init transport layer").Base(err)
	}
	//application layer
	if n.config.Mux.Enabled {
		outboundConn, err = simplesocks.NewOutboundConnSession(req, transport)
	} else {
		outboundConn, err = trojan.NewOutboundConnSession(req, transport, n.config)
	}
	if err != nil {
		return nil, common.NewError("fail to start conn session").Base(err)
	}
	return outboundConn, nil
}

func (n *NAT) handleConn(conn net.Conn) {
	inbound, err := nat.NewInboundConnSession(conn)
	if err != nil {
		log.Error(common.NewError("failed to start inbound session").Base(err))
		return
	}
	req := inbound.GetRequest()
	defer inbound.Close()
	rwc, err := n.transport.DialToServer()
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
	proxy.ProxyConn(n.ctx, inbound, outbound)
}

func (n *NAT) listenUDP(errChan chan error) {
	inboundPacket, err := nat.NewInboundPacketSession(n.config)
	if err != nil {
		log.Fatal(err)
	}
	n.inboundPacket = inboundPacket
	defer inboundPacket.Close()
	req := &protocol.Request{
		Address: &common.Address{
			DomainName:  "UDP_CONN",
			AddressType: common.DomainName,
		},
		Command: protocol.Associate,
	}
	for {
		outboundConn, err := n.openOutboundConn(req)
		if err != nil {
			log.Error(err)
			continue
		}
		outboundPacket, err := trojan.NewPacketSession(outboundConn)
		common.Must(err)
		proxy.ProxyPacket(n.ctx, inboundPacket, outboundPacket)
		outboundPacket.Close()
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
	n.inboundPacket.Close()
	return nil
}

func (n *NAT) Build(config *conf.GlobalConfig) (common.Runnable, error) {
	n.ctx, n.cancel = context.WithCancel(context.Background())
	n.config = config
	if config.Mux.Enabled {
		log.Info("mux enabled")
		n.transport = NewMuxPoolManager(n.ctx, config)
	} else {
		n.transport = NewTLSManager(config)
	}
	return n, nil
}

func init() {
	proxy.RegisterProxy(conf.NAT, &NAT{})
}
