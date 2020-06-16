package socks

import "github.com/p4gefau1t/trojan-go/config"

type Config struct {
	UDPTimeout int `json:"udp_timeout" yaml:"udp-timeout"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			UDPTimeout: 30,
		}
	})
}
