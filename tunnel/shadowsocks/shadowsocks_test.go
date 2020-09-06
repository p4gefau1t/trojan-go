package shadowsocks

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel/freedom"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
)

func TestShadowsocks(t *testing.T) {
	p, err := strconv.ParseInt(util.HTTPPort, 10, 32)
	common.Must(err)

	port := common.PickPort("tcp", "127.0.0.1")
	transportConfig := &transport.Config{
		LocalHost:  "127.0.0.1",
		LocalPort:  port,
		RemoteHost: "127.0.0.1",
		RemotePort: port,
	}
	ctx := config.WithConfig(context.Background(), transport.Name, transportConfig)
	ctx = config.WithConfig(ctx, freedom.Name, &freedom.Config{})
	tcpClient, err := transport.NewClient(ctx, nil)
	common.Must(err)
	tcpServer, err := transport.NewServer(ctx, nil)
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
	ctx = config.WithConfig(ctx, Name, cfg)

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
		conn1.Write(util.GeneratePayload(1024))
		wg.Done()
	}()
	go func() {
		var err error
		conn2, err = s.AcceptConn(nil)
		common.Must(err)
		buf := [1024]byte{}
		conn2.Read(buf[:])
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
	conn1.Close()
	conn3.Close()
	c.Close()
	s.Close()
}
