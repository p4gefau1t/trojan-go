package tunnel

import (
	"context"
	"io"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
)

// Conn is the TCP connection in the tunnel
type Conn interface {
	net.Conn
	Metadata() *Metadata
}

// PacketConn is the UDP packet stream in the tunnel
type PacketConn interface {
	net.PacketConn
	WriteWithMetadata([]byte, *Metadata) (int, error)
	ReadWithMetadata([]byte) (int, *Metadata, error)
}

// ConnDialer creates TCP connections from the tunnel
type ConnDialer interface {
	DialConn(*Address, Tunnel) (Conn, error)
}

// PacketDialer creates UDP packet stream from the tunnel
type PacketDialer interface {
	DialPacket(Tunnel) (PacketConn, error)
}

// ConnListener accept TCP connections
type ConnListener interface {
	AcceptConn(Tunnel) (Conn, error)
}

// PacketListener accept UDP packet stream
// We don't have any tunnel based on packet streams, so AcceptPacket will always recieve a real PacketConn
type PacketListener interface {
	AcceptPacket(Tunnel) (PacketConn, error)
}

// Dialer can dial to original server with a tunnel
type Dialer interface {
	ConnDialer
	PacketDialer
}

// Listener can accept TCP and UDP streams from a tunnel
type Listener interface {
	ConnListener
	PacketListener
}

// Client is the tunnel client based on stream connections
type Client interface {
	Dialer
	io.Closer
}

// Server is the tunnel server based on stream connections
type Server interface {
	Listener
	io.Closer
}

// Tunnel describes a tunnel, allowing creating a tunnel from another tunnel
// We assume that the lower tunnels know exatly how upper tunnels work, and lower tunnels is transparent for the upper tunnels
type Tunnel interface {
	Name() string
	NewClient(context.Context, Client) (Client, error)
	NewServer(context.Context, Server) (Server, error)
}

var tunnels = make(map[string]Tunnel)

// RegisterTunnel register a tunnel by tunnel name
func RegisterTunnel(name string, tunnel Tunnel) {
	tunnels[name] = tunnel
}

func GetTunnel(name string) (Tunnel, error) {
	if t, ok := tunnels[name]; ok {
		return t, nil
	}
	return nil, common.NewError("unknown tunnel name " + string(name))
}
