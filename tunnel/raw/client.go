package raw

import (
	"context"
	"net"

	"github.com/p4gefau1t/trojan-go/tunnel"
)

type Client struct{}

func (c *Client) DialConn(addr *tunnel.Address, t tunnel.Tunnel) (tunnel.Conn, error) {
	tcpConn, err := net.Dial("tcp", addr.String())
	if err != nil {
		return nil, err
	}
	return &Conn{
		TCPConn: tcpConn.(*net.TCPConn),
	}, nil
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	udpConn, err := net.ListenPacket("udp", "")
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

func NewFreeClient(context.Context, tunnel.Client) (*Client, error) {
	return &Client{}, nil
}

// FixedClient will always dial to the FixedAddr
type FixedClient struct {
	FixedAddr *tunnel.Address
	Client
}

func (c *FixedClient) DialConn(addr *tunnel.Address, t tunnel.Tunnel) (tunnel.Conn, error) {
	return c.Client.DialConn(c.FixedAddr, t)
}
