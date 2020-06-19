package custom

import (
	"context"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/proxy"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"gopkg.in/yaml.v2"
	"strings"
)

func buildNodes(ctx context.Context, nodeConfigList []NodeConfig) (map[string]*proxy.Node, error) {
	nodes := make(map[string]*proxy.Node)
	for _, nodeCfg := range nodeConfigList {
		nodeCfg.Protocol = strings.ToUpper(nodeCfg.Protocol)
		if _, err := tunnel.GetTunnel(nodeCfg.Protocol); err != nil {
			return nil, common.NewError("invalid protocol name:" + nodeCfg.Protocol)
		}
		data, err := yaml.Marshal(nodeCfg.Config)
		if err != nil {
			return nil, common.NewError("failed to parse config data for " + nodeCfg.Tag + " with protocol" + nodeCfg.Protocol).Base(err)
		}
		nodeContext, err := config.WithYAMLConfig(ctx, data)
		node := &proxy.Node{
			Name:    nodeCfg.Protocol,
			Next:    make(map[string]*proxy.Node),
			Context: nodeContext,
		}
		nodes[nodeCfg.Tag] = node
	}
	return nodes, nil
}

func init() {
	proxy.RegisterProxyCreator(Name, func(ctx context.Context) (*proxy.Proxy, error) {
		cfg := config.FromContext(ctx, Name).(*Config)

		// inbound
		nodes, err := buildNodes(ctx, cfg.Inbound.Node)
		if err != nil {
			return nil, err
		}

		var root *proxy.Node
		// build server tree
		for _, path := range cfg.Inbound.Path {
			lastNode := root
			for _, tag := range path {
				if _, found := nodes[tag]; !found {
					return nil, common.NewError("invalid node tag: " + tag)
				}
				if lastNode == nil {
					if root != nil {
						panic("root != nil")
					}
					lastNode = nodes[tag]
					root = lastNode
					t, err := tunnel.GetTunnel(root.Name)
					if err != nil {
						return nil, common.NewError("failed to find root tunnel").Base(err)
					}
					s, err := t.NewServer(root.Context, nil)
					if err != nil {
						return nil, common.NewError("failed to init root server").Base(err)
					}
					root.Server = s
				} else {
					lastNode = lastNode.LinkNextNode(nodes[tag])
				}
			}
			lastNode.IsEndpoint = true
		}

		servers := proxy.FindAllEndpoints(root)

		if len(cfg.Outbound.Path) != 1 {
			return nil, common.NewError("there must be only 1 path for outbound protocol stack")
		}

		// outbound
		nodes, err = buildNodes(ctx, cfg.Outbound.Node)
		if err != nil {
			return nil, err
		}

		// build client stack
		var client tunnel.Client
		for _, tag := range cfg.Outbound.Path[0] {
			if _, found := nodes[tag]; !found {
				return nil, common.NewError("invalid node tag: " + tag)
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
