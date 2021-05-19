package router

import (
	"context"
	"net"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

type MockClient struct{}

func (m *MockClient) DialConn(address *tunnel.Address, t tunnel.Tunnel) (tunnel.Conn, error) {
	return nil, common.NewError("mockproxy")
}

func (m *MockClient) DialPacket(t tunnel.Tunnel) (tunnel.PacketConn, error) {
	return MockPacketConn{}, nil
}

func (m MockClient) Close() error {
	return nil
}

type MockPacketConn struct{}

func (m MockPacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	panic("implement me")
}

func (m MockPacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	panic("implement me")
}

func (m MockPacketConn) Close() error {
	panic("implement me")
}

func (m MockPacketConn) LocalAddr() net.Addr {
	panic("implement me")
}

func (m MockPacketConn) SetDeadline(t time.Time) error {
	panic("implement me")
}

func (m MockPacketConn) SetReadDeadline(t time.Time) error {
	panic("implement me")
}

func (m MockPacketConn) SetWriteDeadline(t time.Time) error {
	panic("implement me")
}

func (m MockPacketConn) WriteWithMetadata(bytes []byte, metadata *tunnel.Metadata) (int, error) {
	return 0, common.NewError("mockproxy")
}

func (m MockPacketConn) ReadWithMetadata(bytes []byte) (int, *tunnel.Metadata, error) {
	return 0, nil, common.NewError("mockproxy")
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
    - "regexp:blockreg(.*)"
    - "full:blockfull"
    - "domain:block.com"
    proxy:
    - "regexp:proxyreg(.*)"
    - "full:proxyfull"
    - "domain:proxy.com"
    - "cidr:192.168.1.1/16"
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
		t.Fatal(err)
	}
	_, err = client.DialConn(&tunnel.Address{
		AddressType: tunnel.DomainName,
		DomainName:  "proxyreg123456",
		Port:        80,
	}, nil)
	if err.Error() != "mockproxy" {
		t.Fatal(err)
	}
	_, err = client.DialConn(&tunnel.Address{
		AddressType: tunnel.DomainName,
		DomainName:  "proxyfull",
		Port:        80,
	}, nil)
	if err.Error() != "mockproxy" {
		t.Fatal(err)
	}

	_, err = client.DialConn(&tunnel.Address{
		AddressType: tunnel.IPv4,
		IP:          net.ParseIP("192.168.123.123"),
		Port:        80,
	}, nil)
	if err.Error() != "mockproxy" {
		t.Fatal(err)
	}

	_, err = client.DialConn(&tunnel.Address{
		AddressType: tunnel.DomainName,
		DomainName:  "block.com",
		Port:        80,
	}, nil)
	if !strings.Contains(err.Error(), "block") {
		t.Fatal("block??")
	}
	port, err := strconv.Atoi(util.HTTPPort)
	common.Must(err)

	_, err = client.DialConn(&tunnel.Address{
		AddressType: tunnel.DomainName,
		DomainName:  "localhost",
		Port:        port,
	}, nil)
	if err != nil {
		t.Fatal("dial http failed", err)
	}

	packet, err := client.DialPacket(nil)
	common.Must(err)
	buf := [10]byte{}
	_, err = packet.WriteWithMetadata(buf[:], &tunnel.Metadata{
		Address: &tunnel.Address{
			AddressType: tunnel.DomainName,
			DomainName:  "proxyfull",
			Port:        port,
		},
	})
	if err.Error() != "mockproxy" {
		t.Fail()
	}
}
