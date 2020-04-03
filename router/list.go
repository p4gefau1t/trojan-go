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

func (r *ListRouter) RouteRequest(req *protocol.Request) (Policy, error) {
	switch req.AddressType {
	case protocol.DomainName:
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
			addr, err := net.ResolveIPAddr("tcp", domain)
			if err != nil {
				return Unknown, err
			}
			atype := protocol.IPv4
			if addr.IP.To16() != nil {
				atype = protocol.IPv6
			}
			return r.RouteRequest(&protocol.Request{
				IP:          addr.IP,
				AddressType: atype,
			})
		}
		for _, suffix := range r.domainList {
			if strings.HasSuffix(domain, suffix) {
				return r.matchPolicy, nil
			}
		}
		if r.routeByIPOnNonmatch {
			addr, err := net.ResolveIPAddr("tcp", domain)
			if err != nil {
				return Unknown, err
			}
			atype := protocol.IPv4
			if addr.IP.To16() != nil {
				atype = protocol.IPv6
			}
			return r.RouteRequest(&protocol.Request{
				IP:          addr.IP,
				AddressType: atype,
			})
		}
		return r.nonMatchPolicy, nil
	case protocol.IPv4, protocol.IPv6:
		ip := req.IP
		for _, net := range r.ipList {
			if net.Contains(ip) {
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
