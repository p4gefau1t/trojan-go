package api

import (
	"context"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/stat"
	"google.golang.org/grpc"
)

type ClientAPI struct {
	TrojanClientServiceServer

	auth          stat.Authenticator
	ctx           context.Context
	uploadSpeed   uint64
	downloadSpeed uint64
	lastSent      uint64
	lastRecv      uint64
}

func (s *ClientAPI) GetTraffic(ctx context.Context, req *GetTrafficRequest) (*GetTrafficResponse, error) {
	if req.User == nil {
		return nil, common.NewError("user is unspecified")
	}
	valid, meter := s.auth.AuthUser(req.User.Hash)
	if !valid {
		return nil, common.NewError("user " + req.User.Hash + " not found")
	}
	sent, recv := meter.Get()
	resp := &GetTrafficResponse{
		TrafficTotal: &Traffic{
			UploadTraffic:   sent,
			DownloadTraffic: recv,
		},
	}
	return resp, nil
}

func (s *ClientAPI) GetSpeed(ctx context.Context, req *GetSpeedRequest) (*GetSpeedResponse, error) {
	valid, meter := s.auth.AuthUser(req.User.Hash)
	if !valid {
		return &GetSpeedResponse{}, nil
	}
	sent, recv := meter.GetSpeed()
	resp := &GetSpeedResponse{
		SpeedCurrent: &Speed{
			UploadSpeed:   sent,
			DownloadSpeed: recv,
		},
	}
	return resp, nil
}

func RunClientAPIService(ctx context.Context, config *conf.GlobalConfig, auth stat.Authenticator) error {
	server := grpc.NewServer()
	service := &ClientAPI{
		ctx:  ctx,
		auth: auth,
	}
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
