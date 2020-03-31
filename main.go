package main

import (
	"flag"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/p4gefau1t/trojan-go/cert"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/proxy"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/p4gefau1t/trojan-go/proxy/client"
	_ "github.com/p4gefau1t/trojan-go/proxy/forward"
	_ "github.com/p4gefau1t/trojan-go/proxy/server"
)

var logger = log.New(os.Stdout)

func main() {
	logger.Info("Trojan-Go initializing...")
	configFile := flag.String("config", "config.json", "Config filename")
	certMode := flag.String("cert", "", "Simple letsencrpyt cert acme client. Use \"-cert request\" to request a cert or \"-cert renew\" to renew a cert")
	flag.Parse()
	switch *certMode {
	case "request":
		cert.RequestCertGuide()
		return
	case "renew":
		cert.RenewCertGuide()
		return
	case "":
	default:
		logger.Error("Invalid cert arg")
		return
	}
	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		logger.Fatal(common.NewError("Failed to read config file").Base(err))
	}
	config, err := conf.ParseJSON(data)
	if err != nil {
		logger.Fatal(common.NewError("Failed to parse config file").Base(err))
	}
	proxy, err := proxy.NewProxy(config)
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
	case err := <-errChan:
		logger.Fatal(err)
	}
	logger.Info("Trojan-Go exited")
}
