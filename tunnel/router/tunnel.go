package router

import (
	"context"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

const Name = "ROUTER"

type Tunnel struct {
}

func (t *Tunnel) Name() string {
	return Name
}

func (t *Tunnel) NewClient(ctx context.Context, client tunnel.Client) (tunnel.Client, error) {
	panic("implement me")
}

func (t *Tunnel) NewServer(ctx context.Context, server tunnel.Server) (tunnel.Server, error) {
	panic("not supported")
}
