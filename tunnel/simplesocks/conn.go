package simplesocks

import (
	"bytes"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/trojan"
)

// Conn is a simplesocks connection
type Conn struct {
	tunnel.Conn
	metadata      *tunnel.Metadata
	isOutbound    bool
	headerWritten bool
}

func (c *Conn) Metadata() *tunnel.Metadata {
	return c.metadata
}

func (c *Conn) Write(payload []byte) (int, error) {
	if c.isOutbound && !c.headerWritten {
		buf := bytes.NewBuffer(make([]byte, 0, 4096))
		c.metadata.WriteTo(buf)
		buf.Write(payload)
		_, err := c.Conn.Write(buf.Bytes())
		if err != nil {
			return 0, common.NewError("failed to write simplesocks header").Base(err)
		}
		c.headerWritten = true
		return len(payload), nil
	}
	return c.Conn.Write(payload)
}

// PacketConn is a simplesocks packet connection
// The header syntax is the same as trojan's
type PacketConn struct {
	trojan.PacketConn
}
