package freedom

import "github.com/p4gefau1t/trojan-go/config"

type Config struct {
	LocalHost    string             `json:"local_addr" yaml:"local-addr"`
	LocalPort    int                `json:"local_port" yaml:"local-port"`
	TCP          TCPConfig          `json:"tcp" yaml:"tcp"`
	ForwardProxy ForwardProxyConfig `json:"forward_proxy" yaml:"forward-proxy"`
}

type TCPConfig struct {
	PreferIPV4 bool `json:"prefer_ipv4" yaml:"prefer-ipv4"`
	KeepAlive  bool `json:"keep_alive" yaml:"keep-alive"`
	NoDelay    bool `json:"no_delay" yaml:"no-delay"`
}

type ForwardProxyConfig struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	ProxyHost string `json:"proxy_addr" yaml:"proxy-addr"`
	ProxyPort int    `json:"proxy_port" yaml:"proxy-port"`
	Username  string `json:"username" yaml:"username"`
	Password  string `json:"password" yaml:"password"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			TCP: TCPConfig{
				PreferIPV4: false,
				NoDelay:    true,
				KeepAlive:  true,
			},
		}
	})
}
