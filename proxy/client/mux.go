package client

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"sync"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/p4gefau1t/trojan-go/stat"
	"github.com/xtaci/smux"
)

// HACK stick the smux 8 bytes header to the payload
type smuxStickyReadWriteCloser struct {
	io.ReadWriteCloser
	synQueue chan []byte
	finQueue chan []byte
}

func (rwc *smuxStickyReadWriteCloser) stickToPayload(p []byte) []byte {
	buf := make([]byte, 0, len(p)+16)
	for {
		select {
		case header := <-rwc.synQueue:
			buf = append(buf, header...)
		default:
			goto stick1
		}
	}
stick1:
	buf = append(buf, p...)
	for {
		select {
		case header := <-rwc.finQueue:
			buf = append(buf, header...)
		default:
			goto stick2
		}
	}
stick2:
	return buf
}

func (rwc *smuxStickyReadWriteCloser) Close() error {
	const maxPaddingLength = 512
	padding := [maxPaddingLength + 8]byte{'A', 'B', 'C', 'D', 'E', 'F'} // for debugging
	buf := rwc.stickToPayload(nil)
	rwc.Write(append(buf, padding[:rand.Intn(maxPaddingLength)]...))
	return rwc.ReadWriteCloser.Close()
}

func (rwc *smuxStickyReadWriteCloser) Write(p []byte) (int, error) {
	if len(p) == 8 {
		if p[0] == 1 || p[0] == 2 { //smux 8 bytes header
			switch p[1] {
			// THE CONTENT OF THE BUFFER MIGHT CHANGE
			// NEVER STORE THE POINTER TO HEADER, COPY THE HEADER INSTEAD
			case 0:
				// cmdSYN
				header := make([]byte, 8)
				copy(header, p)
				rwc.synQueue <- header
				return 8, nil
			case 1:
				// cmdFIN
				header := make([]byte, 8)
				copy(header, p)
				rwc.finQueue <- header
				return 8, nil
			}
		} else {
			log.Debug("Unknown 8 bytes")
		}
	}
	_, err := rwc.ReadWriteCloser.Write(rwc.stickToPayload(p))
	return len(p), err
}

func newSmuxStickyReadWriteCloser(rwc io.ReadWriteCloser) *smuxStickyReadWriteCloser {
	return &smuxStickyReadWriteCloser{
		ReadWriteCloser: rwc,
		synQueue:        make(chan []byte, 128),
		finQueue:        make(chan []byte, 128),
	}
}

type muxID uint32

func generateMuxID() muxID {
	return muxID(rand.Uint32())
}

type muxClientInfo struct {
	id             muxID
	client         *smux.Session
	lastActiveTime time.Time
}

type MuxManager struct {
	TransportManager

	sync.Mutex
	muxPool   map[muxID]*muxClientInfo
	config    *conf.GlobalConfig
	auth      stat.Authenticator
	ctx       context.Context
	transport *TLSManager
}

func (m *MuxManager) newMuxClient() (*muxClientInfo, error) {
	id := generateMuxID()
	if _, found := m.muxPool[id]; found {
		return nil, common.NewError("Duplicated id")
	}
	req := &protocol.Request{
		Command: protocol.Mux,
		Address: &common.Address{
			DomainName:  "MUX_CONN",
			AddressType: common.DomainName,
		},
	}
	rwc, err := m.transport.DialToServer()
	if err != nil {
		return nil, common.NewError("Failed to dail to remote server").Base(err)
	}
	trojanConn, err := trojan.NewOutboundConnSession(req, rwc, m.config, m.auth)
	if err != nil {
		rwc.Close()
		log.Error(common.NewError("Failed to dial tls tunnel").Base(err))
		return nil, err
	}

	smuxRWC := newSmuxStickyReadWriteCloser(trojanConn)

	smuxConfig := smux.DefaultConfig()
	smuxConfig.KeepAliveDisabled = true
	client, err := smux.Client(smuxRWC, smuxConfig)
	common.Must(err)
	log.Info(fmt.Sprintf("Mux TLS tunnel established with mux client %x", id))
	return &muxClientInfo{
		client:         client,
		id:             id,
		lastActiveTime: time.Now(),
	}, nil
}

func (m *MuxManager) pickMuxClient() (*muxClientInfo, error) {
	m.Lock()
	defer m.Unlock()

	for _, info := range m.muxPool {
		if info.client.IsClosed() {
			delete(m.muxPool, info.id)
			log.Info(fmt.Sprintf("Mux client %x is closed", info.id))
			continue
		}
		if info.client.NumStreams() < m.config.Mux.Concurrency || m.config.Mux.Concurrency <= 0 {
			info.lastActiveTime = time.Now()
			return info, nil
		}
	}

	select {
	case <-m.ctx.Done():
		return nil, common.NewError("Mux manager closed")
	default:
	}

	// not found
	info, err := m.newMuxClient()
	if err != nil {
		return nil, err
	}
	m.muxPool[info.id] = info
	return info, nil
}

func (m *MuxManager) DialToServer() (io.ReadWriteCloser, error) {
	info, err := m.pickMuxClient()
	if err != nil {
		return nil, err
	}
	stream, err := info.client.OpenStream()
	if err != nil {
		m.Lock()
		defer m.Unlock()
		delete(m.muxPool, info.id)
		info.client.Close()
		log.Warn(common.NewError(fmt.Sprintf("Somthing wrong with mux client %x, closing", info.id)).Base(err))
		return nil, err
	}
	log.Info(fmt.Sprintf("New mux stream established with mux client %x", info.id))
	info.lastActiveTime = time.Now()
	return stream, nil
}

func (m *MuxManager) checkAndCloseIdleMuxClient() {
	var muxIdleDuration, checkDuration time.Duration
	if m.config.Mux.IdleTimeout <= 0 {
		muxIdleDuration = 0
		checkDuration = time.Second * 10
		log.Warn("Invalid mux idle timeout")
	} else {
		muxIdleDuration = time.Duration(m.config.Mux.IdleTimeout) * time.Second
		checkDuration = muxIdleDuration / 4
	}
	for {
		select {
		case <-time.After(checkDuration):
			m.Lock()
			for id, info := range m.muxPool {
				if info.client.IsClosed() {
					delete(m.muxPool, id)
					log.Info("Mux client", id, "is dead")
				} else if info.client.NumStreams() == 0 && time.Now().Sub(info.lastActiveTime) > muxIdleDuration {
					info.client.Close()
					delete(m.muxPool, id)
					log.Info("Mux client", id, "is closed due to inactive")
				}
			}
			log.Debug("Current mux pool clients: ", len(m.muxPool))
			for i, c := range m.muxPool {
				log.Debug(fmt.Sprintf("    Client %x: %d/%d", i, c.client.NumStreams(), m.config.Mux.Concurrency))
			}
			m.Unlock()
		case <-m.ctx.Done():
			log.Debug("Shutting down mux manager..")
			m.Lock()
			for id, info := range m.muxPool {
				info.client.Close()
				log.Debug("Mux client", id, "closed")
			}
			m.Unlock()
			return
		}
	}
}

func NewMuxPoolManager(ctx context.Context, config *conf.GlobalConfig, auth stat.Authenticator) *MuxManager {
	m := &MuxManager{
		ctx:       ctx,
		config:    config,
		muxPool:   make(map[muxID]*muxClientInfo),
		transport: NewTLSManager(config),
		auth:      auth,
	}
	go m.checkAndCloseIdleMuxClient()
	return m
}
