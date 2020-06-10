package websocket

import (
	"context"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/raw"
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

	tcpClient := &raw.FixedClient{
		FixedAddr: tunnel.NewAddressFromHostPort("tcp", "127.0.0.1", common.PickPort("tcp", "127.0.0.1")),
	}
	tcpServer, err := raw.NewServer(tcpClient.FixedAddr)
	common.Must(err)
	c, err := NewClient(ctx, tcpClient)
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
