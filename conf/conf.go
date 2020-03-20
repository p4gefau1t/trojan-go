package conf

import (
	"crypto/tls"
	"crypto/x509"
	"net"
)

type RunType string

const (
	ClientRunType  RunType = "client"
	ServerRunType  RunType = "server"
	NATRunType     RunType = "nat"
	ForwardRunType RunType = "forward"
)

type TLSConfig struct {
	Verify           bool   `json:"verify"`
	VerifyHostname   bool   `json:"verify_hostname"`
	CertPath         string `json:"cert"`
	KeyPath          string `json:"key"`
	CertPool         *x509.CertPool
	KeyPair          []tls.Certificate
	CipherSuites     []uint16
	CipherSuiteTLS13 []uint16
	SNI              string `json:"sni"`
	Alph             []string
	ReuseSession     bool
	SessionTicket    bool
	Curves           string
}

type TCPConfig struct {
	PreferIPV4   bool `json:"prefer_ipv4"`
	KeepAlive    bool `json:"no_delay"`
	FastOpen     bool `json:"fast_open"`
	FastOpenQLen int  `json:"fast_open_qlen"`
	ReusePort    bool `json:"reuse_port"`
	Mux          bool `json:"mux"`
}

type MySQLConfig struct {
	Enabled    bool   `json:"enabled"`
	ServerHost string `json:"server_addr"`
	ServerPort int    `json:"server_port"`
	Database   string `json:"database"`
	Username   string `json:"username"`
	Password   string `json:"password"`
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
	Passwords []string `json:"password"`
	LogLevel  int
	TLS       TLSConfig   `json:"ssl"`
	TCP       TCPConfig   `json:"tcp"`
	MySQL     MySQLConfig `json:"mysql"`
}
