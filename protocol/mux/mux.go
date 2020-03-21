package mux

import (
	"bufio"
	"io"
	"os"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/stat"
	"github.com/withmandala/go-log"
)

var logger = log.New(os.Stdout).WithColor()

type MuxConnSession struct {
	protocol.ConnSession
	protocol.NeedMeter
	protocol.HasRequest

	config        *conf.GlobalConfig
	request       *protocol.Request
	bufReadWriter *bufio.ReadWriter
	conn          io.ReadWriteCloser
	passwordHash  string
	meter         stat.TrafficMeter
	recv          int
	sent          int
}

func (m *MuxConnSession) Read(p []byte) (int, error) {
	n, err := m.bufReadWriter.Read(p)
	m.recv += n
	return n, err
}

func (m *MuxConnSession) Write(p []byte) (int, error) {
	n, err := m.bufReadWriter.Write(p)
	m.bufReadWriter.Flush()
	m.sent += n
	return n, err
}

func (m *MuxConnSession) Close() error {
	m.meter.Count(m.passwordHash, m.sent, m.recv)
	logger.Info("mux conn to", m.request, "closed", "sent:", common.HumanFriendlyTraffic(m.sent), "recv:", common.HumanFriendlyTraffic(m.recv))
	return m.conn.Close()
}

func (m *MuxConnSession) SetMeter(meter stat.TrafficMeter) {
	m.meter = meter
}

func (m *MuxConnSession) GetRequest() *protocol.Request {
	return m.request
}

func (m *MuxConnSession) parseRequest() error {
	req, err := protocol.ParseAddress(m.bufReadWriter)
	if err != nil {
		return err
	}
	req.Command = protocol.Connect
	req.NetworkType = "tcp"
	m.request = req
	return nil
}

func (m *MuxConnSession) writeRequest(req *protocol.Request) error {
	err := protocol.WriteAddress(m.bufReadWriter, req)
	if err != nil {
		return err
	}
	m.request = req
	return m.bufReadWriter.Flush()
}

func NewInboundMuxConnSession(conn io.ReadWriteCloser, passwordHash string) (protocol.ConnSession, error) {
	m := &MuxConnSession{
		conn:          conn,
		bufReadWriter: common.NewBufReadWriter(conn),
		meter:         &stat.EmptyTrafficMeter{},
	}
	if err := m.parseRequest(); err != nil {
		return nil, common.NewError("failed to parse mux request").Base(err)
	}
	return m, nil
}

func NewOutboundMuxConnSession(conn io.ReadWriteCloser, req *protocol.Request) (protocol.ConnSession, error) {
	m := &MuxConnSession{
		conn:          conn,
		bufReadWriter: common.NewBufReadWriter(conn),
		meter:         &stat.EmptyTrafficMeter{},
		passwordHash:  "LOCAL_USER",
	}
	if err := m.writeRequest(req); err != nil {
		return nil, common.NewError("failed to write mux request").Base(err)
	}
	return m, nil
}
