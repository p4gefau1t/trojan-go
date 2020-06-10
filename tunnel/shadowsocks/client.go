package shadowsocks

import (
	"context"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"github.com/shadowsocks/go-shadowsocks2/core"
)

type Client struct {
	underlay tunnel.Client
	core.Cipher
}

func (c *Client) DialConn(address *tunnel.Address, tunnel tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := c.underlay.DialConn(address, tunnel)
	if err != nil {
		return nil, err
	}
	return &transport.Conn{
		Conn: c.Cipher.StreamConn(conn),
	}, nil
}

func (c *Client) DialPacket(tunnel tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

func (c *Client) Close() error {
	return c.underlay.Close()
}

func NewClient(ctx context.Context, underlay tunnel.Client) (*Client, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	cipher, err := core.PickCipher(cfg.Shadowsocks.Method, nil, cfg.Shadowsocks.Password)
	if err != nil {
		return nil, common.NewError("invalid shadowsocks cipher").Base(err)
	}
	log.Info("shadowsocks client created")
	return &Client{
		underlay: underlay,
		Cipher:   cipher,
	}, nil
}
