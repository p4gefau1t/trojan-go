package websocket

import (
	"context"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"golang.org/x/net/websocket"
)

type OutboundConn struct {
	*websocket.Conn
}

func (c *OutboundConn) Metadata() *tunnel.Metadata {
	return nil
}

type InboundConn struct {
	OutboundConn
	ctx    context.Context
	cancel context.CancelFunc
}

func (c *InboundConn) Close() error {
	c.cancel()
	return c.Conn.Close()
}
