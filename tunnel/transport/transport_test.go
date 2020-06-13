package transport

import (
	"context"
	"github.com/p4gefau1t/trojan-go/config"
	"net"
	"sync"
	"testing"

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
