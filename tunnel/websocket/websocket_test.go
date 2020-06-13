package websocket

import (
	"context"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"sync"
	"testing"
)

func TestWebsocket(t *testing.T) {
	cfg := &Config{
		Websocket: WebsocketConfig{
			Enabled:  true,
			Hostname: "localhost",
			Path:     "/ws",
		},
	}

	ctx := config.WithConfig(context.Background(), Name, cfg)

	port := common.PickPort("tcp", "127.0.0.1")
	transportConfig := &transport.Config{
		LocalHost:  "127.0.0.1",
		LocalPort:  port,
		RemoteHost: "127.0.0.1",
		RemotePort: port,
	}
	ctx = config.WithConfig(ctx, transport.Name, transportConfig)
	tcpClient, err := transport.NewClient(ctx, nil)
	common.Must(err)
	tcpServer, err := transport.NewServer(ctx, nil)
	common.Must(err)

	c, err := NewClient(ctx, tcpClient)
	common.Must(err)
	s, err := NewServer(ctx, tcpServer)
	var conn2 tunnel.Conn
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		conn2, err = s.AcceptConn(nil)
		common.Must(err)
		wg.Done()
	}()
	conn1, err := c.DialConn(nil, nil)
	common.Must(err)
	wg.Wait()
	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}
}
