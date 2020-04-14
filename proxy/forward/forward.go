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
	config   *conf.GlobalConfig
	ctx      context.Context
	cancel   context.CancelFunc
	listener net.Listener
}

func (f *Forward) handleConn(conn net.Conn) {
	newConn, err := net.Dial("tcp", f.config.RemoteAddr.String())
	if err != nil {
		log.Error("failed to connect to remote endpoint:", err)
		return
	}
	proxy.ProxyConn(newConn, conn)
}

func (f *Forward) Run() error {
	listener, err := net.Listen("tcp", f.config.LocalAddr.String())
	f.listener = listener
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
			log.Error(err)
			return err
		}
		go f.handleConn(conn)
	}
}

func (f *Forward) Close() error {
	log.Info("shutting down forward..")
	f.listener.Close()
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
