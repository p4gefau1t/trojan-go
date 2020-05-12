package common_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	_ "github.com/p4gefau1t/trojan-go/log/golog"
	"github.com/p4gefau1t/trojan-go/test"
)

func TestBufferedReader(t *testing.T) {
	payload := test.GeneratePayload(1024)
	rawReader := bytes.NewBuffer(payload)
	r := common.NewRewindReader(rawReader)
	r.SetBufferSize(2048)
	buf1 := make([]byte, 512)
	buf2 := make([]byte, 512)
	common.Must2(r.Read(buf1))
	r.Rewind()
	common.Must2(r.Read(buf2))
	if !bytes.Equal(buf1, buf2) {
		t.Fail()
	}
	buf3 := make([]byte, 512)
	common.Must2(r.Read(buf3))
	if !bytes.Equal(buf3, payload[512:]) {
		t.Fail()
	}
	r.Rewind()
	buf4 := make([]byte, 1024)
	common.Must2(r.Read(buf4))
	if !bytes.Equal(payload, buf4) {
		t.Fail()
	}
}

func TestCompressReadWriter(t *testing.T) {
	conn := bytes.NewBuffer(make([]byte, 0, 10000))
	//a, err := common.NewCompReadWriter(buf)
	type mockConn struct {
		io.ReadWriter
		io.Closer
	}
	a, err := common.NewCompReadWriteCloser(mockConn{
		ReadWriter: conn,
		Closer:     ioutil.NopCloser(conn),
	})
	common.Must(err)
	payload := []byte{}
	for i := 0; i < 2048; i++ {
		payload = append(payload, 'A')
	}
	common.Must2(a.Write(payload))
	buf := make([]byte, 2048)
	common.Must2(a.Read(buf[:]))
	if !bytes.Equal(buf, payload) {
		t.Fail()
	}
}
