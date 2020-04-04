package router

import (
	"net"
	"regexp"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"v2ray.com/core/app/router"
)

type GeoRouter struct {
	Router
	domains             []*router.Domain
	cidrs               []*router.CIDR
	matchPolicy         Policy
	nonMatchPolicy      Policy
	routeByIP           bool
	routeByIPOnNonmatch bool
}

func (r *GeoRouter) isSubdomain(fulldomain, domain string) bool {
	if strings.HasSuffix(fulldomain, domain) {
		idx := strings.Index(fulldomain, domain)
		if idx == 0 || fulldomain[idx-1] == '.' {
			return true
		}
	}
	return false
}

func (r *GeoRouter) matchDomain(fulldomain string) bool {
	for _, d := range r.domains {
		switch d.GetType() {
		case router.Domain_Domain, router.Domain_Full:
			if r.isSubdomain(fulldomain, d.GetValue()) {
				return true
			}
		case router.Domain_Plain:
			//keyword
			if strings.Contains(fulldomain, d.GetValue()) {
				return true
			}
		case router.Domain_Regex:
			//expregexp.Compile(site.GetValue())
			matched, err := regexp.Match(d.GetValue(), []byte(fulldomain))
			if err != nil {
				log.DefaultLogger.Error("invalid regex")
			}
			return matched
		default:
		}
	}
	return false
}

func (r *GeoRouter) matchIP(ip net.IP) bool {
	for _, c := range r.cidrs {
		n := int(c.GetPrefix())
		len := net.IPv6len
		if ip.To4() != nil {
			len = net.IPv4len
		}
		mask := net.CIDRMask(n, 8*len)
		subnet := &net.IPNet{IP: ip.Mask(mask), Mask: mask}
		if subnet.Contains(ip) {
			return true
		}
	}
	return false
}

func (r *GeoRouter) routeRequestByIP(domain string) (Policy, error) {
	addr, err := net.ResolveIPAddr("ip", domain)
	if err != nil {
		return Unknown, err
	}
	atype := protocol.IPv6
	if addr.IP.To4() != nil {
		atype = protocol.IPv4
	}
	return r.RouteRequest(&protocol.Request{
		IP:          addr.IP,
		AddressType: atype,
	})
}

func (r *GeoRouter) RouteRequest(req *protocol.Request) (Policy, error) {
	if r.domains == nil || r.cidrs == nil {
		return r.nonMatchPolicy, nil
	}
	switch req.AddressType {
	case protocol.DomainName:
		domain := string(req.DomainName)
		if r.routeByIP {
			return r.routeRequestByIP(domain)
		}
		if r.matchDomain(domain) {
			return r.matchPolicy, nil
		}
		if r.routeByIPOnNonmatch {
			return r.routeRequestByIP(domain)
		}
		return r.nonMatchPolicy, nil
	case protocol.IPv4, protocol.IPv6:
		if r.matchIP(req.IP) {
			return r.matchPolicy, nil
		}
		return r.nonMatchPolicy, nil
	default:
		return Unknown, common.NewError("invalid address type")
	}
}

func (r *GeoRouter) LoadGeoData(geoipData []byte, ipCode []string, geositeData []byte, siteCode []string) error {
	geoip := new(router.GeoIPList)
	if err := proto.Unmarshal(geoipData, geoip); err != nil {
		return err
	}
	for _, e := range geoip.GetEntry() {
		code := e.GetCountryCode()
		found := false
		for _, c := range ipCode {
			if c == code {
				r.cidrs = append(r.cidrs, e.GetCidr()...)
				found = true
				break
			}
		}
		if !found {
			log.DefaultLogger.Warn("ip code", code, "not found")
		}
	}

	geosite := new(router.GeoSiteList)
	if err := proto.Unmarshal(geositeData, geosite); err != nil {
		return err
	}
	for _, s := range geosite.GetEntry() {
		code := s.GetCountryCode()
		found := false
		for _, c := range siteCode {
			if c == code {
				domainList := s.GetDomain()
				r.domains = append(r.domains, domainList...)
				found = true
				break
			}
		}
		if !found {
			log.DefaultLogger.Warn("site code", code, "not found")
		}
	}
	log.DefaultLogger.Info("geoip and geosite loaded")
	return nil
}

func NewProtoRouter(matchPolicy Policy, nonMatchPolicy Policy, routeByIP bool, routeByIPOnNonmatch bool) (*GeoRouter, error) {
	r := GeoRouter{
		matchPolicy:         matchPolicy,
		nonMatchPolicy:      nonMatchPolicy,
		routeByIP:           routeByIP,
		routeByIPOnNonmatch: routeByIPOnNonmatch,
	}
	return &r, nil
}
