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
	valid, user := auth.AuthUser("hash")
	if !valid {
		t.Fail()
	}
	user.AddTraffic(1234, 5678)
	sent, recv := user.GetTraffic()
	if sent != 1234 || recv != 5678 {
		t.Fail()
	}
	go func() {
		for i := 0; i < 100; i++ {
			user.AddTraffic(500, 200)
			time.Sleep(time.Millisecond * 100)
		}
	}()

	for i := 0; i < 15; i++ {
		fmt.Println(user.GetSpeed())
		time.Sleep(time.Millisecond * 1000)
	}
	cancel()
}

func TestLimitSpeed(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &conf.GlobalConfig{
		Hash: map[string]string{
			"hash": "password",
		},
	}
	auth, err := NewMemoryAuth(ctx, config)
	common.Must(err)
	valid, user := auth.AuthUser("hash")
	if !valid {
		t.Fail()
	}
	user.SetSpeedLimit(5000, 6000)
	go func() {
		for {
			user.AddTraffic(50, 0)
		}
	}()
	go func() {
		for {
			user.AddTraffic(0, 100)
		}
	}()
	for i := 0; i < 15; i++ {
		fmt.Println(user.GetSpeed())
		time.Sleep(time.Millisecond * 1000)
	}
	cancel()
}

func TestIPLimit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &conf.GlobalConfig{
		Hash: map[string]string{
			"hash": "password",
		},
	}
	auth, err := NewMemoryAuth(ctx, config)
	common.Must(err)
	valid, user := auth.AuthUser("hash")
	if !valid {
		t.Fail()
	}
	user.SetIPLimit(2)
	ok := user.AddIP("ip1")
	if !ok {
		t.Fail()
	}
	ok = user.AddIP("ip2")
	if !ok {
		t.Fail()
	}
	ok = user.AddIP("ip3")
	if ok {
		t.Fail()
	}
	cancel()
}
