package raw

type Config struct {
	LocalHost  string    `json:"local_addr"`
	LocalPort  int       `json:"local_port"`
	TargetHost string    `json:"target_addr"`
	TargetPort int       `json:"target_port"`
	RemoteHost string    `json:"remote_addr"`
	RemotePort int       `json:"remote_port"`
	DNS        []string  `json:"dns"`
	TCP        TCPConfig `json:"tcp"`
}

type TCPConfig struct {
	PreferIPV4   bool `json:"prefer_ipv4"`
	KeepAlive    bool `json:"keep_alive"`
	FastOpen     bool `json:"fast_open"`
	FastOpenQLen int  `json:"fast_open_qlen"`
	ReusePort    bool `json:"reuse_port"`
	NoDelay      bool `json:"no_delay"`
}
