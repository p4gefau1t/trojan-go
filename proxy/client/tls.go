package client

import (
	"crypto/tls"
	"io"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
)

type TLSManager struct {
	TransportManager

	tlsConfig *tls.Config
	config    *conf.GlobalConfig
}

func (m *TLSManager) DialToServer() (io.ReadWriteCloser, error) {
	network := "tcp"
	if m.config.TCP.PreferIPV4 {
		network = "tcp4"
	}
	tlsConn, err := tls.Dial(network, m.config.RemoteAddress.String(), m.tlsConfig)
	if err != nil {
		return nil, common.NewError("cannot dial to the remote server").Base(err)
	}
	if m.config.LogLevel == 0 {
		state := tlsConn.ConnectionState()
		chain := state.VerifiedChains
		log.Debug("tls handshaked", "cipher:", tls.CipherSuiteName(state.CipherSuite), "resume:", state.DidResume)
		for i := range chain {
			for j := range chain[i] {
				log.Debug("subject:", chain[i][j].Subject, ", issuer:", chain[i][j].Issuer)
			}
		}
	}
	var conn io.ReadWriteCloser = tlsConn
	if m.config.Websocket.Enabled {
		ws, err := trojan.NewOutboundWebosocket(tlsConn, m.config)
		if err != nil {
			return nil, common.NewError("failed to start websocket connection").Base(err)
		}
		conn = ws
	}
	return conn, nil
}

func NewTLSManager(config *conf.GlobalConfig) *TLSManager {
	tlsConfig := &tls.Config{
		CipherSuites:           config.TLS.CipherSuites,
		RootCAs:                config.TLS.CertPool,
		ServerName:             config.TLS.SNI,
		InsecureSkipVerify:     !config.TLS.Verify,
		SessionTicketsDisabled: !config.TLS.SessionTicket,
	}
	if config.TLS.ReuseSession {
		tlsConfig.ClientSessionCache = tls.NewLRUClientSessionCache(-1)
	}
	m := &TLSManager{
		config:    config,
		tlsConfig: tlsConfig,
	}
	return m
}
