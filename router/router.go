package router

import (
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type Policy int
type Strategy int

const (
	Proxy Policy = iota
	Bypass
	Block
	Unknown

	Match
	NonMatch
)

const (
	AsIs Strategy = iota
	IPIfNonMatch
	IPOnDemand
)

type EmptyRouter struct{}

func (r *EmptyRouter) RouteRequest(req *protocol.Request) (Policy, error) {
	return Proxy, nil
}

func NewEmptyRouter(*conf.RouterConfig) (Router, error) {
	return &EmptyRouter{}, nil
}

type Router interface {
	RouteRequest(*protocol.Request) (Policy, error)
}

var NewRouter func(config *conf.RouterConfig) (Router, error) = NewEmptyRouter
