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
	payload1 := [1024]byte{}
	payload2 := [1024]byte{}
	rand.Reader.Read(payload1[:])
	rand.Reader.Read(payload2[:])

	result1 := [1024]byte{}
	result2 := [1024]byte{}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		a.Write(payload1[:])
		a.Read(result2[:])
		wg.Done()
	}()
	go func() {
		b.Read(result1[:])
		b.Write(payload2[:])
		wg.Done()
	}()
	wg.Wait()
	if !bytes.Equal(payload1[:], result1[:]) || !bytes.Equal(payload2[:], result2[:]) {
		return false
	}
	return true
}

// CheckPacketOverConn checks if two PacketConn streaming over a connection work properly
func CheckPacketOverConn(a, b net.PacketConn) bool {
	port := common.PickPort("tcp", "127.0.0.1")
	addr := &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port,
	}
	payload1 := [1024]byte{}
	payload2 := [1024]byte{}
	rand.Reader.Read(payload1[:])
	rand.Reader.Read(payload2[:])

	result1 := [1024]byte{}
	result2 := [1024]byte{}

	common.Must2(a.WriteTo(payload1[:], addr))
	_, addr1, err := b.ReadFrom(result1[:])
	common.Must(err)
	if addr1.String() != addr.String() {
		return false
	}

	common.Must2(a.WriteTo(payload2[:], addr))
	_, addr2, err := b.ReadFrom(result2[:])
	common.Must(err)
	if addr2.String() != addr.String() {
		return false
	}
	if !bytes.Equal(payload1[:], result1[:]) || !bytes.Equal(payload2[:], result2[:]) {
		return false
	}
	return true
}

func CheckPacket(a, b net.PacketConn) bool {
	payload1 := [1024]byte{}
	payload2 := [1024]byte{}
	rand.Reader.Read(payload1[:])
	rand.Reader.Read(payload2[:])

	result1 := [1024]byte{}
	result2 := [1024]byte{}

	_, err := a.WriteTo(payload1[:], b.LocalAddr())
	common.Must(err)
	_, _, err = b.ReadFrom(result1[:])
	common.Must(err)

	_, err = b.WriteTo(payload2[:], a.LocalAddr())
	common.Must(err)
	_, _, err = a.ReadFrom(result2[:])
	common.Must(err)
	if !bytes.Equal(payload1[:], result1[:]) || !bytes.Equal(payload2[:], result2[:]) {
		return false
	}
	return true
}

func GetTestAddr() string {
	port := common.PickPort("tcp", "127.0.0.1")
	return fmt.Sprintf("127.0.0.1:%d", port)
}
