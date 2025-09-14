package http

import (
	"github.com/p4gefau1t/trojan-go/config"
)

const Name = "HTTP"

type Config struct {
	HTTP HTTPConfig `json:"http" yaml:"http"`
}

type HTTPConfig struct {
	Enabled bool   `json:"enabled" yaml:"enabled"`
	APIURL  string `json:"api_url" yaml:"api_url"`
	Timeout int    `json:"timeout" yaml:"timeout"` // 超时时间（秒）
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			HTTP: HTTPConfig{
				Enabled: false,
				APIURL:  "",
				Timeout: 10,
			},
		}
	})
}
