package proxy

import (
	"context"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

type Node struct {
	Name       string
	Next       map[string]*Node
	IsEndpoint bool
	context.Context
	tunnel.Server
}

func (n *Node) BuildNext(name string) *Node {
	if next, found := n.Next[name]; found {
		return next
	}
	t, err := tunnel.GetTunnel(name)
	if err != nil {
		log.Fatal(err)
	}
	s, err := t.NewServer(n.Context, n.Server)
	if err != nil {
		log.Fatal(err)
	}
	newNode := &Node{
		Name:    name,
		Next:    make(map[string]*Node),
		Context: n.Context,
		Server:  s,
	}
	n.Next[name] = newNode
	return newNode
}

func FindAllEndpoints(root *Node) []tunnel.Server {
	list := make([]tunnel.Server, 0)
	if root.IsEndpoint || len(root.Next) == 0 {
		list = append(list, root.Server)
	}
	for _, next := range root.Next {
		list = append(list, FindAllEndpoints(next)...)
	}
	return list
}

func buildServerStacksTree(ctx context.Context, current *Node, parent *Node) ([]tunnel.Server, error) {
	t, err := tunnel.GetTunnel(current.Name)
	if err != nil {
		return nil, err
	}
	current.Server, err = t.NewServer(ctx, parent)
	if err != nil {
		return nil, err
	}
	leaves := make([]tunnel.Server, 0)
	for _, child := range current.Next {
		subTreeLeaves, err := buildServerStacksTree(ctx, child, current)
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, subTreeLeaves...)
	}
	// current node is a leave node
	if len(leaves) == 0 || current.IsEndpoint {
		leaves = append(leaves, current)
	}
	return leaves, nil
}

func CreateServersStacksTree(ctx context.Context, root *Node) ([]tunnel.Server, error) {
	return buildServerStacksTree(ctx, root, nil)
}

// CreateClientStack create client tunnel stacks from lists
func CreateClientStack(ctx context.Context, clientStack []string) (tunnel.Client, error) {
	var client tunnel.Client
	for _, name := range clientStack {
		t, err := tunnel.GetTunnel(name)
		if err != nil {
			return nil, err
		}
		client, err = t.NewClient(ctx, client)
		if err != nil {
			return nil, err
		}
	}
	return client, nil
}

// CreateServerStack create server tunnel stack from list
func CreateServerStack(ctx context.Context, serverStack []string) (tunnel.Server, error) {
	var server tunnel.Server
	for _, name := range serverStack {
		t, err := tunnel.GetTunnel(name)
		if err != nil {
			return nil, err
		}
		server, err = t.NewServer(ctx, server)
		if err != nil {
			return nil, err
		}
	}
	return server, nil
}
