package mux

import (
	"context"
	"fmt"
	"testing"

	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"

	"github.com/p4gefau1t/trojan-go/tunnel/raw"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"gopkg.in/yaml.v2"
)

func TestMux(t *testing.T) {
	cfg := newDefaultConfig().(*Config)
	data, err := yaml.Marshal(cfg)
	common.Must(err)
	fmt.Println(string(data))

	ctx, err := config.WithYAMLConfig(context.Background(), data)
	common.Must(err)

	port := common.PickPort("tcp", "127.0.0.1")
	addr := tunnel.NewAddressFromHostPort("tcp", "127.0.0.1", port)

	tcpClient := &raw.FixedClient{
		FixedAddr: addr,
	}
	tcpServer, err := raw.NewServer(addr)
	common.Must(err)

	muxTunnel := Tunnel{}
	muxClient, _ := muxTunnel.NewClient(ctx, tcpClient)
	muxServer, _ := muxTunnel.NewServer(ctx, tcpServer)

	first := []byte("12345678")
	conn1, err := muxClient.DialConn(addr, nil)
	common.Must2(conn1.Write([]byte(first)))
	common.Must(err)
	buf := [8]byte{}
	conn2, err := muxServer.AcceptConn(nil)
	common.Must(err)
	common.Must2(conn2.Read(buf[:]))
	fmt.Println(string(buf[:]))
	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}
}
