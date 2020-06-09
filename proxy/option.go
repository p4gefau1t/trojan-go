package proxy

import (
	"flag"
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
	if strings.HasSuffix(*o.path, ".json") {
		if err := RunProxy(data, true); err != nil {
			log.Fatal(err)
		}
	} else if strings.HasSuffix(*o.path, ".yaml") {
		if err := RunProxy(data, false); err != nil {
			log.Fatal(err)
		}
	} else {
		log.Fatal("unknown file suffix", *o.path)
	}
	return nil
}

func (o *Option) Priority() int {
	return 0
}

func init() {
	option.RegisterOptionHandler(&Option{
		path: flag.String("config", "config.json", "Trojan-Go config filename (.yaml/.json)"),
	})
}
