package test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/api/service"
	"github.com/p4gefau1t/trojan-go/common"
	"golang.org/x/net/proxy"
	"google.golang.org/grpc"
)

func TestServerAPI(t *testing.T) {
	serverConfig := addAPIConfig(getBasicServerConfig())
	clientConfig := getBasicClientConfig()
	clientConfig.Hash = getHash("apitest")
	clientConfig.Passwords = getPasswords("apitest")

	ctx, cancel := context.WithCancel(context.Background())

	go RunBlackHoleTCPServer(ctx)
	go RunServer(ctx, serverConfig)
	go RunClient(ctx, clientConfig)

	time.Sleep(time.Second * 2)
	grpcConn, err := grpc.Dial("127.0.0.1:10000", grpc.WithInsecure())
	common.Must(err)
	server := service.NewTrojanServerServiceClient(grpcConn)

	listUserStream, err := server.ListUsers(ctx, &service.ListUsersRequest{})
	common.Must(err)
	defer listUserStream.CloseSend()
	for {
		resp, err := listUserStream.Recv()
		if err != nil {
			break
		}
		fmt.Println(resp.User.Hash)
		fmt.Println(*resp.Status.SpeedCurrent)
		fmt.Println(*resp.Status.SpeedLimit)
	}
	listUserStream.CloseSend()
	setUserStream, err := server.SetUsers(ctx)
	setUserStream.Send(&service.SetUsersRequest{
		User: &service.User{
			Hash: common.SHA224String("apitest"),
		},
		SpeedLimit: &service.Speed{
			UploadSpeed: 1024 * 1024 * 2,
		},
		Operation: service.SetUsersRequest_Add,
	})
	resp3, err := setUserStream.Recv()
	if err != nil || !resp3.Success {
		t.Fail()
	}
	setUserStream.CloseSend()

	go func() {
		dialer, err := proxy.SOCKS5("tcp", "127.0.0.1:4444", nil, nil)
		common.Must(err)
		conn, err := dialer.Dial("tcp", "127.0.0.1:5000")
		common.Must(err)
		mbytes := 16
		payload := GeneratePayload(1024 * 1024 * mbytes)
		t1 := time.Now()
		conn.Write(payload)
		t2 := time.Now()
		speed := float64(mbytes) / t2.Sub(t1).Seconds()
		t.Log("single-thread link speed:", speed, "MiB/s")
		conn.Close()
	}()

	time.Sleep(time.Second * 5)
	listUserStream, err = server.ListUsers(ctx, &service.ListUsersRequest{})
	common.Must(err)
	defer listUserStream.CloseSend()
	for {
		resp, err := listUserStream.Recv()
		if err != nil {
			break
		}
		fmt.Println(resp.User.Hash)
		fmt.Println(resp.Status.SpeedCurrent.UploadSpeed)
		fmt.Println(resp.Status.SpeedLimit.UploadSpeed)
	}
	listUserStream.CloseSend()
	cancel()
}
