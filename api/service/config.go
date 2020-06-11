package service

import "github.com/p4gefau1t/trojan-go/config"

const Name = "API_SERVICE"

type APIConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	APIHost string `json:"api_addr" yaml:"api-addr"`
	APIPort int    `json:"api_port" yaml:"api-port"`
}

type Config struct {
	API APIConfig `json,yaml:"api"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
