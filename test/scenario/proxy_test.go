package scenario

import (
	"bytes"
	"fmt"
	"github.com/p4gefau1t/trojan-go/test/util"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	_ "github.com/p4gefau1t/trojan-go/log/golog"
	"github.com/p4gefau1t/trojan-go/proxy"
	_ "github.com/p4gefau1t/trojan-go/proxy/client"
	_ "github.com/p4gefau1t/trojan-go/proxy/forward"
	_ "github.com/p4gefau1t/trojan-go/proxy/nat"
	_ "github.com/p4gefau1t/trojan-go/proxy/server"
	_ "github.com/p4gefau1t/trojan-go/statistic/memory"
	netproxy "golang.org/x/net/proxy"
	_ "net/http/pprof"
)

// test key and cert

var cert = `
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

var key = `
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

func init() {
	ioutil.WriteFile("server.crt", []byte(cert), 0777)
	ioutil.WriteFile("server.key", []byte(key), 0777)
}

func CheckClientServer(clientData, serverData string, socksPort int) (ok bool) {
	server, err := proxy.NewProxyFromConfigData([]byte(clientData), false)
	common.Must(err)
	go server.Run()

	client, err := proxy.NewProxyFromConfigData([]byte(serverData), false)
	common.Must(err)
	go client.Run()

	time.Sleep(time.Second * 2)
	dialer, err := netproxy.SOCKS5("tcp", fmt.Sprintf("127.0.0.1:%d", socksPort), nil, netproxy.Direct)

	ok = true
	const num = 100
	wg := sync.WaitGroup{}
	wg.Add(num)
	for i := 0; i < num; i++ {
		go func() {
			const payloadSize = 1024
			payload := util.GeneratePayload(payloadSize)
			buf := [payloadSize]byte{}

			conn, err := dialer.Dial("tcp", util.EchoAddr)
			common.Must(err)

			common.Must2(conn.Write(payload))
			common.Must2(conn.Read(buf[:]))

			if !bytes.Equal(payload, buf[:]) {
				ok = false
			}
			conn.Close()
			wg.Done()
		}()
	}
	wg.Wait()
	client.Close()
	server.Close()
	return
}

func TestClientServerWebsocketSubTree(t *testing.T) {
	serverPort := common.PickPort("tcp", "127.0.0.1")
	socksPort := common.PickPort("tcp", "127.0.0.1")
	clientData := fmt.Sprintf(`
run-type: client
local-addr: 127.0.0.1
local-port: %d
remote-addr: 127.0.0.1
remote-port: %d
password:
    - password
ssl:
    verify: false
    fingerprint: firefox
    sni: localhost
websocket:
    enabled: true
    path: /ws
    hostname: 127.0.0.1
shadowsocks:
    enabled: true
    method: AEAD_CHACHA20_POLY1305
    password: 12345678
mux:
    enabled: true
`, socksPort, serverPort)
	serverData := fmt.Sprintf(`
run-type: server
local-addr: 127.0.0.1
local-port: %d
remote-addr: 127.0.0.1
remote-port: %s
disable-http-check: true
password:
    - password
ssl:
    verify-hostname: false
    key: server.key
    cert: server.crt
    sni: localhost
shadowsocks:
    enabled: true
    method: AEAD_CHACHA20_POLY1305
    password: 12345678
websocket:
    enabled: true
    path: /ws
    hostname: 127.0.0.1
`, serverPort, util.HTTPPort)

	if !CheckClientServer(clientData, serverData, socksPort) {
		t.Fail()
	}
}

func TestClientServerTrojanSubTree(t *testing.T) {
	serverPort := common.PickPort("tcp", "127.0.0.1")
	socksPort := common.PickPort("tcp", "127.0.0.1")
	clientData := fmt.Sprintf(`
run-type: client
local-addr: 127.0.0.1
local-port: %d
remote-addr: 127.0.0.1
remote-port: %d
password:
    - password
ssl:
    verify: false
    fingerprint: firefox
    sni: localhost
shadowsocks:
    enabled: true
    method: AEAD_CHACHA20_POLY1305
    password: 12345678
mux:
    enabled: true
`, socksPort, serverPort)
	serverData := fmt.Sprintf(`
run-type: server
local-addr: 127.0.0.1
local-port: %d
remote-addr: 127.0.0.1
remote-port: %s
disable-http-check: true
password:
    - password
ssl:
    verify-hostname: false
    key: server.key
    cert: server.crt
    sni: localhost
shadowsocks:
    enabled: true
    method: AEAD_CHACHA20_POLY1305
    password: 12345678
`, serverPort, util.HTTPPort)

	if !CheckClientServer(clientData, serverData, socksPort) {
		t.Fail()
	}
}

func TestWebsocketDetection(t *testing.T) {
	serverPort := common.PickPort("tcp", "127.0.0.1")
	socksPort := common.PickPort("tcp", "127.0.0.1")

	clientData := fmt.Sprintf(`
run-type: client
local-addr: 127.0.0.1
local-port: %d
remote-addr: 127.0.0.1
remote-port: %d
password:
    - password
ssl:
    verify: false
    fingerprint: firefox
    sni: localhost
shadowsocks:
    enabled: true
    method: AEAD_CHACHA20_POLY1305
    password: 12345678
mux:
    enabled: true
`, socksPort, serverPort)
	serverData := fmt.Sprintf(`
run-type: server
local-addr: 127.0.0.1
local-port: %d
remote-addr: 127.0.0.1
remote-port: %s
disable-http-check: true
password:
    - password
ssl:
    verify-hostname: false
    key: server.key
    cert: server.crt
    sni: localhost
shadowsocks:
    enabled: true
    method: AEAD_CHACHA20_POLY1305
    password: 12345678
websocket:
    enabled: true
    path: /ws
    hostname: 127.0.0.1
`, serverPort, util.HTTPPort)

	if !CheckClientServer(clientData, serverData, socksPort) {
		t.Fail()
	}
}

func TestPluginWebsocket(t *testing.T) {
	serverPort := common.PickPort("tcp", "127.0.0.1")
	socksPort := common.PickPort("tcp", "127.0.0.1")

	clientData := fmt.Sprintf(`
run-type: client
local-addr: 127.0.0.1
local-port: %d
remote-addr: 127.0.0.1
remote-port: %d
password:
    - password
transport-plugin:
    enabled: true
    type: plaintext
shadowsocks:
    enabled: true
    method: AEAD_CHACHA20_POLY1305
    password: 12345678
mux:
    enabled: true
websocket:
    enabled: true
    path: /ws
    hostname: 127.0.0.1
`, socksPort, serverPort)
	serverData := fmt.Sprintf(`
run-type: server
local-addr: 127.0.0.1
local-port: %d
remote-addr: 127.0.0.1
remote-port: %s
disable-http-check: true
password:
    - password
transport-plugin:
    enabled: true
    type: plaintext
shadowsocks:
    enabled: true
    method: AEAD_CHACHA20_POLY1305
    password: 12345678
websocket:
    enabled: true
    path: /ws
    hostname: 127.0.0.1
`, serverPort, util.HTTPPort)

	if !CheckClientServer(clientData, serverData, socksPort) {
		t.Fail()
	}
}

func TestForward(t *testing.T) {
	serverPort := common.PickPort("tcp", "127.0.0.1")
	clientPort := common.PickPort("tcp", "127.0.0.1")
	_, targetPort, _ := net.SplitHostPort(util.EchoAddr)
	clientData := fmt.Sprintf(`
run-type: forward
local-addr: 127.0.0.1
local-port: %d
remote-addr: 127.0.0.1
remote-port: %d
target-addr: 127.0.0.1
target-port: %s
password:
    - password
ssl:
    verify: false
    fingerprint: firefox
    sni: localhost
websocket:
    enabled: true
    path: /ws
    hostname: 127.0.0.1
shadowsocks:
    enabled: true
    method: AEAD_CHACHA20_POLY1305
    password: 12345678
mux:
    enabled: true
`, clientPort, serverPort, targetPort)
	go func() {
		proxy, err := proxy.NewProxyFromConfigData([]byte(clientData), false)
		common.Must(err)
		common.Must(proxy.Run())
	}()

	serverData := fmt.Sprintf(`
run-type: server
local-addr: 127.0.0.1
local-port: %d
remote-addr: 127.0.0.1
remote-port: %s
disable-http-check: true
password:
    - password
ssl:
    verify-hostname: false
    key: server.key
    cert: server.crt
    sni: "localhost"
websocket:
    enabled: true
    path: /ws
    hostname: 127.0.0.1
shadowsocks:
    enabled: true
    method: AEAD_CHACHA20_POLY1305
    password: 12345678
`, serverPort, util.HTTPPort)
	go func() {
		proxy, err := proxy.NewProxyFromConfigData([]byte(serverData), false)
		common.Must(err)
		common.Must(proxy.Run())
	}()

	time.Sleep(time.Second * 2)

	payload := util.GeneratePayload(1024)
	buf := [1024]byte{}

	conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", clientPort))
	common.Must(err)

	common.Must2(conn.Write(payload))
	common.Must2(conn.Read(buf[:]))

	if !bytes.Equal(payload, buf[:]) {
		t.Fail()
	}

	packet, err := net.ListenPacket("udp", "")
	common.Must(err)
	common.Must2(packet.WriteTo(payload, &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: clientPort,
	}))
	_, _, err = packet.ReadFrom(buf[:])
	common.Must(err)
	if !bytes.Equal(payload, buf[:]) {
		t.Fail()
	}
}

func SingleThreadBenchmark(clientData, serverData string, socksPort int) {
	server, err := proxy.NewProxyFromConfigData([]byte(clientData), false)
	common.Must(err)
	go server.Run()

	client, err := proxy.NewProxyFromConfigData([]byte(serverData), false)
	common.Must(err)
	go client.Run()

	time.Sleep(time.Second * 2)
	dialer, err := netproxy.SOCKS5("tcp", fmt.Sprintf("127.0.0.1:%d", socksPort), nil, netproxy.Direct)

	const num = 100
	wg := sync.WaitGroup{}
	wg.Add(num)
	const payloadSize = 1024 * 1024 * 1024
	payload := util.GeneratePayload(payloadSize)

	for i := 0; i < 100; i++ {
		conn, err := dialer.Dial("tcp", util.BlackHoleAddr)
		common.Must(err)

		t1 := time.Now()
		common.Must2(conn.Write(payload))
		t2 := time.Now()

		speed := float64(payloadSize) / (float64(t2.Sub(t1).Nanoseconds()) / float64(time.Second))
		fmt.Printf("speed: %f Gbps\n", speed/1024/1024/1024)

		conn.Close()
	}
	client.Close()
	server.Close()
	return
}

func BenchmarkClientServer(b *testing.B) {
	go func() {
		fmt.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	serverPort := common.PickPort("tcp", "127.0.0.1")
	socksPort := common.PickPort("tcp", "127.0.0.1")
	clientData := fmt.Sprintf(`
run-type: client
local-addr: 127.0.0.1
local-port: %d
remote-addr: 127.0.0.1
remote-port: %d
log-level: 0
password:
    - password
ssl:
    verify: false
    fingerprint: firefox
    sni: localhost
`, socksPort, serverPort)
	serverData := fmt.Sprintf(`
run-type: server
local-addr: 127.0.0.1
local-port: %d
remote-addr: 127.0.0.1
remote-port: %s
log-level: 0
disable-http-check: true
password:
    - password
ssl:
    verify-hostname: false
    key: server.key
    cert: server.crt
    sni: localhost
`, serverPort, util.HTTPPort)

	SingleThreadBenchmark(clientData, serverData, socksPort)
}
