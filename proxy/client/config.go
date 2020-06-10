package client

import "github.com/p4gefau1t/trojan-go/config"

type MuxConfig struct {
	Enabled bool `json,yaml:"enabled"`
}

type WebsocketConfig struct {
	Enabled bool `json,yaml:"enabled"`
}

type RouterConfig struct {
	Enabled bool `json,yaml:"enabled"`
}

type Config struct {
	Mux       MuxConfig       `json,yaml:"mux"`
	Websocket WebsocketConfig `json,yaml:"websocket"`
	Router    RouterConfig    `json,yaml:"router"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
