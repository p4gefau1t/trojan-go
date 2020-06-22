package http

import (
	"bufio"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/test/util"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
)

func TestHTTP(t *testing.T) {
	port := common.PickPort("tcp", "127.0.0.1")
	ctx := config.WithConfig(context.Background(), transport.Name, &transport.Config{
		LocalHost: "127.0.0.1",
		LocalPort: port,
	})

	tcpServer, err := transport.NewServer(ctx, nil)
	common.Must(err)
	s, err := NewServer(ctx, tcpServer)
	common.Must(err)

	for i := 0; i < 10; i++ {
		go func() {
			http.Get(fmt.Sprintf("http://127.0.0.1:%d", port))
		}()
		conn, err := s.AcceptConn(nil)
		common.Must(err)
		bufReader := bufio.NewReader(bufio.NewReader(conn))
		req, err := http.ReadRequest(bufReader)
		fmt.Println(req)
		ioutil.ReadAll(req.Body)
		req.Body.Close()
		resp, err := http.Get("http://127.0.0.1:" + util.HTTPPort)
		common.Must(err)
		err = resp.Write(conn)
		common.Must(err)
		buf := [100]byte{}
		_, err = conn.Read(buf[:])
		if err == nil {
			t.Fail()
		}
		conn.Close()
	}

	req, err := http.NewRequest(http.MethodConnect, "https://google.com:443", nil)
	common.Must(err)
	conn1, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	common.Must(err)
	go func() {
		common.Must(req.Write(conn1))
	}()

	conn2, err := s.AcceptConn(nil)
	common.Must(err)

	if conn2.Metadata().Port != 443 || conn2.Metadata().DomainName != "google.com" {
		t.Fail()
	}

	connResp := "HTTP/1.1 200 Connection established\r\n\r\n"
	buf := make([]byte, len(connResp))
	_, err = conn1.Read(buf)
	common.Must(err)
	if string(buf) != connResp {
		t.Fail()
	}

	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}

	conn1.Close()
	conn2.Close()
	s.Close()
}
