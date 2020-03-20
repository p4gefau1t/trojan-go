package direct

import (
	"bufio"
	"context"
	"io"
	"net"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type DirectOutboundConnSession struct {
	protocol.ConnSession
	conn          io.ReadWriteCloser
	bufReadWriter *bufio.ReadWriter
	request       *protocol.Request
}

func (o *DirectOutboundConnSession) Read(p []byte) (int, error) {
	return o.bufReadWriter.Read(p)
}

func (o *DirectOutboundConnSession) Write(p []byte) (int, error) {
	defer o.bufReadWriter.Flush()
	return o.bufReadWriter.Write(p)
}

func (o *DirectOutboundConnSession) Close() error {
	return o.conn.Close()
}

func NewOutboundConnSession(conn io.ReadWriteCloser, req *protocol.Request) (protocol.ConnSession, error) {
	o := &DirectOutboundConnSession{}
	o.request = req
	if conn == nil {
		newConn, err := net.Dial(req.Network(), req.String())
		if err != nil {
			return nil, err
		}
		o.conn = newConn
		o.bufReadWriter = common.NewBufReadWriter(newConn)
	} else {
		o.conn = conn
	}
	return o, nil
}

type packetInfo struct {
	request *protocol.Request
	packet  []byte
}

type DirectOutboundPacketSession struct {
	protocol.PacketSession
	packetChan chan *packetInfo
	ctx        context.Context
	cancel     context.CancelFunc
}

func (o *DirectOutboundPacketSession) listenConn(req *protocol.Request, conn *net.UDPConn) {
	defer conn.Close()
	for {
		buf := make([]byte, protocol.MaxUDPPacketSize)
		conn.SetReadDeadline(time.Now().Add(protocol.UDPTimeout))
		n, addr, err := conn.ReadFromUDP(buf)
		conn.SetReadDeadline(time.Time{})
		if err != nil {
			logger.Info(err)
			return
		}
		if addr.String() != req.String() {
			panic("addr != req, something went wrong")
		}
		info := &packetInfo{
			request: req,
			packet:  buf[0:n],
		}
		o.packetChan <- info
	}
}

func (o *DirectOutboundPacketSession) Close() error {
	o.cancel()
	return nil
}

func (o *DirectOutboundPacketSession) ReadPacket() (*protocol.Request, []byte, error) {
	select {
	case info := <-o.packetChan:
		return info.request, info.packet, nil
	case <-o.ctx.Done():
		return nil, nil, common.NewError("session closed")
	}
}

func (o *DirectOutboundPacketSession) WritePacket(req *protocol.Request, packet []byte) (int, error) {
	remote := &net.UDPAddr{
		IP:   req.IP,
		Port: int(req.Port),
	}
	conn, err := net.DialUDP("udp", nil, remote)
	go o.listenConn(req, conn)
	if err != nil {
		return 0, common.NewError("cannot dial udp").Base(err)
	}
	logger.Info("UDP directly dialing to", remote)
	n, err := conn.Write(packet)
	return n, err
}

func NewOutboundPacketSession() (protocol.PacketSession, error) {
	ctx, cancel := context.WithCancel(context.Background())
	return &DirectOutboundPacketSession{
		ctx:        ctx,
		cancel:     cancel,
		packetChan: make(chan *packetInfo, 256),
	}, nil
}
