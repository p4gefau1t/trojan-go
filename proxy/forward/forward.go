package forward

import (
	"context"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/proxy"
)

type Forward struct {
	common.Runnable
	config *conf.GlobalConfig
	ctx    context.Context
	cancel context.CancelFunc
}

func (f *Forward) handleConn(conn net.Conn) {
	newConn, err := net.Dial("tcp", f.config.RemoteAddr.String())
	if err != nil {
		log.DefaultLogger.Error("failed to connect to remote endpoint:", err)
		return
	}
	proxy.ProxyConn(newConn, conn)
}

func (f *Forward) Run() error {
	listener, err := net.Listen("tcp", f.config.LocalAddr.String())
	if err != nil {
		return common.NewError("failed to listen local address").Base(err)
	}
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-f.ctx.Done():
				return nil
			default:
			}
			log.DefaultLogger.Error(err)
			continue
		}
		go f.handleConn(conn)
	}
}

func (f *Forward) Close() error {
	log.DefaultLogger.Info("shutting down forward..")
	f.cancel()
	return nil
}

func (f *Forward) Build(config *conf.GlobalConfig) (common.Runnable, error) {
	f.ctx, f.cancel = context.WithCancel(context.Background())
	f.config = config
	return f, nil
}

func init() {
	proxy.RegisterProxy(conf.Forward, &Forward{})
}
