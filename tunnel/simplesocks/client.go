package simplesocks

import (
	"context"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/trojan"
)

const (
	Connect   tunnel.Command = 1
	Associate tunnel.Command = 3
)

type Client struct {
	underlay tunnel.Client
}

func (c *Client) DialConn(addr *tunnel.Address, t tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := c.underlay.DialConn(nil, &Tunnel{})
	if err != nil {
		return nil, common.NewError("simplesocks failed to dial using underlying tunnel").Base(err)
	}
	simplesocks := &Conn{
		Conn:       conn,
		isOutbound: true,
		metadata: &tunnel.Metadata{
			Command: Connect,
			Address: addr,
		},
	}
	_, err = simplesocks.Write(make([]byte, 0)) // send simplesocks header, let the server connect to the destination
	if err != nil {
		conn.Close()
		return nil, common.NewError("simplesocks failed to send header").Base(err)
	}

	return simplesocks, nil
}

func (c *Client) DialPacket(t tunnel.Tunnel) (tunnel.PacketConn, error) {
	conn, err := c.underlay.DialConn(nil, &Tunnel{})
	if err != nil {
		return nil, common.NewError("simplesocks failed to dial using underlying tunnel").Base(err)
	}
	metadata := &tunnel.Metadata{
		Command: Associate,
		Address: &tunnel.Address{
			DomainName:  "UDP_CONN",
			AddressType: tunnel.DomainName,
		},
	}
	if err := metadata.WriteTo(conn); err != nil {
		return nil, common.NewError("simplesocks failed to write udp associate").Base(err)
	}
	return &PacketConn{
		PacketConn: trojan.PacketConn{
			Conn: conn,
		},
	}, nil
}

func (c *Client) Close() error {
	return c.underlay.Close()
}

func NewClient(ctx context.Context, underlay tunnel.Client) (*Client, error) {
	log.Debug("simplesocks client created")
	return &Client{
		underlay: underlay,
	}, nil
}
