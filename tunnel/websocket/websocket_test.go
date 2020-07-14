package websocket

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"golang.org/x/net/websocket"
)

func TestWebsocket(t *testing.T) {
	cfg := &Config{
		Websocket: WebsocketConfig{
			Enabled: true,
			Host:    "localhost",
			Path:    "/ws",
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
	time.Sleep(time.Second)
	conn1, err := c.DialConn(nil, nil)
	common.Must(err)
	wg.Wait()
	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}

	if strings.HasPrefix(conn1.RemoteAddr().String(), "ws") {
		t.Fail()
	}
	if strings.HasPrefix(conn2.RemoteAddr().String(), "ws") {
		t.Fail()
	}

	conn1.Close()
	conn2.Close()
	s.Close()
	c.Close()
}

func TestRedirect(t *testing.T) {
	cfg := &Config{
		RemoteHost: "127.0.0.1",
		Websocket: WebsocketConfig{
			Enabled: true,
			Host:    "localhost",
			Path:    "/ws",
		},
	}
	fmt.Sscanf(util.HTTPPort, "%d", &cfg.RemotePort)
	ctx := config.WithConfig(context.Background(), Name, cfg)

	port := common.PickPort("tcp", "127.0.0.1")
	transportConfig := &transport.Config{
		LocalHost: "127.0.0.1",
		LocalPort: port,
	}
	ctx = config.WithConfig(ctx, transport.Name, transportConfig)
	tcpServer, err := transport.NewServer(ctx, nil)
	common.Must(err)

	s, err := NewServer(ctx, tcpServer)
	common.Must(err)

	go func() {
		_, err := s.AcceptConn(nil)
		if err == nil {
			t.Fail()
		}
	}()
	time.Sleep(time.Second)
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	common.Must(err)
	url := "wss://localhost/wrong-path"
	origin := "https://localhost"
	wsConfig, err := websocket.NewConfig(url, origin)
	common.Must(err)
	_, err = websocket.NewClient(wsConfig, conn)
	if err == nil {
		t.Fail()
	}
	conn.Close()

	s.Close()
}
