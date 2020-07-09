package proxy

import (
	"bufio"
	"flag"
	"io/ioutil"
	"os"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/constant"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/option"
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
		path: flag.String("config", "config.json", "Trojan-Go config filename (.yaml/.yml/.json)"),
	})
	option.RegisterHandler(&StdinOption{
		format: flag.String("stdin-format", "yaml", "Read from standard input (yaml/json)"),
	})
}

type StdinOption struct {
	format *string
}

func (o *StdinOption) Name() string {
	return Name + "_STDIN"
}

func (o *StdinOption) Handle() error {
	isJSON, e := o.isFormatJson()
	if e != nil {
		return e
	}

	data, e := ioutil.ReadAll(bufio.NewReader(os.Stdin))
	if e != nil {
		log.Fatalf("Failed to read from stdin: %s", e.Error())
	}

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

func (o *StdinOption) Priority() int {
	return 0
}

func (o *StdinOption) isFormatJson() (isJson bool, e error) {
	if o.format == nil {
		return false, common.NewError("format specifier is nil")
	}
	return strings.ToLower(*o.format) == "json", nil
}
