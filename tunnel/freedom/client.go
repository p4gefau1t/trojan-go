package freedom

import (
	"context"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"golang.org/x/net/proxy"
)

type Client struct {
	preferIPv4   bool
	noDelay      bool
	keepAlive    bool
	ctx          context.Context
	cancel       context.CancelFunc
	forwardProxy bool
	proxyAddr    *tunnel.Address
	username     string
	password     string
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
		noDelay:      cfg.TCP.NoDelay,
		keepAlive:    cfg.TCP.KeepAlive,
		preferIPv4:   cfg.TCP.PreferIPV4,
		forwardProxy: cfg.ForwardProxy.Enabled,
		proxyAddr:    addr,
	}, nil
}
