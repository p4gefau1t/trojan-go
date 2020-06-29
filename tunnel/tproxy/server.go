// +build linux

package tproxy

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

const MaxPacketSize = 1024 * 8

type Server struct {
	tcpListener net.Listener
	udpListener *net.UDPConn
	packetChan  chan tunnel.PacketConn
	timeout     time.Duration
	mappingLock sync.RWMutex
	mapping     map[string]*PacketConn
	ctx         context.Context
	cancel      context.CancelFunc
}

func (s *Server) Close() error {
	s.cancel()
	s.tcpListener.Close()
	return s.udpListener.Close()
}

func (s *Server) AcceptConn(tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := s.tcpListener.Accept()
	if err != nil {
		select {
		case <-s.ctx.Done():
		default:
			log.Fatal(common.NewError("tproxy failed to accept connection").Base(err))
		}
		return nil, common.NewError("tproxy failed to accept conn")
	}
	dst, err := getOriginalTCPDest(conn.(*net.TCPConn))
	if err != nil {
		return nil, common.NewError("tproxy failed to obtain original address of tcp socket").Base(err)
	}
	address, err := tunnel.NewAddressFromAddr("tcp", dst.String())
	common.Must(err)
	log.Info("tproxy connection from", conn.RemoteAddr().String(), "metadata", dst.String())
	return &Conn{
		metadata: &tunnel.Metadata{
			Address: address,
		},
		Conn: conn,
	}, nil
}

func (s *Server) packetDispatchLoop() {
	type tproxyPacketInfo struct {
		src     *net.UDPAddr
		dst     *net.UDPAddr
		payload []byte
	}
	packetQueue := make(chan *tproxyPacketInfo, 1024)

	go func() {
		for {
			buf := make([]byte, MaxPacketSize)
			n, src, dst, err := ReadFromUDP(s.udpListener, buf)
			if err != nil {
				select {
				case <-s.ctx.Done():
				default:
					log.Fatal(common.NewError("tproxy failed to read from udp socket").Base(err))
				}
				s.Close()
				return
			}
			log.Debug("udp packet from", src, "metadata", dst, "size", n)
			packetQueue <- &tproxyPacketInfo{
				src:     src,
				dst:     dst,
				payload: buf[:n],
			}
		}
	}()

	for {
		var info *tproxyPacketInfo
		select {
		case info = <-packetQueue:
		case <-s.ctx.Done():
			log.Debug("exiting")
			return
		}

		s.mappingLock.RLock()
		conn, found := s.mapping[info.src.String()]
		s.mappingLock.RUnlock()

		if !found {
			ctx, cancel := context.WithCancel(s.ctx)
			conn = &PacketConn{
				input:      make(chan *packetInfo, 128),
				output:     make(chan *packetInfo, 128),
				PacketConn: s.udpListener,
				ctx:        ctx,
				cancel:     cancel,
				src:        info.src,
			}

			s.mappingLock.Lock()
			s.mapping[info.src.String()] = conn
			s.mappingLock.Unlock()

			log.Info("new tproxy udp session from", info.src.String(), "metadata", info.dst.String())
			s.packetChan <- conn

			go func(conn *PacketConn) {
				defer conn.Close()
				log.Debug("udp packet daemon for", conn.src.String())
				for {
					select {
					case info := <-conn.output:
						if info.metadata.AddressType != tunnel.IPv4 &&
							info.metadata.AddressType != tunnel.IPv6 {
							log.Error("tproxy invalid response metadata address", info.metadata)
							continue
						}
						back, err := DialUDP(
							"udp",
							&net.UDPAddr{
								IP:   info.metadata.IP,
								Port: info.metadata.Port,
							},
							conn.src.(*net.UDPAddr),
						)
						if err != nil {
							log.Error(common.NewError("failed to dial tproxy udp").Base(err))
							return
						}
						n, err := back.Write(info.payload)
						if err != nil {
							log.Error(common.NewError("tproxy udp write error").Base(err))
							return
						}
						log.Debug("recv packet, send back to", conn.src, "payload", len(info.payload), "sent", n)
						back.Close()
					case <-s.ctx.Done():
						log.Debug("exiting")
						return
					case <-time.After(s.timeout):
						s.mappingLock.Lock()
						delete(s.mapping, conn.src.String())
						s.mappingLock.Unlock()
						log.Debug("packet session ", conn.src.String(), "timeout")
						return
					}
				}
			}(conn)
		}

		newInfo := &packetInfo{
			metadata: &tunnel.Metadata{
				Address: tunnel.NewAddressFromHostPort("udp", info.dst.IP.String(), info.dst.Port),
			},
			payload: info.payload,
		}

		select {
		case conn.input <- newInfo:
			log.Debug("tproxy packet sent with metadata", newInfo.metadata, "size", len(info.payload))
		default:
			// if we got too many packets, simply drop it
			log.Warn("tproxy udp relay queue full!")
		}
	}
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	select {
	case conn := <-s.packetChan:
		log.Info("tproxy packet conn accepted")
		return conn, nil
	case <-s.ctx.Done():
		return nil, io.EOF
	}
}

func NewServer(ctx context.Context, _ tunnel.Server) (*Server, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	ctx, cancel := context.WithCancel(ctx)
	listenAddr := tunnel.NewAddressFromHostPort("tcp", cfg.LocalHost, cfg.LocalPort)
	ip, err := listenAddr.ResolveIP()
	if err != nil {
		return nil, common.NewError("invalid tproxy local address").Base(err)
	}
	tcpListener, err := ListenTCP("tcp", &net.TCPAddr{
		IP:   ip,
		Port: cfg.LocalPort,
	})
	if err != nil {
		return nil, common.NewError("tproxy failed to listen tcp").Base(err)
	}

	udpListener, err := ListenUDP("udp", &net.UDPAddr{
		IP:   ip,
		Port: cfg.LocalPort,
	})
	if err != nil {
		return nil, common.NewError("tproxy failed to listen udp").Base(err)
	}

	server := &Server{
		tcpListener: tcpListener,
		udpListener: udpListener,
		ctx:         ctx,
		cancel:      cancel,
		timeout:     time.Duration(cfg.UDPTimeout) * time.Second,
		mapping:     make(map[string]*PacketConn),
		packetChan:  make(chan tunnel.PacketConn, 32),
	}
	go server.packetDispatchLoop()
	log.Info("tproxy server listening on", tcpListener.Addr(), "(tcp)", udpListener.LocalAddr(), "(udp)")
	log.Debug("tproxy server created")
	return server, nil
}
