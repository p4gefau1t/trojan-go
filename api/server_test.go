package api

import (
	context "context"
	"fmt"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/stat/memory"
	grpc "google.golang.org/grpc"
)

func TestServerAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	auth, err := memory.NewMemoryAuth(ctx, &conf.GlobalConfig{})
	common.Must(err)
	go RunServerAPI(ctx, &conf.GlobalConfig{
		API: conf.APIConfig{
			APIAddress: common.NewAddress("127.0.0.1", 10000, "tcp"),
		},
	}, auth)
	common.Must(auth.AddUser("hash1234"))
	_, meter := auth.AuthUser("hash1234")
	conn, err := grpc.Dial("127.0.0.1:10000", grpc.WithInsecure())
	server := NewTrojanServerServiceClient(conn)
	stream1, err := server.ListUsers(ctx, &ListUserRequest{})
	common.Must(err)
	for {
		resp, err := stream1.Recv()
		if err != nil {
			break
		}
		fmt.Println(resp.User.Hash)
		if resp.User.Hash != "hash1234" {
			t.Fail()
		}
	}
	stream1.CloseSend()

	meter.Count(1234, 5678)
	time.Sleep(time.Millisecond * 400)
	stream2, err := server.GetTraffic(ctx)
	common.Must(err)
	stream2.Send(&GetTrafficRequest{
		User: &User{
			Hash: "hash1234",
		},
	})
	resp2, err := stream2.Recv()
	common.Must(err)
	if resp2.TrafficTotal.DownloadTraffic != 1234 || resp2.TrafficTotal.UploadTraffic != 5678 {
		t.Fail()
	}
	if resp2.SpeedCurrent.DownloadSpeed != 1234 || resp2.TrafficTotal.UploadTraffic != 5678 {
		t.Fail()
	}
	stream2.CloseSend()

	stream3, err := server.SetUsers(ctx)
	stream3.Send(&SetUserRequest{
		User: &User{
			Hash: "hash1234",
		},
		Operation: SetUserRequest_Delete,
	})
	resp3, err := stream3.Recv()
	if err != nil || !resp3.Success {
		t.Fail()
	}
	valid, _ := auth.AuthUser("hash1234")
	if valid {
		t.Fail()
	}
	stream3.Send(&SetUserRequest{
		User: &User{
			Hash: "newhash",
		},
		Operation: SetUserRequest_Add,
	})
	resp3, err = stream3.Recv()
	if err != nil || !resp3.Success {
		t.Fail()
	}
	valid, _ = auth.AuthUser("newhash")
	if !valid {
		t.Fail()
	}
	cancel()
}
