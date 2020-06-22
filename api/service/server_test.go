package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/statistic/memory"
	"google.golang.org/grpc"
)

func TestServerAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ctx = config.WithConfig(ctx, memory.Name,
		&memory.Config{
			Passwords: []string{},
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
	go RunServerAPI(ctx, auth)
	time.Sleep(time.Second * 3)
	common.Must(auth.AddUser("hash1234"))
	_, user := auth.AuthUser("hash1234")
	conn, err := grpc.Dial(fmt.Sprintf("127.0.0.1:%d", port), grpc.WithInsecure())
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
	stream3.Send(&SetUsersRequest{
		User: &User{
			Hash: "hash1234",
		},
		Operation: SetUsersRequest_Delete,
	})
	resp3, err := stream3.Recv()
	if err != nil || !resp3.Success {
		t.Fail()
	}
	valid, _ := auth.AuthUser("hash1234")
	if valid {
		t.Fail()
	}
	stream3.Send(&SetUsersRequest{
		User: &User{
			Hash: "newhash",
		},
		Operation: SetUsersRequest_Add,
	})
	resp3, err = stream3.Recv()
	if err != nil || !resp3.Success {
		t.Fail()
	}
	valid, user = auth.AuthUser("newhash")
	if !valid {
		t.Fail()
	}
	stream3.Send(&SetUsersRequest{
		User: &User{
			Hash: "newhash",
		},
		Operation: SetUsersRequest_Modify,
		SpeedLimit: &Speed{
			DownloadSpeed: 5000,
			UploadSpeed:   3000,
		},
		TrafficTotal: &Traffic{
			DownloadTraffic: 1,
			UploadTraffic:   1,
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
