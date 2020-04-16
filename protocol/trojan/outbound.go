package trojan

import (
	"bufio"
	"io"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type TrojanOutboundConnSession struct {
	protocol.ConnSession
	config        *conf.GlobalConfig
	conn          io.ReadWriteCloser
	bufReadWriter *bufio.ReadWriter
	request       *protocol.Request
	sent          int
	recv          int
}

func (o *TrojanOutboundConnSession) Write(p []byte) (int, error) {
	n, err := o.bufReadWriter.Write(p)
	o.bufReadWriter.Flush()
	o.sent += n
	return n, err
}

func (o *TrojanOutboundConnSession) Read(p []byte) (int, error) {
	n, err := o.bufReadWriter.Read(p)
	o.recv += n
	return n, err
}

func (o *TrojanOutboundConnSession) Close() error {
	log.Info("conn to", o.request, "closed", "sent:", common.HumanFriendlyTraffic(o.sent), "recv:", common.HumanFriendlyTraffic(o.recv))
	return o.conn.Close()
}

func (o *TrojanOutboundConnSession) writeRequest() error {
	hash := ""
	for k := range o.config.Hash {
		hash = k
		break
	}
	crlf := []byte("\r\n")
	o.bufReadWriter.Write([]byte(hash))
	o.bufReadWriter.Write(crlf)
	o.bufReadWriter.WriteByte(byte(o.request.Command))
	err := protocol.WriteAddress(o.bufReadWriter, o.request)
	if err != nil {
		return common.NewError("failed to write address").Base(err)
	}
	o.bufReadWriter.Write(crlf)
	return o.bufReadWriter.Flush()
}

func NewOutboundConnSession(req *protocol.Request, conn io.ReadWriteCloser, config *conf.GlobalConfig) (protocol.ConnSession, error) {
	o := &TrojanOutboundConnSession{
		request:       req,
		config:        config,
		conn:          conn,
		bufReadWriter: common.NewBufReadWriter(conn),
	}
	if err := o.writeRequest(); err != nil {
		return nil, common.NewError("failed to write request").Base(err)
	}
	return o, nil
}
