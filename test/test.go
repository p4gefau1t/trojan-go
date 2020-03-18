package test

import (
	"net"
	"os"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/withmandala/go-log"
)

var logger = log.New(os.Stdout)

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
		logger.Info("echo from", addr)
		conn.WriteToUDP(buf[0:n], addr)
	}
}
