package simplesocks

import (
	"bytes"
	"io"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type SimpleSocksConnSession struct {
	protocol.ConnSession

	request *protocol.Request
	rwc     io.ReadWriteCloser
	recv    uint64
	sent    uint64
	header  []byte
}

func (m *SimpleSocksConnSession) Read(p []byte) (int, error) {
	n, err := m.rwc.Read(p)
	m.recv += uint64(n)
	return n, err
}

func (m *SimpleSocksConnSession) Write(p []byte) (int, error) {
	if m.header != nil {
		_, err := m.rwc.Write(append(m.header, p...))
		m.header = nil
		return len(p), err
	}
	n, err := m.rwc.Write(p)
	m.sent += uint64(n)
	return n, err
}

func (m *SimpleSocksConnSession) Close() error {
	log.Info("SimpleSocks conn to", m.request, "closed", "sent:", common.HumanFriendlyTraffic(m.sent), "recv:", common.HumanFriendlyTraffic(m.recv))
	return m.rwc.Close()
}

func (m *SimpleSocksConnSession) GetRequest() *protocol.Request {
	return m.request
}

func (m *SimpleSocksConnSession) parseRequest() error {
	m.request = new(protocol.Request)
	return m.request.Marshal(m.rwc)
}

func (m *SimpleSocksConnSession) writeRequest(req *protocol.Request) {
	m.request = req
	buf := bytes.NewBuffer(make([]byte, 0, 128))
	m.request.Unmarshal(buf)
	m.header = buf.Bytes()
}

func NewInboundConnSession(conn io.ReadWriteCloser) (protocol.ConnSession, *protocol.Request, error) {
	m := &SimpleSocksConnSession{
		rwc: conn,
	}
	if err := m.parseRequest(); err != nil {
		return nil, nil, common.NewError("Failed to parse mux request").Base(err)
	}
	return m, m.request, nil
}

func NewOutboundConnSession(req *protocol.Request, conn io.ReadWriteCloser) (protocol.ConnSession, error) {
	m := &SimpleSocksConnSession{
		rwc: conn,
	}
	m.writeRequest(req)
	return m, nil
}
