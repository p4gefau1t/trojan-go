package raw

import (
	"context"
	"github.com/p4gefau1t/trojan-go/config"
	"net"

	"github.com/p4gefau1t/trojan-go/tunnel"
)

type Client struct {
	preferIPv4 bool
	noDelay    bool
	keepAlive  bool
}

func (c *Client) DialConn(addr *tunnel.Address, t tunnel.Tunnel) (tunnel.Conn, error) {
	network := "tcp"
	if c.preferIPv4 {
		network = "tcp4"
	}
	tcpConn, err := net.Dial(network, addr.String())
	if err != nil {
		return nil, err
	}

	tcpConn.(*net.TCPConn).SetKeepAlive(c.keepAlive)
	tcpConn.(*net.TCPConn).SetNoDelay(c.noDelay)
	return &Conn{
		TCPConn: tcpConn.(*net.TCPConn),
	}, nil
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	network := "udp"
	if c.preferIPv4 {
		network = "udp4"
	}
	udpConn, err := net.ListenPacket(network, "")
	if err != nil {
		return nil, err
	}
	return &PacketConn{
		UDPConn: udpConn.(*net.UDPConn),
	}, nil
}

func (c *Client) Close() error {
	return nil
}

func NewClient(ctx context.Context, client tunnel.Client) (*Client, error) {
	// TODO implement dns
	cfg := config.FromContext(ctx, Name).(*Config)
	return &Client{
		noDelay:    cfg.TCP.NoDelay,
		keepAlive:  cfg.TCP.KeepAlive,
		preferIPv4: cfg.TCP.PreferIPV4,
	}, nil
}

// FixedClient will always dial to the FixedAddr
type FixedClient struct {
	FixedAddr *tunnel.Address
	Client
}

func (c *FixedClient) DialConn(addr *tunnel.Address, t tunnel.Tunnel) (tunnel.Conn, error) {
	return c.Client.DialConn(c.FixedAddr, t)
}
