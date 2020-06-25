package socks

import (
	"bytes"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

type Conn struct {
	net.Conn
	metadata *tunnel.Metadata
}

func (c *Conn) Metadata() *tunnel.Metadata {
	return c.metadata
}

type PacketConn struct {
	net.PacketConn
	srcAddr net.Addr
}

func (c *PacketConn) Close() error {
	return c.PacketConn.Close()
}

func (c *PacketConn) WriteWithMetadata(payload []byte, metadata *tunnel.Metadata) (int, error) {
	buf := bytes.NewBuffer(make([]byte, 0, MaxPacketSize))
	buf.Write([]byte{0, 0, 0}) //RSV, FRAG
	common.Must(metadata.Address.WriteTo(buf))
	buf.Write(payload)
	var clientAddr net.Addr
	if c.srcAddr != nil {
		clientAddr = c.srcAddr
	} else {
		return 0, common.NewError("client address not found")
	}
	_, err := c.PacketConn.WriteTo(buf.Bytes(), clientAddr)
	if err != nil {
		return 0, err
	}
	log.Debug("sent udp packet to " + clientAddr.String() + " with metadata " + metadata.String())
	return len(payload), nil
}

func (c *PacketConn) ReadWithMetadata(payload []byte) (int, *tunnel.Metadata, error) {
	buf := make([]byte, MaxPacketSize)
	n, from, err := c.PacketConn.ReadFrom(buf)
	if err != nil {
		return 0, nil, err
	}
	log.Debug("recv udp packet from " + from.String())
	addr := new(tunnel.Address)
	c.srcAddr = from
	r := bytes.NewBuffer(buf[3:n])
	if err := addr.ReadFrom(r); err != nil {
		return 0, nil, common.NewError("socks5 failed to parse addr in the packet").Base(err)
	}
	length, err := r.Read(payload)
	if err != nil {
		return 0, nil, err
	}
	return length, &tunnel.Metadata{
		Address: addr,
	}, nil
}
