package memory

import (
	"github.com/p4gefau1t/trojan-go/config"
)

type Config struct {
	Passwords []string `json:"password" yaml:"password"`
	Sqlite    string   `json:"sqlite" yaml:"sqlite"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{}
	})
}
