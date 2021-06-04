package sqlite

import (
	"github.com/p4gefau1t/trojan-go/config"
)

type SQLiteConfig struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Database  string `json:"path" yaml:"path"`
	CheckRate int    `json:"check_rate" yaml:"check-rate"`
}

type Config struct {
	SQLite SQLiteConfig `json:"sqlite" yaml:"sqlite"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			SQLite: SQLiteConfig{
				CheckRate: 30,
			},
		}
	})
}
