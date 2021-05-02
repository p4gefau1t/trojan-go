package memory

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
)

func TestMemoryAuth(t *testing.T) {
	cfg := &Config{
		Passwords: nil,
	}
	ctx := config.WithConfig(context.Background(), Name, cfg)
	auth, err := NewAuthenticator(ctx)
	common.Must(err)
	auth.AddUser("user1")
	valid, user := auth.AuthUser("user1")
	if !valid {
		t.Fatal("add, auth")
	}
	if user.Hash() != "user1" {
		t.Fatal("Hash")
	}
	user.AddTraffic(100, 200)
	sent, recv := user.GetTraffic()
	if sent != 100 || recv != 200 {
		t.Fatal("traffic")
	}
	sent, recv = user.ResetTraffic()
	if sent != 100 || recv != 200 {
		t.Fatal("ResetTraffic")
	}
	sent, recv = user.GetTraffic()
	if sent != 0 || recv != 0 {
		t.Fatal("ResetTraffic")
	}

	user.AddIP("1234")
	user.AddIP("5678")
	if user.GetIP() != 0 {
		t.Fatal("GetIP")
	}

	user.SetIPLimit(2)
	user.AddIP("1234")
	user.AddIP("5678")
	user.DelIP("1234")
	if user.GetIP() != 1 {
		t.Fatal("DelIP")
	}
	user.DelIP("5678")

	user.SetIPLimit(2)
	if !user.AddIP("1") || !user.AddIP("2") {
		t.Fatal("AddIP")
	}
	if user.AddIP("3") {
		t.Fatal("AddIP")
	}
	if !user.AddIP("2") {
		t.Fatal("AddIP")
	}

	user.SetTraffic(1234, 4321)
	if a, b := user.GetTraffic(); a != 1234 || b != 4321 {
		t.Fatal("SetTraffic")
	}

	user.ResetTraffic()
	go func() {
		for {
			k := 100
			time.Sleep(time.Second / time.Duration(k))
			user.AddTraffic(200/k, 100/k)
		}
	}()
	time.Sleep(time.Second * 4)
	if sent, recv := user.GetSpeed(); sent > 300 || sent < 100 || recv > 150 || recv < 50 {
		t.Fatal("GetSpeed", sent, recv)
	}

	user.SetSpeedLimit(30, 20)
	time.Sleep(time.Second * 4)
	if sent, recv := user.GetSpeed(); sent > 45 || recv > 30 {
		t.Fatal("SetSpeedLimit", sent, recv)
	}

	user.SetSpeedLimit(0, 0)
	time.Sleep(time.Second * 4)
	if sent, recv := user.GetSpeed(); sent < 30 || recv < 20 {
		t.Fatal("SetSpeedLimit", sent, recv)
	}

	auth.AddUser("user2")
	valid, _ = auth.AuthUser("user2")
	if !valid {
		t.Fatal()
	}
	auth.DelUser("user2")
	valid, _ = auth.AuthUser("user2")
	if valid {
		t.Fatal()
	}
	auth.AddUser("user3")
	users := auth.ListUsers()
	if len(users) != 2 {
		t.Fatal()
	}
	user.Close()
	auth.Close()
}

func BenchmarkMemoryUsage(b *testing.B) {
	cfg := &Config{
		Passwords: nil,
	}
	ctx := config.WithConfig(context.Background(), Name, cfg)
	auth, err := NewAuthenticator(ctx)
	common.Must(err)
	m1 := runtime.MemStats{}
	m2 := runtime.MemStats{}
	runtime.ReadMemStats(&m1)
	for i := 0; i < 100000; i++ {
		err := auth.AddUser(common.SHA224String("hash" + strconv.FormatInt(int64(i), 10)))
		common.Must(err)
	}
	runtime.ReadMemStats(&m2)
	fmt.Println(float64(m2.Alloc-m1.Alloc)/1024/1024, "MiB")
}
