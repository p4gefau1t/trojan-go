// +build linux,!386

package tproxy

import (
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/LiamHaworth/go-tproxy"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/dokodemo"
)

const MaxPacketSize = 1024 * 8

type Server struct {
	tcpListener net.Listener
	udpListener *net.UDPConn
	packetChan  chan tunnel.PacketConn
	timeout     time.Duration
	mappingLock sync.RWMutex
	mapping     map[string]*dokodemo.PacketConn
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
	addr, err := getOriginalTCPDest(conn.(*tproxy.Conn).TCPConn)
	if err != nil {
		return nil, common.NewError("tproxy failed to obtain original address of tcp socket").Base(err)
	}
	address, err := tunnel.NewAddressFromAddr("tcp", addr.String())
	common.Must(err)
	log.Info("tproxy connection from", conn.RemoteAddr().String(), "metadata", addr.String())
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
			n, src, dst, err := tproxy.ReadFromUDP(s.udpListener, buf)
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
		}

		s.mappingLock.RLock()
		conn, found := s.mapping[info.src.String()+"|"+info.dst.String()]
		s.mappingLock.RUnlock()

		if !found {
			ctx, cancel := context.WithCancel(s.ctx)
			conn = &dokodemo.PacketConn{
				Input:      make(chan []byte, 128),
				Output:     make(chan []byte, 128),
				PacketConn: s.udpListener,
				Context:    ctx,
				Cancel:     cancel,
				Source:     info.src,
				M: &tunnel.Metadata{
					Address: tunnel.NewAddressFromHostPort("udp", info.dst.IP.String(), info.dst.Port),
				},
			}

			s.mappingLock.Lock()
			s.mapping[info.src.String()+"|"+info.dst.String()] = conn
			s.mappingLock.Unlock()

			log.Info("new tproxy udp session, from", info.src, "metadata", info.dst)

			go func(conn *dokodemo.PacketConn) {
				defer conn.Close()
				back, err := tproxy.DialUDP(
					"udp",
					&net.UDPAddr{
						IP:   conn.M.IP,
						Port: conn.M.Port,
					},
					conn.Source.(*net.UDPAddr),
				)
				if err != nil {
					log.Error(common.NewError("failed to dial tproxy udp").Base(err))
					return
				}
				defer back.Close()
				for {
					select {
					case payload := <-conn.Output:
						n, err := back.Write(payload)
						if err != nil {
							log.Error(common.NewError("tproxy udp write error").Base(err))
							return
						}
						log.Debug("recv packet send back to", conn.Source.String(), "payload", len(payload), "sent", n)
					case <-s.ctx.Done():
						log.Debug("exiting")
						return
					case <-time.After(s.timeout):
						s.mappingLock.Lock()
						delete(s.mapping, conn.Source.String()+"|"+conn.M.String())
						s.mappingLock.Unlock()
						log.Debug("packet session timeout. closed", conn.Source.String())
						return
					}
				}
			}(conn)
		}

		select {
		case conn.Input <- info.payload:
			log.Debug("sending tproxy packet payload to", info.dst.String(), "size", len(info.payload))
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
	tcpListener, err := tproxy.ListenTCP("tcp", &net.TCPAddr{
		IP:   ip,
		Port: cfg.LocalPort,
	})
	if err != nil {
		return nil, common.NewError("tproxy failed to listen tcp").Base(err)
	}

	udpListener, err := tproxy.ListenUDP("udp", &net.UDPAddr{
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
		mapping:     make(map[string]*dokodemo.PacketConn),
		packetChan:  make(chan tunnel.PacketConn, 32),
	}
	go server.packetDispatchLoop()
	log.Info("tproxy server listening on", tcpListener.Addr(), "(tcp)", udpListener.LocalAddr(), "(udp)")
	log.Debug("tproxy server created")
	return server, nil
}
