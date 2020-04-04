package router

import (
	"io/ioutil"
	"net"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
)

func TestGeoRouter(t *testing.T) {
	r, err := NewGeoRouter(Bypass, Proxy, false, false)
	common.Must(err)
	geoipData, err := ioutil.ReadFile("geoip.dat")
	common.Must(err)
	geositeData, err := ioutil.ReadFile("geosite.dat")
	common.Must(err)
	common.Must(r.LoadGeoData(geoipData, []string{"CN"}, geositeData, []string{"CN"}))

	p, err := r.RouteRequest(&protocol.Request{
		DomainName:  []byte("mail.google.com"),
		AddressType: protocol.DomainName,
	})
	common.Must(err)
	if p != Proxy {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		DomainName:  []byte("tupian.baidu.com"),
		AddressType: protocol.DomainName,
	})
	common.Must(err)
	if p != Bypass {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		IP:          net.ParseIP("8.8.8.8"),
		AddressType: protocol.IPv4,
	})
	common.Must(err)
	if p != Proxy {
		t.Fatal("wrong result")
	}

	p, err = r.RouteRequest(&protocol.Request{
		IP:          net.ParseIP("114.114.114.114"),
		AddressType: protocol.IPv4,
	})
	common.Must(err)
	if p != Bypass {
		t.Fatal("wrong result")
	}
}
