package client

import (
	"context"
	"net"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/socks"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/p4gefau1t/trojan-go/proxy"
)

type Forward struct {
	common.Runnable
	proxy.Buildable

	config *conf.GlobalConfig
	ctx    context.Context
	cancel context.CancelFunc
	mux    *muxPoolManager
}

func (f *Forward) listenUDP(errChan chan error) {
	localIP, err := f.config.LocalAddress.ResolveIP(false)
	listener, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   localIP,
		Port: f.config.LocalAddress.Port,
	})
	if err != nil {
		errChan <- common.NewError("failed to listen udp")
		return
	}
	inbound, err := socks.NewInboundPacketSession(listener)
	common.Must(err)
	req := &protocol.Request{
		Address: &common.Address{
			DomainName:  "UDP_CONN",
			AddressType: common.DomainName,
		},
		Command: protocol.Associate,
	}
	for {
		rwc, err := DialTLSToServer(f.config)
		if err != nil {
			log.Error(common.NewError("failed dial tls to remote server").Base(err))
			time.Sleep(time.Second)
			continue
		}
		tunnel, err := trojan.NewOutboundConnSession(req, rwc, f.config)
		if err != nil {
			log.Error(common.NewError("failed to open udp tunnel").Base(err))
			continue
		}
		trojanOutbound, err := trojan.NewPacketSession(tunnel)
		common.Must(err)
		proxy.ProxyPacket(inbound, trojanOutbound)
		trojanOutbound.Close()
	}
}

func (f *Forward) listenTCP(errChan chan error) {
	listener, err := net.Listen("tcp", f.config.LocalAddress.String())
	if err != nil {
		errChan <- common.NewError("failed to listen local address").Base(err)
	}
	defer listener.Close()
	log.Info("forward is running at", listener.Addr())
	req := &protocol.Request{
		Address: f.config.TargetAddress,
		Command: protocol.Connect,
	}
	for {
		inboundConn, err := listener.Accept()
		if err != nil {
			errChan <- err
			return
		}
		handle := func(inboundConn net.Conn) {
			tlsConn, err := DialTLSToServer(f.config)
			if err != nil {
				log.Error(common.NewError("failed to dial to remote server").Base(err))
			}
			outboundConn, err := trojan.NewOutboundConnSession(req, tlsConn, f.config)
			if err != nil {
				log.Error(common.NewError("failed to start outbound session").Base(err))
			}
			proxy.ProxyConn(inboundConn, outboundConn)
		}
		go handle(inboundConn)
	}
}

func (f *Forward) Run() error {
	errChan := make(chan error, 2)
	go f.listenUDP(errChan)
	go f.listenTCP(errChan)
	select {
	case <-f.ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}

func (f *Forward) Close() error {
	log.Info("shutting down forward..")
	f.cancel()
	return nil
}

func (f *Forward) Build(config *conf.GlobalConfig) (common.Runnable, error) {
	f.ctx, f.cancel = context.WithCancel(context.Background())
	var err error
	if config.Mux.Enabled {
		log.Info("mux enabled")
		f.mux, err = NewMuxPoolManager(f.ctx, config)
		if err != nil {
			log.Fatal(err)
		}
	}
	f.config = config
	return f, nil
}

func init() {
	proxy.RegisterProxy(conf.Forward, &Forward{})
}
