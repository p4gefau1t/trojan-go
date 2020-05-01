package memory

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
)

func TestMemoryAuth(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &conf.GlobalConfig{
		Hash: map[string]string{
			"hash": "password",
		},
	}
	auth, err := NewMemoryAuth(ctx, config)
	common.Must(err)
	valid, traffic := auth.AuthUser("hash")
	if !valid {
		t.Fail()
	}
	traffic.Count(1234, 5678)
	sent, recv := traffic.Get()
	if sent != 1234 || recv != 5678 {
		t.Fail()
	}
	go func() {
		for i := 0; i < 100; i++ {
			traffic.Count(500, 200)
			time.Sleep(time.Millisecond * 100)
		}
	}()

	for i := 0; i < 100; i++ {
		fmt.Println(traffic.GetSpeed())
		time.Sleep(time.Millisecond * 1000)
	}
	cancel()
}
