package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/p4gefau1t/trojan-go/tunnel/transport/fingerprint"
	utls "github.com/refraction-networking/utls"
	"io"
	"io/ioutil"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"

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
	reuseSession  bool
	curve         []tls.CurveID
	fingerprint   string
	keyLogger     io.WriteCloser
	websocket     bool
	plugin        bool
	cmd           *exec.Cmd
}

func (c *Client) Close() error {
	if c.cmd != nil {
		c.cmd.Process.Kill()
	}
	if c.keyLogger != nil {
		c.keyLogger.Close()
	}
	return nil
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

// DialConn implements tunnel.Client. It will ignore the params and directly dial to remote server
func (c *Client) DialConn(*tunnel.Address, tunnel.Tunnel) (tunnel.Conn, error) {
	if c.plugin {
		conn, err := net.Dial("tcp", c.serverAddress.String())
		if err != nil {
			return nil, common.NewError("transport failed to connect to plugin")
		}
		return &Conn{
			Conn: conn,
		}, nil
	}
	if c.fingerprint != "" {
		tcpConn, err := net.Dial("tcp", c.serverAddress.String())
		if err != nil {
			return nil, err
		}
		tlsConn := utls.UClient(tcpConn, &utls.Config{
			RootCAs:            c.ca,
			ServerName:         c.sni,
			InsecureSkipVerify: !c.verify,
			KeyLogWriter:       c.keyLogger,
		}, utls.HelloCustom)
		spec, err := fingerprint.GetClientHelloSpec(c.fingerprint, c.websocket)
		if err != nil {
			return nil, common.NewError("invalid hello spec").Base(err)
		}
		if err := tlsConn.ApplyPreset(spec); err != nil {
			return nil, common.NewError("transport failed to apply preset fingerprint").Base(err)
		}
		if err := tlsConn.Handshake(); err != nil {
			return nil, common.NewError("transport failed to handshake with remote server").Base(err)
		}
		return &Conn{
			Conn: tlsConn,
		}, nil
	}
	tlsConn, err := tls.Dial("tcp", c.serverAddress.String(), &tls.Config{
		InsecureSkipVerify:     !c.verify,
		ServerName:             c.sni,
		RootCAs:                c.ca,
		KeyLogWriter:           c.keyLogger,
		CipherSuites:           c.cipher,
		SessionTicketsDisabled: !c.sessionTicket,
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
	cfg := config.FromContext(ctx, Name).(*Config)
	serverAddress := tunnel.NewAddressFromHostPort("tcp", cfg.RemoteHost, cfg.RemotePort)

	if cfg.TransportPlugin.Enabled {
		var cmd *exec.Cmd
		log.Warn("trojan-go will use transport plugin and work in plain text mode")
		switch cfg.TransportPlugin.Type {
		case "shadowsocks":
			pluginHost := "127.0.0.1"
			pluginPort := common.PickPort("tcp", pluginHost)
			cfg.TransportPlugin.Env = append(
				cfg.TransportPlugin.Env,
				"SS_LOCAL_HOST="+pluginHost,
				"SS_LOCAL_PORT="+strconv.FormatInt(int64(pluginPort), 10),
				"SS_REMOTE_HOST="+cfg.RemoteHost,
				"SS_REMOTE_PORT="+strconv.FormatInt(int64(cfg.RemotePort), 10),
				"SS_PLUGIN_OPTIONS="+cfg.TransportPlugin.PluginOption,
			)
			cfg.RemoteHost = pluginHost
			cfg.RemotePort = pluginPort
			serverAddress = tunnel.NewAddressFromHostPort("tcp", cfg.RemoteHost, cfg.RemotePort)
			log.Debug("plugin address", serverAddress.String())
			log.Debug("plugin env", cfg.TransportPlugin.Env)

			cmd = exec.Command(cfg.TransportPlugin.Command, cfg.TransportPlugin.Arg...)
			cmd.Env = append(cmd.Env, cfg.TransportPlugin.Env...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			cmd.Start()
		case "other":
			cmd = exec.Command(cfg.TransportPlugin.Command, cfg.TransportPlugin.Arg...)
			cmd.Env = append(cmd.Env, cfg.TransportPlugin.Env...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			cmd.Start()
		case "plaintext":
			// do nothing
		default:
			return nil, common.NewError("invalid plugin type: " + cfg.TransportPlugin.Type)
		}
		client := &Client{
			serverAddress: serverAddress,
			cmd:           cmd,
			plugin:        true,
		}
		return client, nil
	}

	if cfg.TLS.Fingerprint != "" {
		_, err := fingerprint.GetClientHelloSpec(cfg.TLS.Fingerprint, cfg.Websocket.Enabled)
		if err != nil {
			return nil, err
		}
		log.Info("tls fingerprint", cfg.TLS.Fingerprint, "applied")
	}

	if cfg.TLS.SNI == "" {
		cfg.TLS.SNI = cfg.RemoteHost
		log.Warn("tls sni is unspecified. using remote-address")
	}

	client := &Client{
		verify:        cfg.TLS.Verify,
		sni:           cfg.TLS.SNI,
		serverAddress: serverAddress,
		cipher:        fingerprint.ParseCipher(strings.Split(cfg.TLS.Cipher, ":")),
		sessionTicket: cfg.TLS.ReuseSession,
		fingerprint:   cfg.TLS.Fingerprint,
		websocket:     cfg.Websocket.Enabled,
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
	} else if len(cfg.TLS.CertBytes) != 0 {
		client.ca = x509.NewCertPool()
		ok := client.ca.AppendCertsFromPEM(cfg.TLS.CertBytes)
		if !ok {
			log.Warn("invalid cert list")
		}
		log.Info("using custom cert (data)")
	}

	if cfg.TLS.CertPath == "" && len(cfg.TLS.CertBytes) == 0 {
		log.Info("cert is unspecified, using default ca list")
	}

	log.Debug("transport client created")
	return client, nil
}
