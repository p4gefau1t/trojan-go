package conf

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
)

func ConvertToIP(s string) (net.IP, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		ips, err := net.LookupIP(s)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, common.NewError("cannot resolve ip")
		}
		return ips[0], nil
	}
	return ip, nil
}

func ParseJSON(data []byte) (*GlobalConfig, error) {
	var config GlobalConfig
	err := json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	config.Hash = make(map[string]string)
	for _, password := range config.Passwords {
		config.Hash[common.SHA224String(password)] = password
	}
	switch config.RunType {
	case ClientRunType:
		if len(config.Passwords) == 0 {
			return nil, common.NewError("no password found")
		}
		serverCertBytes, err := ioutil.ReadFile(config.TLS.CertPath)
		if err != nil {
			return nil, err
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(serverCertBytes)
		config.TLS.CertPool = pool
	case ServerRunType:
		if len(config.Passwords) == 0 {
			return nil, common.NewError("no password found")
		}
		keyPair, err := tls.LoadX509KeyPair(config.TLS.CertPath, config.TLS.KeyPath)
		if err != nil {
			return nil, err
		}
		config.TLS.KeyPair = []tls.Certificate{keyPair}
	case ForwardRunType:
	default:
		return nil, common.NewError("invalid run type")
	}
	localIP, err := ConvertToIP(config.LocalHost)
	if err != nil {
		return nil, err
	}
	remoteIP, err := ConvertToIP(config.RemoteHost)
	if err != nil {
		return nil, err
	}
	config.LocalIP = localIP
	config.RemoteIP = remoteIP
	config.LocalAddr = &net.TCPAddr{
		IP:   localIP,
		Port: int(config.LocalPort),
	}
	config.RemoteAddr = &net.TCPAddr{
		IP:   remoteIP,
		Port: int(config.RemotePort),
	}
	return &config, nil
}
