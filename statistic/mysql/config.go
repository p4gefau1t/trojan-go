package mysql

import (
	"github.com/p4gefau1t/trojan-go/config"
)

type MySQLConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	ServerHost string `json:"server_addr" yaml:"server-addr"`
	ServerPort int    `json:"server_port" yaml:"server-port"`
	Database   string `json:"database" yaml:"database"`
	Username   string `json:"username" yaml:"username"`
	Password   string `json:"password" yaml:"password"`
	Key        string `json:"key" yaml:"key"`
	Cert       string `json:"cert" yaml:"cert"`
	CA         string `json:"ca" yaml:"ca"`
	CheckRate  int    `json:"check_rate" yaml:"check-rate"`
}

type Config struct {
	MySQL MySQLConfig `json:"mysql" yaml:"mysql"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			MySQL: MySQLConfig{
				ServerPort: 3306,
				CheckRate:  30,
			},
		}
	})
}
