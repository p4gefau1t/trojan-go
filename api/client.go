package api

import (
	"context"
	"time"

	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/stat"
	"google.golang.org/grpc"
	"v2ray.com/core/common/net"
)

type ClientAPIService struct {
	TrojanServiceServer
	meter         stat.TrafficMeter
	uploadSpeed   uint64
	downloadSpeed uint64
	lastSent      uint64
	lastRecv      uint64
	ctx           context.Context
}

func (s *ClientAPIService) QueryStats(ctx context.Context, req *StatsRequest) (*StatsReply, error) {
	log.Debug("query stats, password", req.Password)
	//password := req.Password
	//passwordHash := common.SHA224String(password)
	sent, recv := s.meter.Query("")
	reply := &StatsReply{
		UploadTraffic:   sent,
		DownloadTraffic: recv,
		UploadSpeed:     s.uploadSpeed,
		DownloadSpeed:   s.downloadSpeed,
	}
	return reply, nil
}

func (s *ClientAPIService) calcSpeed() {
	select {
	case <-time.After(time.Second):
		sent, recv := s.meter.Query("")
		s.uploadSpeed = sent - s.lastSent
		s.downloadSpeed = recv - s.lastRecv
		s.lastSent = sent
		s.lastRecv = recv
	case <-s.ctx.Done():
		return
	}
}

func RunClientAPIService(ctx context.Context, config *conf.GlobalConfig, meter stat.TrafficMeter) error {
	server := grpc.NewServer()
	service := &ClientAPIService{
		meter: meter,
		ctx:   ctx,
	}
	go service.calcSpeed()
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
