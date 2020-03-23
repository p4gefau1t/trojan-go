package proxy

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"testing"

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
}

func TestServerWithJSON(t *testing.T) {
	data, err := ioutil.ReadFile("server.json")
	common.Must(err)
	config, err := conf.ParseJSON(data)
	common.Must(err)

	server := Server{
		config: config,
	}
	common.Must(server.Run())
}

func TestServerWithDatabase(t *testing.T) {
	key, err := tls.LoadX509KeyPair("server.crt", "server.key")
	common.Must(err)
	ip := net.IPv4(127, 0, 0, 1)
	port := 4445
	//password := "pass123123"
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
		Hash:       map[string]string{},
		MySQL: conf.MySQLConfig{
			Enabled:    true,
			ServerHost: "127.0.0.1",
			ServerPort: 3306,
			Username:   "root",
			Password:   "password",
			Database:   "trojan",
		},
	}
	config.TLS.KeyPair = []tls.Certificate{key}
	config.TLS.SNI = "localhost"

	server := Server{
		config: config,
	}
	common.Must(server.Run())
}

func TestPortReusingServer(t *testing.T) {
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
		TCP: conf.TCPConfig{
			ReusePort: true,
		},
	}
	config.TLS.KeyPair = []tls.Certificate{key}
	config.TLS.SNI = "localhost"

	server1 := Server{
		config: config,
	}
	server2 := Server{
		config: config,
	}
	go server1.Run()
	server2.Run()
	//common.Must(server2.Run())
	//time.Sleep(time.Hour)
}

func TestServerTCPRedirecting(t *testing.T) {
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
	addr, err := net.ResolveTCPAddr("tcp", "localhost:443")
	common.Must(err)
	config.TLS.FallbackAddr = addr

	server := Server{
		config: config,
	}
	server.Run()
}

func TestServerWithSQLite(t *testing.T) {
	key, err := tls.LoadX509KeyPair("server.crt", "server.key")
	common.Must(err)
	ip := net.IPv4(127, 0, 0, 1)
	port := 4445
	//password := "pass123123"
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
		Hash:       map[string]string{},
		SQLite: conf.SQLiteConfig{
			Enabled:  true,
			Database: "test.db",
		},
	}
	config.TLS.KeyPair = []tls.Certificate{key}
	config.TLS.SNI = "localhost"

	server := Server{
		config: config,
	}
	common.Must(server.Run())
}
