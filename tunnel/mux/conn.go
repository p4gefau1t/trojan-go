package mux

import (
	"io"
	"math/rand"

	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

type stickyConn struct {
	tunnel.Conn
	synQueue chan []byte
	finQueue chan []byte
}

func (c *stickyConn) stickToPayload(p []byte) []byte {
	buf := make([]byte, 0, len(p)+16)
	for {
		select {
		case header := <-c.synQueue:
			buf = append(buf, header...)
		default:
			goto stick1
		}
	}
stick1:
	buf = append(buf, p...)
	for {
		select {
		case header := <-c.finQueue:
			buf = append(buf, header...)
		default:
			goto stick2
		}
	}
stick2:
	return buf
}

func (c *stickyConn) Close() error {
	const maxPaddingLength = 512
	padding := [maxPaddingLength + 8]byte{'A', 'B', 'C', 'D', 'E', 'F'} // for debugging
	buf := c.stickToPayload(nil)
	c.Write(append(buf, padding[:rand.Intn(maxPaddingLength)]...))
	return c.Conn.Close()
}

func (c *stickyConn) Write(p []byte) (int, error) {
	if len(p) == 8 {
		if p[0] == 1 || p[0] == 2 { //smux 8 bytes header
			switch p[1] {
			// THE CONTENT OF THE BUFFER MIGHT CHANGE
			// NEVER STORE THE POINTER TO HEADER, COPY THE HEADER INSTEAD
			case 0:
				// cmdSYN
				header := make([]byte, 8)
				copy(header, p)
				c.synQueue <- header
				return 8, nil
			case 1:
				// cmdFIN
				header := make([]byte, 8)
				copy(header, p)
				c.finQueue <- header
				return 8, nil
			}
		} else {
			log.Debug("Unknown 8 bytes header")
		}
	}
	_, err := c.Conn.Write(c.stickToPayload(p))
	return len(p), err
}

func newStickyConn(conn tunnel.Conn) *stickyConn {
	return &stickyConn{
		Conn:     conn,
		synQueue: make(chan []byte, 128),
		finQueue: make(chan []byte, 128),
	}
}

type Conn struct {
	rwc io.ReadWriteCloser
	tunnel.Conn
}

func (c *Conn) Read(p []byte) (int, error) {
	return c.rwc.Read(p)
}

func (c *Conn) Write(p []byte) (int, error) {
	return c.rwc.Write(p)
}

func (c *Conn) Close() error {
	return c.rwc.Close()
}
