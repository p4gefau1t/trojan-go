// +build windows

package server

import (
	"net"
)

func ListenWithTCPOption(fastOpen, reusePort, noDelay bool, ip net.IP, addr string) (net.Listener, error) {
	panic("this os does not support tcp options")
}
