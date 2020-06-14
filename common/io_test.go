package common

import (
	"bytes"
	"crypto/rand"
	"testing"
	"v2ray.com/core/common"
)

func TestBufferedReader(t *testing.T) {
	payload := [1024]byte{}
	rand.Reader.Read(payload[:])
	rawReader := bytes.NewBuffer(payload[:])
	r := RewindReader{
		rawReader: rawReader,
	}
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
	if !bytes.Equal(payload[:], buf4) {
		t.Fail()
	}
}
