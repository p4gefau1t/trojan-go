package trojan

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/stat"
)

type TrojanInboundConnSession struct {
	protocol.ConnSession
	protocol.NeedAuth
	protocol.NeedMeter
	protocol.HasHash

	config        *conf.GlobalConfig
	request       *protocol.Request
	bufReadWriter *bufio.ReadWriter
	conn          io.ReadWriteCloser
	auth          stat.Authenticator
	meter         stat.TrafficMeter
	sent          int
	recv          int
	passwordHash  string
	ctx           context.Context
	cancel        context.CancelFunc
	readBytes     *bytes.Buffer
}

func (i *TrojanInboundConnSession) Write(p []byte) (int, error) {
	n, err := i.bufReadWriter.Write(p)
	i.bufReadWriter.Flush()
	i.sent += n
	return n, err
}

func (i *TrojanInboundConnSession) Read(p []byte) (int, error) {
	if i.readBytes != nil {
		n, err := i.readBytes.Read(p)
		if err == io.EOF {
			i.readBytes = nil
		}
		return n, err
	}
	n, err := i.bufReadWriter.Read(p)
	i.recv += n
	return n, err
}

func (i *TrojanInboundConnSession) Close() error {
	log.Info("user", i.passwordHash, "conn to", i.request, "closed", "sent:", common.HumanFriendlyTraffic(i.sent), "recv:", common.HumanFriendlyTraffic(i.recv))
	i.meter.Count(i.passwordHash, i.sent, i.recv)
	i.cancel()
	return i.conn.Close()
}

func (i *TrojanInboundConnSession) GetRequest() *protocol.Request {
	return i.request
}

func (i *TrojanInboundConnSession) GetHash() string {
	return i.passwordHash
}

func (i *TrojanInboundConnSession) parseRequest() error {
	userHash, err := i.bufReadWriter.Peek(56)
	if err != nil {
		return common.NewError("failed to read hash").Base(err)
	}
	if !i.auth.CheckHash(string(userHash)) {
		return common.NewError("invalid hash:" + string(userHash))
	}
	i.passwordHash = string(userHash)
	i.bufReadWriter.Discard(56 + 2)

	cmd, err := i.bufReadWriter.ReadByte()
	network := "tcp"
	switch protocol.Command(cmd) {
	case protocol.Connect, protocol.Mux:
		network = "tcp"
	case protocol.Associate:
		network = "udp"
	default:
		return common.NewError("invalid command")
	}
	if err != nil {
		return common.NewError("failed to read cmd").Base(err)
	}

	req, err := protocol.ParseAddress(i.bufReadWriter)
	if err != nil {
		return common.NewError("failed to parse address").Base(err)
	}
	req.Command = protocol.Command(cmd)
	req.NetworkType = network
	i.request = req

	i.bufReadWriter.Discard(2)
	return nil
}

func (i *TrojanInboundConnSession) SetAuth(auth stat.Authenticator) {
	i.auth = auth
}

func (i *TrojanInboundConnSession) SetMeter(meter stat.TrafficMeter) {
	i.meter = meter
}

func NewInboundConnSession(conn net.Conn, config *conf.GlobalConfig, auth stat.Authenticator) (protocol.ConnSession, error) {
	ctx, cancel := context.WithCancel(context.Background())
	i := &TrojanInboundConnSession{
		config:        config,
		conn:          conn,
		bufReadWriter: common.NewBufReadWriter(conn),
		meter:         &stat.EmptyTrafficMeter{},
		auth:          auth,
		passwordHash:  "INVALID_HASH",
		ctx:           ctx,
		cancel:        cancel,
	}
	if i.config.Websocket.Enabled {
		ws, err := NewInboundWebsocket(conn, i.bufReadWriter, i.ctx, config)
		if ws != nil {
			log.Debug("websocket conn")
			i.conn = ws
			i.bufReadWriter = common.NewBufReadWriter(ws)
		}
		if err != nil {
			//no need to continue parsing
			i.request = &protocol.Request{
				IP:          i.config.RemoteIP,
				Port:        i.config.RemotePort,
				NetworkType: "tcp",
			}
			log.Warn("remote", conn.RemoteAddr(), "invalid websocket conn")
			return i, nil
		}
	}
	if err := i.parseRequest(); err != nil {
		i.request = &protocol.Request{
			IP:          i.config.RemoteIP,
			Port:        i.config.RemotePort,
			NetworkType: "tcp",
		}
		log.Warn("remote", conn.RemoteAddr(), "invalid hash or other protocol")
	}
	return i, nil
}
