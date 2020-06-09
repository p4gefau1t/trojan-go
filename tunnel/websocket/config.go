package websocket

import "github.com/p4gefau1t/trojan-go/config"

type WebsocketConfig struct {
	Enabled  bool   `json,yaml:"enabled""`
	Hostname string `json,yaml:"hostname"`
	Path     string `json,yaml:"path"`
}

type Config struct {
	RemoteHost string          `json:"remote_addr" yaml:"remote-addr"`
	RemotePort int             `json:"remote_port" yaml:"remote-port"`
	Websocket  WebsocketConfig `json,yaml:"websocket"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
