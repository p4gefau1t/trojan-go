package tproxy

import (
	"context"
	"fmt"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"os"
	"testing"
)

func TestTProxy(t *testing.T) {
	if os.Getuid() != 0 {
		t.Skip()
	}
	port := common.PickPort("tcp", "127.0.0.1")
	cfg := &Config{
		LocalHost:  "127.0.0.1",
		LocalPort:  port,
		UDPTimeout: 0,
	}
	ctx := config.WithConfig(context.Background(), Name, cfg)
	s, err := NewServer(ctx, nil)
	common.Must(err)
	go func() {
		conn, err := s.AcceptConn(nil)
		common.Must(err)
		fmt.Println(conn.Metadata())
	}()
}
