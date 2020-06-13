package simplesocks

import (
	"context"
	"fmt"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

func TestSimpleSocks(t *testing.T) {
	port := common.PickPort("tcp", "127.0.0.1")
	transportConfig := &transport.Config{
		LocalHost:  "127.0.0.1",
		LocalPort:  port,
		RemoteHost: "127.0.0.1",
		RemotePort: port,
	}
	ctx := config.WithConfig(context.Background(), transport.Name, transportConfig)
	tcpClient, err := transport.NewClient(ctx, nil)
	common.Must(err)
	tcpServer, err := transport.NewServer(ctx, nil)
	common.Must(err)

	c, err := NewClient(ctx, tcpClient)
	common.Must(err)
	s, err := NewServer(ctx, tcpServer)
	common.Must(err)

	conn1, err := c.DialConn(&tunnel.Address{
		DomainName:  "www.baidu.com",
		AddressType: tunnel.DomainName,
		Port:        443,
	}, nil)
	common.Must(err)
	defer conn1.Close()
	conn1.Write(util.GeneratePayload(1024))
	conn2, err := s.AcceptConn(nil)
	common.Must(err)
	defer conn2.Close()
	buf := [1024]byte{}
	common.Must2(conn2.Read(buf[:]))
	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}

	packet1, err := c.DialPacket(nil)
	packet1.WriteWithMetadata([]byte("12345678"), &tunnel.Metadata{
		Address: &tunnel.Address{
			DomainName:  "test.com",
			AddressType: tunnel.DomainName,
			Port:        443,
		},
	})
	defer packet1.Close()
	packet2, err := s.AcceptPacket(nil)
	defer packet2.Close()
	_, m, err := packet2.ReadWithMetadata(buf[:])
	common.Must(err)
	fmt.Println(m)

	if !util.CheckPacketOverConn(packet1, packet2) {
		t.Fail()
	}
	s.Close()
	c.Close()
}
