package websocket

import (
	"context"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

const Name = "WEBSOCKET"

type Tunnel struct {
}

func (*Tunnel) Name() string {
	return Name
}

func (*Tunnel) NewServer(ctx context.Context, underlay tunnel.Server) (tunnel.Server, error) {
	return NewServer(ctx, underlay)
}

func (*Tunnel) NewClient(ctx context.Context, underlay tunnel.Client) (tunnel.Client, error) {
	return NewClient(ctx, underlay)
}

func init() {
	tunnel.RegisterTunnel(Name, &Tunnel{})
}
