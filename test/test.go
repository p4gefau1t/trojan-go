package test

import (
	"crypto/rand"
	"io"
	"io/ioutil"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
)

func RunEchoUDPServer(port int) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: port,
	})
	common.Must(err)
	for {
		buf := make([]byte, protocol.MaxUDPPacketSize)
		n, addr, err := conn.ReadFromUDP(buf[:])
		common.Must(err)
		log.DefaultLogger.Info("echo from", addr)
		conn.WriteToUDP(buf[0:n], addr)
	}
}

func RunBlackHoleTCPServer() net.Addr {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	common.Must(err)
	blackhole := func(conn net.Conn) {
		io.Copy(ioutil.Discard, conn)
		conn.Close()
	}
	serve := func() {
		for {
			conn, _ := listener.Accept()
			go blackhole(conn)
		}
	}
	go serve()
	return listener.Addr()
}

func GeneratePayload(length int) []byte {
	buf := make([]byte, length)
	io.ReadFull(rand.Reader, buf)
	return buf
}
