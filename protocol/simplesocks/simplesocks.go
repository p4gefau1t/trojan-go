package simplesocks

import (
	"bytes"
	"io"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type SimpleSocksConnSession struct {
	protocol.ConnSession

	config  *conf.GlobalConfig
	request *protocol.Request
	rwc     io.ReadWriteCloser
	recv    uint64
	sent    uint64
}

func (m *SimpleSocksConnSession) Read(p []byte) (int, error) {
	n, err := m.rwc.Read(p)
	m.recv += uint64(n)
	return n, err
}

func (m *SimpleSocksConnSession) Write(p []byte) (int, error) {
	n, err := m.rwc.Write(p)
	m.sent += uint64(n)
	return n, err
}

func (m *SimpleSocksConnSession) Close() error {
	log.Info("simplesocks conn to", m.request, "closed", "sent:", common.HumanFriendlyTraffic(m.sent), "recv:", common.HumanFriendlyTraffic(m.recv))
	return m.rwc.Close()
}

func (m *SimpleSocksConnSession) GetRequest() *protocol.Request {
	return m.request
}

func (m *SimpleSocksConnSession) parseRequest() error {
	cmd, err := common.ReadByte(m.rwc)
	if err != nil {
		return common.NewError("failed to read cmd").Base(err)
	}
	addr, err := protocol.ParseAddress(m.rwc, "tcp")
	if err != nil {
		return common.NewError("failed to parse addr").Base(err)
	}
	req := &protocol.Request{
		Address: addr,
		Command: protocol.Command(cmd),
	}
	m.request = req
	return nil
}

func (m *SimpleSocksConnSession) writeRequest(req *protocol.Request) error {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	common.Must(buf.WriteByte(byte(req.Command)))
	common.Must(protocol.WriteAddress(buf, req))
	m.request = req
	_, err := m.rwc.Write(buf.Bytes())
	return err
}

func NewInboundConnSession(conn io.ReadWriteCloser) (protocol.ConnSession, *protocol.Request, error) {
	m := &SimpleSocksConnSession{
		rwc: conn,
	}
	if err := m.parseRequest(); err != nil {
		return nil, nil, common.NewError("failed to parse mux request").Base(err)
	}
	return m, m.request, nil
}

func NewOutboundConnSession(req *protocol.Request, conn io.ReadWriteCloser) (protocol.ConnSession, error) {
	m := &SimpleSocksConnSession{
		rwc: conn,
	}
	if err := m.writeRequest(req); err != nil {
		return nil, common.NewError("failed to write mux request").Base(err)
	}
	return m, nil
}
