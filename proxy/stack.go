package proxy

import (
	"context"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

type Node struct {
	Name       string
	Next       []*Node
	IsEndpoint bool
	tunnel.Server
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
