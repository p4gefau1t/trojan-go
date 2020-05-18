package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/direct"
	"github.com/p4gefau1t/trojan-go/protocol/simplesocks"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/p4gefau1t/trojan-go/proxy"
	"github.com/p4gefau1t/trojan-go/shadow"
	"github.com/p4gefau1t/trojan-go/sockopt"
	"github.com/p4gefau1t/trojan-go/stat"
	"github.com/xtaci/smux"
)

type Server struct {
	common.Runnable
	proxy.Buildable

	listener net.Listener
	auth     stat.Authenticator
	config   *conf.GlobalConfig
	shadow   *shadow.ShadowManager
	ctx      context.Context
	cancel   context.CancelFunc
}

func (s *Server) handleMuxConn(stream *smux.Stream) {
	inboundConn, req, err := simplesocks.NewInboundConnSession(stream)
	if err != nil {
		stream.Close()
		log.Error(common.NewError("Failed to init inbound session").Base(err))
		return
	}
	defer stream.Close()

	switch req.Command {
	case protocol.Connect:
		outboundConn, err := direct.NewOutboundConnSession(s.ctx, req, s.config)
		if err != nil {
			log.Error(err)
			return
		}
		log.Info("Mux conn tunneling to", req.String())
		defer outboundConn.Close()
		proxy.RelayConn(s.ctx, inboundConn, outboundConn, s.config.BufferSize)
	case protocol.Associate:
		outboundPacket, err := direct.NewOutboundPacketSession(s.ctx)
		common.Must(err)
		inboundPacket, err := trojan.NewPacketSession(inboundConn)
		defer inboundPacket.Close()
		proxy.RelayPacket(s.ctx, inboundPacket, outboundPacket)
	default:
		log.Error(fmt.Sprintf("Invalid command %d", req.Command))
		return
	}
}

func (s *Server) handleConn(conn net.Conn) {
	protocol.SetRandomizedTimeout(conn)
	inboundConn, req, err := trojan.NewInboundConnSession(s.ctx, conn, s.config, s.auth, s.shadow)
	if err != nil {
		//once the auth is failed, the conn will be took over by shadow manager. don't close it
		log.Error(common.NewError("Failed to start inbound session, remote:" + conn.RemoteAddr().String()).Base(err))
		return
	}
	protocol.CancelTimeout(conn)
	defer conn.Close()

	if req.Command == protocol.Mux {
		muxServer, err := smux.Server(inboundConn, nil)
		common.Must(err)
		defer muxServer.Close()
		for {
			stream, err := muxServer.AcceptStream()
			if err != nil {
				log.Debug("Mux conn from", conn.RemoteAddr(), "closed:", err)
				return
			}
			go s.handleMuxConn(stream)
		}
	}

	if req.Command == protocol.Associate {
		inboundPacket, err := trojan.NewPacketSession(inboundConn)
		common.Must(err)
		defer inboundPacket.Close()

		outboundPacket, err := direct.NewOutboundPacketSession(s.ctx)
		if err != nil {
			log.Error(err)
			return
		}
		defer outboundPacket.Close()
		log.Info("UDP tunnel established")
		proxy.RelayPacket(s.ctx, inboundPacket, outboundPacket)
		log.Debug("UDP tunnel closed")
		return
	}

	defer inboundConn.Close()
	outboundConn, err := direct.NewOutboundConnSession(s.ctx, req, s.config)
	if err != nil {
		log.Error(err)
		return
	}
	defer outboundConn.Close()

	log.Info("Conn from", conn.RemoteAddr(), "tunneling to", req.String())
	proxy.RelayConn(s.ctx, inboundConn, outboundConn, s.config.BufferSize)
}

func (s *Server) ListenTCP(errChan chan error) {
	log.Info("Trojan-Go server is listening on", s.config.LocalAddress)

	var listener net.Listener
	listener, err := net.Listen("tcp", s.config.LocalAddress.String())
	if err != nil {
		errChan <- err
		return
	}
	s.listener = listener
	defer listener.Close()

	err = sockopt.ApplyTCPListenerOption(listener.(*net.TCPListener), &s.config.TCP)
	if err != nil {
		errChan <- common.NewError(fmt.Sprintf("failed to apply tcp option: %v", &s.config.TCP)).Base(err)
		return
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				errChan <- err
				return
			}
		}
		log.Info("Conn accepted from", conn.RemoteAddr())
		go func(conn net.Conn) {
			if s.config.TLS.ServePlainText {
				s.handleConn(conn)
				return
			}
			//using randomized timeout
			protocol.SetRandomizedTimeout(conn)

			rewindConn := common.NewRewindConn(conn)
			rewindConn.R.SetBufferSize(2048)

			tlsConfig := &tls.Config{
				Certificates:             s.config.TLS.KeyPair,
				CipherSuites:             s.config.TLS.CipherSuites,
				PreferServerCipherSuites: s.config.TLS.PreferServerCipher,
				SessionTicketsDisabled:   !s.config.TLS.SessionTicket,
				NextProtos:               s.config.TLS.ALPN,
			}
			tlsConn := tls.Server(rewindConn, tlsConfig)
			err = tlsConn.Handshake()
			rewindConn.R.StopBuffering()
			protocol.CancelTimeout(conn)

			if s.config.LogLevel == 0 {
				state := tlsConn.ConnectionState()
				log.Trace("TLS handshaked", tls.CipherSuiteName(state.CipherSuite), state.DidResume, state.NegotiatedProtocol)
			}

			if err != nil {
				rewindConn.R.Rewind()
				err = common.NewError("Failed to perform tls handshake with " + conn.RemoteAddr().String()).Base(err)
				log.Warn(err)
				if s.config.TLS.FallbackAddress != nil {
					s.shadow.SubmitScapegoat(&shadow.Scapegoat{
						Conn:          rewindConn,
						ShadowAddress: s.config.TLS.FallbackAddress,
						Info:          err.Error(),
					})
				} else if s.config.TLS.HTTPResponse != nil {
					rewindConn.Write(s.config.TLS.HTTPResponse)
					rewindConn.Close()
				} else {
					rewindConn.Close()
				}
				return
			}
			s.handleConn(tlsConn)
		}(conn)
	}
}

func (s *Server) Run() error {
	errChan := make(chan error, 2)
	if s.config.API.Enabled {
		log.Info("API enabled")
		go func() {
			errChan <- proxy.RunAPIService(conf.Server, s.ctx, s.config, s.auth)
		}()
	}
	go s.ListenTCP(errChan)
	select {
	case <-s.ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}

func (s *Server) Close() error {
	log.Info("Shutting down server..")
	s.cancel()
	s.listener.Close()
	return nil
}

func (*Server) Build(config *conf.GlobalConfig) (common.Runnable, error) {
	ctx, cancel := context.WithCancel(context.Background())
	var err error
	authDriver := "memory"
	if config.MySQL.Enabled {
		authDriver = "mysql"
	} else if config.Redis.Enabled {
		authDriver = "redis"
	}
	auth, err := stat.NewAuth(ctx, authDriver, config)
	if err != nil {
		cancel()
		return nil, err
	}
	s := &Server{
		config: config,
		ctx:    ctx,
		cancel: cancel,
		shadow: shadow.NewShadowManager(ctx, config),
		auth:   auth,
	}
	return s, nil
}

func init() {
	proxy.RegisterProxy(conf.Server, &Server{})
}
