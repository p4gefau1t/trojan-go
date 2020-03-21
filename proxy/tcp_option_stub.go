// +build windows

package proxy

import (
	"net"
)

func ListenWithTCPOption(fastOpen, reusePort, noDelay bool, ip net.IP, addr string) (net.Listener, error) {
	panic("os not support tcp option")
}
