package router

import (
	"net"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
)

func TestMixed(t *testing.T) {
	bypass := []byte("0.0.0.0/8\n10.0.0.0/8\n192.0.0.0/24\nbaidu.com\nqq.com\n")

	r, err := NewMixedRouter(Proxy, false, false, []byte{}, bypass, []byte{})
	common.Must(err)
	p, err := r.RouteRequest(&protocol.Request{
		AddressType: protocol.IPv4,
		IP:          net.ParseIP("10.1.1.1"),
	})
	common.Must(err)
	if p != Bypass {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		AddressType: protocol.IPv4,
		IP:          net.ParseIP("1.1.1.1"),
	})
	common.Must(err)
	if p != Proxy {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		AddressType: protocol.DomainName,
		DomainName:  []byte("www.baidu.com"),
	})
	common.Must(err)
	if p != Bypass {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		AddressType: protocol.DomainName,
		DomainName:  []byte("im.qq.com"),
	})
	common.Must(err)
	if p != Bypass {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		AddressType: protocol.DomainName,
		DomainName:  []byte("www.google.com"),
	})
	common.Must(err)
	if p != Proxy {
		t.Fatal("wrong result")
	}
}
