package url

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/option"
	"github.com/p4gefau1t/trojan-go/proxy"
)

const Name = "URL"

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
	clientConfigFormat := `
{
	"run_type": "client",
	"local_addr": "%s",
	"local_port": %d,
	"remote_addr": "%s",
	"remote_port": %d,
	"password": [
		"%s"
	],
	"tls": {
		"sni": "%s"
	},
	"websocket": {
		"enabled": %t,
		"host": "%s",
		"path": "%s"
	},
	"shadowsocks": {
		"enabled": %t,
		"method": "%s",
		"password": "%s"
	},
	"mux": {
		"enabled": %t
	}
}`
	wsEnabled := false
	if info.Type == ShareInfoTypeWebSocket {
		wsEnabled = true
	}
	ssEnabled := false
	ssPassword := ""
	ssMethod := ""
	if strings.HasPrefix(info.Encryption, "ss;") {
		ssEnabled = true
		ssConfig := strings.Split(info.Encryption, ";")
		if len(ssConfig) != 3 {
			log.Fatalf("invalid shadowsocks config: %s", info.Encryption)
		}
		ssMethod = ssConfig[1]
		ssPassword = ssConfig[2]
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
	clientConfig := fmt.Sprintf(clientConfigFormat, listenHost, listenPort, info.TrojanHost, info.Port, info.TrojanPassword, info.SNI, wsEnabled, info.Host, info.Path, ssEnabled, ssMethod, ssPassword, muxEnabled)
	log.Debug(clientConfig)
	client, err := proxy.NewProxyFromConfigData([]byte(clientConfig), true)
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
