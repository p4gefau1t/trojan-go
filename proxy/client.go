package proxy

import (
	"context"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
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
	sync.Mutex
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

	id := generateMuxID()
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
		if !info.client.IsClosed() && info.client.NumStreams() < c.config.TCP.MuxConcurrency {
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
			logger.Debug("current mux pool conn", len(c.muxPool))
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

func (c *Client) proxyToMuxConn(req *protocol.Request, conn protocol.ConnSession) {
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
	proxyConn(conn, outboundConn)
}

func (c *Client) handleConn(conn net.Conn) {
	inboundConn, err := socks.NewInboundConnSession(conn)
	if err != nil {
		logger.Error("failed to start new inbound session:", err)
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
			logger.Error("failed to listen udp:", err)
			return
		}

		req.IP = c.config.LocalIP
		port, err := protocol.ParsePort(listenConn.LocalAddr())
		common.Must(err)
		req.Port = uint16(port)
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
		logger.Error("failed to respond:", err)
		return
	}

	if c.config.TCP.Mux {
		logger.Info("conn from", conn.RemoteAddr(), "mux tunneling to", req)
		c.proxyToMuxConn(req, inboundConn)
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
			logger.Error("error occured when accpeting conn", err)
			continue
		}
		go c.handleConn(conn)
	}
}

func (c *Client) Close() error {
	logger.Info("shutting down client..")
	c.cancel()
	return nil
}
