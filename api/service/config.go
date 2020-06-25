package service

import "github.com/p4gefau1t/trojan-go/config"

const Name = "API_SERVICE"

type SSLConfig struct {
	Enabled        bool     `json,yaml:"enabled"`
	CertPath       string   `json:"cert" yaml:"cert"`
	KeyPath        string   `json:"key" yaml:"key"`
	VerifyClient   bool     `json:"verify_client" yaml:"verify-client"`
	ClientCertPath []string `json:"client_cert" yaml:"client-cert"`
}

type APIConfig struct {
	Enabled bool      `json:"enabled" yaml:"enabled"`
	APIHost string    `json:"api_addr" yaml:"api-addr"`
	APIPort int       `json:"api_port" yaml:"api-port"`
	SSL     SSLConfig `json,yaml:"ssl"`
}

type Config struct {
	API APIConfig `json,yaml:"api"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return new(Config)
	})
}
