package websocket

import (
	"context"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"golang.org/x/net/websocket"
	"strings"
)

type Client struct {
	underlay tunnel.Client
	hostname string
	path     string
}

func (c *Client) DialConn(*tunnel.Address, tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := c.underlay.DialConn(nil, &Tunnel{})
	if err != nil {
		return nil, common.NewError("websocket cannot dial with underlying client").Base(err)
	}
	url := "wss://" + c.hostname + c.path
	origin := "https://" + c.hostname
	wsConfig, err := websocket.NewConfig(url, origin)
	if err != nil {
		return nil, common.NewError("invalid websocket config").Base(err)
	}
	wsConn, err := websocket.NewClient(wsConfig, conn)
	if err != nil {
		return nil, common.NewError("websocket failed to handshake with server").Base(err)
	}
	return &OutboundConn{
		Conn: wsConn,
	}, nil
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	return nil, common.NewError("not supported by websocket")
}

func (c *Client) Close() error {
	return c.underlay.Close()
}

func NewClient(ctx context.Context, underlay tunnel.Client) (*Client, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	if !strings.HasPrefix(cfg.Websocket.Path, "/") {
		return nil, common.NewError("websocket path must start with \"/\"")
	}
	if cfg.Websocket.Hostname == "" {
		cfg.Websocket.Hostname = cfg.RemoteHost
		log.Warn("empty websocket hostname")
	}
	log.Debug("websocket client created")
	return &Client{
		hostname: cfg.Websocket.Hostname,
		path:     cfg.Websocket.Path,
		underlay: underlay,
	}, nil
}
