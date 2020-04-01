package proxy

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
)

type proxyOption struct {
	args *string
	common.OptionHandler
}

func (*proxyOption) Name() string {
	return "proxy"
}

func (*proxyOption) Priority() int {
	return 0
}

func (c *proxyOption) Handle() error {
	logger.Info("Trojan-Go proxy initializing...")
	data, err := ioutil.ReadFile(*c.args)
	if err != nil {
		logger.Fatal(common.NewError("Failed to read config file").Base(err))
	}
	config, err := conf.ParseJSON(data)
	if err != nil {
		logger.Fatal(common.NewError("Failed to parse config file").Base(err))
	}
	proxy, err := NewProxy(config)
	if err != nil {
		logger.Fatal(err)
	}
	errChan := make(chan error)
	go func() {
		errChan <- proxy.Run()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case <-sigs:
		proxy.Close()
		return nil
	case err := <-errChan:
		logger.Fatal(err)
		return err
	}
}

func init() {
	common.RegisterOptionHandler(&proxyOption{
		args: flag.String("config", "config.json", "Config filename"),
	})
}
