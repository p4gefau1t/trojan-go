package transport

import (
	"context"
	"net"
	"sync"
	"testing"

	"github.com/p4gefau1t/trojan-go/config"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/test/util"
)

func TestTransport(t *testing.T) {
	serverCfg := &Config{
		LocalHost:  "127.0.0.1",
		LocalPort:  common.PickPort("tcp", "127.0.0.1"),
		RemoteHost: "127.0.0.1",
		RemotePort: common.PickPort("tcp", "127.0.0.1"),
	}
	clientCfg := &Config{
		LocalHost:  "127.0.0.1",
		LocalPort:  common.PickPort("tcp", "127.0.0.1"),
		RemoteHost: "127.0.0.1",
		RemotePort: serverCfg.LocalPort,
	}
	sctx := config.WithConfig(context.Background(), Name, serverCfg)
	cctx := config.WithConfig(context.Background(), Name, clientCfg)

	s, err := NewServer(sctx, nil)
	common.Must(err)
	c, err := NewClient(cctx, nil)
	common.Must(err)

	wg := sync.WaitGroup{}
	wg.Add(1)
	var conn1, conn2 net.Conn
	go func() {
		conn2, err = s.AcceptConn(nil)
		common.Must(err)
		wg.Done()
	}()
	conn1, err = c.DialConn(nil, nil)
	common.Must(err)

	common.Must2(conn1.Write([]byte("12345678\r\n")))
	wg.Wait()
	buf := [10]byte{}
	conn2.Read(buf[:])
	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}
	s.Close()
	c.Close()
}

func TestClientPlugin(t *testing.T) {
	clientCfg := &Config{
		LocalHost:  "127.0.0.1",
		LocalPort:  common.PickPort("tcp", "127.0.0.1"),
		RemoteHost: "127.0.0.1",
		RemotePort: 12345,
		TransportPlugin: TransportPluginConfig{
			Enabled:      true,
			Type:         "shadowsocks",
			Command:      "echo $SS_REMOTE_PORT",
			PluginOption: "",
			Arg:          nil,
			Env:          nil,
		},
	}
	ctx := config.WithConfig(context.Background(), Name, clientCfg)
	c, err := NewClient(ctx, nil)
	common.Must(err)
	c.Close()
}

func TestServerPlugin(t *testing.T) {
	cfg := &Config{
		LocalHost:  "127.0.0.1",
		LocalPort:  common.PickPort("tcp", "127.0.0.1"),
		RemoteHost: "127.0.0.1",
		RemotePort: 12345,
		TransportPlugin: TransportPluginConfig{
			Enabled:      true,
			Type:         "shadowsocks",
			Command:      "echo $SS_REMOTE_PORT",
			PluginOption: "",
			Arg:          nil,
			Env:          nil,
		},
	}
	ctx := config.WithConfig(context.Background(), Name, cfg)
	s, err := NewServer(ctx, nil)
	common.Must(err)
	s.Close()
}
