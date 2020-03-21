// +build !windows

package proxy

import (
	"net"

	"github.com/valyala/tcplisten"
)

func ListenWithTCPOption(fastOpen, reusePort, noDelay bool, ip net.IP, addr string) (net.Listener, error) {
	cfg := tcplisten.Config{
		ReusePort:   reusePort,
		FastOpen:    fastOpen,
		DeferAccept: noDelay,
	}
	network := "tcp6"
	if ip.To4() != nil {
		network = "tcp4"
	}
	return cfg.NewListener(network, addr)
}
