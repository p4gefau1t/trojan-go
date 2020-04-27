package sockopt

import (
	"net"

	"github.com/p4gefau1t/trojan-go/conf"
)

func ApplyTCPListenerOption(l *net.TCPListener, config *conf.TCPConfig) error {
	rawConn, err := l.SyscallConn()
	if err != nil {
		return err
	}
	rawConn.Control(func(fd uintptr) {
		err = ApplySocketOption(fd, config, true)
	})
	return err
}

func ApplyTCPConnOption(conn *net.TCPConn, config *conf.TCPConfig) error {
	if err := conn.SetKeepAlive(config.KeepAlive); err != nil {
		return err
	}
	if err := conn.SetNoDelay(config.NoDelay); err != nil {
		return err
	}
	rawConn, err := conn.SyscallConn()
	if err != nil {
		return err
	}
	rawConn.Control(func(fd uintptr) {
		err = ApplySocketOption(fd, config, false)
	})
	return err
}
