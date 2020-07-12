package scenario

import (
	"fmt"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	_ "github.com/p4gefau1t/trojan-go/proxy/custom"
	"github.com/p4gefau1t/trojan-go/test/util"
)

func TestCustom1(t *testing.T) {
	serverPort := common.PickPort("tcp", "127.0.0.1")
	socksPort := common.PickPort("tcp", "127.0.0.1")
	clientData := fmt.Sprintf(`
run-type: custom

inbound:
  node:
    - protocol: adapter
      tag: adapter
      config:
        local-addr: 127.0.0.1
        local-port: %d
    - protocol: socks
      tag: socks
      config:
        local-addr: 127.0.0.1
        local-port: %d
  path:
    -
      - adapter
      - socks

outbound:
  node:
    - protocol: transport
      tag: transport
      config:
        remote-addr: 127.0.0.1
        remote-port: %d

    - protocol: tls
      tag: tls
      config:
        ssl:
          sni: localhost
          key: server.key
          cert: server.crt

    - protocol: trojan
      tag: trojan
      config:
        password:
          - "12345678"

  path:
    - 
      - transport
      - tls
      - trojan

`, socksPort, socksPort, serverPort)
	serverData := fmt.Sprintf(`
run-type: custom

inbound:
  node:
    - protocol: transport
      tag: transport
      config:
        local-addr: 127.0.0.1
        local-port: %d
        remote-addr: 127.0.0.1
        remote-port: %s

    - protocol: tls
      tag: tls
      config:
        ssl:
          sni: localhost
          key: server.key
          cert: server.crt

    - protocol: trojan
      tag: trojan
      config:
        disable-http-check: true
        password:
          - "12345678"

    - protocol: mux
      tag: mux

    - protocol: simplesocks
      tag: simplesocks
     

  path:
    - 
      - transport
      - tls
      - trojan
    - 
      - transport
      - tls
      - trojan
      - mux
      - simplesocks

outbound:
  node:
    - protocol: freedom
      tag: freedom

  path:
    - 
      - freedom

`, serverPort, util.HTTPPort)

	if !CheckClientServer(clientData, serverData, socksPort) {
		t.Fail()
	}
}

func TestCustom2(t *testing.T) {
	serverPort := common.PickPort("tcp", "127.0.0.1")
	socksPort := common.PickPort("tcp", "127.0.0.1")
	clientData := fmt.Sprintf(`
run-type: custom
log-level: 0

inbound:
  node:
    - protocol: adapter
      tag: adapter
      config:
        local-addr: 127.0.0.1
        local-port: %d
    - protocol: socks
      tag: socks
      config:
        local-addr: 127.0.0.1
        local-port: %d
  path:
    -
      - adapter
      - socks

outbound:
  node:
    - protocol: transport
      tag: transport
      config:
        remote-addr: 127.0.0.1
        remote-port: %d

    - protocol: tls
      tag: tls
      config:
        ssl:
          sni: localhost
          key: server.key
          cert: server.crt

    - protocol: trojan
      tag: trojan
      config:
        password:
          - "12345678"

    - protocol: shadowsocks
      tag: shadowsocks
      config:
        remote-addr: 127.0.0.1
        remote-port: 80
        shadowsocks:
          enabled: true
          password: "12345678"

    - protocol: websocket
      tag: websocket
      config:
        websocket:
          host: localhost
          path: /ws

  path:
    - 
      - transport
      - tls
      - websocket
      - shadowsocks 
      - trojan

`, socksPort, socksPort, serverPort)
	serverData := fmt.Sprintf(`
run-type: custom
log-level: 0

inbound:
  node:
    - protocol: transport
      tag: transport
      config:
        local-addr: 127.0.0.1
        local-port: %d
        remote-addr: 127.0.0.1
        remote-port: %s

    - protocol: tls
      tag: tls
      config:
        ssl:
          sni: localhost
          key: server.key
          cert: server.crt

    - protocol: trojan
      tag: trojan
      config:
        disable-http-check: true
        password:
          - "12345678"

    - protocol: trojan
      tag: trojan2
      config:
        disable-http-check: true
        password:
          - "12345678"

    - protocol: websocket
      tag: websocket
      config:
        websocket:
          enabled: true
          host: localhost
          path: /ws

    - protocol: mux
      tag: mux

    - protocol: simplesocks
      tag: simplesocks

    - protocol: shadowsocks
      tag: shadowsocks
      config:
        remote-addr: 127.0.0.1
        remote-port: 80
        shadowsocks:
          enabled: true
          password: "12345678"

    - protocol: shadowsocks
      tag: shadowsocks2
      config:
        remote-addr: 127.0.0.1
        remote-port: 80
        shadowsocks:
          enabled: true
          password: "12345678"
     
  path:
    - 
      - transport
      - tls
      - shadowsocks 
      - trojan
    - 
      - transport
      - tls
      - websocket
      - shadowsocks2
      - trojan2
    - 
      - transport
      - tls
      - shadowsocks
      - trojan
      - mux
      - simplesocks

outbound:
  node:
    - protocol: freedom
      tag: freedom

  path:
    - 
      - freedom

`, serverPort, util.HTTPPort)

	if !CheckClientServer(clientData, serverData, socksPort) {
		t.Fail()
	}
}
