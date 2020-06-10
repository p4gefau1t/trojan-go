package router

import (
	"context"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"strconv"
	"strings"
	"testing"
)

type MockClient struct{}

func (m *MockClient) DialConn(address *tunnel.Address, t tunnel.Tunnel) (tunnel.Conn, error) {
	return nil, common.NewError("mockproxy")
}

func (m *MockClient) DialPacket(t tunnel.Tunnel) (tunnel.PacketConn, error) {
	return nil, common.NewError("mockproxy")
}

func (m MockClient) Close() error {
	return nil
}

func TestRouter(t *testing.T) {
	data := `
router:
    enabled: true
    bypass: 
    - "regex:bypassreg(.*)"
    - "full:bypassfull"
    - "full:localhost"
    - "domain:bypass.com"
    block:
    - "regex:blockreg(.*)"
    - "full:blockfull"
    - "domain:block.com"
    proxy:
    - "regex:proxyreg(.*)"
    - "full:proxyfull"
    - "domain:proxy.com"
`
	ctx, err := config.WithYAMLConfig(context.Background(), []byte(data))
	common.Must(err)
	client, err := NewClient(ctx, &MockClient{})
	common.Must(err)
	_, err = client.DialConn(&tunnel.Address{
		AddressType: tunnel.DomainName,
		DomainName:  "proxy.com",
		Port:        80,
	}, nil)
	if err.Error() != "mockproxy" {
		t.Fail()
	}
	_, err = client.DialConn(&tunnel.Address{
		AddressType: tunnel.DomainName,
		DomainName:  "proxyreg123456",
		Port:        80,
	}, nil)
	if err.Error() != "mockproxy" {
		t.Fail()
	}
	_, err = client.DialConn(&tunnel.Address{
		AddressType: tunnel.DomainName,
		DomainName:  "proxyfull",
		Port:        80,
	}, nil)
	if err.Error() != "mockproxy" {
		t.Fail()
	}

	_, err = client.DialConn(&tunnel.Address{
		AddressType: tunnel.DomainName,
		DomainName:  "block.com",
		Port:        80,
	}, nil)
	if !strings.Contains(err.Error(), "block") {
		t.Fail()
	}
	port, err := strconv.Atoi(util.HTTPPort)
	common.Must(err)
	_, err = client.DialConn(&tunnel.Address{
		AddressType: tunnel.DomainName,
		DomainName:  "localhost",
		Port:        port,
	}, nil)
	if err != nil {
		t.Fail()
	}
}
