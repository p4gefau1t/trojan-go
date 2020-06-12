package dokodemo

import (
	"context"
	"io"
	"net"

	"github.com/p4gefau1t/trojan-go/tunnel"
)

const MaxPacketSize = 1024 * 8

type Conn struct {
	net.Conn
	src            *tunnel.Address
	targetMetadata *tunnel.Metadata
}

func (c *Conn) Metadata() *tunnel.Metadata {
	return c.targetMetadata
}

// PacketConn receive packet info from the packet dispatcher
// TODO implement net.PacketConn
type PacketConn struct {
	net.PacketConn
	M      *tunnel.Metadata //fixed
	Input  chan []byte
	Output chan []byte
	Source net.Addr
	Ctx    context.Context
	Cancel context.CancelFunc
}

func (c *PacketConn) Close() error {
	c.Cancel()
	return c.PacketConn.Close()
}

func (c *PacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	return c.ReadWithMetadata(p)
}

func (c *PacketConn) WriteTo(p []byte, addr net.Addr) (int, error) {
	address, err := tunnel.NewAddressFromAddr("udp", addr.String())
	if err != nil {
		return 0, err
	}
	return c.WriteWithMetadata(p, &tunnel.Metadata{
		Address: address,
	})
}

func (c *PacketConn) ReadWithMetadata(p []byte) (int, *tunnel.Metadata, error) {
	select {
	case payload := <-c.Input:
		n := copy(p, payload)
		return n, c.M, nil
	case <-c.Ctx.Done():
		return 0, nil, io.EOF
	}
}

func (c *PacketConn) WriteWithMetadata(p []byte, m *tunnel.Metadata) (int, error) {
	select {
	case c.Output <- p:
	case <-c.Ctx.Done():
		return 0, io.EOF
	}
	return len(p), nil
}
