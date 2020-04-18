package client

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/mux"
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

	config              *conf.GlobalConfig
	ctx                 context.Context
	cancel              context.CancelFunc
	mux                 *muxPoolManager
	clientPackets       chan *dispatchInfo
	packetOutboundLock  sync.Mutex
	packetOutboundTable map[string]protocol.PacketSession
	udpListener         *net.UDPConn
	tcpListener         net.Listener
}

func (f *Forward) dispatchServerPacket(addr *net.UDPAddr) {
	for {
		f.packetOutboundLock.Lock()
		//use src addr as the key
		outbound, found := f.packetOutboundTable[addr.String()]
		f.packetOutboundLock.Unlock()
		if !found {
			log.Error("addr key not found")
			return
		}
		payloadChan := make(chan []byte)
		go func() {
			_, payload, err := outbound.ReadPacket()
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
			outbound.Close()
			f.packetOutboundLock.Lock()
			delete(f.packetOutboundTable, addr.String())
			f.packetOutboundLock.Unlock()
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
			f.packetOutboundLock.Lock()
			outbound, found := f.packetOutboundTable[packet.addr.String()]
			var err error
			if !found {
				var transport io.ReadWriteCloser
				if f.config.Mux.Enabled {
					muxConn, info, err := f.mux.OpenMuxConn()
					if err != nil {
						log.Error(common.NewError("failed to start mux conn").Base(err))
						continue
					}
					log.Info("mux udp conn id", info.id)
					transport, err = mux.NewOutboundConnSession(muxConn, associateReq)
					if err != nil {
						log.Error(err)
						continue
					}
				} else {
					tlsConn, err := DialTLSToServer(f.config)
					if err != nil {
						log.Error(err)
						continue
					}
					transport, err = trojan.NewOutboundConnSession(associateReq, tlsConn, f.config)
					if err != nil {
						log.Error(err)
						continue
					}
				}
				outbound, err = trojan.NewPacketSession(transport)
				common.Must(err)
				f.packetOutboundTable[packet.addr.String()] = outbound
				go f.dispatchServerPacket(packet.addr)
			}
			f.packetOutboundLock.Unlock()
			outbound.WritePacket(fixedReq, packet.payload)
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
			errChan <- err
			return
		}
		handle := func(inboundConn net.Conn) {
			var transport io.ReadWriteCloser
			if f.config.Mux.Enabled {
				muxConn, info, err := f.mux.OpenMuxConn()
				if err != nil {
					log.Error(err)
					return
				}
				transport = muxConn
				log.Info("conn from", inboundConn.RemoteAddr(), "mux tunneling to", f.config.TargetAddress, "id", info.id)
			} else {
				tlsConn, err := DialTLSToServer(f.config)
				if err != nil {
					log.Error(err)
					return
				}
				transport = tlsConn
				log.Info("conn from", inboundConn.RemoteAddr(), "tunneling to", f.config.TargetAddress)
			}
			outboundConn, err := trojan.NewOutboundConnSession(req, transport, f.config)
			if err != nil {
				log.Error(common.NewError("failed to start outbound session").Base(err))
			}
			proxy.ProxyConn(f.ctx, inboundConn, outboundConn)
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
	f.tcpListener.Close()
	f.udpListener.Close()
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
	f.clientPackets = make(chan *dispatchInfo, 512)
	f.packetOutboundTable = make(map[string]protocol.PacketSession)
	f.config = config
	return f, nil
}

func init() {
	proxy.RegisterProxy(conf.Forward, &Forward{})
}
