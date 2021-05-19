package socks

import (
	"context"

	"github.com/p4gefau1t/trojan-go/tunnel"
)

const Name = "SOCKS"

type Tunnel struct{}

func (*Tunnel) Name() string {
	return Name
}

func (*Tunnel) NewClient(context.Context, tunnel.Client) (tunnel.Client, error) {
	panic("not supported")
}

func (*Tunnel) NewServer(ctx context.Context, server tunnel.Server) (tunnel.Server, error) {
	return NewServer(ctx, server)
}

func init() {
	tunnel.RegisterTunnel(Name, &Tunnel{})
}
