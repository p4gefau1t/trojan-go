package memory

import (
	"context"
	"testing"

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
	cancel()
}
