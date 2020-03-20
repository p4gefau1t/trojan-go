package proxy

import (
	"io/ioutil"
	"sync"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/test"
	"golang.org/x/net/proxy"
)

func TestClientToServer(t *testing.T) {
	go TestServer(t)
	TestClient(t)
}

func TestClientToDatabaseServer(t *testing.T) {
	go TestServerWithDatabse(t)
	TestClient(t)
}

func TestClientToServerWithJSON(t *testing.T) {
	go TestServerWithJSON(t)
	go TestClientWithJSON(t)
	time.Sleep(time.Hour)
}

func TestMuxClientToServer(t *testing.T) {
	go TestMuxClient(t)
	go TestServer(t)
	time.Sleep(time.Hour)
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

	tcpServer := test.RunBlackHoleTCPServer()

	mbytes := 512
	payload := test.GeneratePayload(1024 * 1024 * mbytes)
	dialer, err := proxy.SOCKS5("tcp", clientConfig.LocalAddr.String(), nil, nil)
	common.Must(err)
	conn, err := dialer.Dial("tcp", tcpServer.String())
	common.Must(err)
	b.StartTimer()
	t1 := time.Now()
	conn.Write(payload)
	t2 := time.Now()
	speed := float64(mbytes) / t2.Sub(t1).Seconds()
	logger.Info("Speed: ", speed, "MBytes/s")
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

	tcpServer := test.RunBlackHoleTCPServer()

	connNum := 1024
	mbytes := 1
	payload := test.GeneratePayload(1024 * 1024 * mbytes)
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
	logger.Info("Speed: ", speed, "MBytes/s")
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

	tcpServer := test.RunBlackHoleTCPServer()

	connNum := 1024
	mbytes := 32
	payload := test.GeneratePayload(1024 * 1024 * mbytes)
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

func BenchmarkMuxClientToServerHugePayload(b *testing.B) {
	b.StopTimer()
	data, err := ioutil.ReadFile("client.json")
	common.Must(err)
	clientConfig, err := conf.ParseJSON(data)
	common.Must(err)
	clientConfig.TCP.Mux = true

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

	tcpServer := test.RunBlackHoleTCPServer()

	mbytes := 128
	payload := test.GeneratePayload(1024 * 1024 * mbytes)
	dialer, err := proxy.SOCKS5("tcp", clientConfig.LocalAddr.String(), nil, nil)
	common.Must(err)
	conn, err := dialer.Dial("tcp", tcpServer.String())
	common.Must(err)
	b.StartTimer()
	t1 := time.Now()
	conn.Write(payload)
	t2 := time.Now()
	speed := float64(mbytes) / t2.Sub(t1).Seconds()
	logger.Info("Speed: ", speed, "MBytes/s")
	b.StopTimer()
}

func BenchmarkMuxClientToContinuesHugeConn(b *testing.B) {
	b.StopTimer()
	data, err := ioutil.ReadFile("client.json")
	common.Must(err)
	clientConfig, err := conf.ParseJSON(data)
	common.Must(err)
	clientConfig.TCP.Mux = true

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

	tcpServer := test.RunBlackHoleTCPServer()

	connNum := 256
	mbytes := 16
	payload := test.GeneratePayload(1024 * 1024 * mbytes)
	dialer, err := proxy.SOCKS5("tcp", clientConfig.LocalAddr.String(), nil, nil)
	common.Must(err)
	wg := sync.WaitGroup{}
	wg.Add(connNum)

	sender := func() {
		conn, err := dialer.Dial("tcp", tcpServer.String())
		common.Must(err)
		conn.Write(payload)
		conn.Close()
		wg.Done()
	}
	b.StartTimer()
	for i := 0; i < connNum; i++ {
		go sender()
	}
	wg.Wait()
	b.StopTimer()
}
