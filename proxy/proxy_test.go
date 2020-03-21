package proxy

import (
	"crypto/x509"
	"io/ioutil"
	"net"
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
	go TestServerWithDatabase(t)
	TestClient(t)
}

func TestClientToServerWithJSON(t *testing.T) {
	go TestServerWithJSON(t)
	TestClientWithJSON(t)
}

func TestMuxClientToServer(t *testing.T) {
	go TestMuxClient(t)
	TestServer(t)
}

func TestClientToPortReusingServer(t *testing.T) {
	go TestClient(t)
	TestPortReusingServer(t)
}

func TestSNIConfig(t *testing.T) {
	go ClientWithWrongSNI(t)
	TestServer(t)
}

func ClientWithWrongSNI(t *testing.T) {
	serverCertBytes, err := ioutil.ReadFile("./server.crt")
	common.Must(err)
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(serverCertBytes)
	ip := net.IPv4(127, 0, 0, 1)
	port := 4444
	password := "pass123123"
	config := &conf.GlobalConfig{
		LocalAddr: &net.TCPAddr{
			IP:   ip,
			Port: port,
		},
		LocalIP:   ip,
		LocalPort: uint16(port),
		RemoteAddr: &net.TCPAddr{
			IP:   ip,
			Port: 4445,
		},
		Hash: map[string]string{common.SHA224String(password): password},
	}
	config.TLS.Verify = true
	config.TLS.CertPool = pool
	config.TLS.SNI = "localhost123"
	config.TLS.VerifyHostname = true

	c := Client{
		config: config,
	}
	c.Run()
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
