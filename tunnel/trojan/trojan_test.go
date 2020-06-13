package trojan

import (
	"bytes"
	"context"
	"fmt"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	_ "github.com/p4gefau1t/trojan-go/log/golog"
	"github.com/p4gefau1t/trojan-go/redirector"
	"github.com/p4gefau1t/trojan-go/statistic"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"net"
	"testing"
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
	ctx, cancel := context.WithCancel(ctx)
	s := &Server{
		underlay:   tcpServer,
		auth:       &MockAuth{},
		redirAddr:  tunnel.NewAddressFromHostPort("tcp", "127.0.0.1", util.EchoPort),
		connChan:   make(chan tunnel.Conn, 32),
		muxChan:    make(chan tunnel.Conn, 32),
		packetChan: make(chan tunnel.PacketConn, 32),
		ctx:        ctx,
		cancel:     cancel,
		redir:      redirector.NewRedirector(ctx),
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
	common.Must2(conn1.Write([]byte("87654321")))
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

	//redirecting
	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	common.Must(err)
	sendBuf := util.GeneratePayload(1024)
	recvBuf := [1024]byte{}
	common.Must2(conn.Write(sendBuf))
	common.Must2(conn.Read(recvBuf[:]))
	if !bytes.Equal(sendBuf, recvBuf[:]) {
		fmt.Println(sendBuf)
		fmt.Println(recvBuf[:])
		t.Fail()
	}
	conn1.Close()
	conn2.Close()
	packet1.Close()
	packet2.Close()
	conn.Close()
	c.Close()
	s.Close()
}
