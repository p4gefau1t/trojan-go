package redirector

import (
	"bytes"
	"context"
	"io"
	"net"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

func TestRedirector(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	redir := NewRedirector(ctx)
	redir.Redirect(&Redirection{
		Dial:        nil,
		RedirectTo:  nil,
		InboundConn: nil,
	})
	var fakeAddr net.Addr
	var fakeConn net.Conn
	redir.Redirect(&Redirection{
		Dial:        nil,
		RedirectTo:  fakeAddr,
		InboundConn: fakeConn,
	})
	l, err := net.Listen("tcp", "127.0.0.1:0")
	common.Must(err)
	conn1, err := net.Dial("tcp", l.Addr().String())
	common.Must(err)
	conn2, err := l.Accept()
	common.Must(err)
	redir.Redirect(&Redirection{
		Dial:        nil,
		RedirectTo:  tunnel.NewAddressFromHostPort("tcp", "127.0.0.1", util.EchoPort),
		InboundConn: conn2,
	})
	payload := util.GeneratePayload(128)
	common.Must2(conn1.Write(payload))
	buf := make([]byte, 128)
	n, err := io.ReadFull(conn2, buf)
	if n != 128 || err != nil {
		t.Fatal(n, err)
	}
	if !bytes.Equal(buf, payload) {
		t.Fatal("diff: ", payload, "\n", buf)
	}
	cancel()
	conn1.Close()
	conn2.Close()
}
