package proxy

import "github.com/p4gefau1t/trojan-go/config"

type Config struct {
	RunType         string `json:"run_type" yaml:"run-type"`
	LogLevel        int    `json:"log_level" yaml:"log-level"`
	LogFile         string `json:"log_file" yaml:"log-file"`
	RelayBufferSize int    `json:"relay_buffer_size" yaml:"relay_buffer_size"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			LogLevel:        1,
			RelayBufferSize: 4 * 1024,
		}
	})
}
