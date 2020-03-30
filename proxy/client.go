package proxy

import (
	"bufio"
	"context"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/http"
	"github.com/p4gefau1t/trojan-go/protocol/mux"
	"github.com/p4gefau1t/trojan-go/protocol/socks"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/xtaci/smux"
)

type muxID uint32

func generateMuxID() muxID {
	return muxID(rand.Uint32())
}

type muxClientInfo struct {
	id             muxID
	client         *smux.Session
	lastActiveTime time.Time
}

type Client struct {
	common.Runnable

	config *conf.GlobalConfig
	ctx    context.Context
	cancel context.CancelFunc

	muxLock sync.Mutex
	muxPool map[muxID]*muxClientInfo
}

func (c *Client) newMuxClient() (*muxClientInfo, error) {
	id := generateMuxID()
	if _, found := c.muxPool[id]; found {
		return nil, common.NewError("duplicated id")
	}
	req := &protocol.Request{
		Command:     protocol.Mux,
		DomainName:  []byte("MUX_CONN"),
		AddressType: protocol.DomainName,
	}
	conn, err := trojan.NewOutboundConnSession(req, nil, c.config)
	if err != nil {
		logger.Error(common.NewError("failed to dial tls tunnel").Base(err))
		return nil, err
	}

	client, err := smux.Client(conn, nil)
	common.Must(err)
	logger.Info("mux TLS tunnel established, id:", id)
	return &muxClientInfo{
		client:         client,
		id:             id,
		lastActiveTime: time.Now(),
	}, nil
}

func (c *Client) pickMuxClient() (*muxClientInfo, error) {
	c.muxLock.Lock()
	defer c.muxLock.Unlock()

	for _, info := range c.muxPool {
		if !info.client.IsClosed() && (info.client.NumStreams() < c.config.TCP.MuxConcurrency || c.config.TCP.MuxConcurrency <= 0) {
			info.lastActiveTime = time.Now()
			return info, nil
		}
	}

	//not found
	info, err := c.newMuxClient()
	if err != nil {
		return nil, err
	}
	c.muxPool[info.id] = info
	return info, nil
}

func (c *Client) checkAndCloseIdleMuxClient() {
	muxIdleDuration := time.Duration(c.config.TCP.MuxIdleTimeout) * time.Second
	for {
		select {
		case <-time.After(muxIdleDuration / 4):
			c.muxLock.Lock()
			for id, info := range c.muxPool {
				if info.client.IsClosed() {
					delete(c.muxPool, id)
					logger.Info("mux", id, "is dead")
				} else if info.client.NumStreams() == 0 && time.Now().Sub(info.lastActiveTime) > muxIdleDuration {
					info.client.Close()
					delete(c.muxPool, id)
					logger.Info("mux", id, "is closed due to inactive")
				}
			}
			if len(c.muxPool) != 0 {
				logger.Info("current mux pool conn num", len(c.muxPool))
			}
			c.muxLock.Unlock()
		case <-c.ctx.Done():
			c.muxLock.Lock()
			for id, info := range c.muxPool {
				info.client.Close()
				logger.Info("mux", id, "closed")
			}
			c.muxLock.Unlock()
			return
		}
	}
}

func (c *Client) handleSocksConn(conn net.Conn, rw *bufio.ReadWriter) {
	inboundConn, err := socks.NewInboundConnSession(conn, rw)
	if err != nil {
		logger.Error(common.NewError("failed to start new inbound session:").Base(err))
		return
	}
	defer inboundConn.Close()
	req := inboundConn.GetRequest()

	if req.Command == protocol.Associate {
		outboundConn, err := trojan.NewOutboundConnSession(req, nil, c.config)
		if err != nil {
			logger.Error(common.NewError("failed to start new outbound session for UDP").Base(err))
			return
		}

		listenConn, err := net.ListenUDP("udp", &net.UDPAddr{
			IP: c.config.LocalIP,
		})
		if err != nil {
			logger.Error(common.NewError("failed to listen udp:").Base(err))
			return
		}

		req.IP = c.config.LocalIP
		port, err := protocol.ParsePort(listenConn.LocalAddr())
		common.Must(err)
		req.Port = port
		req.AddressType = protocol.IPv4

		inboundPacket, err := socks.NewInboundPacketSession(listenConn)
		if err != nil {
			logger.Error("failed to start inbound packet session:", err)
			return
		}
		defer inboundPacket.Close()

		outboundPacket, err := trojan.NewPacketSession(outboundConn)
		common.Must(err)
		go proxyPacket(inboundPacket, outboundPacket)

		inboundConn.(protocol.NeedRespond).Respond(nil)
		logger.Info("UDP associated to", req)

		//stop relaying UDP once TCP connection is closed
		var buf [1]byte
		_, err = conn.Read(buf[:])
		logger.Info("UDP conn ends", err)
		return
	}

	if err := inboundConn.(protocol.NeedRespond).Respond(nil); err != nil {
		logger.Error(common.NewError("failed to respond").Base(err))
		return
	}

	if c.config.TCP.Mux {
		info, err := c.pickMuxClient()
		if err != nil {
			logger.Error(common.NewError("failed to pick a mux client").Base(err))
			return
		}

		stream, err := info.client.OpenStream()
		if err != nil {
			logger.Error(err)
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
		proxyConn(conn, outboundConn)
		info.lastActiveTime = time.Now()
	} else {
		outboundConn, err := trojan.NewOutboundConnSession(req, nil, c.config)
		if err != nil {
			logger.Error(common.NewError("failed to start new outbound session").Base(err))
			return
		}
		defer outboundConn.Close()

		logger.Info("conn from", conn.RemoteAddr(), "tunneling to", req)
		proxyConn(inboundConn, outboundConn)
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

		if err := inboundConn.(protocol.NeedRespond).Respond(nil); err != nil {
			logger.Error(common.NewError("failed to respond").Base(err))
			return
		}

		if c.config.TCP.Mux {
			info, err := c.pickMuxClient()
			if err != nil {
				logger.Error(common.NewError("failed to pick a mux client").Base(err))
				return
			}

			stream, err := info.client.OpenStream()
			if err != nil {
				logger.Error(err)
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
			proxyConn(conn, outboundConn)
			info.lastActiveTime = time.Now()
		} else {
			outboundConn, err := trojan.NewOutboundConnSession(req, nil, c.config)
			if err != nil {
				logger.Error(common.NewError("failed to start new outbound session").Base(err))
				return
			}
			defer outboundConn.Close()

			logger.Info("conn from", conn.RemoteAddr(), "tunneling to", req)
			proxyConn(inboundConn, outboundConn)
		}
	} else {
		defer inboundPacket.Close()
		type httpPacket struct {
			request *protocol.Request
			packet  []byte
		}
		packetChan := make(chan *httpPacket, 128)

		readHTTPPackets := func() {
			for {
				req, packet, err := inboundPacket.ReadPacket()
				if err != nil {
					logger.Error(err)
					return
				}
				packetChan <- &httpPacket{
					request: req,
					packet:  packet,
				}
			}
		}

		writeHTTPPackets := func() {
			for {
				select {
				case packet := <-packetChan:
					outboundConn, err := trojan.NewOutboundConnSession(packet.request, nil, c.config)
					if err != nil {
						logger.Error(err)
						continue
					}
					_, err = outboundConn.Write(packet.packet)
					if err != nil {
						logger.Error(err)
						continue
					}
					go func(outboundConn protocol.ConnSession) {
						buf := [1024]byte{}
						for {
							n, err := outboundConn.Read(buf[:])
							if err != nil {
								logger.Error(err)
								return
							}
							if _, err = inboundPacket.WritePacket(nil, buf[0:n]); err != nil {
								logger.Error(err)
								return
							}
						}
					}(outboundConn)
				}
			}
		}

		go readHTTPPackets()
		writeHTTPPackets()
	}
}

func (c *Client) Run() error {
	listener, err := net.Listen("tcp", c.config.LocalAddr.String())
	if err != nil {
		return common.NewError("failed to listen local address").Base(err)
	}
	defer listener.Close()

	ctx, cancel := context.WithCancel(context.Background())
	c.ctx = ctx
	c.cancel = cancel

	if c.config.TCP.MuxIdleTimeout > 0 {
		go c.checkAndCloseIdleMuxClient()
	}
	c.ctx = ctx
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
			logger.Error(err)
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
