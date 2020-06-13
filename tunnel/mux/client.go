package mux

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/xtaci/smux"
)

type muxID uint32

func generateMuxID() muxID {
	return muxID(rand.Uint32())
}

type smuxClientInfo struct {
	id             muxID
	client         *smux.Session
	lastActiveTime time.Time
	underlayConn   tunnel.Conn
}

//Client is a smux client
type Client struct {
	clientPoolLock sync.Mutex
	clientPool     map[muxID]*smuxClientInfo
	underlay       tunnel.Client
	concurrency    int
	timeout        time.Duration
	ctx            context.Context
	cancel         context.CancelFunc
}

func (c *Client) Close() error {
	c.cancel()
	c.clientPoolLock.Lock()
	defer c.clientPoolLock.Unlock()
	for id, info := range c.clientPool {
		info.client.Close()
		log.Debug("mux client", id, "closed")
	}
	return nil
}

func (c *Client) cleanLoop() {
	var checkDuration time.Duration
	if c.timeout <= 0 {
		checkDuration = time.Second * 10
		log.Warn("negative mux timeout")
	} else {
		checkDuration = c.timeout / 4
	}
	log.Debug("check duration:", checkDuration.Seconds(), "s")
	for {
		select {
		case <-time.After(checkDuration):
			c.clientPoolLock.Lock()
			for id, info := range c.clientPool {
				if info.client.IsClosed() {
					info.client.Close()
					info.underlayConn.Close()
					delete(c.clientPool, id)
					log.Info("mux client", id, "is dead")
				} else if info.client.NumStreams() == 0 && time.Now().Sub(info.lastActiveTime) > c.timeout {
					info.client.Close()
					info.underlayConn.Close()
					delete(c.clientPool, id)
					log.Info("mux client", id, "is closed due to inactivity")
				}
			}
			log.Debug("current mux clients: ", len(c.clientPool))
			for id, info := range c.clientPool {
				log.Debug(fmt.Sprintf("  - %x: %d/%d", id, info.client.NumStreams(), c.concurrency))
			}
			c.clientPoolLock.Unlock()
		case <-c.ctx.Done():
			log.Debug("shutting down mux cleaner..")
			c.clientPoolLock.Lock()
			for id, info := range c.clientPool {
				info.client.Close()
				info.underlayConn.Close()
				delete(c.clientPool, id)
				log.Debug("mux client", id, "closed")
			}
			c.clientPoolLock.Unlock()
			return
		}
	}
}

func (c *Client) newMuxClient() (*smuxClientInfo, error) {
	// The mutex should be locked when this function is called
	id := generateMuxID()
	if _, found := c.clientPool[id]; found {
		return nil, common.NewError("duplicated id")
	}

	fakeAddr := &tunnel.Address{
		DomainName:  "MUX_CONN",
		AddressType: tunnel.DomainName,
	}
	conn, err := c.underlay.DialConn(fakeAddr, &Tunnel{})
	if err != nil {
		return nil, common.NewError("mux failed to dial").Base(err)
	}
	conn = newStickyConn(conn)

	smuxConfig := smux.DefaultConfig()
	//smuxConfig.KeepAliveDisabled = true
	client, err := smux.Client(conn, smuxConfig)
	info := &smuxClientInfo{
		client:         client,
		underlayConn:   conn,
		id:             id,
		lastActiveTime: time.Now(),
	}
	c.clientPool[id] = info
	return info, nil
}

func (c *Client) DialConn(*tunnel.Address, tunnel.Tunnel) (tunnel.Conn, error) {
	createNewConn := func(info *smuxClientInfo) (tunnel.Conn, error) {
		rwc, err := info.client.Open()
		info.lastActiveTime = time.Now()
		if err != nil {
			info.underlayConn.Close()
			info.client.Close()
			delete(c.clientPool, info.id)
			return nil, common.NewError("mux failed to open stream from client").Base(err)
		}
		return &Conn{
			rwc:  rwc,
			Conn: info.underlayConn,
		}, nil
	}

	c.clientPoolLock.Lock()
	defer c.clientPoolLock.Unlock()
	for _, info := range c.clientPool {
		if info.client.IsClosed() {
			delete(c.clientPool, info.id)
			log.Info(fmt.Sprintf("Mux client %x is closed", info.id))
			continue
		}
		if info.client.NumStreams() < c.concurrency || c.concurrency <= 0 {
			return createNewConn(info)
		}
	}

	info, err := c.newMuxClient()
	if err != nil {
		return nil, common.NewError("no available mux client found").Base(err)
	}
	return createNewConn(info)
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

func NewClient(ctx context.Context, underlay tunnel.Client) (*Client, error) {
	clientConfig := config.FromContext(ctx, Name).(*Config)
	ctx, cancel := context.WithCancel(ctx)
	client := &Client{
		underlay:    underlay,
		concurrency: clientConfig.Mux.Concurrency,
		timeout:     time.Duration(clientConfig.Mux.Timeout) * time.Second,
		ctx:         ctx,
		cancel:      cancel,
		clientPool:  make(map[muxID]*smuxClientInfo),
	}
	go client.cleanLoop()
	log.Debug("mux client created")
	return client, nil
}
