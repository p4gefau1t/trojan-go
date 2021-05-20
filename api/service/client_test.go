package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/statistic/memory"
)

func TestClientAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = config.WithConfig(ctx, memory.Name,
		&memory.Config{
			Passwords: []string{"useless"},
		})
	port := common.PickPort("tcp", "127.0.0.1")
	ctx = config.WithConfig(ctx, Name, &Config{
		APIConfig{
			Enabled: true,
			APIHost: "127.0.0.1",
			APIPort: port,
		},
	})
	auth, err := memory.NewAuthenticator(ctx)
	common.Must(err)
	go RunClientAPI(ctx, auth)

	time.Sleep(time.Second * 3)
	common.Must(auth.AddUser("hash1234"))
	valid, user := auth.AuthUser("hash1234")
	if !valid {
		t.Fail()
	}
	user.AddTraffic(1234, 5678)
	time.Sleep(time.Second)
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithInsecure())
	common.Must(err)
	client := NewTrojanClientServiceClient(conn)
	resp, err := client.GetTraffic(ctx, &GetTrafficRequest{User: &User{
		Hash: "hash1234",
	}})
	common.Must(err)
	if resp.TrafficTotal.DownloadTraffic != 5678 || resp.TrafficTotal.UploadTraffic != 1234 {
		t.Fail()
	}
	_, err = client.GetTraffic(ctx, &GetTrafficRequest{})
	if err == nil {
		t.Fail()
	}
	cancel()
}
