package server

import (
	"context"
	"crypto/tls"
	"database/sql"
	"net"
	"reflect"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/direct"
	"github.com/p4gefau1t/trojan-go/protocol/mux"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/p4gefau1t/trojan-go/proxy"
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
	ctx      context.Context
	cancel   context.CancelFunc
}

func (s *Server) handleMuxConn(stream *smux.Stream, passwordHash string) {
	inboundConn, err := mux.NewInboundMuxConnSession(stream, passwordHash)
	if err != nil {
		stream.Close()
		log.DefaultLogger.Error(common.NewError("cannot start inbound session").Base(err))
		return
	}
	inboundConn.(protocol.NeedMeter).SetMeter(s.meter)
	defer inboundConn.Close()
	req := inboundConn.GetRequest()
	if req.Command != protocol.Connect {
		log.DefaultLogger.Error("mux only support tcp now")
		return
	}
	outboundConn, err := direct.NewOutboundConnSession(nil, req)
	if err != nil {
		log.DefaultLogger.Error(err)
		return
	}
	log.DefaultLogger.Info("user", passwordHash, "mux tunneling to", req.String())
	defer outboundConn.Close()
	proxy.ProxyConn(inboundConn, outboundConn)
}

func (s *Server) handleConn(conn net.Conn) {
	inboundConn, err := trojan.NewInboundConnSession(conn, s.config, s.auth)
	if err != nil {
		log.DefaultLogger.Error(common.NewError("failed to start inbound session, remote:" + conn.RemoteAddr().String()).Base(err))
		return
	}

	req := inboundConn.GetRequest()
	hash := inboundConn.(protocol.HasHash).GetHash()

	if req.Command == protocol.Mux {
		muxServer, err := smux.Server(conn, nil)
		defer muxServer.Close()
		common.Must(err)
		for {
			stream, err := muxServer.AcceptStream()
			if err != nil {
				log.DefaultLogger.Debug("mux conn from", conn.RemoteAddr(), "closed: ", err)
				return
			}
			go s.handleMuxConn(stream, hash)
		}
	}
	inboundConn.(protocol.NeedMeter).SetMeter(s.meter)

	if req.Command == protocol.Associate {
		inboundPacket, _ := trojan.NewPacketSession(inboundConn)
		defer inboundPacket.Close()

		outboundPacket, err := direct.NewOutboundPacketSession()
		if err != nil {
			log.DefaultLogger.Error(err)
			return
		}
		defer outboundPacket.Close()
		log.DefaultLogger.Info("UDP associated")
		proxy.ProxyPacket(inboundPacket, outboundPacket)
		log.DefaultLogger.Info("UDP tunnel closed")
		return
	}

	defer inboundConn.Close()
	outboundConn, err := direct.NewOutboundConnSession(nil, req)
	if err != nil {
		log.DefaultLogger.Error(err)
		return
	}
	defer outboundConn.Close()

	log.DefaultLogger.Info("conn from", conn.RemoteAddr(), "tunneling to", req.String())
	proxy.ProxyConn(inboundConn, outboundConn)
}

func (s *Server) handleInvalidConn(conn net.Conn, tlsConn *tls.Conn) {
	defer conn.Close()
	if len(s.config.TLS.HTTPResponse) > 0 {
		log.DefaultLogger.Warn("trying to response with a plain http response")
		conn.Write(s.config.TLS.HTTPResponse)
		return
	}

	if s.config.TLS.FallbackAddr != nil {
		defer func() {
			if r := recover(); r != nil {
				log.DefaultLogger.Error("recovered", r)
			}
		}()
		//HACK
		//obtain the bytes buffered by the tls conn
		v := reflect.ValueOf(*tlsConn)
		buf := v.FieldByName("rawInput").FieldByName("buf").Bytes()
		log.DefaultLogger.Debug("payload:" + string(buf))

		remote, err := net.Dial("tcp", s.config.TLS.FallbackAddr.String())
		if err != nil {
			log.DefaultLogger.Warn(common.NewError("failed to dial to tls fallback server").Base(err))
			return
		}
		log.DefaultLogger.Warn("proxying this invalid tls conn to the tls fallback server")
		remote.Write(buf)
		proxy.ProxyConn(conn, remote)
	} else {
		log.DefaultLogger.Warn("fallback port is unspecified, closing")
	}

}

func (s *Server) Run() error {
	var db *sql.DB
	var err error
	if s.config.MySQL.Enabled {
		db, err = common.ConnectDatabase(
			"mysql",
			s.config.MySQL.Username,
			s.config.MySQL.Password,
			s.config.MySQL.ServerHost,
			s.config.MySQL.ServerPort,
			s.config.MySQL.Database,
		)
		if err != nil {
			return common.NewError("failed to connect to database server").Base(err)
		}
	}
	if db == nil {
		s.auth = &stat.ConfigUserAuthenticator{
			Config: s.config,
		}
		s.meter = &stat.EmptyTrafficMeter{}
	} else {
		s.auth, err = stat.NewMixedAuthenticator(s.config, db)
		if err != nil {
			return common.NewError("failed to init auth").Base(err)
		}
		s.meter, err = stat.NewDBTrafficMeter(s.config, db)
		if err != nil {
			return common.NewError("failed to init traffic meter").Base(err)
		}
	}
	defer s.auth.Close()
	defer s.meter.Close()
	log.DefaultLogger.Info("server is running at", s.config.LocalAddr)

	var listener net.Listener
	if s.config.TCP.ReusePort || s.config.TCP.FastOpen || s.config.TCP.NoDelay {
		listener, err = ListenWithTCPOption(
			s.config.TCP.FastOpen,
			s.config.TCP.ReusePort,
			s.config.TCP.NoDelay,
			s.config.LocalIP,
			s.config.LocalAddr.String(),
		)
		if err != nil {
			return err
		}
	} else {
		listener, err = net.Listen("tcp", s.config.LocalAddr.String())
		if err != nil {
			return err
		}
	}
	s.listener = listener
	defer listener.Close()

	tlsConfig := &tls.Config{
		Certificates:             s.config.TLS.KeyPair,
		CipherSuites:             s.config.TLS.CipherSuites,
		PreferServerCipherSuites: s.config.TLS.PreferServerCipher,
		SessionTicketsDisabled:   !s.config.TLS.SessionTicket,
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return nil
			default:
			}
			log.DefaultLogger.Warn(err)
			continue
		}
		tlsConn := tls.Server(conn, tlsConfig)
		err = tlsConn.Handshake()
		if err != nil {
			log.DefaultLogger.Warn(common.NewError("failed to perform tls handshake, remote:" + conn.RemoteAddr().String()).Base(err))
			go s.handleInvalidConn(conn, tlsConn)
			continue
		}
		go s.handleConn(tlsConn)
	}
}

func (s *Server) Close() error {
	log.DefaultLogger.Info("shutting down server..")
	if s.listener != nil {
		s.listener.Close()
	}
	s.cancel()
	return nil
}

func (s *Server) Build(config *conf.GlobalConfig) (common.Runnable, error) {
	s.config = config
	s.ctx, s.cancel = context.WithCancel(context.Background())
	return s, nil
}

func init() {
	proxy.RegisterProxy(conf.Server, &Server{})
}
