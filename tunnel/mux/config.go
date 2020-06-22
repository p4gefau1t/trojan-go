package mux

import "github.com/p4gefau1t/trojan-go/config"

type MuxConfig struct {
	Enabled     bool `json,yaml:"enabled"`
	IdleTimeout int  `json:"idle_timeout" yaml:"idle-timeout"`
	Concurrency int  `json,yaml:"concurrency"`
}

type Config struct {
	Mux MuxConfig `json,yaml:"mux"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			Mux: MuxConfig{
				Enabled:     false,
				IdleTimeout: 30,
				Concurrency: 8,
			},
		}
	})

}
