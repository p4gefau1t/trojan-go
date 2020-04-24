package client

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/simplesocks"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/p4gefau1t/trojan-go/proxy"
)

type dispatchInfo struct {
	addr    *net.UDPAddr
	payload []byte
}

type Forward struct {
	common.Runnable
	proxy.Buildable

	config                  *conf.GlobalConfig
	ctx                     context.Context
	cancel                  context.CancelFunc
	clientPackets           chan *dispatchInfo
	outboundPacketTableLock sync.Mutex
	outboundPacketTable     map[string]protocol.PacketSession
	udpListener             *net.UDPConn
	tcpListener             net.Listener
	transport               TransportManager
}

func (f *Forward) openOutboundConn(req *protocol.Request) (protocol.ConnSession, error) {
	var outboundConn protocol.ConnSession
	//transport layer
	transport, err := f.transport.DialToServer()
	if err != nil {
		return nil, common.NewError("failed to init transport layer").Base(err)
	}
	//application layer
	if f.config.Mux.Enabled {
		outboundConn, err = simplesocks.NewOutboundConnSession(req, transport)
	} else {
		outboundConn, err = trojan.NewOutboundConnSession(req, transport, f.config)
	}
	if err != nil {
		return nil, common.NewError("fail to start conn session").Base(err)
	}
	return outboundConn, nil
}

func (f *Forward) dispatchServerPacket(addr *net.UDPAddr) {
	for {
		f.outboundPacketTableLock.Lock()
		//use src addr as the key
		outboundPacket, found := f.outboundPacketTable[addr.String()]
		f.outboundPacketTableLock.Unlock()
		if !found {
			log.Error("addr key not found")
			return
		}
		payloadChan := make(chan []byte, 64)
		go func() {
			_, payload, err := outboundPacket.ReadPacket()
			if err != nil { //expired
				return
			}
			payloadChan <- payload
		}()
		select {
		case payload := <-payloadChan:
			_, err := f.udpListener.WriteTo(payload, addr)
			if err != nil { //closed
				return
			}
		case <-time.After(protocol.UDPTimeout):
			outboundPacket.Close()
			f.outboundPacketTableLock.Lock()
			delete(f.outboundPacketTable, addr.String())
			f.outboundPacketTableLock.Unlock()
			log.Debug("udp timeout, exiting..")
			return
		case <-f.ctx.Done():
			log.Debug("forward closed, exiting..")
			return
		}
	}
}

func (f *Forward) dispatchClientPacket() {
	fixedReq := &protocol.Request{
		Address: f.config.TargetAddress,
	}
	associateReq := &protocol.Request{
		Address: &common.Address{
			DomainName:  "UDP_CONN",
			AddressType: common.DomainName,
		},
		Command: protocol.Associate,
	}
	for {
		select {
		case packet := <-f.clientPackets:
			f.outboundPacketTableLock.Lock()
			outboundPacket, found := f.outboundPacketTable[packet.addr.String()]
			if !found {
				outboundConn, err := f.openOutboundConn(associateReq)
				if err != nil {
					log.Error(err)
					continue
				}
				outboundPacket, err = trojan.NewPacketSession(outboundConn)
				common.Must(err)
				f.outboundPacketTable[packet.addr.String()] = outboundPacket
				go f.dispatchServerPacket(packet.addr)
			}
			f.outboundPacketTableLock.Unlock()
			outboundPacket.WritePacket(fixedReq, packet.payload)
		case <-f.ctx.Done():
			return
		}
	}
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
	f.udpListener = listener
	go f.dispatchClientPacket()
	for {
		buf := make([]byte, protocol.MaxUDPPacketSize)
		n, addr, err := listener.ReadFromUDP(buf)
		log.Info("packet from", addr, "tunneling to", f.config.TargetAddress)
		if err != nil {
			errChan <- err
			return
		}
		f.clientPackets <- &dispatchInfo{
			addr:    addr,
			payload: buf[0:n],
		}
	}
}

func (f *Forward) listenTCP(errChan chan error) {
	listener, err := net.Listen("tcp", f.config.LocalAddress.String())
	if err != nil {
		errChan <- common.NewError("failed to listen local address").Base(err)
		return
	}
	f.tcpListener = listener
	defer listener.Close()
	req := &protocol.Request{
		Address: f.config.TargetAddress,
		Command: protocol.Connect,
	}
	for {
		inboundConn, err := listener.Accept()
		if err != nil {
			errChan <- common.NewError("error occured when accpeting conn").Base(err)
		}
		handle := func(inboundConn net.Conn) {
			outboundConn, err := f.openOutboundConn(req)
			if err != nil {
				log.Error(common.NewError("failed to start outbound session").Base(err))
				return
			}
			defer outboundConn.Close()
			proxy.ProxyConn(f.ctx, inboundConn, outboundConn, f.config.BufferSize)
		}
		go handle(inboundConn)
	}
}

func (f *Forward) Run() error {
	log.Info("forward is running at", f.config.LocalAddress)
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
	if f.udpListener != nil {
		f.udpListener.Close()
	}
	if f.tcpListener != nil {
		f.tcpListener.Close()
	}
	return nil
}

func (f *Forward) Build(config *conf.GlobalConfig) (common.Runnable, error) {
	f.ctx, f.cancel = context.WithCancel(context.Background())
	if config.Mux.Enabled {
		log.Info("mux enabled")
		f.transport = NewMuxPoolManager(f.ctx, config)
	} else {
		f.transport = NewTLSManager(config)
	}
	f.clientPackets = make(chan *dispatchInfo, 512)
	f.outboundPacketTable = make(map[string]protocol.PacketSession)
	f.config = config
	return f, nil
}

func init() {
	proxy.RegisterProxy(conf.Forward, &Forward{})
}
