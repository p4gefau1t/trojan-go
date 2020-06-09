package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

// Client implements tunnel.Client
type Client struct {
	serverAddress *tunnel.Address
	verify        bool
	sni           string
	ca            *x509.CertPool
	cipher        []uint16
	sessionTicket bool
	curve         []tls.CurveID
	keyLogger     io.WriteCloser
}

func (c *Client) Close() error {
	return c.keyLogger.Close()
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	return nil, common.NewError("Not supported")
}

// DialConn implements tunnel.Client. It will ignore the params and directly dial to remote server
func (c *Client) DialConn(*tunnel.Address, tunnel.Tunnel) (tunnel.Conn, error) {
	tlsConn, err := tls.Dial("tcp", c.serverAddress.String(), &tls.Config{
		InsecureSkipVerify: !c.verify,
		ServerName:         c.sni,
		RootCAs:            c.ca,
		KeyLogWriter:       c.keyLogger,
	})
	if err != nil {
		return nil, err
	}
	return &Conn{
		Conn: tlsConn,
	}, nil
}

// NewClient creates a transport layer client
func NewClient(ctx context.Context, c tunnel.Client) (*Client, error) {
	clientConfig := config.FromContext(ctx, Name).(*Config)
	serverAddress := tunnel.NewAddressFromHostPort("tcp", clientConfig.RemoteHost, clientConfig.RemotePort)
	client := &Client{
		verify:        clientConfig.TLS.Verify,
		sni:           clientConfig.TLS.SNI,
		serverAddress: serverAddress,
	}
	if clientConfig.TLS.CertPath != "" {
		caCertByte, err := ioutil.ReadFile(clientConfig.TLS.CertPath)
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
	} else if len(clientConfig.TLS.CertBytes) != 0 {
		client.ca = x509.NewCertPool()
		ok := client.ca.AppendCertsFromPEM(clientConfig.TLS.CertBytes)
		if !ok {
			log.Warn("invalid cert list")
		}
		log.Info("using custom cert (data)")
	}

	if clientConfig.TLS.CertPath == "" && len(clientConfig.TLS.CertBytes) == 0 {
		log.Info("cert is unspecified, using default ca list")
	}

	log.Debug("transport client created")
	return client, nil
}
