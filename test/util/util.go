package util

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"net"
	"sync"

	"github.com/p4gefau1t/trojan-go/common"
)

// CheckConn checks if two netConn were connected and work properly
func CheckConn(a net.Conn, b net.Conn) bool {
	payload1 := make([]byte, 1024)
	payload2 := make([]byte, 1024)

	result1 := make([]byte, 1024)
	result2 := make([]byte, 1024)

	rand.Reader.Read(payload1)
	rand.Reader.Read(payload2)

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		a.Write(payload1)
		a.Read(result2)
		wg.Done()
	}()

	go func() {
		b.Read(result1)
		b.Write(payload2)
		wg.Done()
	}()

	wg.Wait()

	return bytes.Equal(payload1, result1) && bytes.Equal(payload2, result2)
}

// CheckPacketOverConn checks if two PacketConn streaming over a connection work properly
func CheckPacketOverConn(a, b net.PacketConn) bool {
	port := common.PickPort("tcp", "127.0.0.1")
	addr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port,
	}

	payload1 := make([]byte, 1024)
	payload2 := make([]byte, 1024)

	result1 := make([]byte, 1024)
	result2 := make([]byte, 1024)

	rand.Reader.Read(payload1)
	rand.Reader.Read(payload2)

	common.Must2(a.WriteTo(payload1, addr))
	_, addr1, err := b.ReadFrom(result1)
	common.Must(err)
	if addr1.String() != addr.String() {
		return false
	}

	common.Must2(a.WriteTo(payload2, addr))
	_, addr2, err := b.ReadFrom(result2)
	common.Must(err)
	if addr2.String() != addr.String() {
		return false
	}

	return bytes.Equal(payload1, result1) && bytes.Equal(payload2, result2)
}

func CheckPacket(a, b net.PacketConn) bool {
	payload1 := make([]byte, 1024)
	payload2 := make([]byte, 1024)

	result1 := make([]byte, 1024)
	result2 := make([]byte, 1024)

	rand.Reader.Read(payload1)
	rand.Reader.Read(payload2)

	_, err := a.WriteTo(payload1, b.LocalAddr())
	common.Must(err)
	_, _, err = b.ReadFrom(result1)
	common.Must(err)

	_, err = b.WriteTo(payload2, a.LocalAddr())
	common.Must(err)
	_, _, err = a.ReadFrom(result2)
	common.Must(err)

	return bytes.Equal(payload1, result1) && bytes.Equal(payload2, result2)
}

func GetTestAddr() string {
	port := common.PickPort("tcp", "127.0.0.1")
	return fmt.Sprintf("127.0.0.1:%d", port)
}
