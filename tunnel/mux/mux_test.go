package mux

import (
	"context"
	"testing"

	"github.com/p4gefau1t/trojan-go/tunnel/transport"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/test/util"
)

func TestMux(t *testing.T) {
	muxCfg := &Config{
		Mux: MuxConfig{
			Enabled:     true,
			Concurrency: 8,
			IdleTimeout: 60,
		},
	}
	ctx := config.WithConfig(context.Background(), Name, muxCfg)

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

	common.Must(err)

	muxTunnel := Tunnel{}
	muxClient, _ := muxTunnel.NewClient(ctx, tcpClient)
	muxServer, _ := muxTunnel.NewServer(ctx, tcpServer)

	conn1, err := muxClient.DialConn(nil, nil)
	common.Must2(conn1.Write(util.GeneratePayload(1024)))
	common.Must(err)
	buf := [1024]byte{}
	conn2, err := muxServer.AcceptConn(nil)
	common.Must(err)
	common.Must2(conn2.Read(buf[:]))
	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}
	conn1.Close()
	conn2.Close()
	muxClient.Close()
	muxServer.Close()
}
