package trojan

import (
	"bufio"
	"crypto/tls"
	"io"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type TrojanOutboundConnSession struct {
	protocol.ConnSession
	config     *conf.GlobalConfig
	conn       io.ReadWriteCloser
	request    *protocol.Request
	uploaded   int
	downloaded int
}

func (o *TrojanOutboundConnSession) Write(p []byte) (int, error) {
	n, err := o.conn.Write(p)
	o.uploaded += n
	return n, err
}

func (o *TrojanOutboundConnSession) Read(p []byte) (int, error) {
	n, err := o.conn.Read(p)
	o.downloaded += n
	return n, err
}

func (o *TrojanOutboundConnSession) Close() error {
	logger.Info("conn to", o.request, "closed", "up:", common.HumanFriendlyTraffic(o.uploaded), "down:", common.HumanFriendlyTraffic(o.downloaded))
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

func NewOutboundConnSession(req *protocol.Request, conn io.ReadWriteCloser, config *conf.GlobalConfig) (protocol.ConnSession, error) {
	if conn == nil {
		tlsConfig := &tls.Config{
			CipherSuites: config.TLS.CipherSuites,
			RootCAs:      config.TLS.CertPool,
			ServerName:   config.TLS.SNI,
		}
		tlsConn, err := tls.Dial("tcp", config.RemoteAddr.String(), tlsConfig)
		if err != nil {
			return nil, common.NewError("cannot dial to the remote server").Base(err)
		}
		conn = tlsConn
	}
	o := &TrojanOutboundConnSession{
		request: req,
		config:  config,
		conn:    conn,
	}
	if err := o.writeRequest(); err != nil {
		return nil, common.NewError("failed to write request").Base(err)
	}
	return o, nil
}
