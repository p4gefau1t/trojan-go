// +build linux

package sockopt

import (
	"syscall"

	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"golang.org/x/sys/unix"
)

func ApplySocketOption(fd uintptr, config *conf.TCPConfig, isInbound bool) error {
	if config.ReusePort && isInbound {
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
			return err
		}
		if err := syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1); err != nil {
			return err
		}
		log.Debug("port reusing enabled")
	}

	if config.FastOpen {
		if isInbound {
			if err := syscall.SetsockoptInt(int(fd), syscall.SOL_TCP, unix.TCP_FASTOPEN, config.FastOpenQLen); err != nil {
				return err
			}
		} else {
			//if err := syscall.SetsockoptInt(int(fd), syscall.SOL_TCP, unix.TCP_FASTOPEN_CONNECT, 1); err != nil {
			//return err
			//}
		}
		log.Debug("tcp fast open enabled")
	}
	return nil
}
