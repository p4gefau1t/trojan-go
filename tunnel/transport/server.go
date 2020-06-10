package transport

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"github.com/p4gefau1t/trojan-go/tunnel/websocket"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/redirector"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

// Server is a server of trasport layer
type Server struct {
	tcpListener        net.Listener
	listenAddress      *tunnel.Address
	fallbackAddress    *tunnel.Address
	verifySNI          bool
	sni                string
	alpn               []string
	PreferServerCipher bool
	certPool           *x509.CertPool
	keyPair            []tls.Certificate
	httpResp           []byte
	cipherSuite        []uint16
	sessionTicket      bool
	curve              []tls.CurveID
	keyLogger          io.WriteCloser
	redir              *redirector.Redirector
	connChan           chan tunnel.Conn
	wsChan             chan tunnel.Conn
	ctx                context.Context
	cancel             context.CancelFunc
}

func (s *Server) Close() error {
	s.cancel()
	s.keyLogger.Close()
	return s.tcpListener.Close()
}

func (s *Server) acceptLoop() {
	for {
		tcpConn, err := s.tcpListener.Accept()
		if err != nil {
			s.cancel()
			log.Error(common.NewError("transport accept error"))
			return
		}
		go func(tcpConn net.Conn) {
			sniVerified := true
			tlsConfig := &tls.Config{
				Certificates:             s.keyPair,
				CipherSuites:             s.cipherSuite,
				PreferServerCipherSuites: s.PreferServerCipher,
				SessionTicketsDisabled:   !s.sessionTicket,
				NextProtos:               s.alpn,
				KeyLogWriter:             s.keyLogger,
				GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
					if s.verifySNI && hello.ServerName != s.sni {
						sniVerified = false
						return nil, common.NewError("sni mismatched: " + hello.ServerName + ", expected: " + s.sni)
					}
					return &s.keyPair[0], nil
				},
			}

			// ------------------------ WAR ZONE ----------------------------

			rewindConn := common.NewRewindConn(tcpConn)
			rewindConn.SetBufferSize(2048)

			tlsConn := tls.Server(rewindConn, tlsConfig)
			err = tlsConn.Handshake()
			rewindConn.StopBuffering()

			if err != nil {
				if !sniVerified {
					// close tls conn immediately if the sni is invalid
					tlsConn.Close()
					log.Error(common.NewError("tls client hello with wrong sni").Base(err))
				} else if strings.Contains(err.Error(), "first record does not look like a TLS handshake") {
					// not a valid tls client hello
					rewindConn.Rewind()
					log.Error(common.NewError("failed to perform tls handshake with " + tlsConn.RemoteAddr().String() + ", redirecting").Base(err))
					if s.fallbackAddress != nil {
						s.redir.Redirect(&redirector.Redirection{
							InboundConn: rewindConn,
							RedirectTo:  s.fallbackAddress,
						})
					} else if s.httpResp != nil {
						rewindConn.Write(s.httpResp)
						rewindConn.Close()
					} else {
						rewindConn.Close()
					}
				} else {
					// other cases, simply close it
					tlsConn.Close()
					log.Error(common.NewError("tls handshake failed").Base(err))
				}
				return
			}

			state := tlsConn.ConnectionState()
			log.Trace("tls handshake", tls.CipherSuiteName(state.CipherSuite), state.DidResume, state.NegotiatedProtocol)

			// we use a real http header parser to mimic a real http server
			tlsRewindConn := common.NewRewindConn(tlsConn)
			tlsRewindConn.SetBufferSize(512)
			defer tlsRewindConn.StopBuffering()
			r := bufio.NewReader(tlsRewindConn)
			httpReq, err := http.ReadRequest(r)
			tlsRewindConn.Rewind()
			if err != nil {
				// this is not a http request, pass it to trojan protocol layer for further inspection
				s.connChan <- &Conn{
					Conn: tlsRewindConn,
				}
			} else {
				// this is a http request, pass it to websocket protocol layer
				log.Debug("http req: ", httpReq)
				s.wsChan <- &Conn{
					Conn: tlsRewindConn,
				}
			}
		}(tcpConn)
	}
}

func (s *Server) AcceptConn(overlay tunnel.Tunnel) (tunnel.Conn, error) {
	if _, ok := overlay.(*websocket.Tunnel); ok {
		// websocket overlay
		select {
		case conn := <-s.wsChan:
			return conn, nil
		case <-s.ctx.Done():
			return nil, io.EOF
		}
	}
	// trojan overlay
	select {
	case conn := <-s.connChan:
		return conn, nil
	case <-s.ctx.Done():
		return nil, io.EOF
	}
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

// NewServer creates a transport layer server
func NewServer(ctx context.Context, _ tunnel.Server) (*Server, error) {
	cfg := config.FromContext(ctx, Name).(*Config)

	if cfg.TLS.FallbackHost == "" {
		cfg.TLS.FallbackHost = cfg.RemoteHost
		log.Warn("empty fallback address")
	}
	if cfg.TLS.FallbackPort == 0 {
		cfg.TLS.FallbackPort = cfg.RemotePort
		log.Warn("empty fallback port")
	}

	listenAddress := tunnel.NewAddressFromHostPort("tcp", cfg.LocalHost, cfg.LocalPort)
	fallbackAddress := tunnel.NewAddressFromHostPort("tcp", cfg.TLS.FallbackHost, cfg.TLS.FallbackPort)
	if cfg.TLS.FallbackPort != 0 {
		fallbackConn, err := net.Dial("tcp", fallbackAddress.String())
		if err != nil {
			return nil, common.NewError("invalid fallback address").Base(err)
		}
		fallbackConn.Close()
	}

	tcpListener, err := net.Listen("tcp", listenAddress.String())
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(ctx)
	server := &Server{
		fallbackAddress: fallbackAddress,
		redir:           redirector.NewRedirector(ctx),
		ctx:             ctx,
		cancel:          cancel,
		tcpListener:     tcpListener,
		connChan:        make(chan tunnel.Conn, 32),
		wsChan:          make(chan tunnel.Conn, 32),
		sni:             cfg.TLS.SNI,
		alpn:            cfg.TLS.ALPN,
		verifySNI:       cfg.TLS.VerifyHostName,
	}

	if cfg.TLS.KeyLogPath != "" {
		log.Warn("tls key logging activated. USE OF KEY LOGGING COMPROMISES SECURITY. IT SHOULD ONLY BE USED FOR DEBUGGING.")
		file, err := os.OpenFile(cfg.TLS.KeyLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return nil, common.NewError("failed to open key log file").Base(err)
		}
		server.keyLogger = file
	}
	if cfg.TLS.KeyPassword != "" {
		keyFile, err := ioutil.ReadFile(cfg.TLS.KeyPath)
		if err != nil {
			return nil, common.NewError("failed to load key file").Base(err)
		}
		keyBlock, _ := pem.Decode(keyFile)
		if keyBlock == nil {
			return nil, common.NewError("failed to decode key file").Base(err)
		}
		decryptedKey, err := x509.DecryptPEMBlock(keyBlock, []byte(cfg.TLS.KeyPassword))
		if err == nil {
			return nil, common.NewError("failed to decrypt key").Base(err)
		}

		certFile, err := ioutil.ReadFile(cfg.TLS.CertPath)
		certBlock, _ := pem.Decode(certFile)
		if certBlock == nil {
			return nil, common.NewError("failed to decode cert file").Base(err)
		}

		keyPair, err := tls.X509KeyPair(certBlock.Bytes, decryptedKey)
		if err != nil {
			return nil, err
		}

		server.keyPair = []tls.Certificate{keyPair}
	} else {
		if len(cfg.TLS.CertBytes) != 0 {
			keyPair, err := tls.X509KeyPair(cfg.TLS.CertBytes, cfg.TLS.KeyBytes)
			if err != nil {
				return nil, err
			}
			server.keyPair = []tls.Certificate{keyPair}
		} else {
			keyPair, err := tls.LoadX509KeyPair(cfg.TLS.CertPath, cfg.TLS.KeyPath)
			if err != nil {
				return nil, common.NewError("failed to load key pair").Base(err)
			}
			server.keyPair = []tls.Certificate{keyPair}
		}
	}
	go server.acceptLoop()
	log.Debug("transport server created")
	return server, nil
}
