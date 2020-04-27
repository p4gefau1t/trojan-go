package mixed

import (
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/router"
)

type MixedRouter struct {
	proxyList     *ListRouter
	bypassList    *ListRouter
	blockList     *ListRouter
	proxyGeo      *GeoRouter
	bypassGeo     *GeoRouter
	blockGeo      *GeoRouter
	defaultPolicy router.Policy
}

func (r *MixedRouter) match(rr router.Router, req *protocol.Request) bool {
	policy, err := rr.RouteRequest(req)
	if err != nil {
		log.Warn(common.NewError("match error").Base(err))
		return false
	}
	if policy == router.Match {
		return true
	}
	return false
}

func (r *MixedRouter) RouteRequest(req *protocol.Request) (router.Policy, error) {

	if r.match(r.blockGeo, req) {
		return router.Block, nil
	}
	if r.match(r.blockList, req) {
		return router.Block, nil
	}

	if r.match(r.bypassGeo, req) {
		return router.Bypass, nil
	}
	if r.match(r.bypassList, req) {
		return router.Bypass, nil
	}

	if r.match(r.proxyGeo, req) {
		return router.Proxy, nil
	}
	if r.match(r.proxyList, req) {
		return router.Proxy, nil
	}

	return r.defaultPolicy, nil
}

func NewMixedRouter(config *conf.RouterConfig) (router.Router, error) {
	var defaultPolicy router.Policy

	switch config.DefaultPolicy {
	case "proxy":
		defaultPolicy = router.Proxy
	case "bypass":
		defaultPolicy = router.Bypass
	case "block":
		defaultPolicy = router.Block
	}

	routeByIP := config.RouteByIP
	routeByIPOnNonmatch := config.RouteByIPOnNonmatch

	block := config.BlockList
	bypass := config.BypassList
	proxy := config.ProxyList

	r := &MixedRouter{
		defaultPolicy: defaultPolicy,
	}

	var err error
	if r.blockList, err = NewListRouter(router.Match, router.NonMatch, routeByIP, routeByIPOnNonmatch, block); err != nil {
		return nil, err
	}
	if r.bypassList, err = NewListRouter(router.Match, router.NonMatch, routeByIP, routeByIPOnNonmatch, bypass); err != nil {
		return nil, err
	}
	if r.proxyList, err = NewListRouter(router.Match, router.NonMatch, routeByIP, routeByIPOnNonmatch, proxy); err != nil {
		return nil, err
	}

	r.blockGeo, _ = NewGeoRouter(router.Match, router.NonMatch, routeByIP, false)
	r.bypassGeo, _ = NewGeoRouter(router.Match, router.NonMatch, routeByIP, routeByIPOnNonmatch)
	r.proxyGeo, _ = NewGeoRouter(router.Match, router.NonMatch, routeByIP, routeByIPOnNonmatch)

	if err := r.blockGeo.LoadGeoData(config.GeoIP, config.BlockIPCode, config.GeoSite, config.BlockSiteCode); err != nil {
		log.Warn(err)
	}
	if err := r.bypassGeo.LoadGeoData(config.GeoIP, config.BypassIPCode, config.GeoSite, config.BypassSiteCode); err != nil {
		log.Warn(err)
	}
	if err := r.proxyGeo.LoadGeoData(config.GeoIP, config.ProxyIPCode, config.GeoSite, config.ProxySiteCode); err != nil {
		log.Warn(err)
	}
	return r, nil
}

func init() {
	router.NewRouter = NewMixedRouter
}
