package api

import (
	"context"
	"net"
	"time"

	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/stat"
	"google.golang.org/grpc"
)

type ClientAPI struct {
	TrojanClientServiceServer

	meter         stat.TrafficMeter
	uploadSpeed   uint64
	downloadSpeed uint64
	lastSent      uint64
	lastRecv      uint64
	ctx           context.Context
}

func (s *ClientAPI) GetTraffic(context.Context, *GetTrafficRequest) (*GetTrafficResponse, error) {
	sent, recv := s.meter.Query("")
	resp := &GetTrafficResponse{
		TrafficTotal: &Traffic{
			UploadTraffic:   sent,
			DownloadTraffic: recv,
		},
	}
	return resp, nil
}

func (s *ClientAPI) GetSpeed(context.Context, *GetSpeedRequest) (*GetSpeedResponse, error) {
	resp := &GetSpeedResponse{
		SpeedCurrent: &Speed{
			UploadSpeed:   s.uploadSpeed,
			DownloadSpeed: s.downloadSpeed,
		},
	}
	return resp, nil
}

func (s *ClientAPI) calcSpeed() {
	for {
		select {
		case <-time.After(time.Second):
			// TODO avoid racing
			sent, recv := s.meter.Query("")
			s.uploadSpeed = sent - s.lastSent
			s.downloadSpeed = recv - s.lastRecv
			s.lastSent = sent
			s.lastRecv = recv
		case <-s.ctx.Done():
			return
		}
	}
}

func RunClientAPIService(ctx context.Context, config *conf.GlobalConfig, meter stat.TrafficMeter) error {
	server := grpc.NewServer()
	service := &ClientAPI{
		meter: meter,
		ctx:   ctx,
	}
	go service.calcSpeed()
	RegisterTrojanClientServiceServer(server, service)
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
