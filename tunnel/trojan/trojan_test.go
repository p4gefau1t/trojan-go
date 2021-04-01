package trojan

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/statistic/memory"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/freedom"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
)

func TestTrojan(t *testing.T) {
	port := common.PickPort("tcp", "127.0.0.1")
	transportConfig := &transport.Config{
		LocalHost:  "127.0.0.1",
		LocalPort:  port,
		RemoteHost: "127.0.0.1",
		RemotePort: port,
	}
	ctx, cancel := context.WithCancel(context.Background())
	ctx = config.WithConfig(ctx, transport.Name, transportConfig)
	ctx = config.WithConfig(ctx, freedom.Name, &freedom.Config{})
	tcpClient, err := transport.NewClient(ctx, nil)
	common.Must(err)
	tcpServer, err := transport.NewServer(ctx, nil)
	common.Must(err)

	serverPort := common.PickPort("tcp", "127.0.0.1")
	authConfig := &memory.Config{Passwords: []string{"password"}}
	clientConfig := &Config{
		RemoteHost: "127.0.0.1",
		RemotePort: serverPort,
	}
	serverConfig := &Config{
		LocalHost:  "127.0.0.1",
		LocalPort:  serverPort,
		RemoteHost: "127.0.0.1",
		RemotePort: util.EchoPort,
	}

	ctx = config.WithConfig(ctx, memory.Name, authConfig)
	clientCtx := config.WithConfig(ctx, Name, clientConfig)
	serverCtx := config.WithConfig(ctx, Name, serverConfig)
	c, err := NewClient(clientCtx, tcpClient)
	common.Must(err)
	s, err := NewServer(serverCtx, tcpServer)

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
	common.Must2(io.ReadFull(conn, recvBuf[:]))
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
	cancel()
}
