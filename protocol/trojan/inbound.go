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
	sent          uint64
	recv          uint64
	passwordHash  string
	ctx           context.Context
	cancel        context.CancelFunc
	readBytes     *bytes.Buffer
}

func (i *TrojanInboundConnSession) Write(p []byte) (int, error) {
	n, err := i.bufReadWriter.Write(p)
	if i.meter != nil {
		i.meter.Count(i.passwordHash, uint64(n), 0)
	}
	i.sent += uint64(n)
	i.bufReadWriter.Flush()
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
	if i.meter != nil {
		i.meter.Count(i.passwordHash, 0, uint64(n))
	}
	i.recv += uint64(n)
	return n, err
}

func (i *TrojanInboundConnSession) Close() error {
	log.Info("user", i.passwordHash, "conn to", i.request, "closed", "sent:", common.HumanFriendlyTraffic(i.sent), "recv:", common.HumanFriendlyTraffic(i.recv))
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
	if err != nil {
		return common.NewError("failed to read cmd").Base(err)
	}

	addr, err := protocol.ParseAddress(i.bufReadWriter, "tcp")
	if err != nil {
		return common.NewError("failed to parse address").Base(err)
	}
	req := &protocol.Request{
		Command: protocol.Command(cmd),
		Address: addr,
	}
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

func NewInboundConnSession(ctx context.Context, conn net.Conn, config *conf.GlobalConfig, auth stat.Authenticator) (protocol.ConnSession, error) {
	ctx, cancel := context.WithCancel(context.Background())
	i := &TrojanInboundConnSession{
		config:        config,
		conn:          conn,
		bufReadWriter: common.NewBufReadWriter(conn),
		auth:          auth,
		passwordHash:  "INVALID_HASH",
		ctx:           ctx,
		cancel:        cancel,
	}
	if i.config.Websocket.Enabled {
		ws, err := NewInboundWebsocket(i.ctx, conn, i.bufReadWriter, config)
		if ws != nil {
			log.Debug("websocket conn")
			i.conn = ws
			i.bufReadWriter = common.NewBufReadWriter(ws)
		}
		if err != nil {
			//no need to continue parsing
			i.request = &protocol.Request{
				Address: config.RemoteAddress,
				Command: protocol.Connect,
			}
			log.Warn("remote", conn.RemoteAddr(), "invalid websocket conn")
			return i, nil
		}
	}
	if err := i.parseRequest(); err != nil {
		i.request = &protocol.Request{
			Address: i.config.RemoteAddress,
			Command: protocol.Connect,
		}
		log.Warn("remote", conn.RemoteAddr(), "invalid hash or other protocol")
	}
	return i, nil
}
