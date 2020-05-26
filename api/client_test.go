package api

import (
	"context"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	_ "github.com/p4gefau1t/trojan-go/log/golog"
	"github.com/p4gefau1t/trojan-go/stat/memory"
	"google.golang.org/grpc"
)

func TestClientAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	auth, err := memory.NewMemoryAuth(ctx, &conf.GlobalConfig{})
	common.Must(err)
	go RunClientAPI(ctx, &conf.GlobalConfig{
		API: conf.APIConfig{
			APIAddress: common.NewAddress("127.0.0.1", 10000, "tcp"),
		},
	}, auth)
	common.Must(auth.AddUser("hash1234"))
	valid, user := auth.AuthUser("hash1234")
	if !valid {
		t.Fail()
	}
	user.AddTraffic(1234, 5678)
	time.Sleep(time.Second)
	conn, err := grpc.Dial("127.0.0.1:10000", grpc.WithInsecure())
	common.Must(err)
	client := NewTrojanClientServiceClient(conn)
	resp, err := client.GetTraffic(ctx, &GetTrafficRequest{User: &User{
		Hash: "hash1234",
	}})
	common.Must(err)
	if resp.TrafficTotal.DownloadTraffic != 5678 || resp.TrafficTotal.UploadTraffic != 1234 {
		t.Fail()
	}
	resp, err = client.GetTraffic(ctx, &GetTrafficRequest{})
	if err == nil {
		t.Fail()
	}
	cancel()
}
