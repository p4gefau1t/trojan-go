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
	"os/exec"
	"strconv"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"golang.org/x/crypto/pbkdf2"
)

func setKeyLogger(tlsConfig *TLSConfig) error {
	if tlsConfig.KeyLogPath != "" {
		log.Warn("TLS key logging activated. USE OF KEY LOGGING COMPROMISES SECURITY. IT SHOULD ONLY BE USED FOR DEBUGGING.")
		file, err := os.OpenFile(tlsConfig.KeyLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return common.NewError("Failed to open key log file").Base(err)
		}
		tlsConfig.KeyLogger = file
	}
	return nil
}

func loadCert(tlsConfig *TLSConfig) error {
	err := setKeyLogger(tlsConfig)
	if err != nil {
		return err
	}
	if tlsConfig.CertPath == "" {
		log.Info("Cert of the remote server is unspecified. Using default CA list")
	} else {
		caCertByte, err := ioutil.ReadFile(tlsConfig.CertPath)
		if err != nil {
			return common.NewError("failed to load cert file").Base(err)
		}
		tlsConfig.CertPool = x509.NewCertPool()
		ok := tlsConfig.CertPool.AppendCertsFromPEM(caCertByte)
		if !ok {
			log.Warn("Invalid CA cert list")
		}
		log.Info("Using custom CA list")

		//show info abount the cert
		pemCerts := caCertByte
		for len(pemCerts) > 0 {
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
			log.Trace("Issuer:", cert.Issuer, "Subject:", cert.Subject)
		}
	}
	return nil
}

func loadCertAndKey(tlsConfig *TLSConfig) error {
	err := setKeyLogger(tlsConfig)
	if err != nil {
		return err
	}
	if tlsConfig.KeyPassword != "" {
		keyFile, err := ioutil.ReadFile(tlsConfig.KeyPath)
		if err != nil {
			return common.NewError("Failed to load key file").Base(err)
		}
		keyBlock, _ := pem.Decode(keyFile)
		if keyBlock == nil {
			return common.NewError("Failed to decode key file").Base(err)
		}
		decryptedKey, err := x509.DecryptPEMBlock(keyBlock, []byte(tlsConfig.KeyPassword))
		if err == nil {
			return common.NewError("Failed to decrypt key").Base(err)
		}

		certFile, err := ioutil.ReadFile(tlsConfig.CertPath)
		certBlock, _ := pem.Decode(certFile)
		if certBlock == nil {
			return common.NewError("Failed to decode cert file").Base(err)
		}

		keyPair, err := tls.X509KeyPair(certBlock.Bytes, decryptedKey)
		if err != nil {
			return err
		}

		tlsConfig.KeyPair = []tls.Certificate{keyPair}
	} else {
		keyPair, err := tls.LoadX509KeyPair(tlsConfig.CertPath, tlsConfig.KeyPath)
		if err != nil {
			return common.NewError("Failed to load key pair").Base(err)
		}
		tlsConfig.KeyPair = []tls.Certificate{keyPair}
	}

	tlsConfig.ClientCertPool = x509.NewCertPool()
	for _, path := range tlsConfig.ClientCertPath {
		log.Debug("Loading client cert: " + path)
		certBytes, err := ioutil.ReadFile(path)
		if err != nil {
			return common.NewError("Failed to load cert file").Base(err)
		}
		ok := tlsConfig.ClientCertPool.AppendCertsFromPEM(certBytes)
		if !ok {
			return common.NewError("Invalid client cert")
		}
	}
	return nil
}

func loadCommonConfig(config *GlobalConfig) error {
	//log settigns
	log.SetLogLevel(log.LogLevel(config.LogLevel))
	if config.LogFile != "" {
		log.Info("Log will be written to", config.LogFile)
		file, err := os.OpenFile(config.LogFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return common.NewError("Failed to access the log file").Base(err)
		}
		log.SetOutput(file)
	}

	//buffer size, 4KiB - 16MiB
	if config.BufferSize < 4 || config.BufferSize > 16384 {
		return common.NewError("Invalid buffer size, 4 KiB < buffer_size < 16384 KiB")
	}

	config.BufferSize *= 1024

	//password settings
	if len(config.Passwords) == 0 {
		switch config.RunType {
		case Client, NAT, Forward:
			return common.NewError("No password found")
		default:
			log.Warn("Password is unspecified in config file")
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
		if config.TLS.FallbackHost == "" {
			config.TLS.FallbackAddress = common.NewAddress(config.RemoteHost, config.TLS.FallbackPort, "tcp")
		} else {
			config.TLS.FallbackAddress = common.NewAddress(config.TLS.FallbackHost, config.TLS.FallbackPort, "tcp")
		}
	}

	//api settings
	if config.API.Enabled {
		config.API.APIAddress = common.NewAddress(config.API.APIHost, config.API.APIPort, "tcp")
		if config.API.APITLS {
			if err := loadCertAndKey(&config.API.TLS); err != nil {
				return err
			}
		}
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
				log.Warn("Found invalid cipher ", specified)
				break
			}
		}
		if invalid && len(supportedSuites) >= 1 {
			log.Warn("\"cipher_suite\" contains invalid cipher name, ignored")
			log.Warn("Here is a list of supported ciphers:")
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
		log.Info("Websocket enabled")
		if config.Websocket.Path == "" {
			return common.NewError("Websocket path is empty")
		}
		if config.Websocket.Path[0] != '/' {
			return common.NewError("Websocket path must start with \"/\"")
		}
		if config.Websocket.HostName == "" {
			log.Warn("Websocket hostname is unspecified. Using remote_addr \"", config.RemoteHost, "\" as hostname")
			config.Websocket.HostName = config.RemoteHost
			if ip := net.ParseIP(config.RemoteHost); ip != nil && ip.To4() == nil { //ipv6 address
				config.Websocket.HostName = "[" + config.RemoteHost + "]"
			}
		}
		if config.Websocket.ObfuscationPassword != "" {
			log.Info("Websocket obfuscation enabled")
			password := []byte(config.Websocket.ObfuscationPassword)
			//hardcoded salt
			salt := []byte{48, 149, 6, 18, 13, 193, 247, 116, 197, 135, 236, 175, 190, 209, 146, 48}
			config.Websocket.ObfuscationKey = pbkdf2.Key(password, salt, 32, aes.BlockSize, sha256.New)
		}
	}

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

	var err error
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
	return nil
}

func loadClientConfig(config *GlobalConfig) error {
	//forward proxy settings
	if config.ForwardProxy.Enabled {
		log.Info("Forward proxy enabled")
		config.ForwardProxy.ProxyAddress = common.NewAddress(config.ForwardProxy.ProxyHost, config.ForwardProxy.ProxyPort, "tcp")
		log.Debug("Forward proxy", config.ForwardProxy.ProxyAddress.String())
	}

	if config.TransportPlugin.Enabled {
		log.Warn("Trojan-Go will use transport plugin and work in plain text mode")
		switch config.TransportPlugin.Type {
		case "plaintext":
			// do nothing
		case "shadowsocks":
			pluginHost := "127.0.0.1"
			pluginPort := common.PickPort("tcp", pluginHost)
			config.TransportPlugin.Env = append(
				config.TransportPlugin.Env,
				"SS_LOCAL_HOST="+pluginHost,
				"SS_LOCAL_PORT="+strconv.FormatInt(int64(pluginPort), 10),
				"SS_REMOTE_HOST="+config.RemoteHost,
				"SS_REMOTE_PORT="+strconv.FormatInt(int64(config.RemotePort), 10),
				"SS_PLUGIN_OPTIONS="+config.TransportPlugin.PluginOption,
			)
			config.RemoteHost = pluginHost
			config.RemotePort = pluginPort
			config.RemoteAddress = common.NewAddress(config.RemoteHost, config.RemotePort, "tcp")
			log.Debug("New remote address", config.RemoteAddress.String())
			log.Debug("Plugin env", config.TransportPlugin.Env)

			cmd := exec.Command(config.TransportPlugin.Command, config.TransportPlugin.Arg...)
			cmd.Env = append(cmd.Env, config.TransportPlugin.Env...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			config.TransportPlugin.Cmd = cmd
		case "other":
			cmd := exec.Command(config.TransportPlugin.Command, config.TransportPlugin.Arg...)
			cmd.Env = append(cmd.Env, config.TransportPlugin.Env...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			config.TransportPlugin.Cmd = cmd
		default:
			return common.NewError("Invalid plugin type: " + config.TransportPlugin.Type)
		}
	} else {
		if config.TLS.SNI == "" {
			log.Warn("SNI is unspecified, using remote_addr as SNI")
			config.TLS.SNI = config.RemoteHost
		}
		if err := loadCert(&config.TLS); err != nil {
			return err
		}
		if config.Websocket.Enabled && config.Websocket.DoubleTLS {
			if config.Websocket.TLS.CertPath == "" {
				log.Warn("Empty double TLS settings, using default ssl settings")
				config.Websocket.TLS = config.TLS
			} else {
				if err := loadCert(&config.Websocket.TLS); err != nil {
					return err
				}
			}
		}
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

	// transport plugin settings
	if config.TransportPlugin.Enabled {
		log.Warn("Trojan-Go will use transport plugin and work in plain text mode")
		switch config.TransportPlugin.Type {
		case "shadowsocks":
			trojanHost := "127.0.0.1"
			trojanPort := common.PickPort("tcp", trojanHost)
			config.TransportPlugin.Env = append(
				config.TransportPlugin.Env,
				"SS_REMOTE_HOST="+config.LocalHost,
				"SS_REMOTE_PORT="+strconv.FormatInt(int64(config.LocalPort), 10),
				"SS_LOCAL_HOST="+trojanHost,
				"SS_LOCAL_PORT="+strconv.FormatInt(int64(trojanPort), 10),
				"SS_PLUGIN_OPTIONS="+config.TransportPlugin.PluginOption,
			)

			config.LocalHost = trojanHost
			config.LocalPort = trojanPort
			config.LocalAddress = common.NewAddress(config.LocalHost, config.LocalPort, "tcp")
			log.Debug("New local address", config.RemoteAddress.String())
			log.Debug("Plugin env", config.TransportPlugin.Env)

			cmd := exec.Command(config.TransportPlugin.Command, config.TransportPlugin.Arg...)
			cmd.Env = append(cmd.Env, config.TransportPlugin.Env...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			config.TransportPlugin.Cmd = cmd
		case "other":
			cmd := exec.Command(config.TransportPlugin.Command, config.TransportPlugin.Arg...)
			cmd.Env = append(cmd.Env, config.TransportPlugin.Env...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			config.TransportPlugin.Cmd = cmd
		case "plaintext":
			// do nothing
		default:
			return common.NewError("Invalid plugin type: " + config.TransportPlugin.Type)
		}
	} else {
		if err := loadCertAndKey(&config.TLS); err != nil {
			return err
		}
		// tls settings
		if config.TLS.SNI == "" {
			log.Warn("Empty SNI field. Server will not verify the SNI in client hello request")
			config.TLS.VerifyHostName = false
		}

		if config.TLS.HTTPResponseFileName != "" {
			payload, err := ioutil.ReadFile(config.TLS.HTTPResponseFileName)
			if err != nil {
				return common.NewError("Failed to load http response file").Base(err)
			}
			config.TLS.HTTPResponse = payload
		}

		if config.Websocket.DoubleTLS {
			if config.Websocket.TLS.CertPath == "" {
				log.Warn("Empty double TLS settings, using global TLS settings")
				config.Websocket.TLS = config.TLS
			}
			if err := loadCertAndKey(&config.Websocket.TLS); err != nil {
				return err
			}
		}
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
			Verify:        true,
			SessionTicket: true,
			ReuseSession:  true,
			ALPN: []string{
				"http/1.1",
			},
			Fingerprint: "firefox",
		},
		Mux: MuxConfig{
			IdleTimeout: 60,
			Concurrency: 8,
		},
		Websocket: WebsocketConfig{
			DoubleTLS: true,
			TLS: TLSConfig{
				Verify:         true,
				VerifyHostName: true,
				SessionTicket:  true,
				ReuseSession:   true,
			},
		},
		MySQL: MySQLConfig{
			CheckRate:  60,
			ServerHost: "localhost",
			ServerPort: 3306,
		},
		Router: RouterConfig{
			DefaultPolicy:   "proxy",
			DomainStrategy:  "as_is",
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
		return nil, common.NewError("Invalid run type:" + string(config.RunType))
	}

	return config, nil
}
