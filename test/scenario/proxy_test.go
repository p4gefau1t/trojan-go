package scenario

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	_ "net/http/pprof"
	"sync"
	"testing"
	"time"

	netproxy "golang.org/x/net/proxy"

	_ "github.com/p4gefau1t/trojan-go/api"
	_ "github.com/p4gefau1t/trojan-go/api/service"
	"github.com/p4gefau1t/trojan-go/common"
	_ "github.com/p4gefau1t/trojan-go/log/golog"
	"github.com/p4gefau1t/trojan-go/proxy"
	_ "github.com/p4gefau1t/trojan-go/proxy/client"
	_ "github.com/p4gefau1t/trojan-go/proxy/forward"
	_ "github.com/p4gefau1t/trojan-go/proxy/nat"
	_ "github.com/p4gefau1t/trojan-go/proxy/server"
	_ "github.com/p4gefau1t/trojan-go/statistic/memory"
	"github.com/p4gefau1t/trojan-go/test/util"
)

// test key and cert

var cert = `
-----BEGIN CERTIFICATE-----
MIIC+TCCAeGgAwIBAgIQAZ1MkNXl76ABOPPQ6ci25zANBgkqhkiG9w0BAQsFADAS
MRAwDgYDVQQKEwdBY21lIENvMB4XDTIwMDkwNjAzMTM1NVoXDTIxMDkwNjAzMTM1
NVowEjEQMA4GA1UEChMHQWNtZSBDbzCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC
AQoCggEBAJrgYRlmw0851xp1/OWN4b/gQsKc1TmTaMcN+gDX8w2RWfgOOGymeWDJ
QaTu1G6XHjuvW3sqZGixlRtUJKnmldiwpX0ZY00Ce15fgQHZ85uc3rnFjdkeaYFj
KXN7Xx0QZTQjR5N3W5oVvwKRXe9ATwtOKregSJxCMv8P6OWYH8SwR8GsZnUyvKYR
7JodTXpw7pIL4yNx+QETg537y0TXVFpVt0/H9OoKmY/vsIWVWkKOY4nre9XxNf/p
ABWxYy5n1CKTssLWblJs/lSSPfRxCKUnrBcHwr8ZvLwZSvVktLWr0DnurdfSXOSy
nGvF19q7BpB47ZDTca4V95UtqgfquwUCAwEAAaNLMEkwDgYDVR0PAQH/BAQDAgWg
MBMGA1UdJQQMMAoGCCsGAQUFBwMBMAwGA1UdEwEB/wQCMAAwFAYDVR0RBA0wC4IJ
bG9jYWxob3N0MA0GCSqGSIb3DQEBCwUAA4IBAQBzaEBQs2bjx0trJxDoKK5xFDUX
mhhVOlparYS04WG3q18r9qfcvXDv3DOmzJDAnSldGmHad/ba6uLDuGEtuIYdMK9u
CpQVaLsNsjIeSika7l0fbQ7XBAJzIHkQHF8dGS3qyzagyCLiRuV2qT5v+p6X4tbp
PY2raoobm5hiscLk540mAAboz+IM1nTGuxD+XUh9znnGJhiKVoNnWhhXLHQK3Lwd
Mct/q+LkMaVHgT/r5LBMbk/jPluvgN0VJ6FnEw1JmotduJd+f80Syp4qccZmupEe
zNXfXCPNcNXeSbAwWnsFeiUrU5YNqPobhaiZXMGnoFb4Cufb57AbNPNDch0x
-----END CERTIFICATE-----
`

var key = `
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCa4GEZZsNPOdca
dfzljeG/4ELCnNU5k2jHDfoA1/MNkVn4DjhspnlgyUGk7tRulx47r1t7KmRosZUb
VCSp5pXYsKV9GWNNAnteX4EB2fObnN65xY3ZHmmBYylze18dEGU0I0eTd1uaFb8C
kV3vQE8LTiq3oEicQjL/D+jlmB/EsEfBrGZ1MrymEeyaHU16cO6SC+MjcfkBE4Od
+8tE11RaVbdPx/TqCpmP77CFlVpCjmOJ63vV8TX/6QAVsWMuZ9Qik7LC1m5SbP5U
kj30cQilJ6wXB8K/Gby8GUr1ZLS1q9A57q3X0lzkspxrxdfauwaQeO2Q03GuFfeV
LaoH6rsFAgMBAAECggEBAIEbGtZ5+8ZHiTDdunwB0naJFB33bygX4fhNhmK9ojdl
O4K1GAQ6omQ0YSyEi0HFZ8aJX9FEfX9oycuGUSnwtml0l/+48jZ4Iy+AnaJVdeX6
1xA1xxF/cKQTbbJ+3cL0r+jOoBQmI45HInuZgpy3Fy1tc96vFthrtuc49ASw04q5
vIgA+oX6dt7ex7WpXJexqO/9wVsFdiy01gF+e3n2UX5C7F5mm3m0ZJI0A9LCFIim
caLqgqSkFXujw0JurIwLolc/qRn/HG/gfMKBpf1ESpj3dDoZFaRftqtUjzXswD82
eZ1PbfpEZ6iUr4K0scUcDdYrupth6U2tDiPz5y1kduECgYEAwrvqC4ulazroqe+e
LrzftOwg7J3gGyMl+ZTG8Fa4Gd2sAJ0R5kGmcaVU4LW4Ysm5lhgXlKLAGLdPCreS
pruSz1SNgXgYBEnj4Pz0zluQbPmdgGNQ/pOxtI8pr1NGkLwq76M0M8pGHjytIV8N
w1FGikm0Zk4ZSFAFVU5GKCL4hYsCgYEAy5pLi+AejuQuSlR+9aZLwisw3snqGxxe
ECKtPaHAjp/OI43/TGXQihoZJyAYdlDbIIwf03xV14Vv8MPjgtsCFf913YAWSp+y
x1Ul9kGYtVL8QeMcPs1Tb+0BU9VrTDegLNuDNIsxl3pERXIjwotDvQGiTIW7rTY5
SiPOhrlec68CgYAxf/jfVHEJD+FiiRFpigNHhxpba0ozO70Ec0gagcCseoelZEfP
gvKfQsqPkEG9gs+VEqyz0KcJ4VbLP5ycm2OXJkQOHAvm0y2E3GgSKH5O5SifIR/O
hpaOcjHDamSul9ZGMfMsEwe92eicagAinP9UWaXst39/vS+N3qbAvxrzPwKBgQCS
eumLMq0JhKTBGVVWClRK16QLRR1Gb/xEg4473xmYAuTds5VPM5j7IpeiDHdM+BMO
sYFcOAHSUtAcWfJe/I3dobL8ruBaw9ZtjpcHOl5RZejSxkBV9obm6Y6g79SIOyTj
4PHeZZ5CKtbfV6TenC8Z1gkcIMLLdU12R5iYWNjZRQKBgFrFy2jVQHrKVas2Fu+o
HYLaMfoodHq4RWLMf64jSpXkJt8jB1A8vI0ekMe2gTXaldRvinYjuhzU/zJIkWuA
LYIN/nRkP0BLRwfZklUbdO3h1lvvlxM533luvX5mo41Gjg/b2f36yRXTa01Q+QML
NYpAJoagHIeNLGo4aJFwiVsZ
-----END PRIVATE KEY-----
`

func init() {
	ioutil.WriteFile("server.crt", []byte(cert), 0777)
	ioutil.WriteFile("server.key", []byte(key), 0777)
}

func CheckClientServer(clientData, serverData string, socksPort int) (ok bool) {
	server, err := proxy.NewProxyFromConfigData([]byte(serverData), false)
	common.Must(err)
	go server.Run()

	client, err := proxy.NewProxyFromConfigData([]byte(clientData), false)
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
    host: somedomainname.com
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
    host: 127.0.0.1
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

func TestLeak(t *testing.T) {
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
shadowsocks:
    enabled: true
    method: AEAD_CHACHA20_POLY1305
    password: 12345678
mux:
    enabled: true
api:
    enabled: true
    api-port: 0
`, socksPort, serverPort)
	client, err := proxy.NewProxyFromConfigData([]byte(clientData), false)
	common.Must(err)
	go client.Run()
	time.Sleep(time.Second * 3)
	client.Close()
	time.Sleep(time.Second * 3)
	//http.ListenAndServe("localhost:6060", nil)
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
