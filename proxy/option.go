package proxy

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
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

func HandleHelper(file_path string) error {
	isJSON := false
	switch {
	case strings.HasSuffix(file_path, ".json"):
		isJSON = true
	case strings.HasSuffix(file_path, ".yaml"), strings.HasSuffix(file_path, ".yml"):
		isJSON = false
	default:
		log.Fatalf("unsupported filename suffix %s. use .yaml or .json instead.", file_path)
	}

	data, err := ioutil.ReadFile(file_path)
	if err != nil {
		switch {
		case strings.HasSuffix(err.Error(), "The system cannot find the file specified."),
			strings.HasSuffix(err.Error(), "no such file or directory"):
			return err
		default:
			log.Fatal(err)
		}
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

func (o *Option) Handle() error {
	if *o.path != "Not_Specified" {
		err := HandleHelper(*o.path)
		if err != nil {
			log.Fatal(err)
		} else {
			return nil
		}
	}

	file_paths := [3]string{"config.json", "config.yml", "config.yaml"}

	for i := 0; i < 3; i++ {
		log.Infof("loading config from default path %s", file_paths[i])
		err := HandleHelper(file_paths[i])
		if err == nil {
			return nil
		} else {
			log.Warn(err)
		}
	}

	log.Fatal("no config provided: put a config.json/yml/yaml in the directory or specify path with -config")
	return nil
}

func (o *Option) Priority() int {
	return 1
}

func init() {
	option.RegisterHandler(&Option{
		path: flag.String("config", "Not_Specified", "Trojan-Go config filename (.yaml/.yml/.json)"),
	})
	option.RegisterHandler(&StdinOption{
		format:       flag.String("stdin-format", "disabled", "Read from standard input (yaml/json)"),
		suppressHint: flag.Bool("stdin-suppress-hint", false, "Suppress hint text"),
	})
}

type StdinOption struct {
	format       *string
	suppressHint *bool
}

func (o *StdinOption) Name() string {
	return Name + "_STDIN"
}

func (o *StdinOption) Handle() error {
	isJSON, e := o.isFormatJson()
	if e != nil {
		return e
	}

	if o.suppressHint == nil || !*o.suppressHint {
		fmt.Printf("Trojan-Go %s (%s/%s)\n", constant.Version, runtime.GOOS, runtime.GOARCH)
		if isJSON {
			fmt.Println("Reading in JSON configuration from STDIN.")
		} else {
			fmt.Println("Reading in YAML configuration from STDIN.")
		}
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
	if *o.format == "disabled" {
		return false, common.NewError("reading from stdin is disabled")
	}
	return strings.ToLower(*o.format) == "json", nil
}
