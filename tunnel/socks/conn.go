package socks

import (
	"bytes"
	"context"
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

type packetInfo struct {
	metadata *tunnel.Metadata
	payload  []byte
}

type PacketConn struct {
	net.PacketConn
	input  chan *packetInfo
	output chan *packetInfo
	src    net.Addr
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *PacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	panic("implement me")
}

func (c *PacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	panic("implement me")
}

func (c *PacketConn) Close() error {
	c.cancel()
	return nil
}

func (c *PacketConn) WriteWithMetadata(p []byte, m *tunnel.Metadata) (int, error) {
	select {
	case c.output <- &packetInfo{
		metadata: m,
		payload:  p,
	}:
		return len(p), nil
	case <-c.ctx.Done():
		return 0, common.NewError("socks packet conn closed")
	}
}

func (c *PacketConn) ReadWithMetadata(p []byte) (int, *tunnel.Metadata, error) {
	select {
	case info := <-c.input:
		n := copy(p, info.payload)
		return n, info.metadata, nil
	case <-c.ctx.Done():
		return 0, nil, common.NewError("socks packet conn closed")
	}
}

type WrappedPacketConn struct {
	net.PacketConn
	Destination net.Addr
}

func (c *WrappedPacketConn) Close() error {
	return c.PacketConn.Close()
}

func (c *WrappedPacketConn) WriteWithMetadata(payload []byte, metadata *tunnel.Metadata) (int, error) {
	buf := bytes.NewBuffer(make([]byte, 0, MaxPacketSize))
	buf.Write([]byte{0, 0, 0}) //RSV, FRAG
	common.Must(metadata.Address.WriteTo(buf))
	buf.Write(payload)
	var clientAddr net.Addr
	if c.Destination != nil {
		clientAddr = c.Destination
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

func (c *WrappedPacketConn) ReadWithMetadata(payload []byte) (int, *tunnel.Metadata, error) {
	buf := make([]byte, MaxPacketSize)
	n, from, err := c.PacketConn.ReadFrom(buf)
	if err != nil {
		return 0, nil, err
	}
	log.Debug("recv udp packet from " + from.String())
	addr := new(tunnel.Address)
	c.Destination = from
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
