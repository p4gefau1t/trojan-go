package http

import (
	"bufio"
	"bytes"
	"io"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
)

func NewHTTPInbound(conn io.ReadWriteCloser, rw *bufio.ReadWriter) (protocol.ConnSession, protocol.PacketSession, error) {
	var bufReadWriter *bufio.ReadWriter
	if rw == nil {
		bufReadWriter = common.NewBufReadWriter(conn)
	} else {
		bufReadWriter = rw
	}
	method, err := bufReadWriter.Peek(7)
	if err != nil {
		return nil, nil, err
	}
	if bytes.Equal(method, []byte("CONNECT")) {
		i := &HTTPInboundTunnelConnSession{
			bufReadWriter: bufReadWriter,
			conn:          conn,
		}
		if err := i.parseRequest(); err != nil {
			return nil, nil, common.NewError("failed to parse http header").Base(err)
		}
		return i, nil, nil
	}
	i := &HTTPInboundPacketSession{
		bufReadWriter: bufReadWriter,
		conn:          conn,
	}
	return nil, i, nil
}
