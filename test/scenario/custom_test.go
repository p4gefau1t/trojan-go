package scenario

import (
	"fmt"
	"github.com/p4gefau1t/trojan-go/common"
	_ "github.com/p4gefau1t/trojan-go/proxy/custom"
	"github.com/p4gefau1t/trojan-go/test/util"
	"testing"
)

func TestCustom(t *testing.T) {
	serverPort := common.PickPort("tcp", "127.0.0.1")
	socksPort := common.PickPort("tcp", "127.0.0.1")
	clientData := fmt.Sprintf(`
run-type: custom

inbound:
  node:
    - protocol: transport
      tag: transport
      config:
        local-addr: 127.0.0.1
        local-port: %d
    - protocol: socks
      tag: socks
  path:
    -
      - transport
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
          - 12345678

  path:
    - 
      - transport
      - tls
      - trojan

`, socksPort, serverPort)
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
          - 12345678

  path:
    - 
      - transport
      - tls
      - trojan

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
