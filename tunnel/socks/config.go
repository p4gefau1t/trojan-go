package socks

import "github.com/p4gefau1t/trojan-go/config"

type Config struct {
	LocalHost string `json:"local_addr" yaml:"local-addr"`
	LocalPort int    `json:"local_port" yaml:"local-port"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
