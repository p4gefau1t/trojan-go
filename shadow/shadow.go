package shadow

import (
	"context"
	"io"
	"net"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/proxy"
)

type Scapegoat struct {
	Conn          io.ReadWriteCloser
	ShadowConn    io.ReadWriteCloser
	ShadowAddress *common.Address
	Info          string
}

type ShadowManager struct {
	config        *conf.GlobalConfig
	ctx           context.Context
	scapegoatChan chan *Scapegoat
}

func (m *ShadowManager) SubmitScapegoat(goat *Scapegoat) {
	m.scapegoatChan <- goat
	log.Debug("scapegoat commited")
}

func (m *ShadowManager) handleScapegoat() {
	for {
		select {
		case goat := <-m.scapegoatChan:
			if goat.Info != "" {
				log.Info("Scapegoat: ", goat.Info)
			}
			//cancel the deadline
			if conn, ok := goat.Conn.(net.Conn); ok {
				conn.SetDeadline(time.Time{})
			}
			if goat.ShadowConn == nil {
				if goat.ShadowAddress == nil {
					panic("incorrect shadow server")
				}
				var err error
				goat.ShadowConn, err = net.Dial("tcp", goat.ShadowAddress.String())
				if err != nil {
					log.Error(common.NewError("Failed to dial to shadow server").Base(err))
					continue
				}
			}
			go proxy.ProxyConn(m.ctx, goat.Conn, goat.ShadowConn, m.config.BufferSize)
		case <-m.ctx.Done():
			log.Debug("shadow manager exiting..")
			return
		}
	}
}

func NewShadowManager(ctx context.Context, config *conf.GlobalConfig) *ShadowManager {
	m := &ShadowManager{
		config:        config,
		ctx:           ctx,
		scapegoatChan: make(chan *Scapegoat, 1024),
	}
	go m.handleScapegoat()
	return m
}
