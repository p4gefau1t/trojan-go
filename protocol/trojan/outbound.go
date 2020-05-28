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
	config  *conf.GlobalConfig
	rwc     io.ReadWriteCloser
	request *protocol.Request
	sent    uint64
	recv    uint64
	auth    stat.Authenticator
	meter   stat.TrafficMeter
	header  []byte
}

func (o *TrojanOutboundConnSession) Write(p []byte) (int, error) {
	if o.header != nil {
		//send the payload after the trojan request header
		_, err := o.rwc.Write(append(o.header, p...))
		o.meter.AddTraffic(len(p)+len(o.header), 0)
		o.sent += uint64(len(p) + len(o.header))
		o.header = nil
		return len(p), err
	}
	n, err := o.rwc.Write(p)
	o.meter.AddTraffic(n, 0)
	o.sent += uint64(n)
	return n, err
}

func (o *TrojanOutboundConnSession) Read(p []byte) (int, error) {
	n, err := o.rwc.Read(p)
	o.meter.AddTraffic(0, n)
	o.recv += uint64(n)
	return n, err
}

func (o *TrojanOutboundConnSession) Close() error {
	log.Info("Conn to", o.request, "closed", "sent:", common.HumanFriendlyTraffic(o.sent), "recv:", common.HumanFriendlyTraffic(o.recv))
	return o.rwc.Close()
}

func (o *TrojanOutboundConnSession) writeRequest() {
	user := o.auth.ListUsers()[0]
	hash := user.Hash()
	o.meter = user
	buf := bytes.NewBuffer(make([]byte, 0, 128))
	crlf := []byte{0x0d, 0x0a}
	buf.Write([]byte(hash))
	buf.Write(crlf)
	o.request.Unmarshal(buf)
	buf.Write(crlf)
	o.header = buf.Bytes()
}

func NewOutboundConnSession(req *protocol.Request, rwc io.ReadWriteCloser, config *conf.GlobalConfig, auth stat.Authenticator) (protocol.ConnSession, error) {
	o := &TrojanOutboundConnSession{
		request: req,
		config:  config,
		rwc:     rwc,
		auth:    auth,
	}
	o.writeRequest()
	return o, nil
}
