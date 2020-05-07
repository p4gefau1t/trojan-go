package conf

import (
	"crypto/aes"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"golang.org/x/crypto/pbkdf2"
)

func loadCommonConfig(config *GlobalConfig) error {
	//log settigns
	log.SetLogLevel(log.LogLevel(config.LogLevel))
	if config.LogFile != "" {
		log.Info("log will be written into", config.LogFile)
		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return common.NewError("failed to access log file").Base(err)
		}
		log.SetOutput(file)
	}

	//buffer size, 4KiB to 16MiB
	if config.BufferSize < 4 || config.BufferSize > 16384 {
		return common.NewError("invalid buffer size, 4 KiB < buffer_size < 16384 Kib")
	}

	config.BufferSize *= 1024

	//password settings
	if len(config.Passwords) == 0 {
		if config.RunType == Client {
			return common.NewError("no password found")
		}
		log.Warn("password is not specified in config file")
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
		if config.TLS.FallbackHost == "" {
			config.TLS.FallbackAddress = common.NewAddress(config.RemoteHost, config.TLS.FallbackPort, "tcp")
		} else {
			config.TLS.FallbackAddress = common.NewAddress(config.TLS.FallbackHost, config.TLS.FallbackPort, "tcp")
		}
	}

	//api settings
	if config.API.Enabled {
		config.API.APIAddress = common.NewAddress(config.API.APIHost, config.API.APIPort, "tcp")
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
			log.Warn(list[:len(list)-1])
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
		if config.Websocket.HostName == "" {
			log.Warn("websocket hostname is unspecified, using remote_addr \"", config.RemoteHost, "\" as hostname")
			config.Websocket.HostName = config.RemoteHost
			if ip := net.ParseIP(config.RemoteHost); ip != nil && ip.To4() == nil { //ipv6 address
				config.Websocket.HostName = "[" + config.RemoteHost + "]"
			}
		}
		if config.Websocket.ObfuscationPassword != "" {
			log.Info("websocket obfs enabled")
			password := []byte(config.Websocket.ObfuscationPassword)
			salt := []byte{48, 149, 6, 18, 13, 193, 247, 116, 197, 135, 236, 175, 190, 209, 146, 48}
			config.Websocket.ObfuscationKey = pbkdf2.Key(password, salt, 32, aes.BlockSize, sha256.New)
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
			config.Router.BlockIPCode = append(config.Router.BlockIPCode, s[len("geoip:"):])
			continue
		}
		if strings.HasPrefix(s, "geosite:") {
			config.Router.BlockSiteCode = append(config.Router.BlockSiteCode, s[len("geosite:"):])
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
			config.Router.BypassIPCode = append(config.Router.BypassIPCode, s[len("geoip:"):])
			continue
		}
		if strings.HasPrefix(s, "geosite:") {
			config.Router.BypassSiteCode = append(config.Router.BypassSiteCode, s[len("geosite:"):])
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
			config.Router.ProxyIPCode = append(config.Router.ProxyIPCode, s[len("geoip:"):])
			continue
		}
		if strings.HasPrefix(s, "geosite:") {
			config.Router.ProxySiteCode = append(config.Router.ProxySiteCode, s[len("geosite:"):])
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

	//forward proxy settings
	if config.ForwardProxy.Enabled {
		log.Info("forward proxy enabled")
		config.ForwardProxy.ProxyAddress = common.NewAddress(config.ForwardProxy.ProxyHost, config.ForwardProxy.ProxyPort, "tcp")
	}

	return nil
}

func loadServerConfig(config *GlobalConfig) error {
	//check web server
	if !config.DisableHTTPCheck {
		resp, err := http.Get("http://" + config.RemoteAddress.String())
		if err != nil {
			return common.NewError(config.RemoteAddress.String() + " is not a valid web server").Base(err)
		}
		buf := [128]byte{}
		_, err = resp.Body.Read(buf[:])
		log.Debug("body:\n" + string(buf[:]))
		resp.Body.Close()
	}

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
	//default settings
	config := &GlobalConfig{
		LogLevel:   1,
		BufferSize: 32,
		TCP: TCPConfig{
			FastOpenQLen: 20,
			NoDelay:      true,
			KeepAlive:    true,
		},
		TLS: TLSConfig{
			Verify:         true,
			VerifyHostname: true,
			SessionTicket:  true,
			ReuseSession:   true,
		},
		Mux: MuxConfig{
			IdleTimeout: 60,
			Concurrency: 8,
		},
		Websocket: WebsocketConfig{
			DoubleTLS:       true,
			DoubleTLSVerify: true,
		},
		MySQL: MySQLConfig{
			CheckRate:  60,
			ServerHost: "localhost",
			ServerPort: 3306,
		},
		Router: RouterConfig{
			DefaultPolicy:   "proxy",
			GeoIPFilename:   common.GetProgramDir() + "/geoip.dat",
			GeoSiteFilename: common.GetProgramDir() + "/geosite.dat",
		},
		Redis: RedisConfig{
			ServerHost: "localhost",
			ServerPort: 6379,
		},
	}

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
