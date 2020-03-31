package client

import (
	"context"
	"math/rand"
	"sync"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
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

type muxPoolManager struct {
	sync.Mutex
	muxPool map[muxID]*muxClientInfo
	config  *conf.GlobalConfig
	ctx     context.Context
}

func (m *muxPoolManager) newMuxClient() (*muxClientInfo, error) {
	id := generateMuxID()
	if _, found := m.muxPool[id]; found {
		return nil, common.NewError("duplicated id")
	}
	req := &protocol.Request{
		Command:     protocol.Mux,
		DomainName:  []byte("MUX_CONN"),
		AddressType: protocol.DomainName,
	}
	conn, err := trojan.NewOutboundConnSession(req, nil, m.config)
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

func (m *muxPoolManager) pickMuxClient() (*muxClientInfo, error) {
	m.Lock()
	defer m.Unlock()

	for _, info := range m.muxPool {
		if info.client.IsClosed() {
			delete(m.muxPool, info.id)
			logger.Info("mux", info.id, "is dead")
			continue
		}
		if info.client.NumStreams() < m.config.TCP.MuxConcurrency || m.config.TCP.MuxConcurrency <= 0 {
			info.lastActiveTime = time.Now()
			return info, nil
		}
	}

	//not found
	info, err := m.newMuxClient()
	if err != nil {
		return nil, err
	}
	m.muxPool[info.id] = info
	return info, nil
}

func (m *muxPoolManager) OpenMuxConn() (*smux.Stream, *muxClientInfo, error) {
	info, err := m.pickMuxClient()
	if err != nil {
		return nil, nil, err
	}
	stream, err := info.client.OpenStream()
	if err != nil {
		return nil, nil, err
	}
	info.lastActiveTime = time.Now()
	return stream, info, nil
}

func (m *muxPoolManager) checkAndCloseIdleMuxClient() {
	var muxIdleDuration, checkDuration time.Duration
	if m.config.TCP.MuxIdleTimeout <= 0 {
		muxIdleDuration = 0
		checkDuration = time.Second * 10
		logger.Warn("invalid mux idle timeout")
	} else {
		muxIdleDuration = time.Duration(m.config.TCP.MuxIdleTimeout) * time.Second
		checkDuration = muxIdleDuration / 4
	}
	for {
		select {
		case <-time.After(checkDuration):
			m.Lock()
			for id, info := range m.muxPool {
				if info.client.IsClosed() {
					delete(m.muxPool, id)
					logger.Info("mux", id, "is dead")
				} else if info.client.NumStreams() == 0 && time.Now().Sub(info.lastActiveTime) > muxIdleDuration {
					info.client.Close()
					delete(m.muxPool, id)
					logger.Info("mux", id, "is closed due to inactive")
				}
			}
			if len(m.muxPool) != 0 {
				logger.Info("current mux pool conn num", len(m.muxPool))
			}
			m.Unlock()
		case <-m.ctx.Done():
			m.Lock()
			for id, info := range m.muxPool {
				info.client.Close()
				logger.Info("mux", id, "closed")
			}
			m.Unlock()
			return
		}
	}
}

func NewMuxPoolManager(ctx context.Context, config *conf.GlobalConfig) (*muxPoolManager, error) {
	m := &muxPoolManager{
		ctx:     ctx,
		config:  config,
		muxPool: make(map[muxID]*muxClientInfo),
	}
	go m.checkAndCloseIdleMuxClient()
	return m, nil
}
