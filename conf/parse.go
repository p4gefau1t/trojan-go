package conf

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
)

func ConvertToIP(s string) ([]net.IP, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		ips, err := net.LookupIP(s)
		if err != nil {
			return nil, err
		}
		if len(ips) == 0 {
			return nil, common.NewError("cannot resolve host:" + s)
		}
		return ips, nil
	}
	return []net.IP{ip}, nil
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
	case ClientRunType, NATRunType:
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
	localIPs, err := ConvertToIP(config.LocalHost)
	if err != nil {
		return nil, err
	}
	remoteIPs, err := ConvertToIP(config.RemoteHost)
	if err != nil {
		return nil, err
	}

	config.LocalIP = localIPs[0]
	config.RemoteIP = remoteIPs[0]

	if config.TCP.PreferIPV4 {
		for _, ip := range localIPs {
			if ip.To4() != nil {
				config.LocalIP = ip
				break
			}
		}
		for _, ip := range remoteIPs {
			if ip.To4() != nil {
				config.RemoteIP = ip
				break
			}
		}
	}
	config.LocalAddr = &net.TCPAddr{
		IP:   config.LocalIP,
		Port: int(config.LocalPort),
	}
	config.RemoteAddr = &net.TCPAddr{
		IP:   config.RemoteIP,
		Port: int(config.RemotePort),
	}
	return &config, nil
}
