package url

import (
	"encoding/json"
	"flag"
	"net"
	"strconv"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/option"
	"github.com/p4gefau1t/trojan-go/proxy"
)

const Name = "URL"

type Websocket struct {
	Enabled bool
	Host    string
	Path    string
}

type TLS struct {
	SNI string
}

type Shadowsocks struct {
	Enabled  bool
	Method   string
	Password string
}

type Mux struct {
	Enabled bool
}

type UrlConfig struct {
	RunType     string   `json:"run_type"`
	LocalAddr   string   `json:"local_addr"`
	LocalPort   int      `json:"local_port"`
	RemoteAddr  string   `json:"remote_addr"`
	RemotePort  int      `json:"remote_port"`
	Password    []string `json:"password"`
	Websocket   `json:"websocket"`
	Shadowsocks `json:"shadowsocks"`
	TLS         `json:"ssl"`
	Mux         `json:"mux"`
}

type url struct {
	url    *string
	option *string
}

func (u *url) Name() string {
	return Name
}

func (u *url) Handle() error {
	if u.url == nil || *u.url == "" {
		return common.NewError("")
	}
	info, err := NewShareInfoFromURL(*u.url)
	if err != nil {
		log.Fatal(err)
	}
	wsEnabled := false
	if info.Type == ShareInfoTypeWebSocket {
		wsEnabled = true
	}
	ssEnabled := false
	ssPassword := ""
	ssMethod := ""
	if strings.HasPrefix(info.Encryption, "ss;") {
		ssEnabled = true
		ssConfig := strings.Split(info.Encryption[3:], ":")
		if len(ssConfig) != 2 {
			log.Fatalf("invalid shadowsocks config: %s", info.Encryption)
		}
		ssMethod = ssConfig[0]
		ssPassword = ssConfig[1]
	}
	muxEnabled := false
	listenHost := "127.0.0.1"
	listenPort := 1080
	options := strings.Split(*u.option, ";")
	for _, o := range options {
		key := ""
		val := ""
		l := strings.Split(o, "=")
		if len(l) != 2 {
			log.Fatal("option format error, no \"key=value\" pair found:", o)
		}
		key = l[0]
		val = l[1]
		switch key {
		case "mux":
			muxEnabled, err = strconv.ParseBool(val)
			if err != nil {
				log.Fatal(err)
			}
		case "listen":
			h, p, err := net.SplitHostPort(val)
			if err != nil {
				log.Fatal(err)
			}
			listenHost = h
			lp, err := strconv.ParseUint(p, 10, 16)
			if err != nil {
				log.Fatal(err)
			}
			listenPort = int(lp)
		default:
			log.Fatal("invalid option", o)
		}
	}
	config := UrlConfig{
		RunType:    "client",
		LocalAddr:  listenHost,
		LocalPort:  listenPort,
		RemoteAddr: info.TrojanHost,
		RemotePort: int(info.Port),
		Password:   []string{info.TrojanPassword},
		TLS: TLS{
			SNI: info.SNI,
		},
		Websocket: Websocket{
			Enabled: wsEnabled,
			Path:    info.Path,
			Host:    info.Host,
		},
		Mux: Mux{
			Enabled: muxEnabled,
		},
		Shadowsocks: Shadowsocks{
			Enabled:  ssEnabled,
			Password: ssPassword,
			Method:   ssMethod,
		},
	}
	data, err := json.Marshal(&config)
	if err != nil {
		log.Fatal(err)
	}
	log.Debug(string(data))
	client, err := proxy.NewProxyFromConfigData([]byte(data), true)
	if err != nil {
		log.Fatal(err)
	}
	return client.Run()
}

func (u *url) Priority() int {
	return 10
}

func init() {
	option.RegisterHandler(&url{
		url:    flag.String("url", "", "Setup trojan-go client with a url link"),
		option: flag.String("url-option", "mux=true;listen=127.0.0.1:1080", "URL mode options"),
	})
}
