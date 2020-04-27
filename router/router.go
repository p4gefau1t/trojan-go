package router

import (
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type Policy int

const (
	Proxy Policy = iota
	Bypass
	Block
	Unknown

	Match
	NonMatch
)

type EmptyRouter struct{}

func (r *EmptyRouter) RouteRequest(req *protocol.Request) (Policy, error) {
	return Proxy, nil
}

type Router interface {
	RouteRequest(*protocol.Request) (Policy, error)
}

var NewRouter func(config *conf.RouterConfig) (Router, error) = NewEmptyRouter

func NewEmptyRouter(*conf.RouterConfig) (Router, error) {
	return &EmptyRouter{}, nil
}
