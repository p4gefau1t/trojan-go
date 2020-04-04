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
	Cipher             string `json:"cipher"`
	CipherTLS13        string `json:"cipher_tls13"`
	PreferServerCipher bool   `json:"prefer_server_cipher"`
	SNI                string `json:"sni"`
	HTTPFile           string `json:"plain_http_response"`
	FallbackPort       int    `json:"fallback_port"`

	FallbackAddr     net.Addr
	CertPool         *x509.CertPool
	KeyPair          []tls.Certificate
	HTTPResponse     []byte
	CipherSuites     []uint16
	CipherSuiteTLS13 []uint16
	ReuseSession     bool
	SessionTicket    bool
	Curves           string
}

type TCPConfig struct {
	PreferIPV4   bool `json:"prefer_ipv4"`
	KeepAlive    bool `json:"keep_alive"`
	FastOpen     bool `json:"fast_open"`
	FastOpenQLen int  `json:"fast_open_qlen"`
	ReusePort    bool `json:"reuse_port"`
	NoDelay      bool `json:"no_delay"`
}

type MuxConfig struct {
	Enabled     bool `json:"enabled"`
	IdleTimeout int  `json:"idle_timeout"`
	Concurrency int  `json:"concurrency"`
}

type MySQLConfig struct {
	Enabled    bool   `json:"enabled"`
	ServerHost string `json:"server_addr"`
	ServerPort int    `json:"server_port"`
	Database   string `json:"database"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	CheckRate  int    `json:"check_rate"`
}

type SQLiteConfig struct {
	Enabled  bool   `json:"enabled"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RouterConfig struct {
	Enabled             bool     `json:"enabled"`
	Bypass              []string `json:"bypass"`
	Proxy               []string `json:"proxy"`
	Block               []string `json:"block"`
	DefaultPolicy       string   `json:"default_policy"`
	RouteByIP           bool     `json:"route_by_ip"`
	RouteByIPOnNonmatch bool     `json:"route_by_ip_on_nonmatch"`

	BypassList []byte
	ProxyList  []byte
	BlockList  []byte

	GeoIP          []byte
	BypassIPCode   []string
	ProxyIPCode    []string
	BlockIPCode    []string
	GeoSite        []byte
	BypassSiteCode []string
	ProxySiteCode  []string
	BlockSiteCode  []string
}

type GlobalConfig struct {
	RunType    RunType      `json:"run_type"`
	LogLevel   int          `json:"log_level"`
	LocalHost  string       `json:"local_addr"`
	LocalPort  int          `json:"local_port"`
	RemoteHost string       `json:"remote_addr"`
	RemotePort int          `json:"remote_port"`
	Passwords  []string     `json:"password"`
	TLS        TLSConfig    `json:"ssl"`
	TCP        TCPConfig    `json:"tcp"`
	MySQL      MySQLConfig  `json:"mysql"`
	SQLite     SQLiteConfig `json:"sqlite"`
	Mux        MuxConfig    `json:"mux"`
	Router     RouterConfig `json:"router"`

	LocalAddr  net.Addr
	LocalIP    net.IP
	RemoteAddr net.Addr
	RemoteIP   net.IP
	Hash       map[string]string
}
