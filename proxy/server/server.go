package server

import (
	"context"
	"github.com/p4gefau1t/trojan-go/proxy"
	"github.com/p4gefau1t/trojan-go/tunnel/mux"
	"github.com/p4gefau1t/trojan-go/tunnel/raw"
	"github.com/p4gefau1t/trojan-go/tunnel/simplesocks"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"github.com/p4gefau1t/trojan-go/tunnel/trojan"
	"github.com/p4gefau1t/trojan-go/tunnel/websocket"
)

const Name = "SERVER"

func init() {
	proxy.RegisterProxyCreator(Name, func(ctx context.Context) (*proxy.Proxy, error) {
		clientStack := []string{raw.Name}
		serverTree := &proxy.Node{
			Name: transport.Name,
			Next: []*proxy.Node{
				{
					Name:       trojan.Name,
					IsEndpoint: true,
					Next: []*proxy.Node{
						{
							Name: mux.Name,
							Next: []*proxy.Node{
								{
									Name: simplesocks.Name,
								},
							},
						},
					},
				},
				{
					Name: websocket.Name,
					Next: []*proxy.Node{
						{
							Name:       trojan.Name,
							IsEndpoint: true,
							Next: []*proxy.Node{
								{
									Name: mux.Name,
									Next: []*proxy.Node{
										{
											Name: simplesocks.Name,
										},
									},
								},
							},
						},
					},
				},
			},
		}
		c, err := proxy.CreateClientStack(ctx, clientStack)
		if err != nil {
			return nil, err
		}
		s, err := proxy.CreateServersStacksTree(ctx, serverTree)
		if err != nil {
			return nil, err
		}
		return proxy.NewProxy(ctx, s, c), nil
	})

}
