package trojan

import (
	"bufio"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/stat"
)

type TrojanInboundConnSession struct {
	protocol.ConnSession
	protocol.NeedAuth
	protocol.NeedMeter
	config        *conf.GlobalConfig
	request       *protocol.Request
	bufReadWriter *bufio.ReadWriter
	conn          net.Conn
	auth          stat.Authenticator
	meter         stat.TrafficMeter
	uploaded      int
	downloaded    int
	userHash      string
}

func (i *TrojanInboundConnSession) Write(p []byte) (int, error) {
	n, err := i.bufReadWriter.Write(p)
	i.bufReadWriter.Flush()
	i.uploaded += n
	return n, err
}

func (i *TrojanInboundConnSession) Read(p []byte) (int, error) {
	n, err := i.bufReadWriter.Read(p)
	i.downloaded += n
	return n, err
}

func (i *TrojanInboundConnSession) Close() error {
	logger.Info("user", i.userHash, "conn to", i.request, "closed", "up:", common.HumanFriendlyTraffic(i.uploaded), "down:", common.HumanFriendlyTraffic(i.downloaded))
	i.meter.Count(i.userHash, i.uploaded, i.downloaded)
	return i.conn.Close()
}

func (i *TrojanInboundConnSession) GetRequest() *protocol.Request {
	return i.request
}

func (i *TrojanInboundConnSession) parseRequest() error {
	userHash, err := i.bufReadWriter.Peek(56)
	if err != nil {
		return common.NewError("failed to read hash").Base(err)
	}
	if !i.auth.CheckHash(string(userHash)) {
		i.request = &protocol.Request{
			IP:          i.config.RemoteIP,
			Port:        i.config.RemotePort,
			NetworkType: "tcp",
		}
		logger.Warn("invalid hash or other protocol:", string(userHash))
		return nil
	}
	i.userHash = string(userHash)
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
	i := &TrojanInboundConnSession{
		config:        config,
		conn:          conn,
		bufReadWriter: common.NewBufReadWriter(conn),
		meter:         &stat.EmptyTrafficMeter{},
		auth:          auth,
	}
	if err := i.parseRequest(); err != nil {
		return nil, err
	}
	return i, nil
}
