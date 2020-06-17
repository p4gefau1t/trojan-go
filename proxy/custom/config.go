package custom

import "github.com/p4gefau1t/trojan-go/config"

const Name = "CUSTOM"

type Node struct {
	Protocol string
	Tag      string
	Config   interface{}
}

type StackConfig struct {
	NodeList []Node     `json,yaml:"node"`
	Path     [][]string `json,yaml:"path"`
}

type Config struct {
	Inbound  StackConfig `json,yaml:"inbound"`
	Outbound StackConfig `json,yaml:"outbound"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
