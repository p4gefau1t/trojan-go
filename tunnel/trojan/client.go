package trojan

import (
	"bytes"
	"context"
	"github.com/p4gefau1t/trojan-go/api"
	"github.com/p4gefau1t/trojan-go/statistic/memory"
	"github.com/p4gefau1t/trojan-go/tunnel/mux"
	"net"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/statistic"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

const (
	MaxPacketSize = 1024 * 8
)

const (
	Connect   tunnel.Command = 1
	Associate tunnel.Command = 3
	Mux       tunnel.Command = 0x7f
)

type OutboundConn struct {
	metadata      *tunnel.Metadata
	sent          uint64
	recv          uint64
	auth          statistic.Authenticator
	user          statistic.User
	headerWritten bool
	net.Conn
}

func (c *OutboundConn) Metadata() *tunnel.Metadata {
	return c.metadata
}

func (c *OutboundConn) WriteHeader(payload []byte) error {
	if !c.headerWritten {
		users := c.auth.ListUsers()
		if len(users) == 0 {
			return common.NewError("no password found")
		}
		user := users[0]
		hash := user.Hash()
		c.user = user
		buf := bytes.NewBuffer(make([]byte, 0, MaxPacketSize))
		crlf := []byte{0x0d, 0x0a}
		buf.Write([]byte(hash))
		buf.Write(crlf)
		c.metadata.WriteTo(buf)
		buf.Write(crlf)
		if payload != nil {
			buf.Write(payload)
		}
		_, err := c.Conn.Write(buf.Bytes())
		c.headerWritten = true
		return err
	}
	return common.NewError("header is already written")
}

func (c *OutboundConn) Write(p []byte) (int, error) {
	if !c.headerWritten {
		err := c.WriteHeader(p)
		if err != nil {
			return 0, err
		}
		return len(p), nil
	}
	n, err := c.Conn.Write(p)
	c.user.AddTraffic(n, 0)
	c.sent += uint64(n)
	return n, err
}

func (c *OutboundConn) Read(p []byte) (int, error) {
	n, err := c.Conn.Read(p)
	c.user.AddTraffic(0, n)
	c.recv += uint64(n)
	return n, err
}

func (c *OutboundConn) Close() error {
	log.Info("connection to", c.metadata, "closed", "sent:", common.HumanFriendlyTraffic(c.sent), "recv:", common.HumanFriendlyTraffic(c.recv))
	return c.Conn.Close()
}

type Client struct {
	underlay tunnel.Client
	ctx      context.Context
	auth     statistic.Authenticator
}

func (c *Client) Close() error {
	return c.underlay.Close()
}

func (c *Client) DialConn(addr *tunnel.Address, overlay tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := c.underlay.DialConn(addr, &Tunnel{})
	if err != nil {
		return nil, err
	}
	newConn := &OutboundConn{
		Conn: conn,
		auth: c.auth,
		metadata: &tunnel.Metadata{
			Command: Connect,
			Address: addr,
		},
	}
	if _, ok := overlay.(*mux.Tunnel); ok {
		newConn.metadata.Command = Mux
	}

	go func(newConn *OutboundConn) {
		// if the trojan header is still buffered after 100 ms, the client may expect data from the server
		// so we flush the trojan header
		time.Sleep(time.Millisecond * 100)
		newConn.WriteHeader(nil)
	}(newConn)
	return newConn, nil
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	fakeAddr := &tunnel.Address{
		DomainName:  "UDP_CONN",
		AddressType: tunnel.DomainName,
	}
	conn, err := c.underlay.DialConn(fakeAddr, &Tunnel{})
	if err != nil {
		return nil, err
	}
	return &PacketConn{
		Conn: &OutboundConn{
			Conn: conn,
			auth: c.auth,
			metadata: &tunnel.Metadata{
				Command: Associate,
				Address: fakeAddr,
			},
		},
	}, nil
}

func NewClient(ctx context.Context, client tunnel.Client) (*Client, error) {
	auth, err := statistic.NewAuthenticator(ctx, memory.Name)
	if err != nil {
		return nil, err
	}

	go api.RunService(ctx, Name+"_CLIENT", auth)

	log.Debug("trojan client created")
	return &Client{
		underlay: client,
		ctx:      ctx,
		auth:     auth,
	}, nil
}
