package transport

import (
	"bufio"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"

	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/redirector"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/transport/fingerprint"
	"github.com/p4gefau1t/trojan-go/tunnel/websocket"
)

// Server is a server of trasport layer
type Server struct {
	tcpListener        net.Listener
	fallbackAddress    *tunnel.Address
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
	plugin             bool
	redir              *redirector.Redirector
	cmd                *exec.Cmd
	ctx                context.Context
	cancel             context.CancelFunc
}

func (s *Server) Close() error {
	s.cancel()
	if s.keyLogger != nil {
		s.keyLogger.Close()
	}
	if s.cmd != nil {
		s.cmd.Process.Kill()
	}
	return s.tcpListener.Close()
}

func (s *Server) acceptLoop() {
	for {
		tcpConn, err := s.tcpListener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
			default:
				log.Fatal(common.NewError("transport accept error"))
			}
			return
		}
		log.Info("tcp connection from", tcpConn.RemoteAddr())
		go func(tcpConn net.Conn) {
			var transportConn net.Conn
			if s.plugin {
				// plain text mode
				transportConn = tcpConn
			} else {
				// default tls
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

				handshakeRewindConn := common.NewRewindConn(tcpConn)
				handshakeRewindConn.SetBufferSize(2048)

				tlsConn := tls.Server(handshakeRewindConn, tlsConfig)
				err = tlsConn.Handshake()
				handshakeRewindConn.StopBuffering()

				if err != nil {
					if !sniVerified {
						// close tls conn immediately if the sni is invalid
						tlsConn.Close()
						log.Error(common.NewError("tls client hello with wrong sni").Base(err))
					} else if strings.Contains(err.Error(), "first record does not look like a TLS handshake") {
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
						// other cases, simply close it
						tlsConn.Close()
						log.Error(common.NewError("tls handshake failed").Base(err))
					}
					return
				}

				state := tlsConn.ConnectionState()
				log.Trace("tls handshake", tls.CipherSuiteName(state.CipherSuite), state.DidResume, state.NegotiatedProtocol)
				transportConn = tlsConn
			}

			// we use real http header parser to mimic a real http server
			rewindConn := common.NewRewindConn(transportConn)
			rewindConn.SetBufferSize(512)
			defer rewindConn.StopBuffering()
			r := bufio.NewReader(rewindConn)
			httpReq, err := http.ReadRequest(r)
			rewindConn.Rewind()
			if err != nil {
				// this is not a http request, pass it to trojan protocol layer for further inspection
				s.connChan <- &Conn{
					Conn: rewindConn,
				}
			} else {
				// this is a http request, pass it to websocket protocol layer
				log.Debug("http req: ", httpReq)
				s.wsChan <- &Conn{
					Conn: rewindConn,
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

// NewServer creates a transport layer server
func NewServer(ctx context.Context, _ tunnel.Server) (*Server, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	ctx, cancel := context.WithCancel(ctx)
	listenAddress := tunnel.NewAddressFromHostPort("tcp", cfg.LocalHost, cfg.LocalPort)
	fallbackAddress := tunnel.NewAddressFromHostPort("tcp", cfg.TLS.FallbackHost, cfg.TLS.FallbackPort)
	if cfg.TLS.FallbackPort != 0 {
		fallbackConn, err := net.Dial("tcp", fallbackAddress.String())
		if err != nil {
			return nil, common.NewError("invalid fallback address").Base(err)
		}
		fallbackConn.Close()
	}

	if cfg.TransportPlugin.Enabled {
		log.Warn("transport server will use transport plugin and work in plain text mode")
		var cmd *exec.Cmd
		switch cfg.TransportPlugin.Type {
		case "shadowsocks":
			trojanHost := "127.0.0.1"
			trojanPort := common.PickPort("tcp", trojanHost)
			cfg.TransportPlugin.Env = append(
				cfg.TransportPlugin.Env,
				"SS_REMOTE_HOST="+cfg.LocalHost,
				"SS_REMOTE_PORT="+strconv.FormatInt(int64(cfg.LocalPort), 10),
				"SS_LOCAL_HOST="+trojanHost,
				"SS_LOCAL_PORT="+strconv.FormatInt(int64(trojanPort), 10),
				"SS_PLUGIN_OPTIONS="+cfg.TransportPlugin.PluginOption,
			)

			cfg.LocalHost = trojanHost
			cfg.LocalPort = trojanPort
			listenAddress = tunnel.NewAddressFromHostPort("tcp", cfg.LocalHost, cfg.LocalPort)
			log.Debug("new listen address", listenAddress)
			log.Debug("plugin env", cfg.TransportPlugin.Env)

			cmd = exec.Command(cfg.TransportPlugin.Command, cfg.TransportPlugin.Arg...)
			cmd.Env = append(cmd.Env, cfg.TransportPlugin.Env...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			cmd.Start()
		case "other":
			cmd = exec.Command(cfg.TransportPlugin.Command, cfg.TransportPlugin.Arg...)
			cmd.Env = append(cmd.Env, cfg.TransportPlugin.Env...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			cmd.Start()
		case "plaintext":
			// do nothing
		default:
			return nil, common.NewError("invalid plugin type: " + cfg.TransportPlugin.Type)
		}
		tcpListener, err := net.Listen("tcp", listenAddress.String())
		if err != nil {
			return nil, err
		}
		server := &Server{
			connChan:    make(chan tunnel.Conn, 32),
			wsChan:      make(chan tunnel.Conn, 32),
			tcpListener: tcpListener,
			redir:       redirector.NewRedirector(ctx),
			cmd:         cmd,
			plugin:      true,
			ctx:         ctx,
			cancel:      cancel,
		}
		go server.acceptLoop()
		return server, nil
	}

	if cfg.TLS.FallbackHost == "" {
		cfg.TLS.FallbackHost = cfg.RemoteHost
		log.Warn("empty fallback address")
	}
	if cfg.TLS.FallbackPort == 0 {
		cfg.TLS.FallbackPort = cfg.RemotePort
		log.Warn("empty fallback port")
	}
	if cfg.TLS.SNI == "" && cfg.TLS.VerifyHostName {
		return nil, common.NewError("cannot verify hostname without sni")
	}

	tcpListener, err := net.Listen("tcp", listenAddress.String())
	if err != nil {
		return nil, err
	}

	server := &Server{
		tcpListener:        tcpListener,
		fallbackAddress:    fallbackAddress,
		verifySNI:          cfg.TLS.VerifyHostName,
		sni:                cfg.TLS.SNI,
		alpn:               cfg.TLS.ALPN,
		PreferServerCipher: cfg.TLS.PreferServerCipher,
		sessionTicket:      cfg.TLS.ReuseSession,
		connChan:           make(chan tunnel.Conn, 32),
		wsChan:             make(chan tunnel.Conn, 32),
		plugin:             false,
		redir:              redirector.NewRedirector(ctx),
		ctx:                ctx,
		cancel:             cancel,
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

	if cfg.TLS.KeyLogPath != "" {
		file, err := os.OpenFile(cfg.TLS.KeyLogPath, os.O_WRONLY, 0600)
		if err != nil {
			return nil, common.NewError("failed to open key log file")
		}
		server.keyLogger = file
	}

	if len(cfg.TLS.Cipher) != 0 {
		server.cipherSuite = fingerprint.ParseCipher(strings.Split(cfg.TLS.Cipher, ":"))
	}

	go server.acceptLoop()

	log.Info("transport server is listening on tcp:", tcpListener.Addr().String())
	log.Debug("transport server created")
	return server, nil
}
