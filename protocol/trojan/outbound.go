package trojan

import (
	"bytes"
	"io"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/stat"
)

type TrojanOutboundConnSession struct {
	protocol.ConnSession

	config       *conf.GlobalConfig
	rwc          io.ReadWriteCloser
	request      *protocol.Request
	sent         uint64
	recv         uint64
	auth         stat.Authenticator
	meter        stat.TrafficMeter
	trojanHeader []byte
}

func (o *TrojanOutboundConnSession) SetMeter(meter stat.TrafficMeter) {
	o.meter = meter
}

func (o *TrojanOutboundConnSession) Write(p []byte) (int, error) {
	if o.trojanHeader != nil {
		//glue the payload with the trojan request header
		fullRequest := o.trojanHeader
		fullRequest = append(fullRequest, p...)
		n, err := o.rwc.Write(fullRequest)
		if n >= len(o.trojanHeader) {
			n -= len(o.trojanHeader)
		}
		o.meter.Count(n, 0)
		o.sent += uint64(n)
		o.trojanHeader = nil
		return n, err
	}
	n, err := o.rwc.Write(p)
	o.meter.Count(n, 0)
	o.sent += uint64(n)
	return n, err
}

func (o *TrojanOutboundConnSession) Read(p []byte) (int, error) {
	n, err := o.rwc.Read(p)
	o.meter.Count(0, n)
	o.recv += uint64(n)
	return n, err
}

func (o *TrojanOutboundConnSession) Close() error {
	log.Info("conn to", o.request, "closed", "sent:", common.HumanFriendlyTraffic(o.sent), "recv:", common.HumanFriendlyTraffic(o.recv))
	return o.rwc.Close()
}

func (o *TrojanOutboundConnSession) writeRequest() error {
	user := o.auth.ListUsers()[0]
	hash := user.Hash()
	o.meter = user
	buf := bytes.NewBuffer(make([]byte, 0, 128))
	crlf := []byte("\r\n")
	buf.Write([]byte(hash))
	buf.Write(crlf)
	buf.WriteByte(byte(o.request.Command))
	protocol.WriteAddress(buf, o.request)
	buf.Write(crlf)
	o.trojanHeader = buf.Bytes()
	return nil
}

func NewOutboundConnSession(req *protocol.Request, rwc io.ReadWriteCloser, config *conf.GlobalConfig, auth stat.Authenticator) (protocol.ConnSession, error) {
	o := &TrojanOutboundConnSession{
		request: req,
		config:  config,
		rwc:     rwc,
		auth:    auth,
	}
	if err := o.writeRequest(); err != nil {
		return nil, common.NewError("failed to write request").Base(err)
	}
	return o, nil
}
