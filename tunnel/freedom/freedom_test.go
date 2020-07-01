package freedom

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/txthinking/socks5"
)

func TestConn(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &Client{
		ctx:    ctx,
		cancel: cancel,
	}
	addr, err := tunnel.NewAddressFromAddr("tcp", util.EchoAddr)
	common.Must(err)
	conn1, err := client.DialConn(addr, nil)
	common.Must(err)

	sendBuf := util.GeneratePayload(1024)
	recvBuf := [1024]byte{}

	common.Must2(conn1.Write(sendBuf))
	common.Must2(conn1.Read(recvBuf[:]))

	if !bytes.Equal(sendBuf, recvBuf[:]) {
		t.Fail()
	}
	client.Close()
}

func TestPacket(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &Client{
		ctx:    ctx,
		cancel: cancel,
	}
	addr, err := tunnel.NewAddressFromAddr("udp", util.EchoAddr)
	common.Must(err)
	conn1, err := client.DialPacket(nil)
	common.Must(err)

	sendBuf := util.GeneratePayload(1024)
	recvBuf := [1024]byte{}

	common.Must2(conn1.WriteTo(sendBuf, addr))
	_, _, err = conn1.ReadFrom(recvBuf[:])
	common.Must(err)

	if !bytes.Equal(sendBuf, recvBuf[:]) {
		t.Fail()
	}
}

func TestSocks(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	socksAddr := tunnel.NewAddressFromHostPort("udp", "127.0.0.1", common.PickPort("udp", "127.0.0.1"))
	client := &Client{
		ctx:          ctx,
		cancel:       cancel,
		proxyAddr:    socksAddr,
		forwardProxy: true,
		noDelay:      true,
	}
	target, err := tunnel.NewAddressFromAddr("tcp", util.EchoAddr)

	s, _ := socks5.NewClassicServer(socksAddr.String(), "127.0.0.1", "", "", 0, 0, 0, 0)
	s.Handle = &socks5.DefaultHandle{}
	go s.RunTCPServer()
	go s.RunUDPServer()

	time.Sleep(time.Second * 2)
	conn, err := client.DialConn(target, nil)
	common.Must(err)
	payload := util.GeneratePayload(1024)
	common.Must2(conn.Write(payload))

	recvBuf := [1024]byte{}
	conn.Read(recvBuf[:])
	if !bytes.Equal(recvBuf[:], payload) {
		t.Fail()
	}
	conn.Close()

	packet, err := client.DialPacket(nil)
	common.Must(err)
	common.Must2(packet.WriteWithMetadata(payload, &tunnel.Metadata{
		Address: target,
	}))

	recvBuf = [1024]byte{}
	n, m, err := packet.ReadWithMetadata(recvBuf[:])
	common.Must(err)

	if n != 1024 || !bytes.Equal(recvBuf[:], payload) {
		t.Fail()
	}

	fmt.Println(m)
	packet.Close()
	client.Close()
}
