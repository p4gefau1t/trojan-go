package router

import (
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type MixedRouter struct {
	proxyList     *ListRouter
	bypassList    *ListRouter
	blockList     *ListRouter
	proxyGeo      *GeoRouter
	bypassGeo     *GeoRouter
	blockGeo      *GeoRouter
	defaultPolicy Policy
}

func (r *MixedRouter) match(router Router, req *protocol.Request) bool {
	policy, err := router.RouteRequest(req)
	if err != nil {
		log.DefaultLogger.Warn(common.NewError("match error").Base(err))
		return false
	}
	if policy == match {
		return true
	}
	return false
}

func (r *MixedRouter) RouteRequest(req *protocol.Request) (Policy, error) {

	if r.match(r.blockGeo, req) {
		return Block, nil
	}
	if r.match(r.blockList, req) {
		return Block, nil
	}

	if r.match(r.bypassGeo, req) {
		return Bypass, nil
	}
	if r.match(r.bypassList, req) {
		return Bypass, nil
	}

	if r.match(r.proxyGeo, req) {
		return Proxy, nil
	}
	if r.match(r.proxyList, req) {
		return Proxy, nil
	}

	return r.defaultPolicy, nil
}

func NewMixedRouter(config *conf.GlobalConfig) (Router, error) {
	var defaultPolicy Policy

	switch config.Router.DefaultPolicy {
	case "proxy":
		defaultPolicy = Proxy
	case "bypass":
		defaultPolicy = Bypass
	case "block":
		defaultPolicy = Block
	}

	routeByIP := config.Router.RouteByIP
	routeByIPOnNonmatch := config.Router.RouteByIPOnNonmatch

	block := config.Router.BlockList
	bypass := config.Router.BypassList
	proxy := config.Router.ProxyList

	r := &MixedRouter{
		defaultPolicy: defaultPolicy,
	}

	var err error
	if r.blockList, err = NewListRouter(match, nonMatch, routeByIP, routeByIPOnNonmatch, block); err != nil {
		return nil, err
	}
	if r.bypassList, err = NewListRouter(match, nonMatch, routeByIP, routeByIPOnNonmatch, bypass); err != nil {
		return nil, err
	}
	if r.proxyList, err = NewListRouter(match, nonMatch, routeByIP, routeByIPOnNonmatch, proxy); err != nil {
		return nil, err
	}

	r.blockGeo, _ = NewGeoRouter(match, nonMatch, routeByIP, routeByIPOnNonmatch)
	r.bypassGeo, _ = NewGeoRouter(match, nonMatch, routeByIP, routeByIPOnNonmatch)
	r.proxyGeo, _ = NewGeoRouter(match, nonMatch, routeByIP, routeByIPOnNonmatch)

	if err := r.blockGeo.LoadGeoData(config.Router.GeoIP, config.Router.BlockIPCode, config.Router.GeoSite, config.Router.BlockSiteCode); err != nil {
		//return nil, err
		log.DefaultLogger.Warn(err)
	}
	if err := r.bypassGeo.LoadGeoData(config.Router.GeoIP, config.Router.BypassIPCode, config.Router.GeoSite, config.Router.BypassSiteCode); err != nil {
		//return nil, err
		log.DefaultLogger.Warn(err)
	}
	if err := r.proxyGeo.LoadGeoData(config.Router.GeoIP, config.Router.ProxyIPCode, config.Router.GeoSite, config.Router.ProxySiteCode); err != nil {
		//return nil, err
		log.DefaultLogger.Warn(err)
	}
	return r, nil
}
