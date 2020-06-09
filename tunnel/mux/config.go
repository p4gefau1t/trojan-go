package mux

import "github.com/p4gefau1t/trojan-go/config"

type MuxConfig struct {
	Enabled     bool `json,yaml:"enabled"`
	Timeout     int  `json,yaml:"timeout"`
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
				Timeout:     30,
				Concurrency: 8,
			},
		}
	})

}
