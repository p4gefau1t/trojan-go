package trojan

import "github.com/p4gefau1t/trojan-go/config"

type Config struct {
	LocalHost  string `json:"local_addr" yaml:"local-addr"`
	LocalPort  int    `json:"local_port" yaml:"local-port"`
	RemoteHost string `json:"remote_addr" yaml:"remote-addr"`
	RemotePort int    `json:"remote_port" yaml:"remote-port"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{}
	})
}
