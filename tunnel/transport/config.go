package transport

import (
	"github.com/p4gefau1t/trojan-go/config"
)

type Config struct {
	LocalHost       string                `json:"local_addr" yaml:"local-addr"`
	LocalPort       int                   `json:"local_port" yaml:"local-port"`
	RemoteHost      string                `json:"remote_addr" yaml:"remote-addr"`
	RemotePort      int                   `json:"remote_port" yaml:"remote-port"`
	TLS             TLSConfig             `json:"ssl" yaml:"ssl"`
	TransportPlugin TransportPluginConfig `json:"transport_plugin" yaml:"transport-plugin"`
	Websocket       WebsocketConfig       `json:"websocket" yaml:"websocket"`
}

type WebsocketConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"'`
}

type TLSConfig struct {
	Verify               bool     `json:"verify" yaml:"verify"`
	VerifyHostName       bool     `json:"verify_hostname" yaml:"verify-hostname"`
	CertPath             string   `json:"cert" yaml:"cert"`
	KeyPath              string   `json:"key" yaml:"key"`
	KeyPassword          string   `json:"key_password" yaml:"key-password"`
	Cipher               string   `json:"cipher" yaml:"cipher"`
	PreferServerCipher   bool     `json:"prefer_server_cipher" yaml:"prefer-server-cipher"`
	SNI                  string   `json:"sni" yaml:"sni"`
	HTTPResponseFileName string   `json:"plain_http_response" yaml:"plain-http-response"`
	FallbackHost         string   `json:"fallback_addr" yaml:"fallback-addr"`
	FallbackPort         int      `json:"fallback_port" yaml:"fallback-port"`
	ReuseSession         bool     `json:"reuse_session" yaml:"reuse-session"`
	ALPN                 []string `json:"alpn" yaml:"alpn"`
	Curves               string   `json:"curves" yaml:"curves"`
	Fingerprint          string   `json:"fingerprint" yaml:"fingerprint"`
	KeyLogPath           string   `json:"key_log" yaml:"key-log"`
	KeyBytes             []byte   `json:"key_bytes" yaml:"key-bytes"`
	CertBytes            []byte   `json:"cert_bytes" yaml:"cert-bytes"`
}

type TransportPluginConfig struct {
	Enabled      bool     `json:"enabled" yaml:"enabled"`
	Type         string   `json:"type" yaml:"type"`
	Command      string   `json:"command" yaml:"command"`
	PluginOption string   `json:"plugin_option" yaml:"plugin-option"`
	Arg          []string `json:"arg" yaml:"arg"`
	Env          []string `json:"env" yaml:"env"`
}

func init() {
	config.RegisterConfigCreator(Name, func() interface{} {
		return &Config{
			TLS: TLSConfig{
				Verify:         true,
				VerifyHostName: true,
				Fingerprint:    "firefox",
				ALPN:           []string{"http/1.1"},
			},
		}
	})
}
