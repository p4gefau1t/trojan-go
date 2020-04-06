package test

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	_ "net/http/pprof"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	_ "github.com/p4gefau1t/trojan-go/log/golog"
	"github.com/p4gefau1t/trojan-go/proxy/client"
	"github.com/p4gefau1t/trojan-go/proxy/server"
	"golang.org/x/net/proxy"
)

var cert string = `
-----BEGIN CERTIFICATE-----
MIIDZTCCAk0CFFphZh018B5iAD9F5fV4y0AlD0LxMA0GCSqGSIb3DQEBCwUAMG8x
CzAJBgNVBAYTAlVTMQ0wCwYDVQQIDARNYXJzMRMwEQYDVQQHDAppVHJhbnN3YXJw
MRMwEQYDVQQKDAppVHJhbnN3YXJwMRMwEQYDVQQLDAppVHJhbnN3YXJwMRIwEAYD
VQQDDAlsb2NhbGhvc3QwHhcNMjAwMzMxMTAwMDUxWhcNMzAwMzI5MTAwMDUxWjBv
MQswCQYDVQQGEwJVUzENMAsGA1UECAwETWFyczETMBEGA1UEBwwKaVRyYW5zd2Fy
cDETMBEGA1UECgwKaVRyYW5zd2FycDETMBEGA1UECwwKaVRyYW5zd2FycDESMBAG
A1UEAwwJbG9jYWxob3N0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA
ml44fThYMkCcT627o7ibEs7mq2WOhImjDwYijYJ1684BatrCsHJNcw8PJGTuP+tg
GdngmALjA3l+RipjaE/UK4FJrAjruphA/hOCjZfWqk8KBR4qk0OltxCMWJlp/XCM
9ny1ogFdWUlBbqThs4NWSOUESgxf/Be2njeiOrngGR31qxSiLCLBvafIhKqq/4av
Rlx0Ht770uvF97MlAj1ASAvzTZICHAfUZxEdWl0J4MBbG7SNcnMBbyAF+s60eFTa
4RGMfRGnUa2Fzz/gfjhvfSIGeLQ3JRG6sl6jkc5xe0PZzhq3UNpK0gtQ48yy9CSP
neZnrynoKks7XC2bizsr3QIDAQABMA0GCSqGSIb3DQEBCwUAA4IBAQAHS/xuG5+F
yGU3N6V4kv+HbKqHaXNOq4zKVsCc1k7vg4MFFpKUJKxtJYooCI8n2ypp5XRUTIGQ
bmEbVcIPqm9Rf/4vHtF0falNCwieAbXDkiEHoykRmmU1UE/ccPA7X8NO9aVLJAJO
N2Li8MH0Ixgs02pQH56eyGKoRBWPR5C3ETQ9Leqvazg6Dn1iJWvmfF0mOte5228s
mZJOntF9t8MZOJdIWGdrUHn6euRfhd0btkmL/NUDzeCTwJcuPORLxkBbCP5mTC6G
GnLS5Z4oRYgCgvT2pLtcM0r48hYjwgjXFQ4zalkW6YI9LPpqwwMhhOzINlXjBaDi
Haz8uKI4EciU
-----END CERTIFICATE-----
`

var key string = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEAml44fThYMkCcT627o7ibEs7mq2WOhImjDwYijYJ1684BatrC
sHJNcw8PJGTuP+tgGdngmALjA3l+RipjaE/UK4FJrAjruphA/hOCjZfWqk8KBR4q
k0OltxCMWJlp/XCM9ny1ogFdWUlBbqThs4NWSOUESgxf/Be2njeiOrngGR31qxSi
LCLBvafIhKqq/4avRlx0Ht770uvF97MlAj1ASAvzTZICHAfUZxEdWl0J4MBbG7SN
cnMBbyAF+s60eFTa4RGMfRGnUa2Fzz/gfjhvfSIGeLQ3JRG6sl6jkc5xe0PZzhq3
UNpK0gtQ48yy9CSPneZnrynoKks7XC2bizsr3QIDAQABAoIBAFpYUo9W7qdakSFA
+NS1Mm0rkm01nteLBlfAq3BOrl030DSNm+xQuWthoOcX+yiFxVTb40qURfC+plzC
ajOepPphTJDXF7+5ZDBPktTzzLsYTzD3mstdiBtAICOqhhHCUX3hNxx91/htm1H6
Re4eK921y3DbFUIhTswCm3vrVXDc4yTXtURGllVzo40K/1Of39CpufKFdpJ81HV+
h/VW++h3o+sFV4KqcqIjClxBfDxoJpBaRlOCunTiHqZNvqO+EPqPR5zdn34werjU
xQEvPzmz+ClwnaEXQxYWgIcYQii9VNsHogDxEw4R31S7lVrUt0f0atDmGJip1lPb
E7IomAECgYEAzKQ3PzBV46nUNfVO9SODpf14Z+xYfLKouPC+Qnepwp0V0JS6zY1+
Wzskyb80drjnoQraWSEvGsX+tEWeLcnjN7JuMu/U8DPKRcQ+Q2dsVo/q4sfBOgvl
VhPNMZLfa7NIkRUx2KXku++Ep0Xtak0dskrfQrZnvhymRPyWuIMM6IECgYEAwRwL
Gt/ZZdUueE/hwT3c1hNn6igeDLOwK2t6frib+Ofw5oCAQxtTROvP1ljlnWUPkeIS
uzTusmqucalcK3lCHIsyHLwApOI/B31M971pxMVBRZ0wIbBaoarCGND7gi8JUPFR
VErGcAB5YnpRlmfLPEgw2o7DpjsDc2KmdE9oNV0CgYEAmfNEWLYtNztxGTK1treD
96ELLutf2lexlIgQKgLJ5E22tpbdPXwfvdRtpZTBjDsojj+S6hCL1lFzfv0MtZe2
5xTF0G4avKXJmti6moy4tRpJ81ehZuDCJBJ7gLrkd6qFghf2yuxqenQDUK/Lnvfq
ylGHSjHdM+lrsGRxotd8I4ECgYBoo4GA9nseqv2bQ+3YgGUBu1I7l7FwwI1decfO
ksoxfb0Tqd3WfyAH4J+mTlVdjD17lzz/JBeTpisQe+ztwa8JOIPW/ih7L/1nWYYz
V/fQH/LWfe5u0tjJcXXrbJJcYJBzw8+GFV6hoiAkNJOxJF0ENToDtAhgMuoTxAje
TYjyIQKBgQCmHkLLq0Bj3FpIOVrwo2gNvQteNPa7jkkGp4lljO8JQUHhCHDGWKEH
MUJ0EFsxS/EaQa+rW6jHhs3GyBA2TxmC783stAOOEX+hO/zpcbzdCWgp6eZ0aGMW
WS94/5WE/lwHJi8ZPSjH1AURCzXhUi4fGvBrNBtry95e+jcEvP5c0g==
-----END RSA PRIVATE KEY-----
`

func getKeyPair() []tls.Certificate {
	cert, err := tls.X509KeyPair([]byte(cert), []byte(key))
	common.Must(err)
	return []tls.Certificate{cert}
}

func getTLSConfig() conf.TLSConfig {
	KeyPair := getKeyPair()
	pool := x509.NewCertPool()
	if ok := pool.AppendCertsFromPEM([]byte(cert)); !ok {
		panic("invalid cert")
	}
	c := conf.TLSConfig{
		CertPool:       pool,
		KeyPair:        KeyPair,
		Verify:         true,
		VerifyHostname: true,
		SNI:            "localhost",
	}
	return c
}

func getLocalAddr(port int) net.Addr {
	return &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: port,
	}
}

func getLocalIP() net.IP {
	return net.IPv4(127, 0, 0, 1)
}

func getHash(password string) map[string]string {
	hash := common.SHA224String(password)
	m := make(map[string]string)
	m[hash] = password
	return m
}

func TestClientJSON(t *testing.T) {
	data, err := ioutil.ReadFile("client.json")
	common.Must(err)
	config, err := conf.ParseJSON(data)
	common.Must(err)
	c := client.Client{}
	c.Build(config)
	c.Run()
}

func TestServerJSON(t *testing.T) {
	data, err := ioutil.ReadFile("server.json")
	common.Must(err)
	config, err := conf.ParseJSON(data)
	common.Must(err)
	c := client.Client{}
	c.Build(config)
	c.Run()
}

func TestClient(t *testing.T) {
	config := &conf.GlobalConfig{
		LocalIP:    getLocalIP(),
		LocalPort:  4444,
		LocalAddr:  getLocalAddr(4444),
		RemoteIP:   getLocalIP(),
		RemotePort: 4445,
		RemoteAddr: getLocalAddr(4445),
		TLS:        getTLSConfig(),
		Hash:       getHash("pass123"),
	}
	c := client.Client{}
	c.Build(config)
	common.Must(c.Run())
}

func TestServer(t *testing.T) {
	config := &conf.GlobalConfig{
		LocalIP:    getLocalIP(),
		LocalPort:  4445,
		LocalAddr:  getLocalAddr(4445),
		RemoteIP:   getLocalIP(),
		RemotePort: 80,
		RemoteAddr: getLocalAddr(80),
		TLS:        getTLSConfig(),
		Hash:       getHash("pass123"),
	}
	s := server.Server{}
	s.Build(config)
	common.Must(s.Run())
}

func TestNAT(t *testing.T) {
	config := &conf.GlobalConfig{
		LocalIP:    getLocalIP(),
		LocalPort:  4445,
		LocalAddr:  getLocalAddr(4445),
		RemoteIP:   getLocalIP(),
		RemotePort: 80,
		RemoteAddr: getLocalAddr(80),
		TLS:        getTLSConfig(),
		Hash:       getHash("pass123"),
	}
	n := client.NAT{}
	n.Build(config)
	common.Must(n.Run())
}

func TestMuxClient(t *testing.T) {
	config := &conf.GlobalConfig{
		LocalIP:    getLocalIP(),
		LocalPort:  4444,
		LocalAddr:  getLocalAddr(4444),
		RemoteIP:   getLocalIP(),
		RemotePort: 4445,
		RemoteAddr: getLocalAddr(4445),
		TLS:        getTLSConfig(),
		Hash:       getHash("pass123"),
		Mux: conf.MuxConfig{
			Enabled:     true,
			Concurrency: 8,
			IdleTimeout: 30,
		},
	}
	client := client.Client{}
	client.Build(config)
	client.Run()
}

func TestRouterClient(t *testing.T) {
	config := &conf.GlobalConfig{
		LocalIP:    getLocalIP(),
		LocalPort:  4444,
		LocalAddr:  getLocalAddr(4444),
		RemoteIP:   getLocalIP(),
		RemotePort: 4445,
		RemoteAddr: getLocalAddr(4445),
		TLS:        getTLSConfig(),
		Hash:       getHash("pass123"),
		Router: conf.RouterConfig{
			Enabled:       true,
			BypassList:    []byte("baidu.com\nqq.com\n\n192.168.0.0/16\n"),
			DefaultPolicy: "proxy",
		},
	}
	c := client.Client{}
	c.Build(config)
	common.Must(c.Run())
}

func TestClientAndServer(t *testing.T) {
	go func() {
		err := http.ListenAndServe("0.0.0.0:8000", nil)
		log.Error(err)
	}()
	go TestClient(t)
	TestServer(t)
}

func TestMuxClientAndServer(t *testing.T) {
	go func() {
		err := http.ListenAndServe("0.0.0.0:8000", nil)
		log.Error(err)
	}()
	go TestMuxClient(t)
	TestServer(t)
}

func TestRouterClientAndServer(t *testing.T) {
	go func() {
		err := http.ListenAndServe("0.0.0.0:8000", nil)
		log.Error(err)
	}()
	go TestRouterClient(t)
	TestServer(t)
}

func TestClientServerJSON(t *testing.T) {
	go TestServerJSON(t)
	TestClientJSON(t)
}

func BenchmarkNormalClientToServer(b *testing.B) {
	config1 := &conf.GlobalConfig{
		LogLevel:   5,
		LocalIP:    getLocalIP(),
		LocalPort:  4444,
		LocalAddr:  getLocalAddr(4444),
		RemoteIP:   getLocalIP(),
		RemotePort: 4445,
		RemoteAddr: getLocalAddr(4445),
		TLS:        getTLSConfig(),
		Hash:       getHash("pass123"),
	}
	c := client.Client{}
	c.Build(config1)
	go c.Run()

	config2 := &conf.GlobalConfig{
		LocalIP:    getLocalIP(),
		LocalPort:  4445,
		LocalAddr:  getLocalAddr(4445),
		RemoteIP:   getLocalIP(),
		RemotePort: 80,
		RemoteAddr: getLocalAddr(80),
		TLS:        getTLSConfig(),
		Hash:       getHash("pass123"),
	}
	s := server.Server{}
	s.Build(config2)
	go s.Run()

	target := RunBlackHoleTCPServer()
	dialer, err := proxy.SOCKS5("tcp", getLocalAddr(4444).String(), nil, nil)
	common.Must(err)
	conn, err := dialer.Dial("tcp", target.String())
	common.Must(err)
	mbytes := 512
	payload := GeneratePayload(1024 * 1024 * mbytes)
	t1 := time.Now()
	conn.Write(payload)
	t2 := time.Now()
	speed := float64(mbytes) / t2.Sub(t1).Seconds()
	b.Log("Speed: ", speed, "MB/s")
	conn.Close()
}

func BenchmarkMuxClientToServer(b *testing.B) {
	config1 := &conf.GlobalConfig{
		LogLevel:   5,
		LocalIP:    getLocalIP(),
		LocalPort:  4444,
		LocalAddr:  getLocalAddr(4444),
		RemoteIP:   getLocalIP(),
		RemotePort: 4445,
		RemoteAddr: getLocalAddr(4445),
		TLS:        getTLSConfig(),
		Hash:       getHash("pass123"),
		Mux: conf.MuxConfig{
			Enabled:     true,
			Concurrency: 8,
			IdleTimeout: 30,
		},
	}
	c := client.Client{}
	c.Build(config1)
	go c.Run()

	config2 := &conf.GlobalConfig{
		LocalIP:    getLocalIP(),
		LocalPort:  4445,
		LocalAddr:  getLocalAddr(4445),
		RemoteIP:   getLocalIP(),
		RemotePort: 80,
		RemoteAddr: getLocalAddr(80),
		TLS:        getTLSConfig(),
		Hash:       getHash("pass123"),
	}
	s := server.Server{}
	s.Build(config2)
	go s.Run()

	target := RunBlackHoleTCPServer()
	dialer, err := proxy.SOCKS5("tcp", getLocalAddr(4444).String(), nil, nil)
	common.Must(err)
	conn, err := dialer.Dial("tcp", target.String())
	common.Must(err)
	mbytes := 512
	payload := GeneratePayload(1024 * 1024 * mbytes)
	t1 := time.Now()
	conn.Write(payload)
	t2 := time.Now()
	speed := float64(mbytes) / t2.Sub(t1).Seconds()
	b.Log("Speed: ", speed, "MB/s")
	conn.Close()
}

func BenchmarkNormalClientToServerHighConcurrency(b *testing.B) {
	config1 := &conf.GlobalConfig{
		LogLevel:   5,
		LocalIP:    getLocalIP(),
		LocalPort:  4444,
		LocalAddr:  getLocalAddr(4444),
		RemoteIP:   getLocalIP(),
		RemotePort: 4445,
		RemoteAddr: getLocalAddr(4445),
		TLS:        getTLSConfig(),
		Hash:       getHash("pass123"),
	}
	c := client.Client{}
	c.Build(config1)
	go c.Run()

	config2 := &conf.GlobalConfig{
		LocalIP:    getLocalIP(),
		LocalPort:  4445,
		LocalAddr:  getLocalAddr(4445),
		RemoteIP:   getLocalIP(),
		RemotePort: 80,
		RemoteAddr: getLocalAddr(80),
		TLS:        getTLSConfig(),
		Hash:       getHash("pass123"),
	}
	s := server.Server{}
	s.Build(config2)
	go s.Run()

	target := RunBlackHoleTCPServer()
	dialer, err := proxy.SOCKS5("tcp", getLocalAddr(4444).String(), nil, nil)
	common.Must(err)

	connNum := 128
	mbytes := 128
	payload := GeneratePayload(1024 * 1024 * mbytes)

	wg := sync.WaitGroup{}
	sender := func(wg *sync.WaitGroup) {
		conn, err := dialer.Dial("tcp", target.String())
		common.Must(err)
		conn.Write(payload)
		conn.Close()
		wg.Done()
	}

	wg.Add(connNum)

	t1 := time.Now()
	for i := 0; i < connNum; i++ {
		go sender(&wg)
	}
	wg.Wait()
	t2 := time.Now()
	speed := float64(mbytes) * float64(connNum) / t2.Sub(t1).Seconds()
	b.Log("Speed: ", speed, "MB/s")
}

func BenchmarkMuxClientToServerHighConcurrency(b *testing.B) {
	config1 := &conf.GlobalConfig{
		LogLevel:   5,
		LocalIP:    getLocalIP(),
		LocalPort:  4444,
		LocalAddr:  getLocalAddr(4444),
		RemoteIP:   getLocalIP(),
		RemotePort: 4445,
		RemoteAddr: getLocalAddr(4445),
		TLS:        getTLSConfig(),
		Hash:       getHash("pass123"),
		Mux: conf.MuxConfig{
			Enabled:     true,
			Concurrency: 8,
			IdleTimeout: 30,
		},
	}
	c := client.Client{}
	c.Build(config1)
	go c.Run()

	config2 := &conf.GlobalConfig{
		LocalIP:    getLocalIP(),
		LocalPort:  4445,
		LocalAddr:  getLocalAddr(4445),
		RemoteIP:   getLocalIP(),
		RemotePort: 80,
		RemoteAddr: getLocalAddr(80),
		TLS:        getTLSConfig(),
		Hash:       getHash("pass123"),
	}
	s := server.Server{}
	s.Build(config2)
	go s.Run()

	target := RunBlackHoleTCPServer()
	dialer, err := proxy.SOCKS5("tcp", getLocalAddr(4444).String(), nil, nil)
	common.Must(err)

	connNum := 128
	mbytes := 128
	payload := GeneratePayload(1024 * 1024 * mbytes)

	wg := sync.WaitGroup{}
	sender := func(wg *sync.WaitGroup) {
		conn, err := dialer.Dial("tcp", target.String())
		common.Must(err)
		conn.Write(payload)
		conn.Close()
		wg.Done()
	}

	wg.Add(connNum)

	t1 := time.Now()
	for i := 0; i < connNum; i++ {
		go sender(&wg)
	}
	wg.Wait()
	t2 := time.Now()
	speed := float64(mbytes) * float64(connNum) / t2.Sub(t1).Seconds()
	b.Log("Speed: ", speed, "MB/s")
}
