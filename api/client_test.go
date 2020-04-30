package api

import (
	"context"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/stat"
	"google.golang.org/grpc"
)

func TestClientAPI(t *testing.T) {
	meter := &stat.MemoryTrafficMeter{}
	go RunClientAPIService(context.Background(), &conf.GlobalConfig{
		API: conf.APIConfig{
			APIAddress: common.NewAddress("127.0.0.1", 10000, "tcp"),
		},
	}, meter)
	meter.Count("test", 123, 456)
	time.Sleep(time.Second)
	conn, err := grpc.Dial("127.0.0.1:10000", grpc.WithInsecure())
	common.Must(err)
	client := NewTrojanClientServiceClient(conn)
	resp, err := client.GetTraffic(context.Background(), &GetTrafficRequest{})
	common.Must(err)
	if resp.TrafficTotal.DownloadTraffic != 456 || resp.TrafficTotal.UploadTraffic != 123 {
		t.Fail()
	}
}
