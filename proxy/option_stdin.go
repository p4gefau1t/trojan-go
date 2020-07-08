package proxy

import (
	"bufio"
	"errors"
	"flag"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/option"
	"io/ioutil"
	"os"
	"strings"
)

type StdinOption struct {
	format *string
}

func (o *StdinOption) Name() string {
	return Name
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

func init() {
	option.RegisterHandler(&StdinOption{
		format: flag.String("stdin-format", "yaml", "Read From Standard Input (yaml/json)"),
	})
}

func (o *StdinOption) isFormatJson() (isJson bool, e error) {
	if o.format == nil {
		return false, errors.New("format specifier is nil")
	}
	return strings.ToLower(*o.format) == "json", nil
}
