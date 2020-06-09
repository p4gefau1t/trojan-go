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

// NewClient creates a raw client
func (*Tunnel) NewClient(ctx context.Context, _ tunnel.Client) (tunnel.Client, error) {
	return &Client{}, nil
}

// NewServer creates a raw server, which is used by "Forward"
func (*Tunnel) NewServer(ctx context.Context, _ tunnel.Server) (tunnel.Server, error) {
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
