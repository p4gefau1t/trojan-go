package client

import "github.com/p4gefau1t/trojan-go/config"

type MuxConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type WebsocketConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
}

type Config struct {
	Mux       MuxConfig       `json:"mux" yaml:"mux"`
	Websocket WebsocketConfig `json:"websocket" yaml:"websocket"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
