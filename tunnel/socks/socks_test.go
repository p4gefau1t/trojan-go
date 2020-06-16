package socks

import (
	"context"
	"fmt"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"golang.org/x/net/proxy"
	"net"
	"sync"
	"testing"
	"time"
)

func TestSocks(t *testing.T) {
	port := common.PickPort("tcp", "127.0.0.1")
	ctx := config.WithConfig(context.Background(), transport.Name, &transport.Config{
		LocalHost: "127.0.0.1",
		LocalPort: port,
	})
	ctx = config.WithConfig(ctx, Name, &Config{
		UDPTimeout: 30,
	})
	tcpServer, err := transport.NewServer(ctx, nil)
	common.Must(err)
	addr := tunnel.NewAddressFromHostPort("tcp", "127.0.0.1", port)
	s, err := NewServer(ctx, tcpServer)
	common.Must(err)
	socksClient, err := proxy.SOCKS5("tcp", addr.String(), nil, proxy.Direct)
	common.Must(err)
	var conn1, conn2 net.Conn
	wg := sync.WaitGroup{}
	wg.Add(2)

	time.Sleep(time.Second * 2)
	go func() {
		conn2, err = s.AcceptConn(nil)
		common.Must(err)
		wg.Done()
	}()

	time.Sleep(time.Second * 1)
	go func() {
		conn1, err = socksClient.Dial("tcp", util.EchoAddr)
		common.Must(err)
		wg.Done()
	}()

	wg.Wait()
	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}
	fmt.Println(conn2.(tunnel.Conn).Metadata())
}
