package trojan

import (
	"net"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"golang.org/x/net/context"
)

func TestHTTP2(t *testing.T) {
	ctx := context.Background()
	go func() {
		l, err := net.Listen("tcp", "127.0.0.1:4444")
		common.Must(err)
		conn, err := l.Accept()
		common.Must(err)
		rwc, err := NewH2InboundConn(ctx, conn)
		common.Must(err)
		common.Must2(rwc.Write([]byte("HelloImServer")))
		buf := [256]byte{}
		n, err := rwc.Read(buf[:])
		common.Must(err)
		if string(buf[:n]) != "HelloImClient" {
			t.Fail()
		}
		rwc.Close()
	}()
	time.Sleep(time.Second)
	conn, err := net.Dial("tcp", "127.0.0.1:4444")
	common.Must(err)
	rwc, err := NewH2OutboundConn(ctx, conn)
	buf := [256]byte{}
	n, err := rwc.Read(buf[:])
	if string(buf[:n]) != "HelloImServer" {
		t.Fail()
	}
	common.Must2(rwc.Write([]byte("HelloImClient")))
	rwc.Close()
}
