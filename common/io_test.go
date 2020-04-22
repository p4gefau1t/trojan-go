package common_test

import (
	"bytes"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
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
		t.Fatal("somthing wrong")
	}
	buf3 := make([]byte, 512)
	common.Must2(r.Read(buf3))
	if !bytes.Equal(buf3, payload[512:]) {
		t.Fatal("somthing wrong")
	}
	r.Rewind()
	buf4 := make([]byte, 1024)
	common.Must2(r.Read(buf4))
	if !bytes.Equal(payload, buf4) {
		t.Fatal("somthing wrong")
	}
}
