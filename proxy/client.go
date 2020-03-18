package proxy

import (
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/protocol/socks"
	"github.com/p4gefau1t/trojan-go/protocol/trojan"
)

type Client struct {
	config *conf.GlobalConfig
	common.Runnable
}

func (c *Client) handleConn(conn net.Conn) {
	inboundConn, err := socks.NewInboundConnSession(conn)
	if err != nil {
		logger.Error("failed to start new inbound session:", err)
		return
	}
	defer inboundConn.Close()
	req := inboundConn.GetRequest()

	if err := inboundConn.(protocol.NeedRespond).Respond(nil); err != nil {
		logger.Error("failed to respond:", err)
		return
	}
	outboundConn, err := trojan.NewOutboundConnSession(req, c.config)
	if err != nil {
		logger.Error("failed to start new outbound session:", err)
		return
	}

	if req.Command == protocol.Associate {
		listenConn, err := net.ListenUDP("udp", &net.UDPAddr{
			IP: c.config.LocalIP,
		})
		if err != nil {
			logger.Error("failed to listen udp:", err)
			return
		}

		req.IP = c.config.LocalIP
		port, err := protocol.ParsePort(listenConn.LocalAddr())
		common.Must(err)
		req.Port = uint16(port)
		req.AddressType = protocol.IPv4

		inboundPacket, err := socks.NewInboundPacketSession(listenConn)
		if err != nil {
			logger.Error("failed to start inbound packet session:", err)
			return
		}
		defer inboundPacket.Close()

		outboundPacket, _ := trojan.NewPacketSession(outboundConn)
		defer outboundPacket.Close()
		go proxyPacket(inboundPacket, outboundPacket)

		logger.Info("UDP associated to", req.String())
		inboundConn.(protocol.NeedRespond).Respond(nil)

		var buf [1]byte
		_, err = conn.Read(buf[:])
		logger.Info("UDP conn ends")
		return
	}

	logger.Info("conn from", conn.RemoteAddr(), "tunneling to", req)
	defer outboundConn.Close()
	proxyConn(inboundConn, outboundConn)
}

func (c *Client) Run() error {
	listener, err := net.Listen("tcp", c.config.LocalAddr.String())
	if err != nil {
		return err
	}
	logger.Info("running client at", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("error occured when accpeting conn", err)
			continue
		}
		go c.handleConn(conn)
	}
}
