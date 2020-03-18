package proxy

import (
	"crypto/tls"
	"fmt"
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

func (c *Server) handleConn(conn net.Conn) {
	inboundConn, err := trojan.NewInboundConnSession(conn, c.config)
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
		logger.Info("UDP associated to", req.String())

		inboundConn.(protocol.NeedRespond).Respond(nil)
		proxyPacket(inboundPacket, outboundPacket)
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

func (c *Server) Run() error {
	tlsConfig := &tls.Config{
		Certificates: c.config.TLS.KeyPair,
		CipherSuites: c.config.TLS.CipherSuites,
	}
	listener, err := tls.Listen("tcp", c.config.LocalAddr.String(), tlsConfig)
	if err != nil {
		return err
	}
	fmt.Println("running server at", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error(err)
			continue
		}
		go c.handleConn(conn)
	}
}
