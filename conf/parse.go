package conf

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/withmandala/go-log"
)

var logger = log.New(os.Stdout).WithColor()

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

	//default settings
	config.TLS.Verify = true
	config.TLS.VerifyHostname = true
	config.TLS.SessionTicket = true
	config.TCP.MuxIdleTimeout = 5

	err := json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}
	config.Hash = make(map[string]string)
	for _, password := range config.Passwords {
		config.Hash[common.SHA224String(password)] = password
	}
	switch config.RunType {
	case Client, NAT:
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
	case Server:
		if len(config.Passwords) == 0 {
			return nil, common.NewError("no password found")
		}
		keyPair, err := tls.LoadX509KeyPair(config.TLS.CertPath, config.TLS.KeyPath)
		if err != nil {
			return nil, err
		}
		config.TLS.KeyPair = []tls.Certificate{keyPair}
	case Forward:
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

	if config.TLS.Cipher != "" || config.TLS.CipherTLS13 != "" {
		specifiedSuites := strings.Split(config.TLS.Cipher+":"+config.TLS.CipherTLS13, ":")
		supportedSuites := tls.CipherSuites()
		invalid := false
		for _, specified := range specifiedSuites {
			found := false
			if specified == "" {
				continue
			}
			for _, supported := range supportedSuites {
				if supported.Name == specified {
					config.TLS.CipherSuites = append(config.TLS.CipherSuites, supported.ID)
					found = true
					break
				}
			}
			if !found {
				invalid = true
				logger.Warn("found invalid cipher name", specified)
				break
			}
		}
		if invalid && len(supportedSuites) >= 1 {
			logger.Warn("cipher list contains invalid cipher name, ignored")
			logger.Warn("here's a list of supported ciphers:")
			list := ""
			for _, c := range supportedSuites {
				list += c.Name + ":"
			}
			logger.Warn(list[0 : len(list)-1])
			config.TLS.CipherSuites = nil
		}
	}

	if config.TLS.HTTPFile != "" {
		payload, err := ioutil.ReadFile(config.TLS.HTTPFile)
		if err != nil {
			logger.Warn("failed to load http response file", err)
		}
		config.TLS.HTTPResponse = payload
	}

	return &config, nil
}
