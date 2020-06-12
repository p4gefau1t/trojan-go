package service

import (
	"context"
	"fmt"
	"github.com/p4gefau1t/trojan-go/api"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/statistic"
	"github.com/p4gefau1t/trojan-go/tunnel/trojan"
	"google.golang.org/grpc"
	"net"
)

type ClientAPI struct {
	TrojanClientServiceServer

	auth          statistic.Authenticator
	ctx           context.Context
	uploadSpeed   uint64
	downloadSpeed uint64
	lastSent      uint64
	lastRecv      uint64
}

func (s *ClientAPI) GetTraffic(ctx context.Context, req *GetTrafficRequest) (*GetTrafficResponse, error) {
	log.Debug("API: GetTraffic")
	if req.User == nil {
		return nil, common.NewError("User is unspecified")
	}
	valid := false
	var user statistic.User
	if req.User.Hash == "" {
		if req.User.Password == "" {
			user = s.auth.ListUsers()[0]
			valid = true
		} else {
			req.User.Hash = common.SHA224String(req.User.Password)
		}
	}
	if !valid {
		valid, user = s.auth.AuthUser(req.User.Hash)
	}
	if !valid {
		return nil, common.NewError("User " + req.User.Hash + " not found")
	}
	sent, recv := user.GetTraffic()
	sentSpeed, recvSpeed := user.GetSpeed()
	resp := &GetTrafficResponse{
		Success: true,
		TrafficTotal: &Traffic{
			UploadTraffic:   sent,
			DownloadTraffic: recv,
		},
		SpeedCurrent: &Speed{
			UploadSpeed:   sentSpeed,
			DownloadSpeed: recvSpeed,
		},
	}
	return resp, nil
}

func RunClientAPI(ctx context.Context, auth statistic.Authenticator) error {
	cfg := config.FromContext(ctx, Name).(*Config)
	if !cfg.API.Enabled {
		return nil
	}
	server := grpc.NewServer()
	service := &ClientAPI{
		ctx:  ctx,
		auth: auth,
	}
	RegisterTrojanClientServiceServer(server, service)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.API.APIHost, cfg.API.APIPort))
	if err != nil {
		return err
	}
	log.Info("client-side api service is listening on", listener.Addr().String())
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

func init() {
	api.RegisterHandler(trojan.Name+"_CLIENT", RunClientAPI)
}
