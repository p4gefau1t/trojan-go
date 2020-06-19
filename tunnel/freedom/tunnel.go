package freedom

import (
	"context"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

const Name = "FREEDOM"

type Tunnel struct{}

func (*Tunnel) Name() string {
	return Name
}

func (*Tunnel) NewClient(ctx context.Context, client tunnel.Client) (tunnel.Client, error) {
	return NewClient(ctx, client)
}

func (*Tunnel) NewServer(ctx context.Context, client tunnel.Server) (tunnel.Server, error) {
	panic("not supported")
}

func init() {
	tunnel.RegisterTunnel(Name, &Tunnel{})
}
