package simplesocks

import (
	"context"
	"fmt"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/raw"
)

func TestSimpleSocks(t *testing.T) {
	addr := tunnel.NewAddressFromHostPort("tcp", "127.0.0.1", common.PickPort("tcp", "127.0.0.1"))
	rawClient := &raw.FixedClient{
		FixedAddr: addr,
	}
	rawServer, err := raw.NewServer(addr)
	common.Must(err)
	c, err := NewClient(context.Background(), rawClient)
	common.Must(err)
	s, err := NewServer(context.Background(), rawServer)
	common.Must(err)

	conn1, err := c.DialConn(&tunnel.Address{
		DomainName:  "www.baidu.com",
		AddressType: tunnel.DomainName,
		Port:        443,
	}, nil)
	common.Must(err)
	conn1.Write([]byte("12345678"))
	conn2, err := s.AcceptConn(nil)
	common.Must(err)
	buf := [8]byte{}
	common.Must2(conn2.Read(buf[:]))
	if string(buf[:]) != "12345678" {
		t.Fail()
	}
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
	packet2, err := s.AcceptPacket(nil)
	_, m, err := packet2.ReadWithMetadata(buf[:])
	common.Must(err)
	fmt.Println(m)

	if !util.CheckPacketOverConn(packet1, packet2) {
		t.Fail()
	}
}
