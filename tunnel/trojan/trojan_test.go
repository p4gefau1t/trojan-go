package trojan

import (
	"context"
	"fmt"
	"testing"

	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/raw"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/statistic"
)

type MockUser struct {
	statistic.User
}

func (*MockUser) AddIP(string) bool {
	return true
}
func (*MockUser) DelIP(string) bool {
	return true
}

func (*MockUser) Hash() string {
	return common.SHA224String("user")
}

func (*MockUser) AddTraffic(sent, recv int) {}

type MockAuth struct {
	statistic.Authenticator
}

func (*MockAuth) AuthUser(hash string) (bool, statistic.User) {
	if hash == common.SHA224String("user") {
		return true, &MockUser{}
	}
	return false, nil
}

func (*MockAuth) ListUsers() []statistic.User {
	return []statistic.User{&MockUser{}}
}

func TestTrojan(t *testing.T) {
	addr := tunnel.NewAddressFromHostPort("tcp", "127.0.0.1", common.PickPort("tcp", "127.0.0.1"))
	tcpClient := &raw.FixedClient{
		FixedAddr: addr,
	}
	tcpServer, err := raw.NewServer(addr)
	ctx := context.Background()
	s := &Server{
		underlay:   tcpServer,
		auth:       &MockAuth{},
		ctx:        ctx,
		redirAddr:  nil,
		connChan:   make(chan tunnel.Conn, 32),
		muxChan:    make(chan tunnel.Conn, 32),
		packetChan: make(chan tunnel.PacketConn, 32),
	}
	go s.acceptLoop()
	c := &Client{
		underlay: tcpClient,
		ctx:      ctx,
		user:     &MockUser{},
	}

	conn1, err := c.DialConn(&tunnel.Address{
		DomainName:  "example.com",
		AddressType: tunnel.DomainName,
	}, nil)
	common.Must2(conn1.Write([]byte("12345678")))
	conn2, err := s.AcceptConn(nil)
	buf := [8]byte{}
	conn2.Read(buf[:])
	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}

	packet1, err := c.DialPacket(nil)
	common.Must(err)
	packet1.WriteWithMetadata([]byte("12345678"), &tunnel.Metadata{
		Address: &tunnel.Address{
			DomainName:  "example.com",
			AddressType: tunnel.DomainName,
			Port:        80,
		},
	})
	packet2, err := s.AcceptPacket(nil)
	common.Must(err)

	_, m, err := packet2.ReadWithMetadata(buf[:])
	common.Must(err)

	fmt.Println(m)

	if !util.CheckPacketOverConn(packet1, packet2) {
		t.Fail()
	}
}
