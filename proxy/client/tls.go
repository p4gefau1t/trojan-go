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

	fingerprints       []string
	workingFingerprint string
	fingerprintsLock   sync.Mutex
	config             *conf.GlobalConfig
	sessionCache       tls.ClientSessionCache
}

func (m *TLSManager) genClientSpec(name string) (*utls.ClientHelloSpec, error) {
	var spec *utls.ClientHelloSpec
	switch name {
	case "chrome":
		spec = &utls.ClientHelloSpec{
			TLSVersMin: utls.VersionTLS10,
			TLSVersMax: utls.VersionTLS13,
			CipherSuites: []uint16{
				utls.GREASE_PLACEHOLDER,
				utls.TLS_AES_128_GCM_SHA256,
				utls.TLS_AES_256_GCM_SHA384,
				utls.TLS_CHACHA20_POLY1305_SHA256,
				utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				utls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				utls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				utls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				utls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				utls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				utls.TLS_RSA_WITH_AES_128_CBC_SHA,
				utls.TLS_RSA_WITH_AES_256_CBC_SHA,
				utls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
			},
			CompressionMethods: []byte{
				0x00, // compressionNone
			},
			Extensions: []utls.TLSExtension{
				&utls.UtlsGREASEExtension{},
				&utls.SNIExtension{},
				&utls.UtlsExtendedMasterSecretExtension{},
				&utls.RenegotiationInfoExtension{Renegotiation: utls.RenegotiateOnceAsClient},
				&utls.SupportedCurvesExtension{[]utls.CurveID{
					utls.CurveID(utls.GREASE_PLACEHOLDER),
					utls.X25519,
					utls.CurveP256,
					utls.CurveP384,
				}},
				&utls.SupportedPointsExtension{SupportedPoints: []byte{
					0x00, // pointFormatUncompressed
				}},
				&utls.SessionTicketExtension{},
				&utls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
				&utls.StatusRequestExtension{},
				&utls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{
					utls.ECDSAWithP256AndSHA256,
					utls.PSSWithSHA256,
					utls.PKCS1WithSHA256,
					utls.ECDSAWithP384AndSHA384,
					utls.PSSWithSHA384,
					utls.PKCS1WithSHA384,
					utls.PSSWithSHA512,
					utls.PKCS1WithSHA512,
					utls.PKCS1WithSHA1,
				}},
				&utls.SCTExtension{},
				&utls.KeyShareExtension{[]utls.KeyShare{
					{Group: utls.CurveID(utls.GREASE_PLACEHOLDER), Data: []byte{0}},
					{Group: utls.X25519},
				}},
				&utls.PSKKeyExchangeModesExtension{[]uint8{
					utls.PskModeDHE,
				}},
				&utls.SupportedVersionsExtension{[]uint16{
					utls.GREASE_PLACEHOLDER,
					utls.VersionTLS13,
					utls.VersionTLS12,
					utls.VersionTLS11,
					utls.VersionTLS10,
				}},
				&utls.FakeCertCompressionAlgsExtension{[]utls.CertCompressionAlgo{
					utls.CertCompressionBrotli,
				}},
				&utls.UtlsGREASEExtension{},
				&utls.UtlsPaddingExtension{GetPaddingLen: utls.BoringPaddingStyle},
			},
		}
	case "firefox":
		spec = &utls.ClientHelloSpec{
			TLSVersMin: utls.VersionTLS10,
			TLSVersMax: utls.VersionTLS13,
			CipherSuites: []uint16{
				utls.TLS_AES_128_GCM_SHA256,
				utls.TLS_CHACHA20_POLY1305_SHA256,
				utls.TLS_AES_256_GCM_SHA384,
				utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				utls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				utls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				utls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				utls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				utls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				utls.FAKE_TLS_DHE_RSA_WITH_AES_128_CBC_SHA,
				utls.FAKE_TLS_DHE_RSA_WITH_AES_256_CBC_SHA,
				utls.TLS_RSA_WITH_AES_128_CBC_SHA,
				utls.TLS_RSA_WITH_AES_256_CBC_SHA,
				utls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
			},
			CompressionMethods: []byte{
				0, //compressionNone,
			},
			Extensions: []utls.TLSExtension{
				&utls.SNIExtension{},
				&utls.UtlsExtendedMasterSecretExtension{},
				&utls.RenegotiationInfoExtension{Renegotiation: utls.RenegotiateOnceAsClient},
				&utls.SupportedCurvesExtension{[]utls.CurveID{
					utls.X25519,
					utls.CurveP256,
					utls.CurveP384,
					utls.CurveP521,
					utls.CurveID(utls.FakeFFDHE2048),
					utls.CurveID(utls.FakeFFDHE3072),
				}},
				&utls.SupportedPointsExtension{SupportedPoints: []byte{
					0, //pointFormatUncompressed,
				}},
				&utls.SessionTicketExtension{},
				&utls.ALPNExtension{AlpnProtocols: []string{"h2", "http/1.1"}},
				&utls.StatusRequestExtension{},
				&utls.KeyShareExtension{[]utls.KeyShare{
					{Group: utls.X25519},
					{Group: utls.CurveP256},
				}},
				&utls.SupportedVersionsExtension{[]uint16{
					utls.VersionTLS13,
					utls.VersionTLS12,
					utls.VersionTLS11,
					utls.VersionTLS10}},
				&utls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{
					utls.ECDSAWithP256AndSHA256,
					utls.ECDSAWithP384AndSHA384,
					utls.ECDSAWithP521AndSHA512,
					utls.PSSWithSHA256,
					utls.PSSWithSHA384,
					utls.PSSWithSHA512,
					utls.PKCS1WithSHA256,
					utls.PKCS1WithSHA384,
					utls.PKCS1WithSHA512,
					utls.ECDSAWithSHA1,
					utls.PKCS1WithSHA1,
				}},
				&utls.PSKKeyExchangeModesExtension{[]uint8{utls.PskModeDHE}},
				&utls.FakeRecordSizeLimitExtension{0x4001},
				&utls.UtlsPaddingExtension{GetPaddingLen: utls.BoringPaddingStyle},
			},
		}
	case "ios":
		spec = &utls.ClientHelloSpec{
			TLSVersMin: utls.VersionTLS10,
			TLSVersMax: utls.VersionTLS13,
			CipherSuites: []uint16{
				utls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				utls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				utls.DISABLED_TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA384,
				utls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
				utls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
				utls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
				utls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
				utls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				utls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
				utls.DISABLED_TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA384,
				utls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
				utls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
				utls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
				utls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
				utls.TLS_RSA_WITH_AES_256_GCM_SHA384,
				utls.TLS_RSA_WITH_AES_128_GCM_SHA256,
				utls.DISABLED_TLS_RSA_WITH_AES_256_CBC_SHA256,
				utls.TLS_RSA_WITH_AES_128_CBC_SHA256,
				utls.TLS_RSA_WITH_AES_256_CBC_SHA,
				utls.TLS_RSA_WITH_AES_128_CBC_SHA,
				0xc008,
				utls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
				utls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
			},
			CompressionMethods: []byte{
				0, //compressionNone,
			},
			Extensions: []utls.TLSExtension{
				&utls.RenegotiationInfoExtension{Renegotiation: utls.RenegotiateOnceAsClient},
				&utls.SNIExtension{},
				&utls.UtlsExtendedMasterSecretExtension{},
				&utls.SignatureAlgorithmsExtension{SupportedSignatureAlgorithms: []utls.SignatureScheme{
					utls.ECDSAWithP256AndSHA256,
					utls.PSSWithSHA256,
					utls.PKCS1WithSHA256,
					utls.ECDSAWithP384AndSHA384,
					utls.ECDSAWithSHA1,
					utls.PSSWithSHA384,
					utls.PSSWithSHA384,
					utls.PKCS1WithSHA384,
					utls.PSSWithSHA512,
					utls.PKCS1WithSHA512,
					utls.PKCS1WithSHA1,
				}},
				&utls.StatusRequestExtension{},
				&utls.NPNExtension{},
				&utls.SCTExtension{},
				&utls.ALPNExtension{AlpnProtocols: []string{"h2", "h2-16", "h2-15", "h2-14", "spdy/3.1", "spdy/3", "http/1.1"}},
				&utls.SupportedPointsExtension{SupportedPoints: []byte{
					0, //pointFormatUncompressed,
				}},
				&utls.SupportedCurvesExtension{[]utls.CurveID{
					utls.X25519,
					utls.CurveP256,
					utls.CurveP384,
					utls.CurveP521,
				}},
			},
		}
	}
	if spec == nil {
		return nil, common.NewError("Invalid fingerprint:" + name)
	}
	if m.config.Websocket.Enabled {
		for i := range spec.Extensions {
			if alpn, ok := spec.Extensions[i].(*utls.ALPNExtension); ok {
				alpn.AlpnProtocols = []string{"http/1.1"}
				spec.Extensions[i] = alpn
				log.Debug("Force http/1.1")
			}
		}
	}
	return spec, nil
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
		log.Trace("TLS handshaked", tls.CipherSuiteName(state.CipherSuite), state.DidResume, state.NegotiatedProtocol)
		for i := range chain {
			for j := range chain[i] {
				log.Trace("Subject:", chain[i][j].Subject, "Issuer:", chain[i][j].Issuer)
			}
		}
	case *utls.UConn:
		tlsConn := conn.(*utls.UConn)
		state := tlsConn.ConnectionState()
		chain := state.VerifiedChains
		log.Trace("uTLS handshaked", tls.CipherSuiteName(state.CipherSuite), state.DidResume, state.NegotiatedProtocol)
		for i := range chain {
			for j := range chain[i] {
				log.Trace("Subject:", chain[i][j].Subject, "Issuer:", chain[i][j].Issuer)
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
		return nil, common.NewError("Failed to dial to remote server").Base(err)
	}
	if err := sockopt.ApplyTCPConnOption(conn.(*net.TCPConn), &m.config.TCP); err != nil {
		log.Warn(common.NewError("Failed to apply tcp options").Base(err))
	}
	return conn, nil
}

func (m *TLSManager) dialTLSWithFakeFingerprint() (*utls.UConn, error) {
	m.fingerprintsLock.Lock()
	workingFingerprint := m.workingFingerprint
	m.fingerprintsLock.Unlock()

	utlsConfig := &utls.Config{
		RootCAs:            m.config.TLS.CertPool,
		ServerName:         m.config.TLS.SNI,
		InsecureSkipVerify: !m.config.TLS.Verify,
		KeyLogWriter:       m.config.TLS.KeyLogger,
	}
	if workingFingerprint != "" {
		spec, err := m.genClientSpec(workingFingerprint)
		if err != nil {
			return nil, err
		}
		tcpConn, err := m.dialTCP()
		if err != nil {
			return nil, err // on tcp Dial failure return with error right away
		}
		tlsConn := utls.UClient(tcpConn, utlsConfig, utls.HelloCustom)
		if err := tlsConn.ApplyPreset(spec); err != nil {
			m.fingerprintsLock.Lock()
			workingFingerprint = ""
			m.fingerprintsLock.Unlock()
			log.Error(common.NewError("Failed to apply working fingerprint").Base(err))
		} else {
			protocol.SetRandomizedTimeout(tlsConn)
			err = tlsConn.Handshake()
			protocol.CancelTimeout(tlsConn)
			if err != nil {
				log.Error("Working hello id is no longer working, err:", err)
			} else {
				return tlsConn, nil
			}
		}
	}

	for _, name := range m.fingerprints {
		spec, err := m.genClientSpec(name)
		if err != nil {
			return nil, err
		}

		tcpConn, err := m.dialTCP()
		if err != nil {
			return nil, err // on tcp Dial failure return with error right away
		}

		tlsConn := utls.UClient(tcpConn, utlsConfig, utls.HelloCustom)

		if err := tlsConn.ApplyPreset(spec); err != nil {
			log.Error(common.NewError("Failed to apply fingerprint:" + name).Base(err))
			continue
		}

		protocol.SetRandomizedTimeout(tlsConn)
		err = tlsConn.Handshake()
		protocol.CancelTimeout(tlsConn)
		if err != nil {
			log.Info("Handshaking with fingerprint:", name, "failed:", err)
			continue // on tls Dial error keep trying
		}

		log.Info("Avaliable hello id found:", name)
		m.fingerprintsLock.Lock()
		m.workingFingerprint = name
		m.fingerprintsLock.Unlock()
		return tlsConn, err
	}
	return nil, common.NewError("All client hello IDs tried but failed")
}

func (m *TLSManager) DialToServer() (io.ReadWriteCloser, error) {
	if m.config.TransportPlugin.Enabled {
		// plain text
		return m.dialTCP()
	}
	var transport net.Conn
	if m.config.TLS.Fingerprint != "" {
		// use utls fingerprints
		tlsConn, err := m.dialTLSWithFakeFingerprint()
		if err != nil {
			return nil, err
		}
		m.printConnInfo(tlsConn)
		transport = tlsConn
	} else {
		// default golang tls library
		tcpConn, err := m.dialTCP()
		if err != nil {
			return nil, err
		}
		tlsConfig := &tls.Config{
			CipherSuites:           m.config.TLS.CipherSuites,
			RootCAs:                m.config.TLS.CertPool,
			ServerName:             m.config.TLS.SNI,
			InsecureSkipVerify:     !m.config.TLS.Verify,
			SessionTicketsDisabled: !m.config.TLS.SessionTicket,
			CurvePreferences:       m.config.TLS.CurvePreferences,
			NextProtos:             m.config.TLS.ALPN,
			ClientSessionCache:     m.sessionCache,
			KeyLogWriter:           m.config.TLS.KeyLogger,
		}
		tlsConn := tls.Client(tcpConn, tlsConfig)
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
			return nil, common.NewError("Failed to start websocket connection").Base(err)
		}
		return ws, nil
	}
	return transport, nil
}

func NewTLSManager(config *conf.GlobalConfig) *TLSManager {
	m := &TLSManager{
		config: config,
	}

	if config.TLS.Fingerprint != "" {
		m.fingerprints = []string{config.TLS.Fingerprint}
	}
	if config.TLS.Fingerprint == "auto" {
		m.fingerprints = []string{"chrome", "firefox", "ios"}
		rand.Shuffle(len(m.fingerprints), func(i, j int) {
			m.fingerprints[i], m.fingerprints[j] = m.fingerprints[j], m.fingerprints[i]
		})
	}
	return m
}
