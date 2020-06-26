package custom

import "github.com/p4gefau1t/trojan-go/config"

const Name = "CUSTOM"

type NodeConfig struct {
	Protocol string      `json:"protocol" yaml:"protocol"`
	Tag      string      `json:"tag" yaml:"tag"`
	Config   interface{} `json:"config" yaml:"config"`
}

type StackConfig struct {
	Path [][]string   `json:"path" yaml:"path"`
	Node []NodeConfig `json:"node" yaml:"node"`
}

type Config struct {
	Inbound  StackConfig `json:"inbound" yaml:"inbound"`
	Outbound StackConfig `json:"outbound" yaml:"outbound"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
