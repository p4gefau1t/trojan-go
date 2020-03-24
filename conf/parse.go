package conf

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
)

var logger = log.New(os.Stdout)

func convertToAddr(preferV4 bool, host string, port uint16) (*net.TCPAddr, error) {
	ip := net.ParseIP(host)
	if ip != nil {
		return &net.TCPAddr{
			IP:   ip,
			Port: int(port),
		}, nil
	}
	if preferV4 {
		return net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", host, port))
	}
	return net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
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

	log.LogLevel = config.LogLevel

	config.Hash = make(map[string]string)
	for _, password := range config.Passwords {
		config.Hash[common.SHA224String(password)] = password
	}
	switch config.RunType {
	case Client, NAT:
		if len(config.Passwords) == 0 {
			return nil, common.NewError("no password found")
		}
		if config.TLS.CertPath == "" {
			logger.Warn("cert of the remote server is not specified. using default CA list.")
			break
		}
		serverCertBytes, err := ioutil.ReadFile(config.TLS.CertPath)
		if err != nil {
			return nil, common.NewError("failed to load cert file").Base(err)
		}
		pool := x509.NewCertPool()
		pool.AppendCertsFromPEM(serverCertBytes)
		config.TLS.CertPool = pool
	case Server:
		if len(config.Passwords) == 0 {
			return nil, common.NewError("no password found")
		}
		if config.TLS.KeyPassword != "" {
			keyFile, err := ioutil.ReadFile(config.TLS.KeyPath)
			if err != nil {
				return nil, common.NewError("failed to load key file").Base(err)
			}
			keyBlock, _ := pem.Decode(keyFile)
			if keyBlock == nil {
				return nil, common.NewError("failed to decode key file").Base(err)
			}
			decryptedKey, err := x509.DecryptPEMBlock(keyBlock, []byte(config.TLS.KeyPassword))
			if err == nil {
				return nil, common.NewError("failed to decrypt key").Base(err)
			}

			certFile, err := ioutil.ReadFile(config.TLS.CertPath)
			certBlock, _ := pem.Decode(certFile)
			if certBlock == nil {
				return nil, common.NewError("failed to decode cert file").Base(err)
			}

			keyPair, err := tls.X509KeyPair(certBlock.Bytes, decryptedKey)
			if err != nil {
				return nil, err
			}

			config.TLS.KeyPair = []tls.Certificate{keyPair}
		} else {
			keyPair, err := tls.LoadX509KeyPair(config.TLS.CertPath, config.TLS.KeyPath)
			if err != nil {
				return nil, common.NewError("failed to load key pair").Base(err)
			}
			config.TLS.KeyPair = []tls.Certificate{keyPair}
		}
	case Forward:
	default:
		return nil, common.NewError("invalid run type")
	}

	localAddr, err := convertToAddr(config.TCP.PreferIPV4, config.LocalHost, config.LocalPort)
	if err != nil {
		return nil, common.NewError("invalid local address").Base(err)
	}
	config.LocalAddr = localAddr
	config.LocalIP = localAddr.IP

	remoteAddr, err := convertToAddr(config.TCP.PreferIPV4, config.RemoteHost, config.RemotePort)
	if err != nil {
		return nil, common.NewError("invalid remote address").Base(err)
	}
	config.RemoteAddr = remoteAddr
	config.RemoteIP = remoteAddr.IP

	if len(config.TLS.ALPN) != 0 || config.TLS.ALPHPortOverride != 0 {
		if config.TLS.ALPHPortOverride == 0 {
			logger.Warn("alpn port override is unspecified. using remote port")
			config.TLS.ALPHPortOverride = config.RemotePort
		}
		fallbackAddr, err := convertToAddr(config.TCP.PreferIPV4, config.RemoteHost, config.TLS.ALPHPortOverride)
		if err != nil {
			return nil, common.NewError("invalid tls fallback address").Base(err)
		}
		config.TLS.FallbackAddr = fallbackAddr
		for _, s := range config.TLS.ALPN {
			if strings.Contains(s, "http") || strings.Contains(s, "HTTP") {
				config.TLS.FallbackHTTP = true
			}
			if s == "h2" {
				config.TLS.FallbackHTTP2 = true
			}
		}
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
