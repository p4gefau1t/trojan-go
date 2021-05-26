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
	"sync"
	"time"

	"github.com/huandu/go-clone"

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
	fallbackAddress    *tunnel.Address
	verifySNI          bool
	sni                string
	alpn               []string
	PreferServerCipher bool
	keyPair            []tls.Certificate
	keyPairLock        sync.RWMutex
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
			tlsConfig := &tls.Config{
				CipherSuites:             s.cipherSuite,
				PreferServerCipherSuites: s.PreferServerCipher,
				SessionTicketsDisabled:   !s.sessionTicket,
				NextProtos:               s.alpn,
				KeyLogWriter:             s.keyLogger,
				GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
					s.keyPairLock.RLock()
					defer s.keyPairLock.RUnlock()
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
					keyPairCopied := clone.Clone(&s.keyPair[0]).(*tls.Certificate)
					return keyPairCopied, nil
				},
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
					switch {
					case s.fallbackAddress != nil:
						s.redir.Redirect(&redirector.Redirection{
							InboundConn: handshakeRewindConn,
							RedirectTo:  s.fallbackAddress,
						})
					case s.httpResp != nil:
						handshakeRewindConn.Write(s.httpResp)
						handshakeRewindConn.Close()
					default:
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
			if err != nil {
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
		log.Debug("checking cert...")
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
			s.keyPairLock.Lock()
			s.keyPair = []tls.Certificate{*keyPair}
			s.keyPairLock.Unlock()
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

	var fallbackAddress *tunnel.Address
	var httpResp []byte
	if cfg.TLS.FallbackPort != 0 {
		if cfg.TLS.FallbackHost == "" {
			cfg.TLS.FallbackHost = cfg.RemoteHost
			log.Warn("empty tls fallback address")
		}
		fallbackAddress = tunnel.NewAddressFromHostPort("tcp", cfg.TLS.FallbackHost, cfg.TLS.FallbackPort)
		fallbackConn, err := net.Dial("tcp", fallbackAddress.String())
		if err != nil {
			return nil, common.NewError("invalid fallback address").Base(err)
		}
		fallbackConn.Close()
	} else {
		log.Warn("empty tls fallback port")
		if cfg.TLS.HTTPResponseFileName != "" {
			httpRespBody, err := ioutil.ReadFile(cfg.TLS.HTTPResponseFileName)
			if err != nil {
				return nil, common.NewError("invalid response file").Base(err)
			}
			httpResp = httpRespBody
		} else {
			log.Warn("empty tls http response")
		}
	}

	keyPair, err := loadKeyPair(cfg.TLS.KeyPath, cfg.TLS.CertPath, cfg.TLS.KeyPassword)
	if err != nil {
		return nil, common.NewError("tls failed to load key pair")
	}

	var keyLogger io.WriteCloser
	if cfg.TLS.KeyLogPath != "" {
		log.Warn("tls key logging activated. USE OF KEY LOGGING COMPROMISES SECURITY. IT SHOULD ONLY BE USED FOR DEBUGGING.")
		file, err := os.OpenFile(cfg.TLS.KeyLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return nil, common.NewError("failed to open key log file").Base(err)
		}
		keyLogger = file
	}

	var cipherSuite []uint16
	if len(cfg.TLS.Cipher) != 0 {
		cipherSuite = fingerprint.ParseCipher(strings.Split(cfg.TLS.Cipher, ":"))
	}

	ctx, cancel := context.WithCancel(ctx)
	server := &Server{
		underlay:           underlay,
		fallbackAddress:    fallbackAddress,
		httpResp:           httpResp,
		verifySNI:          cfg.TLS.VerifyHostName,
		sni:                cfg.TLS.SNI,
		alpn:               cfg.TLS.ALPN,
		PreferServerCipher: cfg.TLS.PreferServerCipher,
		sessionTicket:      cfg.TLS.ReuseSession,
		connChan:           make(chan tunnel.Conn, 32),
		wsChan:             make(chan tunnel.Conn, 32),
		redir:              redirector.NewRedirector(ctx),
		keyPair:            []tls.Certificate{*keyPair},
		keyLogger:          keyLogger,
		cipherSuite:        cipherSuite,
		ctx:                ctx,
		cancel:             cancel,
	}

	go server.acceptLoop()
	if cfg.TLS.CertCheckRate > 0 {
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
