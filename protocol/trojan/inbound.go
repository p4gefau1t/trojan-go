package trojan

import (
	"bufio"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type TrojanInboundConnSession struct {
	protocol.ConnSession
	config        *conf.GlobalConfig
	request       *protocol.Request
	bufReadWriter *bufio.ReadWriter
	conn          net.Conn
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
	_, found := i.config.Hash[string(userHash)]
	i.userHash = string(userHash[0:16])
	if !found {
		i.request = &protocol.Request{
			IP:          i.config.RemoteIP,
			Port:        i.config.RemotePort,
			NetworkType: "tcp",
		}
		logger.Warn("invalid hash or other protocol:", string(userHash))
		return nil
	}
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

func NewInboundConnSession(conn net.Conn, config *conf.GlobalConfig) (protocol.ConnSession, error) {
	i := &TrojanInboundConnSession{
		config:        config,
		conn:          conn,
		bufReadWriter: common.NewBufReadWriter(conn),
	}
	if err := i.parseRequest(); err != nil {
		return nil, err
	}
	return i, nil
}
