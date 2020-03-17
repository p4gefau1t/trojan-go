package proxy

import (
	"log"
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
		log.Println(err)
		return
	}
	defer inboundConn.Close()
	req := inboundConn.GetRequest()

	if err := inboundConn.(protocol.NeedRespond).Respond(nil); err != nil {
		log.Println(err)
		return
	}
	outboundConn, err := trojan.NewOutboundConnSession(req, c.config)
	if err != nil {
		log.Println(err)
		return
	}

	if req.Command == protocol.Associate {
		listenConn, err := net.ListenUDP("udp", &net.UDPAddr{
			IP: c.config.LocalIP,
		})
		if err != nil {
			log.Println(err)
			return
		}

		req.IP = c.config.LocalIP
		port, err := protocol.ParsePort(listenConn.LocalAddr())
		common.Must(err)
		req.Port = uint16(port)
		req.AddressType = protocol.IPv4

		inboundPacket, err := socks.NewInboundPacketSession(listenConn)
		if err != nil {
			log.Println(err)
			return
		}
		defer inboundPacket.Close()

		if err != nil {
			log.Println(err)
			return
		}

		outboundPacket, _ := trojan.NewPacketSession(outboundConn)
		defer outboundPacket.Close()
		go proxyPacket(inboundPacket, outboundPacket)

		log.Println("UDP associated to", req.String())
		inboundConn.(protocol.NeedRespond).Respond(nil)

		var buf [1]byte
		n, err := conn.Read(buf[:])
		log.Println("UDP association ends", err, n)
		return
	}

	log.Println("conn from", conn.RemoteAddr(), "tunneling to", req.String())
	defer outboundConn.Close()
	proxyConn(inboundConn, outboundConn)
}

func (c *Client) Run() error {
	listener, err := net.Listen("tcp", c.config.LocalAddr.String())
	if err != nil {
		return err
	}
	log.Println("running client at", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go c.handleConn(conn)
	}
}
