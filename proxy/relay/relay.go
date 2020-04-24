package relay

import (
	"context"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/proxy"
)

type Relay struct {
	common.Runnable
	config   *conf.GlobalConfig
	ctx      context.Context
	cancel   context.CancelFunc
	listener net.Listener
}

func (f *Relay) handleConn(conn net.Conn) {
	defer conn.Close()
	newConn, err := net.Dial("tcp", f.config.RemoteAddress.String())
	if err != nil {
		log.Error("failed to connect to remote endpoint:", err)
		return
	}
	defer newConn.Close()
	proxy.ProxyConn(f.ctx, newConn, conn, f.config.BufferSize)
}

func (f *Relay) Run() error {
	log.Info("relay is running at", f.config.LocalAddress)
	listener, err := net.Listen("tcp", f.config.LocalAddress.String())
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

func (f *Relay) Close() error {
	log.Info("shutting down relay..")
	f.cancel()
	f.listener.Close()
	return nil
}

func (f *Relay) Build(config *conf.GlobalConfig) (common.Runnable, error) {
	f.ctx, f.cancel = context.WithCancel(context.Background())
	f.config = config
	return f, nil
}

func init() {
	proxy.RegisterProxy(conf.Relay, &Relay{})
}
