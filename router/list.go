package router

import (
	"bytes"
	"net"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type ListRouter struct {
	Router
	domainList          []string
	ipList              []*net.IPNet
	matchPolicy         Policy
	nonMatchPolicy      Policy
	routeByIP           bool
	routeByIPOnNonmatch bool
}

func (r *ListRouter) isSubdomain(fulldomain, domain string) bool {
	if strings.HasSuffix(fulldomain, domain) {
		idx := strings.Index(fulldomain, domain)
		if idx == 0 || fulldomain[idx-1] == '.' {
			return true
		}
	}
	return false
}

func (r *ListRouter) RouteRequest(req *protocol.Request) (Policy, error) {
	switch req.AddressType {
	case common.DomainName:
		domain := string(req.DomainName)
		if ip := net.ParseIP(domain); ip != nil {
			for _, net := range r.ipList {
				if net.Contains(ip) {
					return r.matchPolicy, nil
				}
			}
			return r.nonMatchPolicy, nil
		}
		if r.routeByIP {
			addr, err := net.ResolveIPAddr("ip", domain)
			if err != nil {
				return Unknown, err
			}
			atype := common.IPv6
			if addr.IP.To4() != nil {
				atype = common.IPv4
			}
			return r.RouteRequest(&protocol.Request{
				Address: &common.Address{
					IP:          addr.IP,
					AddressType: atype,
				},
			})
		}
		for _, d := range r.domainList {
			if r.isSubdomain(domain, d) {
				return r.matchPolicy, nil
			}
		}
		if r.routeByIPOnNonmatch {
			addr, err := net.ResolveIPAddr("ip", domain)
			if err != nil {
				return Unknown, err
			}
			atype := common.IPv6
			if addr.IP.To4() != nil {
				atype = common.IPv4
			}
			return r.RouteRequest(&protocol.Request{
				Address: &common.Address{
					IP:          addr.IP,
					AddressType: atype,
				},
			})
		}
		return r.nonMatchPolicy, nil
	case common.IPv4, common.IPv6:
		ip := req.IP
		for _, ipNet := range r.ipList {
			if ipNet.Contains(ip) {
				return r.matchPolicy, nil
			}
		}
		return r.nonMatchPolicy, nil
	default:
		return Unknown, common.NewError("invalid address type")
	}
}

func (r *ListRouter) LoadList(data []byte) error {
	buf := bytes.NewBuffer(data)
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil {
			break
		}
		if line[0] == '\n' || line[0] == '\r' {
			continue
		}
		record := string(line)
		record = strings.Replace(string(record), "\r\n", "", -1)
		record = strings.Replace(string(record), "\n", "", -1)
		_, ipNet, err := net.ParseCIDR(record)
		if err != nil {
			r.domainList = append(r.domainList, record)
			continue
		}
		r.ipList = append(r.ipList, ipNet)
	}
	return nil
}

func NewListRouter(matchPolicy Policy, nonMatchPolicy Policy, routeByIP bool, routeByIPOnNonmatch bool, list []byte) (*ListRouter, error) {
	r := ListRouter{
		matchPolicy:         matchPolicy,
		nonMatchPolicy:      nonMatchPolicy,
		routeByIP:           routeByIP,
		routeByIPOnNonmatch: routeByIPOnNonmatch,
	}
	if err := r.LoadList(list); err != nil {
		return nil, err
	}
	return &r, nil
}
