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
	allResolveToIP      bool
	nonMatchResolveToIP bool
}

func (r *ListRouter) RouteRequest(req *protocol.Request) (Policy, error) {
	switch req.AddressType {
	case protocol.DomainName:
		domain := string(req.DomainName)
		if r.allResolveToIP {
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
		if r.nonMatchResolveToIP {
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

func (r *ListRouter) LoadIPList(data []byte) error {
	buf := bytes.NewBuffer(data)
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil || line[0] == '\n' {
			break
		}
		_, ipNet, err := net.ParseCIDR(string(line[0 : len(line)-1]))
		if err != nil {
			return err
		}
		r.ipList = append(r.ipList, ipNet)
	}
	return nil
}

func (r *ListRouter) LoadDomainList(data []byte) error {
	buf := bytes.NewBuffer(data)
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil {
			break
		}
		r.domainList = append(r.domainList, string(line[0:len(line)-1]))
	}
	return nil
}

func NewListRouter(matchPolicy Policy, nonMatchPolicy Policy, allResolveToIP bool, nonMatchResolveToIP bool, ipList []byte, domainList []byte) (*ListRouter, error) {
	r := ListRouter{
		matchPolicy:         matchPolicy,
		nonMatchPolicy:      nonMatchPolicy,
		allResolveToIP:      allResolveToIP,
		nonMatchResolveToIP: nonMatchResolveToIP,
	}
	if err := r.LoadIPList(ipList); err != nil {
		return nil, err
	}
	if err := r.LoadDomainList(domainList); err != nil {
		return nil, err
	}
	return &r, nil
}
