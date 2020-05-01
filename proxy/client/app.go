package client

import (
	"context"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/simplesocks"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/p4gefau1t/trojan-go/stat"
)

type AppManager struct {
	auth      stat.Authenticator
	config    *conf.GlobalConfig
	transport TransportManager
	ctx       context.Context
}

func (m *AppManager) OpenAppConn(req *protocol.Request) (protocol.ConnSession, error) {
	var outboundConn protocol.ConnSession
	//transport layer
	transport, err := m.transport.DialToServer()
	if err != nil {
		return nil, common.NewError("failed to init transport layer").Base(err)
	}
	//application layer
	if m.config.Mux.Enabled {
		outboundConn, err = simplesocks.NewOutboundConnSession(req, transport)
	} else {
		outboundConn, err = trojan.NewOutboundConnSession(req, transport, m.config, m.auth)
	}
	if err != nil {
		return nil, common.NewError("fail to start conn session").Base(err)
	}
	return outboundConn, nil
}

func NewAppManager(ctx context.Context, config *conf.GlobalConfig, auth stat.Authenticator) *AppManager {
	c := &AppManager{
		ctx:    ctx,
		config: config,
		auth:   auth,
	}
	if config.Mux.Enabled {
		log.Info("mux enabled")
		c.transport = NewMuxPoolManager(ctx, config, auth)
	} else {
		c.transport = NewTLSManager(config)
	}
	return c
}
