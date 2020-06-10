package shadowsocks

import (
	"context"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/raw"
	"net"
	"testing"
)

func TestShadowsocks(t *testing.T) {
	cfg := &Config{
		Shadowsocks: ShadowsocksConfig{
			Enabled:  true,
			Method:   "AES-128-GCM",
			Password: "password",
		},
	}
	ctx := config.WithConfig(context.Background(), Name, cfg)
	port := common.PickPort("tcp", "127.0.0.1")
	addr := &tunnel.Address{
		AddressType: tunnel.IPv4,
		IP:          net.ParseIP("127.0.0.1"),
		Port:        port,
	}
	tcpServer, err := raw.NewServer(addr)
	common.Must(err)
	tcpClient := &raw.FixedClient{
		FixedAddr: addr,
	}
	c, err := NewClient(ctx, tcpClient)
	common.Must(err)
	s, err := NewServer(ctx, tcpServer)
	common.Must(err)
	conn1, err := c.DialConn(nil, nil)
	common.Must(err)
	conn2, err := s.AcceptConn(nil)
	common.Must(err)
	util.CheckConn(conn1, conn2)
}
