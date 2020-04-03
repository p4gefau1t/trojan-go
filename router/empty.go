package router

import "github.com/p4gefau1t/trojan-go/protocol"

type EmptyRouter struct {
	DefaultPolicy Policy
}

func (r *EmptyRouter) RouteRequest(req *protocol.Request) (Policy, error) {
	return r.DefaultPolicy, nil
}
