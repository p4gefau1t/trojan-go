package shadowsocks

import "github.com/p4gefau1t/trojan-go/config"

type ShadowsocksConfig struct {
	Enabled  bool   `json:"enabled" yaml:"enabled"`
	Method   string `json:"method" yaml:"method"`
	Password string `json:"password" yaml:"password"`
}

type Config struct {
	RemoteHost  string            `json:"remote_addr" yaml:"remote-addr"`
	RemotePort  int               `json:"remote_port" yaml:"remote-port"`
	Shadowsocks ShadowsocksConfig `json:"shadowsocks" yaml:"shadowsocks"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			Shadowsocks: ShadowsocksConfig{
				Method: "AES-128-GCM",
			},
		}
	})
}
