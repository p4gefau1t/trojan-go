package api

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	_ "github.com/p4gefau1t/trojan-go/log/golog"
	"github.com/p4gefau1t/trojan-go/stat/memory"
	"google.golang.org/grpc"
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
	_, user := auth.AuthUser("hash1234")
	conn, err := grpc.Dial("127.0.0.1:10000", grpc.WithInsecure())
	common.Must(err)
	server := NewTrojanServerServiceClient(conn)
	stream1, err := server.ListUsers(ctx, &ListUsersRequest{})
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
		fmt.Println(resp.Status.SpeedCurrent)
		fmt.Println(resp.Status.SpeedLimit)
	}
	stream1.CloseSend()
	user.AddTraffic(1234, 5678)
	time.Sleep(time.Millisecond * 1000)
	stream2, err := server.GetUsers(ctx)
	common.Must(err)
	stream2.Send(&GetUsersRequest{
		User: &User{
			Hash: "hash1234",
		},
	})
	resp2, err := stream2.Recv()
	common.Must(err)
	if resp2.Status.TrafficTotal.DownloadTraffic != 1234 || resp2.Status.TrafficTotal.UploadTraffic != 5678 {
		t.Fail()
	}
	if resp2.Status.SpeedCurrent.DownloadSpeed != 1234 || resp2.Status.TrafficTotal.UploadTraffic != 5678 {
		t.Fail()
	}

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
	valid, user = auth.AuthUser("newhash")
	if !valid {
		t.Fail()
	}
	stream3.Send(&SetUserRequest{
		User: &User{
			Hash: "newhash",
		},
		Operation: SetUserRequest_Modify,
		SpeedLimit: &Speed{
			DownloadSpeed: 5000,
			UploadSpeed:   3000,
		},
	})
	go func() {
		for {
			user.AddTraffic(200, 0)
		}
	}()
	go func() {
		for {
			user.AddTraffic(0, 300)
		}
	}()
	time.Sleep(time.Second * 3)
	for i := 0; i < 3; i++ {
		stream2.Send(&GetUsersRequest{
			User: &User{
				Hash: "newhash",
			},
		})
		resp2, err = stream2.Recv()
		fmt.Println(resp2.Status.SpeedCurrent)
		fmt.Println(resp2.Status.SpeedLimit)
		time.Sleep(time.Second)
	}
	stream2.CloseSend()
	cancel()
}
