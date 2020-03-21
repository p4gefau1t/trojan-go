package proxy

import (
	"crypto/x509"
	"io/ioutil"
	"net"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
)

func TestClient(t *testing.T) {
	serverCertBytes, err := ioutil.ReadFile("./server.crt")
	common.Must(err)
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(serverCertBytes)
	ip := net.IPv4(127, 0, 0, 1)
	port := 4444
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
			Port: 4445,
		},
		Hash: map[string]string{common.SHA224String(password): password},
	}
	config.TLS.CertPool = pool
	config.TLS.SNI = "localhost"

	c := Client{
		config: config,
	}
	c.Run()
}

func TestMuxClient(t *testing.T) {
	serverCertBytes, err := ioutil.ReadFile("./server.crt")
	common.Must(err)
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(serverCertBytes)
	ip := net.IPv4(127, 0, 0, 1)
	port := 4444
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
			Port: 4445,
		},
		TCP: conf.TCPConfig{
			Mux: true,
		},
		Hash: map[string]string{common.SHA224String(password): password},
	}
	config.TCP.MuxIdleTimeout = 1
	config.TLS.CertPool = pool
	config.TLS.SNI = "localhost"

	c := Client{
		config: config,
	}
	c.Run()
}

func TestClientWithJSON(t *testing.T) {
	data, err := ioutil.ReadFile("client.json")
	common.Must(err)
	config, err := conf.ParseJSON(data)
	common.Must(err)

	client := Client{
		config: config,
	}
	client.Run()
}
