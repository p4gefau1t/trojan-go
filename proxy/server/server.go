package server

import (
	"context"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/proxy"
	"github.com/p4gefau1t/trojan-go/proxy/client"
	"github.com/p4gefau1t/trojan-go/tunnel/mux"
	"github.com/p4gefau1t/trojan-go/tunnel/raw"
	"github.com/p4gefau1t/trojan-go/tunnel/router"
	"github.com/p4gefau1t/trojan-go/tunnel/shadowsocks"
	"github.com/p4gefau1t/trojan-go/tunnel/simplesocks"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"github.com/p4gefau1t/trojan-go/tunnel/trojan"
	"github.com/p4gefau1t/trojan-go/tunnel/websocket"
)

const Name = "SERVER"

func init() {
	proxy.RegisterProxyCreator(Name, func(ctx context.Context) (*proxy.Proxy, error) {
		cfg := config.FromContext(ctx, Name).(*client.Config)
		s, err := transport.NewServer(ctx, nil)
		if err != nil {
			return nil, err
		}
		clientStack := []string{raw.Name}
		if cfg.Router.Enabled {
			clientStack = []string{raw.Name, router.Name}
		}
		root := &proxy.Node{
			Name:       transport.Name,
			Next:       make(map[string]*proxy.Node),
			IsEndpoint: false,
			Context:    ctx,
			Server:     s,
		}

		trojanSubTree := root
		if cfg.Shadowsocks.Enabled {
			trojanSubTree = root.BuildNext(shadowsocks.Name)
		}
		trojanSubTree.BuildNext(trojan.Name).BuildNext(mux.Name).BuildNext(simplesocks.Name).IsEndpoint = true
		trojanSubTree.BuildNext(trojan.Name).IsEndpoint = true

		if cfg.Websocket.Enabled {
			wsSubTree := root.BuildNext(websocket.Name)
			if cfg.Shadowsocks.Enabled {
				wsSubTree = wsSubTree.BuildNext(shadowsocks.Name)
			}
			wsSubTree.BuildNext(trojan.Name).BuildNext(mux.Name).BuildNext(simplesocks.Name).IsEndpoint = true
			wsSubTree.BuildNext(trojan.Name).IsEndpoint = true
		}

		serverList := proxy.FindAllEndpoints(root)
		clientList, err := proxy.CreateClientStack(ctx, clientStack)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
		return proxy.NewProxy(ctx, serverList, clientList), nil
	})

}
