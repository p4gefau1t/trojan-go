package proxy

import (
	"crypto/tls"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/direct"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
)

type Server struct {
	config *conf.GlobalConfig
	common.Runnable
}

func (s *Server) handleConn(conn net.Conn) {
	inboundConn, err := trojan.NewInboundConnSession(conn, s.config)
	if err != nil {
		logger.Error(err)
		return
	}
	req := inboundConn.GetRequest()

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

	defer inboundConn.Close()

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
	//listener, err := net.ListenTCP("tcp", s.config.LocalAddr)
	listener, err := net.Listen("tcp", s.config.LocalAddr.String())
	if err != nil {
		return err
	}
	logger.Info("running server at", listener.Addr())
	for {
		conn, err := listener.Accept()
		tlsConn := tls.Server(conn, tlsConfig)
		if err := tlsConn.Handshake(); err != nil {
			err = common.NewError("a non-tls conn accepted").Base(err)
			logger.Warn(err)
			remoteConn, err := net.Dial("tcp", s.config.RemoteAddr.String())
			if err != nil {
				err = common.NewError("failed to dial to remote endpoint").Base(err)
				logger.Error(err)
				continue
			}
			go proxyConn(tlsConn, remoteConn)
			continue
		}
		if err != nil {
			logger.Error(err)
			continue
		}
		go s.handleConn(tlsConn)
	}
}
