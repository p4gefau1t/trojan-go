package proxy

import (
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
)

type Forward struct {
	common.Runnable
	config *conf.GlobalConfig
}

func (f *Forward) handleConn(conn net.Conn) {
	newConn, err := net.Dial("tcp", f.config.RemoteAddr.String())
	if err != nil {
		logger.Error("failed to connect to remote endpoint:", err)
		return
	}
	proxyConn(newConn, conn)
}

func (f *Forward) Run() error {
	listener, err := net.Listen("tcp", f.config.LocalAddr.String())
	if err != nil {
		return common.NewError("failed to listen local address").Base(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error(err)
			continue
		}
		go f.handleConn(conn)
	}
}
