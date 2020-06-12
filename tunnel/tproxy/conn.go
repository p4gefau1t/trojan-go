// +build linux,!386

package tproxy

import (
	"github.com/p4gefau1t/trojan-go/tunnel"
	"net"
)

type Conn struct {
	net.Conn
	metadata *tunnel.Metadata
}

func (c *Conn) Metadata() *tunnel.Metadata {
	return c.metadata
}
