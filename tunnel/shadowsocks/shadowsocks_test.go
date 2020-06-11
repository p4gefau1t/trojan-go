package shadowsocks

import (
	"context"
	"fmt"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/raw"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestShadowsocks(t *testing.T) {
	p, err := strconv.ParseInt(util.HTTPPort, 10, 32)
	common.Must(err)
	cfg := &Config{
		RemoteHost: "127.0.0.1",
		RemotePort: int(p),
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

	wg := sync.WaitGroup{}
	wg.Add(2)
	var conn1, conn2 net.Conn
	go func() {
		var err error
		conn1, err = c.DialConn(nil, nil)
		common.Must(err)
		conn1.Write([]byte("12345678"))
		wg.Done()
	}()
	go func() {
		var err error
		conn2, err = s.AcceptConn(nil)
		buf := [8]byte{}
		conn2.Read(buf[:])
		common.Must(err)
		wg.Done()
	}()
	wg.Wait()
	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}

	go func() {
		var err error
		conn2, err = s.AcceptConn(nil)
		if err == nil {
			t.Fail()
		}
	}()
	// test redirection
	conn3, err := tcpClient.DialConn(nil, nil)
	common.Must(err)
	n, err := conn3.Write(util.GeneratePayload(1024))
	common.Must(err)
	fmt.Println("write:", n)
	buf := [1024]byte{}
	n, err = conn3.Read(buf[:])
	common.Must(err)
	fmt.Println("read:", n)
	if !strings.Contains(string(buf[:n]), "Bad Request") {
		t.Fail()
	}
}
