package websocket

import "github.com/p4gefau1t/trojan-go/config"

type WebsocketConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	Host    string `json:"host" yaml:"host"`
	Path    string `json:"path" yaml:"path"`
}

type Config struct {
	RemoteHost string          `json:"remote_addr" yaml:"remote-addr"`
	RemotePort int             `json:"remote_port" yaml:"remote-port"`
	Websocket  WebsocketConfig `json:"websocket" yaml:"websocket"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
