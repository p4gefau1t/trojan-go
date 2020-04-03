package router

import (
	"github.com/p4gefau1t/trojan-go/protocol"
)

type MixedRouter struct {
	proxyList     *ListRouter
	bypassList    *ListRouter
	blockList     *ListRouter
	defaultPolicy Policy
}

func (r *MixedRouter) RouteRequest(req *protocol.Request) (Policy, error) {
	policy, err := r.bypassList.RouteRequest(req)
	if err != nil {
		return Unknown, err
	}
	if policy == match {
		return Bypass, nil
	}

	policy, err = r.blockList.RouteRequest(req)
	if err != nil {
		return Unknown, err
	}
	if policy == match {
		return Block, nil
	}

	policy, err = r.proxyList.RouteRequest(req)
	if err != nil {
		return Unknown, err
	}
	if policy == match {
		return Proxy, nil
	}
	return r.defaultPolicy, nil
}

func NewMixedRouter(defaultPolicy Policy, allResolveToIP bool, resolveToIPOnNonMatch bool, proxyIP []byte, proxyDomain []byte, bypassIP []byte, bypassDomain []byte, blockIP []byte, blockDomain []byte) (Router, error) {
	r := &MixedRouter{
		defaultPolicy: defaultPolicy,
	}
	var err error
	if r.blockList, err = NewListRouter(match, nonMatch, allResolveToIP, resolveToIPOnNonMatch, blockIP, blockDomain); err != nil {
		return nil, err
	}
	if r.bypassList, err = NewListRouter(match, nonMatch, allResolveToIP, resolveToIPOnNonMatch, bypassIP, bypassDomain); err != nil {
		return nil, err
	}
	if r.proxyList, err = NewListRouter(match, nonMatch, allResolveToIP, resolveToIPOnNonMatch, proxyIP, proxyIP); err != nil {
		return nil, err
	}
	return r, nil
}
