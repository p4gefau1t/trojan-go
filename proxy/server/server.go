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
	meter    stat.TrafficMeter
	config   *conf.GlobalConfig
	shadow   *shadow.ShadowManager
	ctx      context.Context
	cancel   context.CancelFunc
}

func (s *Server) handleMuxConn(stream *smux.Stream) {
	inboundConn, req, err := simplesocks.NewInboundConnSession(stream)
	if err != nil {
		stream.Close()
		log.Error(common.NewError("cannot start inbound session").Base(err))
		return
	}
	inboundConn.(protocol.NeedMeter).SetMeter(s.meter)
	switch req.Command {
	case protocol.Connect:
		outboundConn, err := direct.NewOutboundConnSession(req)
		if err != nil {
			log.Error(err)
			return
		}
		log.Info("mux tunneling to", req.String())
		defer outboundConn.Close()
		proxy.ProxyConn(s.ctx, inboundConn, outboundConn, s.config.BufferSize)
	case protocol.Associate:
		outboundPacket, err := direct.NewOutboundPacketSession(s.ctx)
		common.Must(err)
		inboundPacket, err := trojan.NewPacketSession(inboundConn)
		proxy.ProxyPacket(s.ctx, inboundPacket, outboundPacket)
	default:
		log.Error(fmt.Sprintf("invalid command %d", req.Command))
		return
	}
}

func (s *Server) handleConn(conn *tls.Conn) {
	protocol.SetRandomizedTimeout(conn)
	inboundConn, req, err := trojan.NewInboundConnSession(s.ctx, conn, s.config, s.auth, s.shadow)
	if err != nil {
		//once the auth is failed, the conn will be took over by shadow manager. don't close it
		log.Error(common.NewError("failed to start inbound session, remote:" + conn.RemoteAddr().String()).Base(err))
		return
	}
	protocol.CancelTimeout(conn)

	if req.Command == protocol.Mux {
		muxServer, err := smux.Server(inboundConn, nil)
		defer muxServer.Close()
		common.Must(err)
		for {
			stream, err := muxServer.AcceptStream()
			if err != nil {
				log.Debug("mux conn from", conn.RemoteAddr(), "closed:", err)
				return
			}
			go s.handleMuxConn(stream)
		}
	}
	inboundConn.(protocol.NeedMeter).SetMeter(s.meter)

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
		log.Info("udp tunnel established")
		proxy.ProxyPacket(s.ctx, inboundPacket, outboundPacket)
		log.Debug("udp tunnel closed")
		return
	}

	defer inboundConn.Close()
	outboundConn, err := direct.NewOutboundConnSession(req)
	if err != nil {
		log.Error(err)
		return
	}
	defer outboundConn.Close()

	log.Info("conn from", conn.RemoteAddr(), "tunneling to", req.String())
	proxy.ProxyConn(s.ctx, inboundConn, outboundConn, s.config.BufferSize)
}

func (s *Server) Run() error {
	var err error
	if s.config.MySQL.Enabled {
		s.auth, err = stat.NewMixedAuthenticator(s.config)
		if err != nil {
			return common.NewError("failed to init auth").Base(err)
		}
		s.meter, err = stat.NewDBTrafficMeter(s.config)
		if err != nil {
			return common.NewError("failed to init traffic meter").Base(err)
		}
	} else {
		s.auth = &stat.ConfigUserAuthenticator{
			Config: s.config,
		}
	}
	defer s.auth.Close()
	if s.meter != nil {
		defer s.meter.Close()
	}
	log.Info("server is running at", s.config.LocalAddress)

	var listener net.Listener
	listener, err = net.Listen("tcp", s.config.LocalAddress.String())
	if err != nil {
		return err
	}
	s.listener = listener
	defer listener.Close()

	err = sockopt.ApplyTCPListenerOption(listener.(*net.TCPListener), &s.config.TCP)
	if err != nil {
		return common.NewError(fmt.Sprintf("failed to apply tcp option: %v", &s.config.TCP)).Base(err)
	}

	tlsConfig := &tls.Config{
		Certificates:             s.config.TLS.KeyPair,
		CipherSuites:             s.config.TLS.CipherSuites,
		PreferServerCipherSuites: s.config.TLS.PreferServerCipher,
		SessionTicketsDisabled:   !s.config.TLS.SessionTicket,
		NextProtos:               s.config.TLS.ALPN,
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return nil
			default:
			}
			return err
		}
		log.Info("conn accepted from", conn.RemoteAddr())
		go func(conn net.Conn) {
			//using randomized timeout
			protocol.SetRandomizedTimeout(conn)

			rewindConn := common.NewRewindConn(conn)
			rewindConn.R.SetBufferSize(512)

			tlsConn := tls.Server(rewindConn, tlsConfig)
			err = tlsConn.Handshake()
			rewindConn.R.StopBuffering()
			protocol.CancelTimeout(conn)

			if err != nil {
				rewindConn.R.Rewind()
				err = common.NewError("failed to perform tls handshake with " + conn.RemoteAddr().String()).Base(err)
				log.Warn(err)
				if s.config.TLS.FallbackAddress != nil {
					s.shadow.CommitScapegoat(&shadow.Scapegoat{
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

func (s *Server) Close() error {
	log.Info("shutting down server..")
	s.cancel()
	s.listener.Close()
	return nil
}

func (s *Server) Build(config *conf.GlobalConfig) (common.Runnable, error) {
	s.config = config
	s.ctx, s.cancel = context.WithCancel(context.Background())
	s.shadow = shadow.NewShadowManager(s.ctx, s.config)
	return s, nil
}

func init() {
	proxy.RegisterProxy(conf.Server, &Server{})
}
