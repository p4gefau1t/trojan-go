package socks

import (
	"bufio"
	"bytes"
	"io"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type SocksConnInboundSession struct {
	protocol.ConnSession
	protocol.NeedRespond
	request   *protocol.Request
	conn      io.ReadWriteCloser
	bufReader *bufio.Reader
}

func (i *SocksConnInboundSession) checkVersion() error {
	version, err := i.bufReader.ReadByte()
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
	nmethods, err := i.bufReader.ReadByte()
	if err != nil {
		return err
	}
	i.bufReader.Discard(int(nmethods))
	i.conn.Write([]byte{0x5, 0x0})
	return nil
}

func (i *SocksConnInboundSession) parseRequest() error {
	if err := i.checkVersion(); err != nil {
		return err
	}
	cmd, err := i.bufReader.ReadByte()
	if err != nil {
		return common.NewError("cannot read cmd").Base(err)
	}
	i.bufReader.Discard(1)

	switch protocol.Command(cmd) {
	case protocol.Connect, protocol.Associate:
	default:
		return common.NewError("invalid command")
	}

	request, err := protocol.ParseAddress(i.bufReader)
	if err != nil {
		return common.NewError("cannot read request").Base(err)
	}
	request.Command = protocol.Command(cmd)
	request.NetworkType = "tcp"

	i.request = request
	return nil
}

func (i *SocksConnInboundSession) Respond(r io.Reader) error {
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
	return i.bufReader.Read(p)
}

func (i *SocksConnInboundSession) Write(p []byte) (int, error) {
	return i.conn.Write(p)
}

func (i *SocksConnInboundSession) Close() error {
	return i.conn.Close()
}

func (i *SocksConnInboundSession) GetRequest() *protocol.Request {
	return i.request
}

func NewInboundConnSession(conn io.ReadWriteCloser) (protocol.ConnSession, error) {
	i := &SocksConnInboundSession{}
	i.conn = conn
	i.bufReader = bufio.NewReader(conn)
	if err := i.auth(); err != nil {
		return nil, err
	}
	if err := i.parseRequest(); err != nil {
		return nil, err
	}
	return i, nil
}

type SocksInboundPacketSession struct {
	protocol.PacketSession
	conn         *net.UDPConn
	socks5Client *net.UDPAddr
}

func (i *SocksInboundPacketSession) parsePacketHeader(rawPacket []byte) (*protocol.Request, []byte, error) {
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

func (i *SocksInboundPacketSession) ReadPacket() (*protocol.Request, []byte, error) {
	buf := make([]byte, protocol.MaxUDPPacketSize)
	n, remote, err := i.conn.ReadFromUDP(buf)
	i.socks5Client = remote
	if err != nil {
		return nil, nil, err
	}
	return i.parsePacketHeader(buf[0:n])
}

func (i *SocksInboundPacketSession) WritePacket(req *protocol.Request, packet []byte) (int, error) {
	w := bytes.NewBuffer(make([]byte, 0))
	if err := i.writePacketHeader(w, req); err != nil {
		return 0, err
	}
	w.Write(packet)
	return i.conn.WriteToUDP(w.Bytes(), i.socks5Client)
}

func (i *SocksInboundPacketSession) Close() error {
	return i.conn.Close()
}

func NewInboundPacketSession(conn *net.UDPConn) (*SocksInboundPacketSession, error) {
	i := &SocksInboundPacketSession{}
	conn.SetWriteBuffer(0)
	i.conn = conn
	return i, nil
}
