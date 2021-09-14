//go:build custom || full
// +build custom full

package build

import (
	_ "github.com/p4gefau1t/trojan-go/proxy/custom"
	_ "github.com/p4gefau1t/trojan-go/tunnel/adapter"
	_ "github.com/p4gefau1t/trojan-go/tunnel/dokodemo"
	_ "github.com/p4gefau1t/trojan-go/tunnel/freedom"
	_ "github.com/p4gefau1t/trojan-go/tunnel/http"
	_ "github.com/p4gefau1t/trojan-go/tunnel/mux"
	_ "github.com/p4gefau1t/trojan-go/tunnel/router"
	_ "github.com/p4gefau1t/trojan-go/tunnel/shadowsocks"
	_ "github.com/p4gefau1t/trojan-go/tunnel/simplesocks"
	_ "github.com/p4gefau1t/trojan-go/tunnel/socks"
	_ "github.com/p4gefau1t/trojan-go/tunnel/tls"
	_ "github.com/p4gefau1t/trojan-go/tunnel/tproxy"
	_ "github.com/p4gefau1t/trojan-go/tunnel/transport"
	_ "github.com/p4gefau1t/trojan-go/tunnel/trojan"
	_ "github.com/p4gefau1t/trojan-go/tunnel/websocket"
)
