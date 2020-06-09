package raw

import (
	"testing"

	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"

	"github.com/p4gefau1t/trojan-go/common"
)

func TestConn(t *testing.T) {
	port := common.PickPort("tcp", "127.0.0.1")
	addr := tunnel.NewAddressFromHostPort("tcp", "127.0.0.1", port)
	server, err := NewServer(addr)
	common.Must(err)
	client := &Client{}

	conn1, err := client.DialConn(addr, nil)
	common.Must(err)
	conn2, err := server.AcceptConn(nil)
	common.Must(err)

	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}
}

func TestPacket(t *testing.T) {
	port := common.PickPort("tcp", "127.0.0.1")
	addr := tunnel.NewAddressFromHostPort("tcp", "127.0.0.1", port)
	server, err := NewServer(addr)
	common.Must(err)
	client := &Client{}

	conn1, err := client.DialPacket(nil)
	common.Must(err)
	conn2, err := server.AcceptPacket(nil)
	common.Must(err)

	buf := [100]byte{}

	str1 := "1234567890zxvbouehrofisjad;fowqiooewuroqpe"
	common.Must2(conn1.WriteTo([]byte(str1), addr))
	n, _, err := conn2.ReadFrom(buf[:])
	common.Must(err)
	if string(buf[:n]) != str1 {
		t.Fail()
	}
}
