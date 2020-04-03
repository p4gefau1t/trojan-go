package client

import (
	"bufio"
	"context"
	"net"
	"os"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/direct"
	"github.com/p4gefau1t/trojan-go/protocol/http"
	"github.com/p4gefau1t/trojan-go/protocol/mux"
	"github.com/p4gefau1t/trojan-go/protocol/socks"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/p4gefau1t/trojan-go/proxy"
	"github.com/p4gefau1t/trojan-go/router"
)

var logger = log.New(os.Stdout)

type packetInfo struct {
	request *protocol.Request
	packet  []byte
}

type Client struct {
	common.Runnable
	proxy.Buildable

	config         *conf.GlobalConfig
	ctx            context.Context
	cancel         context.CancelFunc
	mux            *muxPoolManager
	associatedChan chan int
	router         router.Router
}

func (c *Client) listenUDP() {
	for {
	start:
		listener, err := net.ListenUDP("udp", &net.UDPAddr{
			IP:   c.config.LocalIP,
			Port: int(c.config.LocalPort),
		})
		if err != nil {
			logger.Fatal(common.NewError("failed to listen udp").Base(err))
		}
		inbound, err := socks.NewInboundPacketSession(listener)
		<-c.associatedChan
		common.Must(err)
		logger.Debug("associated signal")
		req := protocol.Request{
			DomainName:  []byte("UDP_CONN"),
			AddressType: protocol.DomainName,
			Command:     protocol.Associate,
		}
		tunnel, err := trojan.NewOutboundConnSession(&req, nil, c.config)
		if err != nil {
			logger.Error(err)
			continue
		}
		outbound, err := trojan.NewPacketSession(tunnel)
		common.Must(err)
		alive := make(chan int)
		go proxy.ProxyPacketWithAliveChan(inbound, outbound, alive)
		for {
			select {
			case <-alive:
				logger.Debug("keep alive..(alive)")
			case <-c.associatedChan:
				logger.Debug("keep alive..(associated)")
			case <-time.After(protocol.UDPTimeout):
				logger.Debug("time out, closing UDP tunnel")
				outbound.Close()
				inbound.Close()
				goto start
			case <-c.ctx.Done():
				outbound.Close()
				inbound.Close()
				return
			}
		}
	}
}

func (c *Client) handleSocksConn(conn net.Conn, rw *bufio.ReadWriter) {
	inboundConn, err := socks.NewInboundConnSession(conn, rw)
	if err != nil {
		logger.Error(common.NewError("failed to start new inbound session").Base(err))
		return
	}
	defer inboundConn.Close()
	req := inboundConn.GetRequest()

	if req.Command == protocol.Associate {
		//setting up the bind address to respond
		//listenUDP() will handle the incoming udp packets
		req.IP = c.config.LocalIP
		req.Port = c.config.LocalPort
		if c.config.LocalIP.To16() != nil {
			req.AddressType = protocol.IPv6
		} else {
			req.AddressType = protocol.IPv4
		}
		//notify listenUDP to get ready for relaying udp packets
		c.associatedChan <- 1
		logger.Info("UDP associated to", req)
		if err := inboundConn.(protocol.NeedRespond).Respond(); err != nil {
			logger.Error("failed to repsond")
		}

		//stop relaying UDP once TCP connection is closed
		var buf [1]byte
		_, err = conn.Read(buf[:])
		logger.Debug(common.NewError("UDP conn ends").Base(err))
		return
	}

	if err := inboundConn.(protocol.NeedRespond).Respond(); err != nil {
		logger.Error(common.NewError("failed to respond").Base(err))
		return
	}

	policy, err := c.router.RouteRequest(req)
	if err != nil {
		logger.Error(err)
		return
	}
	if policy == router.Bypass {
		outboundConn, err := direct.NewOutboundConnSession(nil, req)
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Info("[bypass]conn from", conn.RemoteAddr(), "to", req)
		proxy.ProxyConn(inboundConn, outboundConn)
		return
	} else if policy == router.Block {
		logger.Info("[block]conn from", conn.RemoteAddr(), "to", req)
		return
	}

	if c.config.Mux.Enabled {
		stream, info, err := c.mux.OpenMuxConn()
		if err != nil {
			logger.Error(common.NewError("failed to open mux stream").Base(err))
			return
		}

		outboundConn, err := mux.NewOutboundMuxConnSession(stream, req)
		if err != nil {
			stream.Close()
			logger.Error(common.NewError("fail to start trojan session over mux conn").Base(err))
			return
		}
		defer outboundConn.Close()
		logger.Info("conn from", conn.RemoteAddr(), "mux tunneling to", req, "mux id", info.id)
		proxy.ProxyConn(inboundConn, outboundConn)
	} else {
		outboundConn, err := trojan.NewOutboundConnSession(req, nil, c.config)
		if err != nil {
			logger.Error(common.NewError("failed to start new outbound session").Base(err))
			return
		}
		defer outboundConn.Close()

		logger.Info("conn from", conn.RemoteAddr(), "tunneling to", req)
		proxy.ProxyConn(inboundConn, outboundConn)
	}
}

func (c *Client) handleHTTPConn(conn net.Conn, rw *bufio.ReadWriter) {
	inboundConn, inboundPacket, err := http.NewHTTPInbound(conn, rw)
	if err != nil {
		logger.Error(common.NewError("failed to start new inbound session:").Base(err))
		return
	}
	if inboundConn != nil {
		defer inboundConn.Close()
		req := inboundConn.GetRequest()

		if err := inboundConn.(protocol.NeedRespond).Respond(); err != nil {
			logger.Error(common.NewError("failed to respond").Base(err))
			return
		}

		policy, err := c.router.RouteRequest(req)
		if err != nil {
			logger.Error(err)
			return
		}
		if policy == router.Bypass {
			outboundConn, err := direct.NewOutboundConnSession(nil, req)
			if err != nil {
				logger.Error(err)
				return
			}
			logger.Info("[bypass]conn from", conn.RemoteAddr(), "to", req)
			proxy.ProxyConn(inboundConn, outboundConn)
			return
		} else if policy == router.Block {
			logger.Info("[block]conn from", conn.RemoteAddr(), "to", req)
			return
		}

		if c.config.Mux.Enabled {
			stream, info, err := c.mux.OpenMuxConn()
			if err != nil {
				logger.Error(common.NewError("failed to open mux stream").Base(err))
				return
			}
			defer stream.Close()
			outboundConn, err := mux.NewOutboundMuxConnSession(stream, req)
			if err != nil {
				logger.Error(common.NewError("fail to start trojan session over mux conn").Base(err))
				return
			}
			defer outboundConn.Close()
			logger.Info("conn from", conn.RemoteAddr(), "mux tunneling to", req, "mux id", info.id)
			proxy.ProxyConn(inboundConn, outboundConn)
		} else {
			outboundConn, err := trojan.NewOutboundConnSession(req, nil, c.config)
			if err != nil {
				logger.Error(common.NewError("failed to start new outbound session").Base(err))
				return
			}
			defer outboundConn.Close()

			logger.Info("conn from", conn.RemoteAddr(), "tunneling to", req)
			proxy.ProxyConn(inboundConn, outboundConn)
		}
	} else {
		defer inboundPacket.Close()
		packetChan := make(chan *packetInfo, 128)

		readHTTPPackets := func() {
			for {
				req, packet, err := inboundPacket.ReadPacket()
				if err != nil {
					logger.Error(err)
					return
				}
				packetChan <- &packetInfo{
					request: req,
					packet:  packet,
				}
			}
		}

		writeHTTPPackets := func() {
			for {
				select {
				case packet := <-packetChan:
					var outboundConn protocol.ConnSession
					if c.config.Mux.Enabled {
						stream, info, err := c.mux.OpenMuxConn()
						if err != nil {
							logger.Error(common.NewError("failed to open mux stream").Base(err))
							continue
						}
						outboundConn, err = mux.NewOutboundMuxConnSession(stream, packet.request)
						if err != nil {
							logger.Error(common.NewError("fail to start trojan session over mux conn").Base(err))
							continue
						}
						logger.Info("conn from", conn.RemoteAddr(), "mux tunneling to", packet.request, "mux id", info.id)
					} else {
						outboundConn, err = trojan.NewOutboundConnSession(packet.request, nil, c.config)
						if err != nil {
							logger.Error(err)
							continue
						}
					}
					_, err = outboundConn.Write(packet.packet)
					if err != nil {
						logger.Error(err)
						continue
					}
					go func(outboundConn protocol.ConnSession) {
						buf := [4096]byte{}
						defer outboundConn.Close()
						for {
							n, err := outboundConn.Read(buf[:])
							if err != nil {
								logger.Debug(err)
								return
							}
							if _, err = inboundPacket.WritePacket(nil, buf[0:n]); err != nil {
								logger.Debug(err)
								return
							}
						}
					}(outboundConn)
				case <-c.ctx.Done():
					return
				}
			}
		}

		go readHTTPPackets()
		writeHTTPPackets()
	}
}

func (c *Client) Run() error {
	go c.listenUDP()
	listener, err := net.Listen("tcp", c.config.LocalAddr.String())
	if err != nil {
		return common.NewError("failed to listen local address").Base(err)
	}
	defer listener.Close()

	logger.Info("client is running at", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-c.ctx.Done():
			default:
			}
			logger.Error(common.NewError("error occured when accpeting conn").Base(err))
			continue
		}
		rw := common.NewBufReadWriter(conn)
		tmp, err := rw.Peek(1)
		if err != nil {
			logger.Error(common.NewError("failed to obtain proxy type").Base(err))
			conn.Close()
			continue
		}
		if tmp[0] == 0x05 {
			go c.handleSocksConn(conn, rw)
		} else {
			go c.handleHTTPConn(conn, rw)
		}
	}
}

func (c *Client) Close() error {
	logger.Info("shutting down client..")
	c.cancel()
	return nil
}

func (c *Client) Build(config *conf.GlobalConfig) (common.Runnable, error) {
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.router = &router.EmptyRouter{
		DefaultPolicy: router.Proxy,
	}
	c.associatedChan = make(chan int)
	var err error
	if config.Mux.Enabled {
		logger.Info("mux enabled")
		c.mux, err = NewMuxPoolManager(c.ctx, config)
		if err != nil {
			logger.Fatal(err)
		}
	}
	if config.Router.Enabled {
		logger.Info("router enabled")
		var defaultPolicy router.Policy
		switch config.Router.DefaultPolicy {
		case "proxy":
			defaultPolicy = router.Proxy
		case "bypass":
			defaultPolicy = router.Bypass
		case "block":
			defaultPolicy = router.Block
		}
		c.router, err = router.NewMixedRouter(
			defaultPolicy,
			false,
			false,
			config.Router.Proxy,
			config.Router.Bypass,
			config.Router.Block,
		)
		if err != nil {
			logger.Fatal(common.NewError("invalid list").Base(err))
		}
	}
	c.config = config
	return c, nil
}

func init() {
	proxy.RegisterProxy(conf.Client, &Client{})
}
