package raw

import (
	"context"
	"net"

	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

const Name = "RAW"

type Tunnel struct{}

func (*Tunnel) Name() string {
	return Name
}

func (*Tunnel) NewClient(ctx context.Context, client tunnel.Client) (tunnel.Client, error) {
	return NewClient(ctx, client)
}

func (*Tunnel) NewServer(ctx context.Context, client tunnel.Server) (tunnel.Server, error) {
	serverConfig := config.FromContext(ctx, Name).(*Config)
	addr := tunnel.NewAddressFromHostPort("tcp", serverConfig.LocalHost, serverConfig.LocalPort)

	l, err := net.Listen("tcp", addr.String())
	if err != nil {
		return nil, err
	}
	return &Server{
		addr:        addr,
		tcpListener: l,
	}, nil
}

func init() {
	tunnel.RegisterTunnel(Name, &Tunnel{})
}
