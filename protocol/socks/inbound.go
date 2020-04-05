package socks

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type SocksConnInboundSession struct {
	protocol.ConnSession
	protocol.NeedRespond
	request       *protocol.Request
	conn          io.ReadWriteCloser
	bufReadWriter *bufio.ReadWriter
}

func (i *SocksConnInboundSession) checkVersion() error {
	version, err := i.bufReadWriter.ReadByte()
	if err != nil {
		return err
	}
	if version != 0x5 {
		return common.NewError("unsupported version")
	}
	return nil
}

func (i *SocksConnInboundSession) auth() error {
	if err := i.checkVersion(); err != nil {
		return err
	}
	nmethods, err := i.bufReadWriter.ReadByte()
	if err != nil {
		return err
	}
	i.bufReadWriter.Discard(int(nmethods))
	i.conn.Write([]byte{0x5, 0x0})
	return nil
}

func (i *SocksConnInboundSession) parseRequest() error {
	if err := i.checkVersion(); err != nil {
		return err
	}
	cmd, err := i.bufReadWriter.ReadByte()
	if err != nil {
		return common.NewError("cannot read cmd").Base(err)
	}
	i.bufReadWriter.Discard(1)

	switch protocol.Command(cmd) {
	case protocol.Connect, protocol.Associate:
	default:
		return common.NewError("invalid command")
	}

	request, err := protocol.ParseAddress(i.bufReadWriter)
	if err != nil {
		return common.NewError("cannot read request").Base(err)
	}
	request.Command = protocol.Command(cmd)
	request.NetworkType = "tcp"

	i.request = request
	return nil
}

func (i *SocksConnInboundSession) Respond() error {
	if i.request.Command == protocol.Connect {
		i.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return nil
	} else {
		resp := bytes.NewBuffer([]byte{0x05, 0x00, 0x00})
		common.Must(protocol.WriteAddress(resp, i.request))
		_, err := i.Write(resp.Bytes())
		return err
	}
}

func (i *SocksConnInboundSession) Read(p []byte) (int, error) {
	return i.bufReadWriter.Read(p)
}

func (i *SocksConnInboundSession) Write(p []byte) (int, error) {
	n, err := i.bufReadWriter.Write(p)
	i.bufReadWriter.Flush()
	return n, err
}

func (i *SocksConnInboundSession) Close() error {
	return i.conn.Close()
}

func (i *SocksConnInboundSession) GetRequest() *protocol.Request {
	return i.request
}

func NewInboundConnSession(conn io.ReadWriteCloser, rw *bufio.ReadWriter) (protocol.ConnSession, error) {
	i := &SocksConnInboundSession{}
	i.conn = conn
	if rw == nil {
		i.bufReadWriter = common.NewBufReadWriter(conn)
	} else {
		i.bufReadWriter = rw
	}
	if err := i.auth(); err != nil {
		return nil, err
	}
	if err := i.parseRequest(); err != nil {
		return nil, err
	}
	return i, nil
}

type udpSession struct {
	src    *net.UDPAddr
	req    *protocol.Request
	expire time.Time
}

type SocksInboundPacketSession struct {
	protocol.PacketSession

	conn         *net.UDPConn
	sessionTable map[string]*udpSession
	tableMutex   sync.Mutex
	ctx          context.Context
	cancel       context.CancelFunc
}

func (i *SocksInboundPacketSession) parsePacket(rawPacket []byte) (*protocol.Request, []byte, error) {
	if len(rawPacket) <= 4 {
		return nil, nil, common.NewError("too short")
	}
	buf := bytes.NewBuffer(rawPacket)
	buf.Next(2)
	frag, _ := buf.ReadByte()
	if frag != 0 {
		return nil, nil, common.NewError("fragment is not supported")
	}
	request, err := protocol.ParseAddress(buf)
	if err != nil {
		return nil, nil, common.NewError("cannot parse udp request").Base(err)
	}
	request.NetworkType = "udp"
	return request, buf.Bytes(), nil
}

func (i *SocksInboundPacketSession) writePacketHeader(w io.Writer, req *protocol.Request) error {
	w.Write([]byte{0, 0, 0})
	if err := protocol.WriteAddress(w, req); err != nil {
		return err
	}
	return nil
}

func (i *SocksInboundPacketSession) cleanExpiredSession() {
	for {
		i.tableMutex.Lock()
		now := time.Now()
		for k, v := range i.sessionTable {
			if now.After(v.expire) {
				log.Debug("deleting expired session", v.src, "req:", v.req)
				delete(i.sessionTable, k)
			}
		}
		i.tableMutex.Unlock()
		select {
		case <-time.After(protocol.UDPTimeout):
		case <-i.ctx.Done():
			i.conn.Close()
			return
		}
	}
}

func (i *SocksInboundPacketSession) ReadPacket() (*protocol.Request, []byte, error) {
	buf := make([]byte, protocol.MaxUDPPacketSize)
	i.conn.SetDeadline(time.Now().Add(protocol.UDPTimeout))
	n, src, err := i.conn.ReadFromUDP(buf)
	i.conn.SetDeadline(time.Time{})
	if err != nil {
		return nil, nil, err
	}
	req, payload, err := i.parsePacket(buf[0:n])
	if err != nil {
		return nil, nil, err
	}
	session := &udpSession{
		src:    src,
		req:    req,
		expire: time.Now().Add(protocol.UDPTimeout),
	}
	i.tableMutex.Lock()
	i.sessionTable[req.String()] = session
	i.tableMutex.Unlock()
	log.Debug("UDP read from", src, "req", req)
	return req, payload, err
}

func (i *SocksInboundPacketSession) WritePacket(req *protocol.Request, packet []byte) (int, error) {
	w := bytes.NewBuffer(make([]byte, 0))
	if err := i.writePacketHeader(w, req); err != nil {
		return 0, err
	}
	w.Write(packet)
	i.tableMutex.Lock()
	defer i.tableMutex.Unlock()
	client, found := i.sessionTable[req.String()]
	if !found {
		return 0, common.NewError("session not found")
	}
	client.expire = time.Now().Add(protocol.UDPTimeout)
	log.Debug("UDP write to", client.src, "req", req)
	return i.conn.WriteToUDP(w.Bytes(), client.src)
}

func (i *SocksInboundPacketSession) Close() error {
	i.cancel()
	return i.conn.Close()
}

func NewInboundPacketSession(conn *net.UDPConn) (*SocksInboundPacketSession, error) {
	ctx, cancel := context.WithCancel(context.Background())
	conn.SetWriteBuffer(0)
	i := &SocksInboundPacketSession{
		ctx:          ctx,
		cancel:       cancel,
		sessionTable: make(map[string]*udpSession),
		conn:         conn,
	}
	go i.cleanExpiredSession()
	return i, nil
}
