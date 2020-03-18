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
	config    *conf.GlobalConfig
	request   *protocol.Request
	bufReader *bufio.Reader
	conn      net.Conn
}

func (i *TrojanInboundConnSession) Write(p []byte) (int, error) {
	return i.conn.Write(p)
}

func (i *TrojanInboundConnSession) Read(p []byte) (int, error) {
	return i.bufReader.Read(p)
}

func (i *TrojanInboundConnSession) Close() error {
	return i.conn.Close()
}

func (i *TrojanInboundConnSession) GetRequest() *protocol.Request {
	return i.request
}

func (i *TrojanInboundConnSession) parseRequest() error {
	userHash, err := i.bufReader.Peek(56)
	if err != nil {
		return common.NewError("failed to read hash").Base(err)
	}
	_, found := i.config.Hash[string(userHash)]
	if !found {
		i.request = &protocol.Request{
			IP:          i.config.RemoteIP,
			Port:        i.config.RemotePort,
			NetworkType: "tcp",
		}
		logger.Warn("invalid hash or other protocol", string(userHash))
		return nil
	}
	i.bufReader.Discard(56 + 2)

	cmd, err := i.bufReader.ReadByte()
	network := "tcp"
	switch protocol.Command(cmd) {
	case protocol.Connect:
		network = "tcp"
	case protocol.Associate:
		network = "udp"
	default:
		return common.NewError("invalid command")
	}
	if err != nil {
		return common.NewError("failed to read cmd").Base(err)
	}

	req, err := protocol.ParseAddress(i.bufReader)
	if err != nil {
		return common.NewError("failed to parse address").Base(err)
	}
	req.Command = protocol.Command(cmd)
	req.NetworkType = network
	i.request = req

	i.bufReader.Discard(2)
	return nil
}

func NewInboundConnSession(conn net.Conn, config *conf.GlobalConfig) (protocol.ConnSession, error) {
	i := &TrojanInboundConnSession{
		config:    config,
		conn:      conn,
		bufReader: bufio.NewReader(conn),
	}
	if err := i.parseRequest(); err != nil {
		return nil, err
	}
	return i, nil
}
