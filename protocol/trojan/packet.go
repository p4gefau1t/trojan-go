package trojan

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type TrojanPacketSession struct {
	conn io.ReadWriteCloser
}

func (i *TrojanPacketSession) ReadPacket() (*protocol.Request, []byte, error) {
	addr := &common.Address{
		NetworkType: "udp",
	}
	if err := addr.Marshal(i.conn); err != nil {
		return nil, nil, common.NewError("Failed to parse addr").Base(err)
	}
	req := &protocol.Request{
		Address: addr,
	}
	lengthBuf := [2]byte{}
	_, err := io.ReadFull(i.conn, lengthBuf[:])
	if err != nil {
		return req, nil, common.NewError("Failed to read length")
	}
	length := binary.BigEndian.Uint16(lengthBuf[:])

	crlf := [2]byte{}
	io.ReadFull(i.conn, crlf[:])

	packet := make([]byte, length)
	_, err = io.ReadFull(i.conn, packet)
	if err != nil {
		return req, nil, common.NewError("Failed to read payload")
	}
	return req, packet[:], err
}

func (i *TrojanPacketSession) WritePacket(req *protocol.Request, packet []byte) (int, error) {
	buf := bytes.NewBuffer(make([]byte, 0, len(packet)+32))
	common.Must(req.Address.Unmarshal(buf))

	length := len(packet)
	lengthBuf := [2]byte{}
	binary.BigEndian.PutUint16(lengthBuf[:], uint16(length))
	buf.Write(lengthBuf[:])

	crlf := [2]byte{0x0d, 0x0a}
	buf.Write(crlf[:])

	buf.Write(packet)

	return i.conn.Write(buf.Bytes())
}

func (i *TrojanPacketSession) Close() error {
	return i.conn.Close()
}

func NewPacketSession(conn io.ReadWriteCloser) (protocol.PacketSession, error) {
	i := &TrojanPacketSession{
		conn: conn,
	}
	return i, nil
}
