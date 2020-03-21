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
	config        *conf.GlobalConfig
	conn          io.ReadWriteCloser
	bufReadWriter *bufio.ReadWriter
	request       *protocol.Request
	sent          int
	recv          int
}

func (o *TrojanOutboundConnSession) Write(p []byte) (int, error) {
	n, err := o.bufReadWriter.Write(p)
	o.bufReadWriter.Flush()
	o.sent += n
	return n, err
}

func (o *TrojanOutboundConnSession) Read(p []byte) (int, error) {
	n, err := o.bufReadWriter.Read(p)
	o.recv += n
	return n, err
}

func (o *TrojanOutboundConnSession) Close() error {
	logger.Info("conn to", o.request, "closed", "sent:", common.HumanFriendlyTraffic(o.sent), "recv:", common.HumanFriendlyTraffic(o.recv))
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
			CipherSuites:           config.TLS.CipherSuites,
			RootCAs:                config.TLS.CertPool,
			ServerName:             config.TLS.SNI,
			InsecureSkipVerify:     !config.TLS.Verify,
			SessionTicketsDisabled: !config.TLS.SessionTicket,
			ClientSessionCache:     tls.NewLRUClientSessionCache(-1),
		}
		tlsConn, err := tls.Dial("tcp", config.RemoteAddr.String(), tlsConfig)
		if err != nil {
			return nil, common.NewError("cannot dial to the remote server").Base(err)
		}
		if config.TLS.VerifyHostname {
			if err := tlsConn.VerifyHostname(config.TLS.SNI); err != nil {
				return nil, common.NewError("failed to verify hostname").Base(err)
			}
		}
		conn = tlsConn
	}
	o := &TrojanOutboundConnSession{
		request:       req,
		config:        config,
		conn:          conn,
		bufReadWriter: common.NewBufReadWriter(conn),
	}
	if err := o.writeRequest(); err != nil {
		return nil, common.NewError("failed to write request").Base(err)
	}
	return o, nil
}
