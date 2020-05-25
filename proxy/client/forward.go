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
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/p4gefau1t/trojan-go/proxy"
	"github.com/p4gefau1t/trojan-go/stat"
)

type dispatchInfo struct {
	addr    net.Addr
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
	udpListener             net.PacketConn
	tcpListener             net.Listener
	auth                    stat.Authenticator
	appMan                  *AppManager
}

func (f *Forward) dispatchServerPacket(addr net.Addr) {
	for {
		f.outboundPacketTableLock.Lock()
		//use src addr as the key
		outboundPacket, found := f.outboundPacketTable[addr.String()]
		f.outboundPacketTableLock.Unlock()
		if !found {
			log.Error("Address key not found, expired?", addr.String())
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
			log.Debug("UDP timeout, exiting..")
			return
		case <-f.ctx.Done():
			log.Debug("Forward closed, exiting..")
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
				outboundConn, err := f.appMan.OpenAppConn(associateReq)
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
	listener, err := net.ListenPacket("udp", f.config.LocalAddress.String())
	if err != nil {
		errChan <- common.NewError("Failed to listen udp")
		return
	}
	f.udpListener = listener
	go f.dispatchClientPacket()
	for {
		buf := make([]byte, protocol.MaxUDPPacketSize)
		n, addr, err := listener.ReadFrom(buf)
		log.Info("Packet from", addr, "tunneling to", f.config.TargetAddress)
		if err != nil {
			errChan <- err
			return
		}
		f.clientPackets <- &dispatchInfo{
			addr:    addr,
			payload: buf[:n],
		}
	}
}

func (f *Forward) listenTCP(errChan chan error) {
	listener, err := net.Listen("tcp", f.config.LocalAddress.String())
	if err != nil {
		errChan <- common.NewError("Failed to listen local address").Base(err)
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
			errChan <- common.NewError("Error occured when accepting conn").Base(err)
		}
		handle := func(inboundConn net.Conn) {
			outboundConn, err := f.appMan.OpenAppConn(req)
			if err != nil {
				log.Error(common.NewError("Failed to start outbound session").Base(err))
				return
			}
			defer outboundConn.Close()
			proxy.RelayConn(f.ctx, inboundConn, outboundConn, f.config.BufferSize)
		}
		go handle(inboundConn)
	}
}

func (f *Forward) Run() error {
	log.Info("Trojan-Go forward is listening on", f.config.LocalAddress)
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
	log.Info("Shutting down forward..")
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
	ctx, cancel := context.WithCancel(context.Background())

	authDriver := "memory"
	auth, err := stat.NewAuth(ctx, authDriver, config)
	if err != nil {
		cancel()
		return nil, err
	}
	appMan := NewAppManager(ctx, config, auth)

	newForward := &Forward{
		ctx:                 ctx,
		cancel:              cancel,
		config:              config,
		auth:                auth,
		appMan:              appMan,
		clientPackets:       make(chan *dispatchInfo, 1024),
		outboundPacketTable: make(map[string]protocol.PacketSession),
	}
	return newForward, nil
}

func init() {
	proxy.RegisterProxy(conf.Forward, &Forward{})
}
