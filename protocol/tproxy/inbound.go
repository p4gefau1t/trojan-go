// +build linux

package tproxy

import (
	"context"
	"net"
	"sync"
	"time"

	"github.com/LiamHaworth/go-tproxy"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type TProxyInboundConnSession struct {
	protocol.ConnSession
	reqeust *protocol.Request
	conn    net.Conn
}

func (i *TProxyInboundConnSession) Read(p []byte) (int, error) {
	return i.conn.Read(p)
}

func (i *TProxyInboundConnSession) Write(p []byte) (int, error) {
	return i.conn.Write(p)
}

func (i *TProxyInboundConnSession) Close() error {
	return i.conn.Close()
}

func (i *TProxyInboundConnSession) GetRequest() *protocol.Request {
	return i.reqeust
}

func (i *TProxyInboundConnSession) parseRequest() error {
	addr, err := getOriginalTCPDest(i.conn.(*net.TCPConn))
	if err != nil {
		return common.NewError("failed to get original dst").Base(err)
	}
	req := &protocol.Request{
		Address: &common.Address{
			IP:   addr.IP,
			Port: addr.Port,
		},
		Command: protocol.Connect,
	}
	if addr.IP.To4() != nil {
		req.AddressType = common.IPv4
	} else {
		req.AddressType = common.IPv6
	}
	i.reqeust = req
	return nil
}

func NewInboundConnSession(conn net.Conn) (protocol.ConnSession, *protocol.Request, error) {
	i := &TProxyInboundConnSession{
		conn: conn,
	}
	if err := i.parseRequest(); err != nil {
		return nil, nil, common.NewError("failed to parse request").Base(err)
	}
	return i, i.reqeust, nil
}

type udpSession struct {
	src    *net.UDPAddr
	dst    *net.UDPAddr
	expire time.Time
}

type NATInboundPacketSession struct {
	protocol.PacketSession
	request      *protocol.Request
	conn         *net.UDPConn
	tableMutex   sync.Mutex
	sessionTable map[string]*udpSession
	ctx          context.Context
	cancel       context.CancelFunc
}

func (i *NATInboundPacketSession) cleanExpiredSession() {
	for {
		i.tableMutex.Lock()
		now := time.Now()
		for k, v := range i.sessionTable {
			if now.After(v.expire) {
				delete(i.sessionTable, k)
			}
		}
		i.tableMutex.Unlock()
		select {
		case <-time.After(protocol.UDPTimeout):
		case <-i.ctx.Done():
			i.conn.Close()
			return
		}
	}
}

func (i *NATInboundPacketSession) WritePacket(req *protocol.Request, packet []byte) (int, error) {
	i.tableMutex.Lock()
	defer i.tableMutex.Unlock()
	session, found := i.sessionTable[req.String()]
	if !found {
		return 0, common.NewError("session not found " + req.String())
	}
	conn, err := tproxy.DialUDP("udp", session.dst, session.src)
	if err != nil {
		return 0, common.NewError("cannot dial to source").Base(err)
	}
	defer conn.Close()
	return conn.Write(packet)
}

func (i *NATInboundPacketSession) ReadPacket() (*protocol.Request, []byte, error) {
	buf := [protocol.MaxUDPPacketSize]byte{}
	n, src, dst, err := tproxy.ReadFromUDP(i.conn, buf[:])
	if err != nil {
		return nil, nil, err
	}

	if err != nil {
		return nil, nil, err
	}
	i.tableMutex.Lock()
	i.sessionTable[dst.String()] = &udpSession{
		src:    src,
		dst:    dst,
		expire: time.Now().Add(protocol.UDPTimeout),
	}
	i.tableMutex.Unlock()
	log.Debug("tproxy udp packet from", src, "to", dst)
	req := &protocol.Request{
		Address: &common.Address{
			IP:          dst.IP,
			Port:        dst.Port,
			NetworkType: "udp",
		},
	}
	if dst.IP.To4() != nil {
		req.AddressType = common.IPv4
	} else {
		req.AddressType = common.IPv6
	}
	return req, buf[0:n], nil
}

func (i *NATInboundPacketSession) Close() error {
	i.cancel()
	return nil
}

func NewInboundPacketSession(ctx context.Context, config *conf.GlobalConfig) (protocol.PacketSession, error) {
	localIP, err := config.LocalAddress.ResolveIP(false)
	if err != nil {
		return nil, err
	}
	addr := &net.UDPAddr{
		IP:   localIP,
		Port: int(config.LocalAddress.Port),
	}
	conn, err := tproxy.ListenUDP("udp", addr)
	if err != nil {
		return nil, common.NewError("failed to listen udp addr").Base(err)
	}
	ctx, cancel := context.WithCancel(ctx)
	i := &NATInboundPacketSession{
		conn:         conn,
		sessionTable: make(map[string]*udpSession, 1024),
		ctx:          ctx,
		cancel:       cancel,
	}
	go i.cleanExpiredSession()
	return i, nil
}
