package custom

import (
	"context"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/proxy"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/freedom"
	"github.com/p4gefau1t/trojan-go/tunnel/http"
	"github.com/p4gefau1t/trojan-go/tunnel/simplesocks"
	"github.com/p4gefau1t/trojan-go/tunnel/socks"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"github.com/p4gefau1t/trojan-go/tunnel/trojan"
	"gopkg.in/yaml.v2"
	"strings"
)

func buildNodes(ctx context.Context, nodeConfigList []NodeConfig) (map[string]*proxy.Node, *proxy.Node, error) {
	nodes := make(map[string]*proxy.Node)
	var root *proxy.Node
	for _, nodeCfg := range nodeConfigList {
		nodeCfg.Protocol = strings.ToUpper(nodeCfg.Protocol)
		if _, err := tunnel.GetTunnel(nodeCfg.Protocol); err != nil {
			return nil, nil, common.NewError("invalid protocol name:" + nodeCfg.Protocol)
		}
		data, err := yaml.Marshal(nodeCfg.Config)
		if err != nil {
			return nil, nil, common.NewError("failed to parse config data for " + nodeCfg.Tag + " with protocol" + nodeCfg.Protocol).Base(err)
		}
		nodeContext, err := config.WithYAMLConfig(ctx, data)
		node := &proxy.Node{
			Name:    nodeCfg.Protocol,
			Next:    make(map[string]*proxy.Node),
			Context: nodeContext,
		}
		nodes[nodeCfg.Tag] = node
		if nodeCfg.Protocol == transport.Name || nodeCfg.Protocol == freedom.Name {
			if root != nil {
				return nil, nil, common.NewError("transport layer is defined for twice")
			}
			log.Debug("root found:" + nodeCfg.Tag)
			root = node
		}
	}
	if root == nil {
		return nil, nil, common.NewError("no transport layer found")
	}
	return nodes, root, nil
}

func init() {
	proxy.RegisterProxyCreator(Name, func(ctx context.Context) (*proxy.Proxy, error) {
		cfg := config.FromContext(ctx, Name).(*Config)

		// inbound
		nodes, root, err := buildNodes(ctx, cfg.Inbound.Node)
		if err != nil {
			return nil, err
		}

		transportServer, err := transport.NewServer(root.Context, nil)
		if err != nil {
			return nil, common.NewError("failed to initialize transport server").Base(err)
		}
		root.Server = transportServer

		// build server tree
		for _, path := range cfg.Inbound.Path {
			lastNode := root
			for i, tag := range path {
				if _, found := nodes[tag]; !found {
					return nil, common.NewError("invalid node tag: " + tag)
				}
				if i == len(path)-1 {
					switch nodes[tag].Name {
					case trojan.Name, simplesocks.Name, socks.Name, http.Name:
					default:
						return nil, common.NewError("inbound path must end with protocol trojan/simplesocks/http/socks")
					}
				}
				if i == 0 {
					if nodes[tag].Name != transport.Name {
						return nil, common.NewError("inbound path must start with protocol transport")
					}
					continue
				}
				lastNode = lastNode.LinkNextNode(nodes[tag])
			}
			lastNode.IsEndpoint = true
		}

		servers := proxy.FindAllEndpoints(root)

		if len(cfg.Outbound.Path) != 1 {
			return nil, common.NewError("there must be only 1 path for outbound protocol stack")
		}

		// outbound
		nodes, _, err = buildNodes(ctx, cfg.Outbound.Node)
		if err != nil {
			return nil, err
		}

		// build client stack
		var client tunnel.Client
		for i, tag := range cfg.Outbound.Path[0] {
			if _, found := nodes[tag]; !found {
				return nil, common.NewError("invalid node tag: " + tag)
			}
			if i == 0 && nodes[tag].Name != freedom.Name && nodes[tag].Name != transport.Name {
				return nil, common.NewError("outbound path must start with protocol freedom/transport")
			}
			t, err := tunnel.GetTunnel(nodes[tag].Name)
			if err != nil {
				return nil, common.NewError("invalid tunnel name").Base(err)
			}
			client, err = t.NewClient(nodes[tag].Context, client)
			if err != nil {
				return nil, common.NewError("failed to create client").Base(err)
			}
		}
		return proxy.NewProxy(ctx, servers, client), nil
	})
}
