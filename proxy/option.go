package proxy

import (
	"flag"
	"github.com/p4gefau1t/trojan-go/constant"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/option"
	"io/ioutil"
	"strings"
)

type Option struct {
	path *string
}

func (o *Option) Name() string {
	return Name
}

func (o *Option) Handle() error {
	data, err := ioutil.ReadFile(*o.path)
	if err != nil {
		log.Fatal(err)
	}
	isJSON := false
	if strings.HasSuffix(*o.path, ".json") {
		isJSON = true
	} else if strings.HasSuffix(*o.path, ".yaml") || strings.HasSuffix(*o.path, ".yml") {
		isJSON = false
	} else {
		log.Fatal("unsupported filename suffix", *o.path, ". use .yaml or .json instead.")
	}
	log.Info("trojan-go", constant.Version, "initializing")
	proxy, err := NewProxyFromConfigData(data, isJSON)
	if err != nil {
		log.Fatal(err)
	}
	err = proxy.Run()
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

func (o *Option) Priority() int {
	return 0
}

func init() {
	option.RegisterHandler(&Option{
		path: flag.String("config", "config.json", "Trojan-Go config filename (.yaml/.json)"),
	})
}
