package test

import (
	"context"
	"crypto/rand"
	"io"
	"io/ioutil"
	"net"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
)

func RunEchoUDPServer(ctx context.Context) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 5000,
	})
	common.Must(err)
	defer conn.Close()
	go func() {
		for {
			buf := make([]byte, 2048)
			n, addr, err := conn.ReadFromUDP(buf[:])
			if err != nil {
				return
			}
			log.Info("echo from", addr)
			conn.WriteToUDP(buf[0:n], addr)
		}
	}()
	<-ctx.Done()
}

func RunEchoTCPServer(ctx context.Context) {
	listener, err := net.Listen("tcp", "127.0.0.1:5000")
	common.Must(err)
	defer listener.Close()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				for {
					conn.SetDeadline(time.Now().Add(time.Second))
					buf := make([]byte, 2048)
					n, err := conn.Read(buf)
					if err != nil {
						return
					}
					_, err = conn.Write(buf[0:n])
					if err != nil {
						return
					}
				}
			}(conn)
		}
	}()
	<-ctx.Done()
}

func RunBlackHoleTCPServer(ctx context.Context) {
	listener, err := net.Listen("tcp", "127.0.0.1:5000")
	common.Must(err)
	go func() {
		for {
			conn, _ := listener.Accept()
			go func(conn net.Conn) {
				io.Copy(ioutil.Discard, conn)
				conn.Close()
			}(conn)
		}
	}()
	<-ctx.Done()
}

func GeneratePayload(length int) []byte {
	buf := make([]byte, length)
	io.ReadFull(rand.Reader, buf)
	return buf
}
