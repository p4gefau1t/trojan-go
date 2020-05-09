package client

import (
	"crypto/tls"
	"io"
	"math/rand"
	"net"
	"sync"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/p4gefau1t/trojan-go/sockopt"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/proxy"
)

type TLSManager struct {
	TransportManager

	helloIDs       []utls.ClientHelloID
	helloIDLock    sync.Mutex
	workingHelloID *utls.ClientHelloID
	utlsConfig     *utls.Config
	tlsConfig      *tls.Config
	config         *conf.GlobalConfig
}

func (m *TLSManager) printConnInfo(conn net.Conn) {
	if m.config.LogLevel != 0 {
		return
	}
	switch conn.(type) {
	case *tls.Conn:
		tlsConn := conn.(*tls.Conn)
		state := tlsConn.ConnectionState()
		chain := state.VerifiedChains
		log.Trace("tls handshaked", "cipher:", tls.CipherSuiteName(state.CipherSuite), "resume:", state.DidResume)
		for i := range chain {
			for j := range chain[i] {
				log.Trace("subject:", chain[i][j].Subject, ", issuer:", chain[i][j].Issuer)
			}
		}
	case *utls.UConn:
		tlsConn := conn.(*utls.UConn)
		state := tlsConn.ConnectionState()
		chain := state.VerifiedChains
		log.Trace("utls handshaked", "cipher:", tls.CipherSuiteName(state.CipherSuite), "resume:", state.DidResume)
		for i := range chain {
			for j := range chain[i] {
				log.Trace("subject:", chain[i][j].Subject, ", issuer:", chain[i][j].Issuer)
			}
		}
	default:
		panic(conn)
	}
}

func (m *TLSManager) dialTCP() (net.Conn, error) {
	network := "tcp"
	if m.config.TCP.PreferIPV4 {
		network = "tcp4"
	}
	if m.config.ForwardProxy.Enabled {
		var auth *proxy.Auth
		if m.config.ForwardProxy.Username != "" {
			auth = &proxy.Auth{
				User:     m.config.ForwardProxy.Username,
				Password: m.config.ForwardProxy.Password,
			}
		}
		dialer, err := proxy.SOCKS5(network, m.config.ForwardProxy.ProxyAddress.String(), auth, nil)
		if err != nil {
			return nil, err
		}
		return dialer.Dial(network, m.config.RemoteAddress.String())
	}
	conn, err := net.DialTimeout(network, m.config.RemoteAddress.String(), protocol.GetRandomTimeoutDuration())
	if err != nil {
		return nil, common.NewError("failed to dial to remote server").Base(err)
	}
	if err := sockopt.ApplyTCPConnOption(conn.(*net.TCPConn), &m.config.TCP); err != nil {
		log.Warn(common.NewError("failed to apply tcp options").Base(err))
	}
	return conn, nil
}

func (m *TLSManager) dialTLSWithFakeFingerprint() (*utls.UConn, error) {
	helloIDs := make([]utls.ClientHelloID, len(m.helloIDs))
	copy(helloIDs, m.helloIDs)
	rand.Shuffle(len(m.helloIDs), func(i, j int) {
		helloIDs[i], helloIDs[j] = helloIDs[j], helloIDs[i]
	})

	m.helloIDLock.Lock()
	workingHelloID := m.workingHelloID // keep using same helloID, if it works
	m.helloIDLock.Unlock()
	if workingHelloID != nil {
		helloIDFound := false
		for i, ID := range helloIDs {
			if ID == *workingHelloID {
				helloIDs[i] = helloIDs[0]
				helloIDs[0] = *workingHelloID // push working hello ID first
				helloIDFound = true
				break
			}
		}
		if !helloIDFound {
			helloIDs = append([]utls.ClientHelloID{*workingHelloID}, helloIDs...)
			helloIDs[0], helloIDs[len(helloIDs)-1] = helloIDs[len(helloIDs)-1], helloIDs[0]
		}
	}
	for _, helloID := range helloIDs {
		tcpConn, err := m.dialTCP()
		if err != nil {
			return nil, err // on tcp Dial failure return with error right away
		}

		client := utls.UClient(tcpConn, m.utlsConfig, helloID)
		if m.config.Websocket.Enabled {
			// HACK disable alpn (http/1.1, h2) to support websocket
			client.HandshakeState.Hello.AlpnProtocols = []string{}
		}
		protocol.SetRandomizedTimeout(client)
		err = client.Handshake()
		protocol.CancelTimeout(client)
		if err != nil {
			log.Debug("hello id", helloID.Str(), "failed, err:", err)
			continue // on tls Dial error keep trying HelloIDs
		}

		log.Debug("found avaliable hello id:", helloID.Str())
		m.helloIDLock.Lock()
		m.workingHelloID = &client.ClientHelloID
		m.helloIDLock.Unlock()
		return client, err
	}
	return nil, common.NewError("all client hello id tried but failed")
}

func (m *TLSManager) DialToServer() (io.ReadWriteCloser, error) {
	var transport net.Conn
	if m.config.TLS.Fingerprint != "" {
		//use utls fingerprints
		tlsConn, err := m.dialTLSWithFakeFingerprint()
		if err != nil {
			return nil, err
		}
		m.printConnInfo(tlsConn)
		transport = tlsConn
	} else {
		//normal golang tls
		tcpConn, err := m.dialTCP()
		if err != nil {
			return nil, err
		}
		tlsConn := tls.Client(tcpConn, m.tlsConfig)
		err = tlsConn.Handshake()
		if err != nil {
			return nil, err
		}
		transport = tlsConn
		m.printConnInfo(tlsConn)
	}
	if m.config.Websocket.Enabled {
		ws, err := trojan.NewOutboundWebosocket(transport, m.config)
		if err != nil {
			transport.Close()
			return nil, common.NewError("failed to start websocket connection").Base(err)
		}
		return ws, nil
	}
	return transport, nil
}

func NewTLSManager(config *conf.GlobalConfig) *TLSManager {
	utlsConfig := &utls.Config{
		RootCAs:            config.TLS.CertPool,
		ServerName:         config.TLS.SNI,
		InsecureSkipVerify: !config.TLS.Verify,
	}
	tlsConfig := &tls.Config{
		CipherSuites:           config.TLS.CipherSuites,
		RootCAs:                config.TLS.CertPool,
		ServerName:             config.TLS.SNI,
		InsecureSkipVerify:     !config.TLS.Verify,
		SessionTicketsDisabled: !config.TLS.SessionTicket,
		CurvePreferences:       config.TLS.CurvePreferences,
		NextProtos:             config.TLS.ALPN,
		ClientSessionCache:     tls.NewLRUClientSessionCache(192),
	}

	m := &TLSManager{
		config:     config,
		utlsConfig: utlsConfig,
		tlsConfig:  tlsConfig,
	}

	if config.TLS.Fingerprint == "auto" {
		m.helloIDs = []utls.ClientHelloID{
			utls.HelloChrome_Auto,
			utls.HelloFirefox_Auto,
			utls.HelloIOS_Auto,
			utls.HelloRandomizedNoALPN,
		}
	} else if config.TLS.Fingerprint != "" {
		table := map[string]*utls.ClientHelloID{
			"chrome":     &utls.HelloChrome_Auto,
			"firefox":    &utls.HelloFirefox_Auto,
			"ios":        &utls.HelloIOS_Auto,
			"randomized": &utls.HelloRandomizedNoALPN,
		}
		id, found := table[config.TLS.Fingerprint]
		if found {
			log.Debug("tls fingerprint loaded:", id.Str())
			m.helloIDs = []utls.ClientHelloID{*id}
		} else {
			log.Warn("invalid tls fingerprint:", config.TLS.Fingerprint, ", using default fingerprint")
			config.TLS.Fingerprint = ""
		}
	}

	return m
}
