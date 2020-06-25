package dokodemo

import (
	"context"
	"net"

	"github.com/p4gefau1t/trojan-go/common"

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
	M       *tunnel.Metadata
	Input   chan []byte
	Output  chan []byte
	Source  net.Addr
	Context context.Context
	Cancel  context.CancelFunc
}

func (c *PacketConn) Close() error {
	c.Cancel()
	// don't close the underlying udp socket
	return nil
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
	case <-c.Context.Done():
		return 0, nil, common.NewError("dokodemo packet conn closed")
	}
}

func (c *PacketConn) WriteWithMetadata(p []byte, m *tunnel.Metadata) (int, error) {
	select {
	case c.Output <- p:
		return len(p), nil
	case <-c.Context.Done():
		return 0, common.NewError("dokodemo packet conn failed to write")
	}
}
