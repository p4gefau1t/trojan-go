package trojan

import (
	"bufio"
	"io"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/stat"
)

type TrojanOutboundConnSession struct {
	protocol.ConnSession
	protocol.NeedMeter

	config        *conf.GlobalConfig
	conn          io.ReadWriteCloser
	bufReadWriter *bufio.ReadWriter
	request       *protocol.Request
	sent          uint64
	recv          uint64
	meter         stat.TrafficMeter
}

func (o *TrojanOutboundConnSession) SetMeter(meter stat.TrafficMeter) {
	o.meter = meter
}

func (o *TrojanOutboundConnSession) Write(p []byte) (int, error) {
	n, err := o.bufReadWriter.Write(p)
	if o.meter != nil {
		o.meter.Count("", uint64(n), 0)
	}
	o.sent += uint64(n)
	o.bufReadWriter.Flush()
	return n, err
}

func (o *TrojanOutboundConnSession) Read(p []byte) (int, error) {
	n, err := o.bufReadWriter.Read(p)
	if o.meter != nil {
		o.meter.Count("", 0, uint64(n))
	}
	o.recv += uint64(n)
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
	protocol.WriteAddress(o.bufReadWriter, o.request)
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
