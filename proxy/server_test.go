package proxy

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
)

func TestServer(t *testing.T) {
	key, err := tls.LoadX509KeyPair("server.crt", "server.key")
	common.Must(err)
	ip := net.IPv4(127, 0, 0, 1)
	port := 4445
	password := "pass123123"
	config := &conf.GlobalConfig{
		LocalAddr: &net.TCPAddr{
			IP:   ip,
			Port: port,
		},
		LocalIP:   ip,
		LocalPort: uint16(port),
		RemoteAddr: &net.TCPAddr{
			IP:   ip,
			Port: 80,
		},
		RemoteIP:   ip,
		RemotePort: 80,
		Hash:       map[string]string{common.SHA224String(password): password},
	}
	config.TLS.KeyPair = []tls.Certificate{key}
	config.TLS.SNI = "localhost"

	server := Server{
		config: config,
	}
	server.Run()
	time.Sleep(time.Hour)
}

func TestServerWithJSON(t *testing.T) {
	data, err := ioutil.ReadFile("server.json")
	common.Must(err)
	config, err := conf.ParseJSON(data)
	common.Must(err)

	server := Server{
		config: config,
	}
	server.Run()
	time.Sleep(time.Hour)
}
