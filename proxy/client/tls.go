package client

import (
	"crypto/tls"
	"io"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/p4gefau1t/trojan-go/sockopt"
	utls "github.com/refraction-networking/utls"
)

type Roller struct {
	HelloIDs            []utls.ClientHelloID
	HelloIDMu           sync.Mutex
	WorkingHelloID      *utls.ClientHelloID
	TCPDialTimeout      time.Duration
	TLSHandshakeTimeout time.Duration
	TLSConfig           *utls.Config
}

// NewRoller creates Roller object with default range of HelloIDs to cycle through until a
// working/unblocked one is found.
func NewRoller(config *utls.Config) *Roller {
	tcpDialTimeoutInc := rand.Intn(14)
	tcpDialTimeoutInc = 7 + tcpDialTimeoutInc

	tlsHandshakeTimeoutInc := rand.Intn(20)
	tlsHandshakeTimeoutInc = 11 + tlsHandshakeTimeoutInc

	return &Roller{
		HelloIDs: []utls.ClientHelloID{
			utls.HelloChrome_Auto,
			utls.HelloFirefox_Auto,
			utls.HelloIOS_Auto,
			utls.HelloRandomized,
		},
		TCPDialTimeout:      time.Second * time.Duration(tcpDialTimeoutInc),
		TLSHandshakeTimeout: time.Second * time.Duration(tlsHandshakeTimeoutInc),
		TLSConfig:           config,
	}
}

func (c *Roller) Dial(network, addr, serverName string) (*utls.UConn, error) {
	helloIDs := make([]utls.ClientHelloID, len(c.HelloIDs))
	copy(helloIDs, c.HelloIDs)
	rand.Shuffle(len(c.HelloIDs), func(i, j int) {
		helloIDs[i], helloIDs[j] = helloIDs[j], helloIDs[i]
	})

	c.HelloIDMu.Lock()
	workingHelloID := c.WorkingHelloID // keep using same helloID, if it works
	c.HelloIDMu.Unlock()
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
		}
	}

	var tcpConn net.Conn
	var err error
	for _, helloID := range helloIDs {
		tcpConn, err = net.DialTimeout(network, addr, c.TCPDialTimeout)
		if err != nil {
			return nil, err // on tcp Dial failure return with error right away
		}

		client := utls.UClient(tcpConn, c.TLSConfig, helloID)
		client.SetSNI(serverName)
		client.SetDeadline(time.Now().Add(c.TLSHandshakeTimeout))
		err = client.Handshake()
		client.SetDeadline(time.Time{}) // unset timeout
		if err != nil {
			log.Debug("hello id", helloID.Str(), "failed, err:", err)
			continue // on tls Dial error keep trying HelloIDs
		}

		log.Debug("found avaliable hello id:", helloID.Str())
		c.HelloIDMu.Lock()
		c.WorkingHelloID = &client.ClientHelloID
		c.HelloIDMu.Unlock()
		return client, err
	}
	return nil, err
}

type TLSManager struct {
	TransportManager

	utlsConfig        *utls.Config
	tlsConfig         *tls.Config
	autoClientHelloID *utls.ClientHelloID
	config            *conf.GlobalConfig
	roller            *Roller
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
		log.Debug("tls handshaked", "cipher:", tls.CipherSuiteName(state.CipherSuite), "resume:", state.DidResume)
		for i := range chain {
			for j := range chain[i] {
				log.Debug("subject:", chain[i][j].Subject, ", issuer:", chain[i][j].Issuer)
			}
		}
	case *utls.UConn:
		tlsConn := conn.(*utls.UConn)
		state := tlsConn.ConnectionState()
		chain := state.VerifiedChains
		log.Debug("tls handshaked", "cipher:", tls.CipherSuiteName(state.CipherSuite), "resume:", state.DidResume)
		for i := range chain {
			for j := range chain[i] {
				log.Debug("subject:", chain[i][j].Subject, ", issuer:", chain[i][j].Issuer)
			}
		}
	default:
		panic(conn)
	}
}

func (m *TLSManager) DialToServer() (io.ReadWriteCloser, error) {
	network := "tcp"
	if m.config.TCP.PreferIPV4 {
		network = "tcp4"
	}
	var tlsConn net.Conn
	var err error
	if m.config.TLS.Fingerprint == "auto" {
		//use utls roller
		tlsConn, err = m.roller.Dial(network, m.config.RemoteAddress.String(), m.config.TLS.SNI)
	} else if m.config.TLS.ClientHelloID != nil {
		//use utls fixed fingerprint
		log.Debug("using fingerprint", m.config.TLS.ClientHelloID.Str())
		var conn net.Conn
		conn, err = net.Dial(network, m.config.RemoteAddress.String())
		tlsConn = utls.UClient(conn, m.utlsConfig, *m.config.TLS.ClientHelloID)
	} else {
		//normal golang tls
		conn, err := net.Dial(network, m.config.RemoteAddress.String())
		if err != nil {
			return nil, err
		}
		err = sockopt.ApplyTCPConnOption(conn.(*net.TCPConn), &m.config.TCP)
		if err != nil {
			return nil, common.NewError("failed to apply tcp option").Base(err)
		}
		tlsConn = tls.Client(conn, m.tlsConfig)
		err = tlsConn.(*tls.Conn).Handshake()
	}
	if err != nil {
		return nil, common.NewError("cannot dial to the remote server").Base(err)
	}
	m.printConnInfo(tlsConn)
	var transport io.ReadWriteCloser = tlsConn
	if m.config.Websocket.Enabled {
		ws, err := trojan.NewOutboundWebosocket(tlsConn, m.config)
		if err != nil {
			return nil, common.NewError("failed to start websocket connection").Base(err)
		}
		transport = ws
	}
	return transport, nil
}

func NewTLSManager(config *conf.GlobalConfig) *TLSManager {
	utlsConfig := &utls.Config{
		RootCAs:                config.TLS.CertPool,
		ServerName:             config.TLS.SNI,
		InsecureSkipVerify:     !config.TLS.Verify,
		SessionTicketsDisabled: !config.TLS.SessionTicket,
		ClientSessionCache:     utls.NewLRUClientSessionCache(-1),
	}
	tlsConfig := &tls.Config{
		CipherSuites:           config.TLS.CipherSuites,
		RootCAs:                config.TLS.CertPool,
		ServerName:             config.TLS.SNI,
		InsecureSkipVerify:     !config.TLS.Verify,
		SessionTicketsDisabled: !config.TLS.SessionTicket,
		CurvePreferences:       config.TLS.CurvePreferences,
		NextProtos:             config.TLS.ALPN,
		ClientSessionCache:     tls.NewLRUClientSessionCache(-1),
	}
	m := &TLSManager{
		config:     config,
		utlsConfig: utlsConfig,
		tlsConfig:  tlsConfig,
		roller:     NewRoller(utlsConfig),
	}
	return m
}
