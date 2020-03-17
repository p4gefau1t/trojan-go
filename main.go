package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/proxy"
)

func main() {
	configFile := flag.String("config", "config.json", "Config file name")
	flag.Parse()
	data, err := ioutil.ReadFile(*configFile)
	if err != nil {
		log.Fatal("cannot read config:", *configFile)
	}
	config, err := conf.ParseJSON(data)
	if err != nil {
		log.Fatal("cannot parfse config:", err)
	}
	err = proxy.NewProxy(config).Run()
	log.Fatal(err)
}
