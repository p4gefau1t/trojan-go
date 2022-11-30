package tls

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"
	"strings"

	tls "github.com/refraction-networking/utls"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/tls/fingerprint"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
)

// Client is a tls client
type Client struct {
	verify        bool
	sni           string
	ca            *x509.CertPool
	cipher        []uint16
	sessionTicket bool
	reuseSession  bool
	fingerprint   string
	helloID       tls.ClientHelloID
	keyLogger     io.WriteCloser
	underlay      tunnel.Client
}

func (c *Client) Close() error {
	if c.keyLogger != nil {
		c.keyLogger.Close()
	}
	return c.underlay.Close()
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

func (c *Client) DialConn(_ *tunnel.Address, overlay tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := c.underlay.DialConn(nil, &Tunnel{})
	if err != nil {
		return nil, common.NewError("tls failed to dial conn").Base(err)
	}

	// utls fingerprint
	tlsConn := tls.UClient(conn, &tls.Config{
		RootCAs:            c.ca,
		ServerName:         c.sni,
		InsecureSkipVerify: !c.verify,
		KeyLogWriter:       c.keyLogger,
	}, c.helloID)
	if err := tlsConn.Handshake(); err != nil {
		return nil, common.NewError("tls failed to handshake with remote server").Base(err)
	}
	return &transport.Conn{
		Conn: tlsConn,
	}, nil
}

// NewClient creates a tls client
func NewClient(ctx context.Context, underlay tunnel.Client) (*Client, error) {
	cfg := config.FromContext(ctx, Name).(*Config)

	helloID := tls.ClientHelloID{}
	// keep the parameter name consistent with upstream
	// https://github.com/refraction-networking/utls/blob/35e5b05fc4b6f8c4351d755f2570bc293f30aaf6/u_common.go#L114-L132
	if cfg.TLS.Fingerprint != "" {
		switch strings.ToLower(cfg.TLS.Fingerprint) {
		case "chrome":
			helloID = tls.HelloChrome_Auto
		case "ios":
			helloID = tls.HelloIOS_Auto
		case "firefox":
			helloID = tls.HelloFirefox_Auto
		case "edge":
			helloID = tls.HelloEdge_Auto
		case "safari":
			helloID = tls.HelloSafari_Auto
		case "360browser":
			helloID = tls.Hello360_Auto
		case "qqbrowser":
			helloID = tls.HelloQQ_Auto
		default:
			return nil, common.NewError("Invalid 'fingerprint' value in configuration: '" + cfg.TLS.Fingerprint + "'. Possible values are 'chrome' (default), 'ios', 'firefox', 'edge', 'safari', '360browser', or 'qqbrowser'.")
		}
		log.Info("Your trojan's TLS fingerprint will look like", cfg.TLS.Fingerprint)
	} else {
		helloID = tls.HelloChrome_Auto
		log.Info("No 'fingerprint' value specified in your configuration. Your trojan's TLS fingerprint will look like Chrome by default.")
	}

	if cfg.TLS.SNI == "" {
		cfg.TLS.SNI = cfg.RemoteHost
		log.Warn("tls sni is unspecified")
	}

	client := &Client{
		underlay:      underlay,
		verify:        cfg.TLS.Verify,
		sni:           cfg.TLS.SNI,
		cipher:        fingerprint.ParseCipher(strings.Split(cfg.TLS.Cipher, ":")),
		sessionTicket: cfg.TLS.ReuseSession,
		fingerprint:   cfg.TLS.Fingerprint,
		helloID:       helloID,
	}

	if cfg.TLS.CertPath != "" {
		caCertByte, err := ioutil.ReadFile(cfg.TLS.CertPath)
		if err != nil {
			return nil, common.NewError("failed to load cert file").Base(err)
		}
		client.ca = x509.NewCertPool()
		ok := client.ca.AppendCertsFromPEM(caCertByte)
		if !ok {
			log.Warn("invalid cert list")
		}
		log.Info("using custom cert")

		// print cert info
		pemCerts := caCertByte
		for len(pemCerts) > 0 {
			var block *pem.Block
			block, pemCerts = pem.Decode(pemCerts)
			if block == nil {
				break
			}
			if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
				continue
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				continue
			}
			log.Trace("issuer:", cert.Issuer, "subject:", cert.Subject)
		}
	}

	if cfg.TLS.CertPath == "" {
		log.Info("cert is unspecified, using default ca list")
	}

	log.Debug("tls client created")
	return client, nil
}
