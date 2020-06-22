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
	tunnel.Client
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

func (n *Node) LinkNextNode(next *Node) *Node {
	if next, found := n.Next[next.Name]; found {
		return next
	}
	n.Next[next.Name] = next
	t, err := tunnel.GetTunnel(next.Name)
	if err != nil {
		log.Fatal(err)
	}
	s, err := t.NewServer(next.Context, n.Server) // context of the child nodes have been initialized
	if err != nil {
		log.Fatal(err)
	}
	next.Server = s
	return next
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
