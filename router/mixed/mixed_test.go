package mixed

import (
	"io/ioutil"
	"log"
	"net"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/router"
)

func TestSimpleMixedRouter(t *testing.T) {
	bypass := []byte("0.0.0.0/8\n10.0.0.0/8\n192.0.0.0/24\nbaidu.com\nqq.com\n")

	r, err := NewMixedRouter(
		&conf.RouterConfig{
			BypassList:    bypass,
			DefaultPolicy: "proxy",
		},
	)
	common.Must(err)
	p, err := r.RouteRequest(&protocol.Request{
		Address: &common.Address{
			AddressType: common.IPv4,
			IP:          net.ParseIP("10.1.1.1"),
		},
	})
	common.Must(err)
	if p != router.Bypass {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		Address: &common.Address{
			AddressType: common.IPv4,
			IP:          net.ParseIP("1.1.1.1"),
		},
	})
	common.Must(err)
	if p != router.Proxy {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		Address: &common.Address{
			AddressType: common.DomainName,
			DomainName:  "www.baidu.com",
		},
	})
	common.Must(err)
	if p != router.Bypass {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		Address: &common.Address{
			AddressType: common.DomainName,
			DomainName:  "im.qq.com",
		},
	})
	common.Must(err)
	if p != router.Bypass {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		Address: &common.Address{
			AddressType: common.DomainName,
			DomainName:  "www.google.com",
		},
	})
	common.Must(err)
	if p != router.Proxy {
		t.Fatal("wrong result")
	}
}

func TestMixedRouter(t *testing.T) {
	bypass := ""
	buf, err := ioutil.ReadFile("../data/cn-domain.txt")
	common.Must(err)
	bypass += string(buf)
	buf, err = ioutil.ReadFile("../data/cn-ip.txt")
	common.Must(err)
	bypass += string(buf)

	r, err := NewMixedRouter(
		&conf.RouterConfig{
			BypassList:    []byte(bypass),
			DefaultPolicy: "proxy",
		},
	)

	policy, err := r.RouteRequest(&protocol.Request{
		Address: &common.Address{
			AddressType: common.DomainName,
			DomainName:  "baidu.com",
		},
	})
	if policy != router.Bypass {
		log.Fatal("wrong result")
	}

	policy, err = r.RouteRequest(&protocol.Request{
		Address: &common.Address{
			AddressType: common.DomainName,
			DomainName:  "api.github.com",
		},
	})
	if policy != router.Proxy {
		log.Fatal("wrong result")
	}
}
