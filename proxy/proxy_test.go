package proxy

import (
	"crypto/rand"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"golang.org/x/net/proxy"
)

func TestClientToServer(t *testing.T) {
	go TestServer(t)
	go TestClient(t)
	time.Sleep(time.Hour)
}

func TestClientToServerWithJSON(t *testing.T) {
	go TestServerWithJSON(t)
	go TestClientWithJSON(t)
	time.Sleep(time.Hour)
}

func BlackHoleTCPServer() net.Addr {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	common.Must(err)
	blackhole := func(conn net.Conn) {
		io.Copy(ioutil.Discard, conn)
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

func BenchmarkClientToServerHugePayload(b *testing.B) {
	b.StopTimer()
	data, err := ioutil.ReadFile("client.json")
	common.Must(err)
	clientConfig, err := conf.ParseJSON(data)
	common.Must(err)

	client := Client{
		config: clientConfig,
	}
	go client.Run()

	data, err = ioutil.ReadFile("server.json")
	common.Must(err)
	serverConfig, err := conf.ParseJSON(data)
	common.Must(err)

	server := Server{
		config: serverConfig,
	}
	go server.Run()

	tcpServer := BlackHoleTCPServer()

	mbytes := 512
	payload := GeneratePayload(1024 * 1024 * mbytes)
	dialer, err := proxy.SOCKS5("tcp", clientConfig.LocalAddr.String(), nil, nil)
	common.Must(err)
	conn, err := dialer.Dial("tcp", tcpServer.String())
	common.Must(err)
	b.StartTimer()
	t1 := time.Now()
	conn.Write(payload)
	t2 := time.Now()
	speed := float64(mbytes) / t2.Sub(t1).Seconds()
	log.Println("Speed: ", speed, "MBytes/s")
	b.StopTimer()
}

func BenchmarkClientToServerHugeConn(b *testing.B) {
	b.StopTimer()
	data, err := ioutil.ReadFile("client.json")
	common.Must(err)
	clientConfig, err := conf.ParseJSON(data)
	common.Must(err)

	client := Client{
		config: clientConfig,
	}
	go client.Run()

	data, err = ioutil.ReadFile("server.json")
	common.Must(err)
	serverConfig, err := conf.ParseJSON(data)
	common.Must(err)

	server := Server{
		config: serverConfig,
	}
	go server.Run()

	tcpServer := BlackHoleTCPServer()

	connNum := 1024
	mbytes := 1
	payload := GeneratePayload(1024 * 1024 * mbytes)
	dialer, err := proxy.SOCKS5("tcp", clientConfig.LocalAddr.String(), nil, nil)
	common.Must(err)

	wg := sync.WaitGroup{}
	sender := func(wg *sync.WaitGroup) {
		conn, err := dialer.Dial("tcp", tcpServer.String())
		common.Must(err)
		conn.Write(payload)
		conn.Close()
		wg.Done()
	}
	b.StartTimer()
	wg.Add(connNum)
	t1 := time.Now()
	for i := 0; i < connNum; i++ {
		go sender(&wg)
	}
	wg.Wait()
	t2 := time.Now()
	speed := float64(mbytes) * float64(connNum) / t2.Sub(t1).Seconds()
	log.Println("Speed: ", speed, "MBytes/s")
	b.StopTimer()
}

func BenchmarkClientToContinuesHugeConn(b *testing.B) {
	b.StopTimer()
	data, err := ioutil.ReadFile("client.json")
	common.Must(err)
	clientConfig, err := conf.ParseJSON(data)
	common.Must(err)

	client := Client{
		config: clientConfig,
	}
	go client.Run()

	data, err = ioutil.ReadFile("server.json")
	common.Must(err)
	serverConfig, err := conf.ParseJSON(data)
	common.Must(err)

	server := Server{
		config: serverConfig,
	}
	go server.Run()

	tcpServer := BlackHoleTCPServer()

	connNum := 1024
	mbytes := 32
	payload := GeneratePayload(1024 * 1024 * mbytes)
	dialer, err := proxy.SOCKS5("tcp", clientConfig.LocalAddr.String(), nil, nil)
	common.Must(err)

	sender := func() {
		conn, err := dialer.Dial("tcp", tcpServer.String())
		common.Must(err)
		conn.Write(payload)
		conn.Close()
	}
	b.StartTimer()
	for i := 0; i < 100; i++ {
		for i := 0; i < connNum; i++ {
			go sender()
		}
		time.Sleep(time.Second / 10)
	}
	b.StopTimer()
}
