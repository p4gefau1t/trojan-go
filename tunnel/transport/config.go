package transport

import (
	"github.com/p4gefau1t/trojan-go/config"
)

type Config struct {
	LocalHost       string                `json:"local_addr" yaml:"local-addr"`
	LocalPort       int                   `json:"local_port" yaml:"local-port"`
	RemoteHost      string                `json:"remote_addr" yaml:"remote-addr"`
	RemotePort      int                   `json:"remote_port" yaml:"remote-port"`
	TransportPlugin TransportPluginConfig `json:"transport_plugin" yaml:"transport-plugin"`
}

type TransportPluginConfig struct {
	Enabled bool     `json:"enabled" yaml:"enabled"`
	Type    string   `json:"type" yaml:"type"`
	Command string   `json:"command" yaml:"command"`
	Option  string   `json:"option" yaml:"option"`
	Arg     []string `json:"arg" yaml:"arg"`
	Env     []string `json:"env" yaml:"env"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
