// +build linux,!386

package tproxy

import (
	"net"

	"github.com/p4gefau1t/trojan-go/tunnel"
)

type Conn struct {
	net.Conn
	metadata *tunnel.Metadata
}

func (c *Conn) Metadata() *tunnel.Metadata {
	return c.metadata
}
