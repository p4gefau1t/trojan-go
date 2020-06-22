package memory

import (
	"context"
	"fmt"
	"runtime"
	"strconv"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
)

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
