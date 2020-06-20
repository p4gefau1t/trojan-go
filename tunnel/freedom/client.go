package freedom

import (
	"context"
	"crypto/tls"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"golang.org/x/net/proxy"
	"net"
	"net/url"
	"strconv"
	"time"
)

type Client struct {
	preferIPv4   bool
	noDelay      bool
	keepAlive    bool
	dns          []string
	ctx          context.Context
	cancel       context.CancelFunc
	forwardProxy bool
	proxyAddr    *tunnel.Address
	username     string
	password     string
}

func (c *Client) resolveIP(addr *tunnel.Address) ([]net.IPAddr, error) {
	for _, s := range c.dns {
		var dnsAddr string
		var dnsHost, dnsType string
		var err error

		dnsURL, err := url.Parse(s)
		if err != nil || dnsURL.Scheme == "" {
			dnsType = "udp"
			dnsAddr = s
		} else {
			dnsType = dnsURL.Scheme
			dnsAddr = dnsURL.Host
		}

		dnsHost, tmp, err := net.SplitHostPort(dnsAddr)
		dnsPort, err := strconv.ParseInt(tmp, 10, 32)
		common.Must(err)

		if err != nil {
			dnsHost = dnsAddr
			switch dnsType {
			case "dot":
				dnsPort = 853
			case "tcp", "udp":
				dnsPort = 53
			}
		}

		resolver := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				dnsAddress := tunnel.NewAddressFromHostPort("tcp", dnsHost, int(dnsPort))
				switch dnsType {
				case "udp", "tcp":
					d := net.Dialer{
						Timeout: time.Second * 5,
					}
					conn, err := d.DialContext(ctx, dnsType, dnsAddress.String())
					if err != nil {
						return nil, err
					}
					return conn, nil
				case "dot":
					tlsConn, err := tls.Dial("tcp", dnsAddress.String(), nil)
					if err != nil {
						return nil, err
					}
					return tlsConn, nil
				}
				return nil, common.NewError("invalid dns type:" + dnsType)
			},
		}
		ip, err := resolver.LookupIPAddr(c.ctx, addr.String())
		if err != nil {
			log.Error(err)
			continue
		}
		return ip, nil
	}
	return nil, common.NewError("address not found")
}

func (c *Client) DialConn(addr *tunnel.Address, t tunnel.Tunnel) (tunnel.Conn, error) {
	network := "tcp"
	if c.preferIPv4 {
		network = "tcp4"
	}

	// forward proxy
	if c.forwardProxy {
		var auth *proxy.Auth
		if c.username != "" {
			auth = &proxy.Auth{
				User:     c.username,
				Password: c.password,
			}
		}
		dialer, err := proxy.SOCKS5(network, c.proxyAddr.String(), auth, proxy.Direct)
		if err != nil {
			return nil, err
		}
		socksConn, err := dialer.Dial(network, addr.String())
		if err != nil {
			return nil, err
		}
		return &Conn{
			Conn: socksConn,
		}, nil
	}
	dialer := new(net.Dialer)
	tcpConn, err := dialer.DialContext(c.ctx, network, addr.String())
	if err != nil {
		return nil, err
	}

	tcpConn.(*net.TCPConn).SetKeepAlive(c.keepAlive)
	tcpConn.(*net.TCPConn).SetNoDelay(c.noDelay)
	return &Conn{
		Conn: tcpConn,
	}, nil
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	network := "udp"
	if c.preferIPv4 {
		network = "udp4"
	}
	udpConn, err := net.ListenPacket(network, "")
	if err != nil {
		return nil, common.NewError("freedom failed to listen udp socket").Base(err)
	}
	return &PacketConn{
		UDPConn: udpConn.(*net.UDPConn),
	}, nil
}

func (c *Client) Close() error {
	c.cancel()
	return nil
}

func NewClient(ctx context.Context, _ tunnel.Client) (*Client, error) {
	// TODO implement dns
	// TODO socks5 udp
	cfg := config.FromContext(ctx, Name).(*Config)
	addr := tunnel.NewAddressFromHostPort("tcp", cfg.ForwardProxy.ProxyHost, cfg.ForwardProxy.ProxyPort)
	ctx, cancel := context.WithCancel(ctx)
	return &Client{
		ctx:          ctx,
		cancel:       cancel,
		dns:          cfg.DNS,
		noDelay:      cfg.TCP.NoDelay,
		keepAlive:    cfg.TCP.KeepAlive,
		preferIPv4:   cfg.TCP.PreferIPV4,
		forwardProxy: cfg.ForwardProxy.Enabled,
		proxyAddr:    addr,
	}, nil
}
