package custom

import "github.com/p4gefau1t/trojan-go/config"

const Name = "CUSTOM"

type NodeConfig struct {
	Protocol string      `json,yaml:"protocol"`
	Tag      string      `json,yaml:"tag"`
	Config   interface{} `json,yaml:"config"`
}

type StackConfig struct {
	Path [][]string   `json,yaml:"path"`
	Node []NodeConfig `json,yaml:"node"`
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
