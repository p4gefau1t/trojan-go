package proxy

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
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
	log.Info("Trojan-Go", common.Version, "initializing")
	log.Info("Loading config file from", *c.args)

	//exit code 23 stands for initializing error, and systemd will not trying to restart it
	data, err := ioutil.ReadFile(*c.args)
	if err != nil {
		log.Error(common.NewError("Failed to read config file").Base(err))
		os.Exit(23)
	}
	config, err := conf.ParseJSON(data)
	if err != nil {
		log.Error(common.NewError("Failed to parse config file").Base(err))
		os.Exit(23)
	}
	proxy, err := NewProxy(config)
	if err != nil {
		log.Error(common.NewError("Failed to launch proxy").Base(err))
		os.Exit(23)
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
		log.Fatal(err)
		return err
	}
}

func init() {
	common.RegisterOptionHandler(&proxyOption{
		args: flag.String("config", common.GetProgramDir()+"/config.json", "Config filename"),
	})
}
