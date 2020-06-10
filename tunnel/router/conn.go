package router

import (
	"context"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"io"
	"net"
)

type packetInfo struct {
	src     *tunnel.Metadata
	payload []byte
}

type PacketConn struct {
	proxy tunnel.PacketConn
	net.PacketConn
	packetChan chan *packetInfo
	*Client
	context.Context
	context.CancelFunc
}

func (c *PacketConn) packetLoop() {
	go func() {
		for {
			buf := make([]byte, MaxPacketSize)
			n, addr, err := c.proxy.ReadWithMetadata(buf)
			if err != nil {
				select {
				case <-c.Done():
					return
				default:
					log.Error("router packetConn error", err)
					continue
				}
			}
			c.packetChan <- &packetInfo{
				src:     addr,
				payload: buf[:n],
			}
		}
	}()
	for {
		buf := make([]byte, MaxPacketSize)
		n, addr, err := c.PacketConn.ReadFrom(buf)
		if err != nil {
			select {
			case <-c.Done():
				return
			default:
				log.Error("router packetConn error", err)
				continue
			}
		}
		address, err := tunnel.NewAddressFromAddr("udp", addr.String())
		c.packetChan <- &packetInfo{
			src: &tunnel.Metadata{
				Address: address,
			},
			payload: buf[:n],
		}
	}
}

func (c *PacketConn) Close() error {
	c.CancelFunc()
	c.proxy.Close()
	return c.PacketConn.Close()
}

func (c *PacketConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	panic("implement me")
}

func (c *PacketConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	panic("implement me")
}

func (c *PacketConn) WriteWithMetadata(p []byte, m *tunnel.Metadata) (int, error) {
	policy := c.Route(m.Address)
	switch policy {
	case Proxy:
		return c.proxy.WriteWithMetadata(p, m)
	case Block:
		return 0, common.NewError("router blocked address (udp): " + m.Address.String())
	case Bypass:
		ip, err := m.Address.ResolveIP()
		if err != nil {
			return 0, common.NewError("router failed to resolve udp address").Base(err)
		}
		return c.PacketConn.WriteTo(p, &net.UDPAddr{
			IP:   ip,
			Port: m.Address.Port,
		})
	default:
		panic("unknown policy")
	}
}

func (c *PacketConn) ReadWithMetadata(p []byte) (int, *tunnel.Metadata, error) {
	select {
	case info := <-c.packetChan:
		n := copy(p, info.payload)
		return n, info.src, nil
	case <-c.Done():
		return 0, nil, io.EOF
	}
}
