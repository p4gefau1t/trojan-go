package redirector

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/test/util"
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
	redir.Redirect(&Redirection{
		Dial:        nil,
		RedirectTo:  nil,
		InboundConn: fakeConn,
	})
	redir.Redirect(&Redirection{
		Dial:        nil,
		RedirectTo:  fakeAddr,
		InboundConn: nil,
	})
	l, err := net.Listen("tcp", "127.0.0.1:0")
	common.Must(err)
	conn1, err := net.Dial("tcp", l.Addr().String())
	common.Must(err)
	conn2, err := l.Accept()
	common.Must(err)
	redirAddr, err := net.ResolveTCPAddr("tcp", util.HTTPAddr)
	common.Must(err)
	redir.Redirect(&Redirection{
		Dial:        nil,
		RedirectTo:  redirAddr,
		InboundConn: conn2,
	})
	req, err := http.NewRequest("GET", "http://localhost/", nil)
	common.Must(err)
	req.Write(conn1)
	buf := make([]byte, 1024)
	conn1.Read(buf)
	fmt.Println(string(buf))
	if !strings.HasPrefix(string(buf), "HTTP/1.1 200 OK") {
		t.Fail()
	}
	cancel()
	conn1.Close()
	conn2.Close()
}
