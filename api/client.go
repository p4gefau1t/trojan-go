package api

import (
	"context"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/stat"
	"google.golang.org/grpc"
	"v2ray.com/core/common/net"
)

type ClientAPIService struct {
	TrojanServiceServer
	meter stat.TrafficMeter
}

func (s *ClientAPIService) QueryStats(ctx context.Context, req *StatsRequest) (*StatsReply, error) {
	log.Debug("query stats, password", req.Password)
	password := req.Password
	passwordHash := common.SHA224String(password)
	sent, recv := s.meter.Query(passwordHash)
	reply := &StatsReply{
		Upload:   sent,
		Download: recv,
	}
	return reply, nil
}

func RunClientAPIService(ctx context.Context, config *conf.GlobalConfig, meter stat.TrafficMeter) error {
	server := grpc.NewServer()
	service := &ClientAPIService{
		meter: meter,
	}
	RegisterTrojanServiceServer(server, service)
	listener, err := net.Listen("tcp", config.API.APIAddress.String())
	if err != nil {
		return err
	}
	log.Info("client api service is running at", config.API.APIAddress)
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Serve(listener)
	}()
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		server.Stop()
		return nil
	}
}
