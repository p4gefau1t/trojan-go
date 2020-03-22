package conf

import (
	"crypto/tls"
	"crypto/x509"
	"net"
)

type RunType string

const (
	Client  RunType = "client"
	Server  RunType = "server"
	NAT     RunType = "nat"
	Forward RunType = "forward"
)

type TLSConfig struct {
	Verify             bool   `json:"verify"`
	VerifyHostname     bool   `json:"verify_hostname"`
	CertPath           string `json:"cert"`
	KeyPath            string `json:"key"`
	KeyPassword        string `json:"key_password"`
	CertPool           *x509.CertPool
	KeyPair            []tls.Certificate
	Cipher             string `json:"cipher"`
	CipherTLS13        string `json:"cipher_tls13"`
	HTTPFile           string `json:"plain_http_response"`
	HTTPResponse       []byte
	CipherSuites       []uint16
	CipherSuiteTLS13   []uint16
	PreferServerCipher bool   `json:"prefer_server_cipher"`
	SNI                string `json:"sni"`
	Alph               []string
	ReuseSession       bool
	SessionTicket      bool
	Curves             string
}

type TCPConfig struct {
	PreferIPV4     bool `json:"prefer_ipv4"`
	KeepAlive      bool `json:"keep_alive"`
	FastOpen       bool `json:"fast_open"`
	FastOpenQLen   int  `json:"fast_open_qlen"`
	ReusePort      bool `json:"reuse_port"`
	NoDelay        bool `json:"no_delay"`
	Mux            bool `json:"mux"`
	MuxIdleTimeout int  `json:"mux_idle_timeout"`
}

type MySQLConfig struct {
	Enabled    bool   `json:"enabled"`
	ServerHost string `json:"server_addr"`
	ServerPort int    `json:"server_port"`
	Database   string `json:"database"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

type SQLiteConfig struct {
	Enabled  bool   `json:"enabled"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type GlobalConfig struct {
	RunType RunType `json:"run_type"`

	LocalAddr net.Addr
	LocalHost string `json:"local_addr"`
	LocalIP   net.IP
	LocalPort uint16 `json:"local_port"`

	RemoteHost string `json:"remote_addr"`
	RemoteAddr net.Addr
	RemoteIP   net.IP
	RemotePort uint16 `json:"remote_port"`

	Hash      map[string]string
	Passwords []string     `json:"password"`
	LogLevel  int          `json:"log_level"`
	TLS       TLSConfig    `json:"ssl"`
	TCP       TCPConfig    `json:"tcp"`
	MySQL     MySQLConfig  `json:"mysql"`
	SQLite    SQLiteConfig `json:"mysql"`
}
