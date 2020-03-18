package trojan

import (
	"bufio"
	"crypto/tls"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type TrojanOutboundConnSession struct {
	protocol.ConnSession
	config  *conf.GlobalConfig
	conn    net.Conn
	request *protocol.Request
}

func (o *TrojanOutboundConnSession) Write(p []byte) (int, error) {
	return o.conn.Write(p)
}

func (o *TrojanOutboundConnSession) Read(p []byte) (int, error) {
	return o.conn.Read(p)
}

func (o *TrojanOutboundConnSession) Close() error {
	return o.conn.Close()
}

func (o *TrojanOutboundConnSession) writeRequest() error {
	w := bufio.NewWriter(o.conn)
	hash := ""
	for k := range o.config.Hash {
		hash = k
		break
	}
	crlf := []byte("\r\n")
	w.Write([]byte(hash))
	w.Write(crlf)
	w.WriteByte(byte(o.request.Command))
	err := protocol.WriteAddress(w, o.request)
	if err != nil {
		return common.NewError("failed to write address").Base(err)
	}
	w.Write(crlf)
	err = w.Flush()
	return err
}

func NewOutboundConnSession(req *protocol.Request, config *conf.GlobalConfig) (protocol.ConnSession, error) {
	tlsConfig := &tls.Config{
		CipherSuites: config.TLS.CipherSuites,
		RootCAs:      config.TLS.CertPool,
		ServerName:   config.TLS.SNI,
	}
	tlsConn, err := tls.Dial("tcp", config.RemoteAddr.String(), tlsConfig)
	if err != nil {
		return nil, common.NewError("cannot dial to the remote server").Base(err)
	}
	o := &TrojanOutboundConnSession{
		request: req,
		config:  config,
		conn:    tlsConn,
	}
	if err := o.writeRequest(); err != nil {
		return nil, common.NewError("failed to write request").Base(err)
	}
	return o, nil

}
