package proxy

import (
	"crypto/tls"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/direct"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
	"github.com/xtaci/smux"
)

type Server struct {
	config *conf.GlobalConfig
	common.Runnable
}

func (s *Server) handleMuxConn(stream *smux.Stream) {
	inboundConn, err := trojan.NewInboundConnSession(stream, s.config)
	if err != nil {
		stream.Close()
		logger.Error(common.NewError("cannot start inbound session").Base(err))
		return
	}
	defer inboundConn.Close()
	req := inboundConn.GetRequest()
	if req.Command != protocol.Connect {
		logger.Error("mux only support tcp now")
		return
	}
	outboundConn, err := direct.NewOutboundConnSession(nil, req)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("mux tunneling to", req.String())
	defer outboundConn.Close()
	proxyConn(inboundConn, outboundConn)
}

func (s *Server) handleConn(conn net.Conn) {
	inboundConn, err := trojan.NewInboundConnSession(conn, s.config)

	if err != nil {
		logger.Error(err)
		inboundConn.Close()
		return
	}
	req := inboundConn.GetRequest()

	if req.Command == protocol.Mux {
		muxServer, err := smux.Server(conn, nil)
		defer muxServer.Close()
		common.Must(err)
		for {
			stream, err := muxServer.AcceptStream()
			if err != nil {
				logger.Error(err)
				return
			}
			go s.handleMuxConn(stream)
		}
	}

	if req.Command == protocol.Associate {
		inboundPacket, _ := trojan.NewPacketSession(inboundConn)
		defer inboundPacket.Close()

		outboundPacket, err := direct.NewOutboundPacketSession()
		if err != nil {
			logger.Error(err)
			return
		}
		defer outboundPacket.Close()
		logger.Info("UDP associated")
		proxyPacket(inboundPacket, outboundPacket)
		logger.Info("UDP tunnel closed")
		return
	}

	outboundConn, err := direct.NewOutboundConnSession(nil, req)
	if err != nil {
		logger.Error(err)
		return
	}
	defer outboundConn.Close()

	logger.Info("conn from", conn.RemoteAddr(), "tunneling to", req.String())
	proxyConn(inboundConn, outboundConn)
}

func (s *Server) Run() error {
	tlsConfig := &tls.Config{
		Certificates: s.config.TLS.KeyPair,
		CipherSuites: s.config.TLS.CipherSuites,
	}
	listener, err := tls.Listen("tcp", s.config.LocalAddr.String(), tlsConfig)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			err = common.NewError("tls handshake failed").Base(err)
			logger.Warn(err)
			conn.Close()
			continue
		}
		go s.handleConn(conn)
	}
}
