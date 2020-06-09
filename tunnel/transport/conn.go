package transport

import (
	"net"

	"github.com/p4gefau1t/trojan-go/tunnel"
)

// Conn is the TLS connections
type Conn struct {
	net.Conn
}

// Metadata implements tunnel.Conn. We don't need and metadata here
func (c *Conn) Metadata() *tunnel.Metadata {
	return nil
}
