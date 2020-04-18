package conf

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
)

func loadCommonConfig(config *GlobalConfig) error {
	//log level
	log.SetLogLevel(log.LogLevel(config.LogLevel))

	//password settings
	if len(config.Passwords) == 0 {
		if config.RunType == Client {
			return common.NewError("no password found")
		} else {
			log.Warn("password is not specified in config file")
		}
	}
	config.Hash = make(map[string]string)
	for _, password := range config.Passwords {
		config.Hash[common.SHA224String(password)] = password
	}

	//address settings
	config.LocalAddress = common.NewAddress(config.LocalHost, config.LocalPort, "tcp")
	config.RemoteAddress = common.NewAddress(config.RemoteHost, config.RemotePort, "tcp")
	config.TargetAddress = common.NewAddress(config.TargetHost, config.TargetPort, "tcp")

	if config.TLS.FallbackPort != 0 {
		config.TLS.FallbackAddress = common.NewAddress(config.RemoteHost, config.TLS.FallbackPort, "tcp")
	}

	//tls settings
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
				log.Warn("found invalid cipher name", specified)
				break
			}
		}
		if invalid && len(supportedSuites) >= 1 {
			log.Warn("cipher list contains invalid cipher name, ignored")
			log.Warn("here's a list of supported ciphers:")
			list := ""
			for _, c := range supportedSuites {
				list += c.Name + ":"
			}
			log.Warn(list[0 : len(list)-1])
			config.TLS.CipherSuites = nil
		}
	} else {
		config.TLS.CipherSuites = nil
	}

	//websocket settings
	if config.Websocket.Enabled {
		log.Info("websocket enabled")
		if config.Websocket.Path == "" {
			return common.NewError("websocket path is empty")
		}
		if config.Websocket.Path[0] != '/' {
			return common.NewError("websocket path must start with \"/\"")
		}
		if config.RunType == Client && config.Websocket.HostName == "" {
			log.Warn("Client websocket host_name is unspecified, using remote_addr \"", config.RemoteHost, "\" as host_name")
			config.Websocket.HostName = config.RemoteHost
			if ip := net.ParseIP(config.RemoteHost); ip != nil && ip.To4() == nil { //ipv6 address
				config.Websocket.HostName = "[" + config.RemoteHost + "]"
			}
		}
	}
	return nil
}

func loadClientConfig(config *GlobalConfig) error {
	var err error

	//router settings
	config.Router.BlockList = []byte{}
	config.Router.ProxyList = []byte{}
	config.Router.BypassList = []byte{}

	for _, s := range config.Router.Block {
		if strings.HasPrefix(s, "geoip:") {
			config.Router.BlockIPCode = append(config.Router.BlockIPCode, s[len("geoip:"):len(s)])
			continue
		}
		if strings.HasPrefix(s, "geosite:") {
			config.Router.BlockSiteCode = append(config.Router.BlockSiteCode, s[len("geosite:"):len(s)])
			continue
		}
		data, err := ioutil.ReadFile(s)
		if err != nil {
			return err
		}
		config.Router.BlockList = append(config.Router.BlockList, data...)
		config.Router.BlockList = append(config.Router.BlockList, byte('\n'))
	}

	for _, s := range config.Router.Bypass {
		if strings.HasPrefix(s, "geoip:") {
			config.Router.BypassIPCode = append(config.Router.BypassIPCode, s[len("geoip:"):len(s)])
			continue
		}
		if strings.HasPrefix(s, "geosite:") {
			config.Router.BypassSiteCode = append(config.Router.BypassSiteCode, s[len("geosite:"):len(s)])
			continue
		}
		data, err := ioutil.ReadFile(s)
		if err != nil {
			return err
		}
		config.Router.BypassList = append(config.Router.BypassList, data...)
		config.Router.BypassList = append(config.Router.BypassList, byte('\n'))
	}

	for _, s := range config.Router.Proxy {
		if strings.HasPrefix(s, "geoip:") {
			config.Router.ProxyIPCode = append(config.Router.ProxyIPCode, s[len("geoip:"):len(s)])
			continue
		}
		if strings.HasPrefix(s, "geosite:") {
			config.Router.ProxySiteCode = append(config.Router.ProxySiteCode, s[len("geosite:"):len(s)])
			continue
		}
		data, err := ioutil.ReadFile(s)
		if err != nil {
			return err
		}
		config.Router.ProxyList = append(config.Router.ProxyList, data...)
		config.Router.ProxyList = append(config.Router.ProxyList, byte('\n'))
	}

	config.Router.GeoIP, err = ioutil.ReadFile(config.Router.GeoIPFilename)
	if err != nil {
		config.Router.GeoIP = []byte{}
		log.Warn(err)
	}
	config.Router.GeoSite, err = ioutil.ReadFile(config.Router.GeoSiteFilename)
	if err != nil {
		config.Router.GeoSite = []byte{}
		log.Warn(err)
	}

	//tls settings
	if config.TLS.SNI == "" {
		log.Warn("SNI is unspecified, using remote_addr as SNI")
		config.TLS.SNI = config.RemoteHost
	}
	if config.TLS.CertPath == "" {
		log.Info("cert of the remote server is not specified, using default CA list")
		return nil
	}

	caCertByte, err := ioutil.ReadFile(config.TLS.CertPath)
	if err != nil {
		return common.NewError("failed to load cert file").Base(err)
	}
	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(caCertByte)
	if !ok {
		log.Warn("invalid CA cert list")
	}
	log.Info("using custom CA list")
	pemCerts := caCertByte
	for len(pemCerts) > 0 {
		config.TLS.CertPool = pool
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}
		log.Debug("issuer:", cert.Issuer, ", subject:", cert.Subject)
	}

	return nil
}

func loadServerConfig(config *GlobalConfig) error {
	if config.TLS.KeyPassword != "" {
		keyFile, err := ioutil.ReadFile(config.TLS.KeyPath)
		if err != nil {
			return common.NewError("failed to load key file").Base(err)
		}
		keyBlock, _ := pem.Decode(keyFile)
		if keyBlock == nil {
			return common.NewError("failed to decode key file").Base(err)
		}
		decryptedKey, err := x509.DecryptPEMBlock(keyBlock, []byte(config.TLS.KeyPassword))
		if err == nil {
			return common.NewError("failed to decrypt key").Base(err)
		}

		certFile, err := ioutil.ReadFile(config.TLS.CertPath)
		certBlock, _ := pem.Decode(certFile)
		if certBlock == nil {
			return common.NewError("failed to decode cert file").Base(err)
		}

		keyPair, err := tls.X509KeyPair(certBlock.Bytes, decryptedKey)
		if err != nil {
			return err
		}

		config.TLS.KeyPair = []tls.Certificate{keyPair}
	} else {
		keyPair, err := tls.LoadX509KeyPair(config.TLS.CertPath, config.TLS.KeyPath)
		if err != nil {
			return common.NewError("failed to load key pair").Base(err)
		}
		config.TLS.KeyPair = []tls.Certificate{keyPair}
	}
	if config.TLS.HTTPFile != "" {
		payload, err := ioutil.ReadFile(config.TLS.HTTPFile)
		if err != nil {
			log.Warn("failed to load http response file", err)
		}
		config.TLS.HTTPResponse = payload
	}
	return nil
}

func ParseJSON(data []byte) (*GlobalConfig, error) {
	config := &GlobalConfig{}

	//default settings
	config.TLS.Verify = true
	config.TLS.VerifyHostname = true
	config.TLS.SessionTicket = true
	config.Mux.IdleTimeout = 60
	config.Mux.Concurrency = 8
	config.MySQL.CheckRate = 60
	config.Router.DefaultPolicy = "proxy"
	config.Router.GeoIPFilename = common.GetProgramDir() + "/geoip.dat"
	config.Router.GeoSiteFilename = common.GetProgramDir() + "/geosite.dat"
	config.Websocket.DoubleTLS = true

	err := json.Unmarshal(data, config)
	if err != nil {
		return nil, err
	}

	if err := loadCommonConfig(config); err != nil {
		return nil, err
	}

	switch config.RunType {
	case Client, NAT, Forward:
		if err := loadClientConfig(config); err != nil {
			return nil, err
		}
	case Server:
		if err := loadServerConfig(config); err != nil {
			return nil, err
		}
	case Relay:
	default:
		return nil, common.NewError("invalid run type:" + string(config.RunType))
	}

	return config, nil
}
