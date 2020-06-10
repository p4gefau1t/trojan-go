package shadowsocks

import "github.com/p4gefau1t/trojan-go/config"

type ShadowsocksConfig struct {
	Enabled  bool   `json,yaml:"enabled"`
	Method   string `json,yaml:"method"`
	Password string `json,yaml:"password"`
}

type Config struct {
	Shadowsocks ShadowsocksConfig `json,yaml:"shadowsocks"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
