package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"net"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
)

func TestForward(t *testing.T) {
	serverCertBytes, err := ioutil.ReadFile("./server.crt")
	common.Must(err)
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(serverCertBytes)
	ip := net.IPv4(127, 0, 0, 1)
	forwardPort := 5000
	password := "pass123123"
	clientConfig := &conf.GlobalConfig{
		LocalAddr: &net.TCPAddr{
			IP:   ip,
			Port: 4444,
		},
		LocalIP:   ip,
		LocalPort: uint16(forwardPort),
		RemoteAddr: &net.TCPAddr{
			IP:   ip,
			Port: forwardPort,
		},
		Hash: map[string]string{common.SHA224String(password): password},
	}
	clientConfig.TLS.CertPool = pool
	clientConfig.TLS.SNI = "localhost"

	c := Client{
		config: clientConfig,
	}
	go c.Run()

	key, err := tls.LoadX509KeyPair("server.crt", "server.key")

	common.Must(err)
	serverPort := 4445
	serverConfig := &conf.GlobalConfig{
		LocalAddr: &net.TCPAddr{
			IP:   ip,
			Port: serverPort,
		},
		LocalIP:   ip,
		LocalPort: uint16(forwardPort),
		RemoteAddr: &net.TCPAddr{
			IP:   ip,
			Port: 80,
		},
		RemoteIP:   ip,
		RemotePort: 80,
		Hash:       map[string]string{common.SHA224String(password): password},
	}
	serverConfig.TLS.KeyPair = []tls.Certificate{key}
	serverConfig.TLS.SNI = "localhost"

	server := Server{
		config: serverConfig,
	}
	go server.Run()

	forwardConfig := &conf.GlobalConfig{
		LocalAddr: &net.TCPAddr{
			IP:   ip,
			Port: forwardPort,
		},
		LocalIP:   ip,
		LocalPort: uint16(forwardPort),
		RemoteAddr: &net.TCPAddr{
			IP:   ip,
			Port: serverPort,
		},
		RemoteIP:   ip,
		RemotePort: uint16(serverPort),
	}

	forward := Forward{
		config: forwardConfig,
	}
	forward.Run()

}
