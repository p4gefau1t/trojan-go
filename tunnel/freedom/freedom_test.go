package freedom

import (
	"bytes"
	"context"
	"testing"

	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"

	"github.com/p4gefau1t/trojan-go/common"
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
