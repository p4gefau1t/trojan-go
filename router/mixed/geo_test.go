package mixed

import (
	"io/ioutil"
	"net"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/router"
)

func TestGeoRouter(t *testing.T) {
	r, err := NewGeoRouter(router.Bypass, router.Proxy, router.IPIfNonMatch)
	common.Must(err)
	geoipData, err := ioutil.ReadFile("geoip.dat")
	common.Must(err)
	geositeData, err := ioutil.ReadFile("geosite.dat")
	common.Must(err)
	common.Must(r.LoadGeoData(geoipData, []string{"CN"}, geositeData, []string{"CN"}))

	p, err := r.RouteRequest(&protocol.Request{
		Address: &common.Address{
			DomainName:  "mail.google.com",
			AddressType: common.DomainName,
		},
	})
	common.Must(err)
	if p != router.Proxy {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		Address: &common.Address{
			DomainName:  "tupian.baidu.com",
			AddressType: common.DomainName,
		},
	})
	common.Must(err)
	if p != router.Bypass {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		Address: &common.Address{
			IP:          net.ParseIP("8.8.8.8"),
			AddressType: common.IPv4,
		},
	})
	common.Must(err)
	if p != router.Proxy {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		Address: &common.Address{
			IP:          net.ParseIP("114.114.114.114"),
			AddressType: common.IPv4,
		},
	})
	common.Must(err)
	if p != router.Bypass {
		t.Fatal("wrong result")
	}
}
