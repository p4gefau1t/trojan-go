// +build windows

package sockopt

import (
	"syscall"

	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
)

const (
	TCP_FASTOPEN = 15
)

func ApplySocketOption(fd uintptr, config *conf.TCPConfig, isInbound bool) error {
	if config.FastOpen {
		if err := syscall.SetsockoptInt(syscall.Handle(fd), syscall.IPPROTO_TCP, TCP_FASTOPEN, 1); err != nil {
			return err
		}
		log.Debug("tcp fast open enabled")
	}
	return nil
}
