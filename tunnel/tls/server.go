package tls

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/caddyserver/certmagic"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/redirector"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/tls/fingerprint"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"github.com/p4gefau1t/trojan-go/tunnel/websocket"
)

// Server is a tls server
type Server struct {
	fallbackAddress    net.Addr
	verifySNI          bool
	sni                string
	alpn               []string
	PreferServerCipher bool
	keyPair            []tls.Certificate
	httpResp           []byte
	cipherSuite        []uint16
	sessionTicket      bool
	curve              []tls.CurveID
	keyLogger          io.WriteCloser
	connChan           chan tunnel.Conn
	wsChan             chan tunnel.Conn
	redir              *redirector.Redirector
	ctx                context.Context
	cancel             context.CancelFunc
	underlay           tunnel.Server
	nextHTTP           bool
	portOverrider      map[string]int
	tlsConfig          *tls.Config
	matchSNI           func(string) bool
}

func (s *Server) Close() error {
	s.cancel()
	if s.keyLogger != nil {
		s.keyLogger.Close()
	}
	return s.underlay.Close()
}

func isDomainNameMatched(pattern string, domainName string) bool {
	if strings.HasPrefix(pattern, "*.") {
		suffix := pattern[2:]
		domainPrefixLen := len(domainName) - len(suffix) - 1
		return strings.HasSuffix(domainName, suffix) && domainPrefixLen > 0 && !strings.Contains(domainName[:domainPrefixLen], ".")
	}
	return pattern == domainName
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.underlay.AcceptConn(&Tunnel{})
		if err != nil {
			select {
			case <-s.ctx.Done():
			default:
				log.Fatal(common.NewError("transport accept error"))
			}
			return
		}
		go func(conn net.Conn) {

			var tlsConfig *tls.Config
			if s.tlsConfig != nil {
				tlsConfig = s.tlsConfig
			} else {
				tlsConfig = &tls.Config{
					CipherSuites:             s.cipherSuite,
					PreferServerCipherSuites: s.PreferServerCipher,
					SessionTicketsDisabled:   !s.sessionTicket,
					NextProtos:               s.alpn,
					KeyLogWriter:             s.keyLogger,
					GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
						sni := s.keyPair[0].Leaf.Subject.CommonName
						dnsNames := s.keyPair[0].Leaf.DNSNames
						if s.sni != "" {
							sni = s.sni
						}
						matched := isDomainNameMatched(sni, hello.ServerName)
						for _, name := range dnsNames {
							if isDomainNameMatched(name, hello.ServerName) {
								matched = true
								break
							}
						}
						if s.verifySNI && !matched {
							return nil, common.NewError("sni mismatched: " + hello.ServerName + ", expected: " + s.sni)
						}
						return &s.keyPair[0], nil
					},
				}
			}

			// ------------------------ WAR ZONE ----------------------------

			handshakeRewindConn := common.NewRewindConn(conn)
			handshakeRewindConn.SetBufferSize(2048)

			tlsConn := tls.Server(handshakeRewindConn, tlsConfig)
			err = tlsConn.Handshake()
			handshakeRewindConn.StopBuffering()

			if err != nil {
				if strings.Contains(err.Error(), "first record does not look like a TLS handshake") {
					// not a valid tls client hello
					handshakeRewindConn.Rewind()
					log.Error(common.NewError("failed to perform tls handshake with " + tlsConn.RemoteAddr().String() + ", redirecting").Base(err))
					if s.fallbackAddress != nil {
						s.redir.Redirect(&redirector.Redirection{
							InboundConn: handshakeRewindConn,
							RedirectTo:  s.fallbackAddress,
						})
					} else if s.httpResp != nil {
						handshakeRewindConn.Write(s.httpResp)
						handshakeRewindConn.Close()
					} else {
						handshakeRewindConn.Close()
					}
				} else {
					// in other cases, simply close it
					tlsConn.Close()
					log.Error(common.NewError("tls handshake failed").Base(err))
				}
				return
			}

			log.Info("tls connection from", conn.RemoteAddr())
			state := tlsConn.ConnectionState()
			log.Trace("tls handshake", tls.CipherSuiteName(state.CipherSuite), state.DidResume, state.NegotiatedProtocol)

			// we use a real http header parser to mimic a real http server
			rewindConn := common.NewRewindConn(tlsConn)
			rewindConn.SetBufferSize(1024)
			r := bufio.NewReader(rewindConn)
			httpReq, err := http.ReadRequest(r)
			rewindConn.Rewind()
			rewindConn.StopBuffering()
			if err != nil && s.matchSNI(state.ServerName) {
				// this is not a http request. pass it to trojan protocol layer for further inspection
				s.connChan <- &transport.Conn{
					Conn: rewindConn,
				}
			} else {
				if !s.nextHTTP {
					// there is no websocket layer waiting for connections, redirect it
					log.Error("incoming http request, but no websocket server is listening")
					s.redir.Redirect(&redirector.Redirection{
						InboundConn: rewindConn,
						RedirectTo:  s.fallbackAddress,
					})
					return
				}
				// this is a http request, pass it to websocket protocol layer
				log.Debug("http req: ", httpReq)
				s.wsChan <- &transport.Conn{
					Conn: rewindConn,
				}
			}
		}(conn)
	}
}

func (s *Server) AcceptConn(overlay tunnel.Tunnel) (tunnel.Conn, error) {
	if _, ok := overlay.(*websocket.Tunnel); ok {
		s.nextHTTP = true
		log.Debug("next proto http")
		// websocket overlay
		select {
		case conn := <-s.wsChan:
			return conn, nil
		case <-s.ctx.Done():
			return nil, common.NewError("transport server closed")
		}
	}
	// trojan overlay
	select {
	case conn := <-s.connChan:
		return conn, nil
	case <-s.ctx.Done():
		return nil, common.NewError("transport server closed")
	}
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

func (s *Server) checkKeyPairLoop(checkRate time.Duration, keyPath string, certPath string, password string) {
	var lastKeyBytes, lastCertBytes []byte
	for {
		log.Debug("checking cert..")
		keyBytes, err := ioutil.ReadFile(keyPath)
		if err != nil {
			log.Error(common.NewError("tls failed to check key").Base(err))
			continue
		}
		certBytes, err := ioutil.ReadFile(certPath)
		if err != nil {
			log.Error(common.NewError("tls failed to check cert").Base(err))
			continue
		}
		if !bytes.Equal(keyBytes, lastKeyBytes) || !bytes.Equal(lastCertBytes, certBytes) {
			log.Info("new key pair detected")
			keyPair, err := loadKeyPair(keyPath, certPath, password)
			if err != nil {
				log.Error(common.NewError("tls failed to load new key pair").Base(err))
				continue
			}
			// TODO fix race
			s.keyPair = []tls.Certificate{*keyPair}
			lastKeyBytes = keyBytes
			lastCertBytes = certBytes
		}
		select {
		case <-time.After(checkRate):
			continue
		case <-s.ctx.Done():
			log.Debug("exiting")
			return
		}
	}
}

func loadKeyPair(keyPath string, certPath string, password string) (*tls.Certificate, error) {
	if password != "" {
		keyFile, err := ioutil.ReadFile(keyPath)
		if err != nil {
			return nil, common.NewError("failed to load key file").Base(err)
		}
		keyBlock, _ := pem.Decode(keyFile)
		if keyBlock == nil {
			return nil, common.NewError("failed to decode key file").Base(err)
		}
		decryptedKey, err := x509.DecryptPEMBlock(keyBlock, []byte(password))
		if err == nil {
			return nil, common.NewError("failed to decrypt key").Base(err)
		}

		certFile, err := ioutil.ReadFile(certPath)
		certBlock, _ := pem.Decode(certFile)
		if certBlock == nil {
			return nil, common.NewError("failed to decode cert file").Base(err)
		}

		keyPair, err := tls.X509KeyPair(certBlock.Bytes, decryptedKey)
		if err != nil {
			return nil, err
		}
		keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
		if err != nil {
			return nil, common.NewError("failed to parse leaf certificate").Base(err)
		}

		return &keyPair, nil
	}
	keyPair, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, common.NewError("failed to load key pair").Base(err)
	}
	keyPair.Leaf, err = x509.ParseCertificate(keyPair.Certificate[0])
	if err != nil {
		return nil, common.NewError("failed to parse leaf certificate").Base(err)
	}
	return &keyPair, nil
}

// NewServer creates a tls layer server
func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
	cfg := config.FromContext(ctx, Name).(*Config)

	var tlsConfig *tls.Config
	var fallbackAddress net.Addr
	if len(cfg.TLS.CertmagicDomains) > 0 {
		var storage = &certmagic.FileStorage{cfg.TLS.CertmagicStoragePath}
		certmagic.Default.Storage = storage
		certmagic.DefaultACME.Agreed = true
		certmagic.Default.DefaultServerName = cfg.TLS.CertmagicDefaultSNI

		if !cfg.TLS.AutoRedirect {
			certmagic.DefaultACME.DisableHTTPChallenge = true
		}

		cmCfg := certmagic.NewDefault()
		err := cmCfg.ManageSync(cfg.TLS.CertmagicDomains)
		if err != nil {
			return nil, err
		}

		tlsConfig = cmCfg.TLSConfig()
		tlsConfig.NextProtos = append(cfg.TLS.ALPN, tlsConfig.NextProtos...)

		os.Remove("/tmp/trojan_tls.socket")
		httpsLn, _ := tls.Listen("unix", "/tmp/trojan_tls.socket", tlsConfig)
		httpsServer := &http.Server{
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      2 * time.Minute,
			IdleTimeout:       5 * time.Minute,
			Handler:           http.DefaultServeMux,
		}
		go httpsServer.Serve(httpsLn)
		fallbackAddress = &net.UnixAddr{
			Name: "/tmp/trojan_tls.socket",
			Net:  "unix",
		}

		if cfg.TLS.AutoRedirect {
			hostOnly := func(hostport string) string {
				host, _, err := net.SplitHostPort(hostport)
				if err != nil {
					return hostport // OK; probably had no port to begin with
				}
				return host
			}
			httpRedirectHandler := func(w http.ResponseWriter, r *http.Request) {
				toURL := "https://"

				// since we redirect to the standard HTTPS port, we
				// do not need to include it in the redirect URL
				requestHost := hostOnly(r.Host)

				toURL += requestHost
				toURL += r.URL.RequestURI()

				// get rid of this disgusting unencrypted HTTP connection 🤢
				w.Header().Set("Connection", "close")

				http.Redirect(w, r, toURL, http.StatusMovedPermanently)
			}

			httpLn, _ := net.Listen("tcp", ":80")
			httpServer := &http.Server{
				ReadHeaderTimeout: 5 * time.Second,
				ReadTimeout:       5 * time.Second,
				WriteTimeout:      5 * time.Second,
				IdleTimeout:       5 * time.Second,
			}

			httpServer.Handler = cmCfg.Issuer.(*certmagic.ACMEManager).HTTPChallengeHandler(http.HandlerFunc(httpRedirectHandler))
			go httpServer.Serve(httpLn)
		}
	}
	if cfg.TLS.FallbackPort != 0 && len(cfg.TLS.CertmagicDomains) == 0 {
		if cfg.TLS.FallbackHost == "" {
			cfg.TLS.FallbackHost = cfg.RemoteHost
			log.Warn("empty tls fallback address")
		}
		fallbackAddress = tunnel.NewAddressFromHostPort("tcp", cfg.TLS.FallbackHost, cfg.TLS.FallbackPort)
	} else {
		log.Warn("empty tls fallback port")
	}
	if fallbackAddress != nil {
		fallbackConn, err := net.Dial(fallbackAddress.Network(), fallbackAddress.String())
		if err != nil {
			return nil, common.NewError("invalid fallback address").Base(err)
		}
		fallbackConn.Close()
	}

	keyPair, err := loadKeyPair(cfg.TLS.KeyPath, cfg.TLS.CertPath, cfg.TLS.KeyPassword)
	if err != nil && len(cfg.TLS.CertmagicDomains) == 0 {
		return nil, common.NewError("tls failed to load key pair")
	}

	var keyLogger io.WriteCloser
	if cfg.TLS.KeyLogPath != "" {
		log.Warn("tls key logging activated. USE OF KEY LOGGING COMPROMISES SECURITY. IT SHOULD ONLY BE USED FOR DEBUGGING.")
		file, err := os.OpenFile(cfg.TLS.KeyLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, common.NewError("failed to open key log file").Base(err)
		}
		keyLogger = file
	}

	var cipherSuite []uint16
	if len(cfg.TLS.Cipher) != 0 {
		cipherSuite = fingerprint.ParseCipher(strings.Split(cfg.TLS.Cipher, ":"))
		if tlsConfig != nil {
			tlsConfig.CipherSuites = cipherSuite
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	server := &Server{
		underlay:           underlay,
		fallbackAddress:    fallbackAddress,
		verifySNI:          cfg.TLS.VerifyHostName,
		sni:                cfg.TLS.SNI,
		alpn:               cfg.TLS.ALPN,
		PreferServerCipher: cfg.TLS.PreferServerCipher,
		sessionTicket:      cfg.TLS.ReuseSession,
		connChan:           make(chan tunnel.Conn, 32),
		wsChan:             make(chan tunnel.Conn, 32),
		redir:              redirector.NewRedirector(ctx),
		keyLogger:          keyLogger,
		cipherSuite:        cipherSuite,
		ctx:                ctx,
		cancel:             cancel,
		tlsConfig:          tlsConfig,
		matchSNI: func(s string) bool {
			if len(cfg.TLS.MatchSNI) == 0 {
				return true
			}
			for _, v := range cfg.TLS.MatchSNI {
				if s == v {
					return true
				}
			}
			return false
		},
	}

	if keyPair != nil {
		server.keyPair = []tls.Certificate{*keyPair}
	}

	go server.acceptLoop()
	if cfg.TLS.CertCheckRate > 0 && len(cfg.TLS.CertmagicDomains) == 0 {
		go server.checkKeyPairLoop(
			time.Second*time.Duration(cfg.TLS.CertCheckRate),
			cfg.TLS.KeyPath,
			cfg.TLS.CertPath,
			cfg.TLS.KeyPassword,
		)
	}

	log.Debug("tls server created")
	return server, nil
}
