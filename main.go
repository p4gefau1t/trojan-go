package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/guide"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/proxy"
)

var logger = log.New(os.Stdout)

func main() {
	logger.Info("Trojan-Go initializing...")
	configFile := flag.String("config", "config.json", "Config filename")
	guideMode := flag.String("guide", "", "guide mode, use -guide cert to request a cert from letsencrypt")
	flag.Parse()
	switch *guideMode {
	case "cert":
		guide.CertGuide()
		return
	default:
	}
	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		logger.Fatal("Failed to read config file", err)
	}
	config, err := conf.ParseJSON(data)
	if err != nil {
		logger.Fatal("Failed to parse config file", err)
	}
	proxy := proxy.NewProxy(config)
	errChan := make(chan error)
	go func() {
		errChan <- proxy.Run()
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	select {
	case <-sigs:
		proxy.Close()
	case err := <-errChan:
		logger.Fatal(err)
	}
	logger.Info("Trojan-Go exited")
}
