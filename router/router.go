package router

import (
	"github.com/p4gefau1t/trojan-go/protocol"
)

type Policy int

const (
	Proxy Policy = iota
	Bypass
	Block
	Unknown

	match
	nonMatch
)

type Router interface {
	RouteRequest(*protocol.Request) (Policy, error)
}
