package api

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/proxy"
	"github.com/p4gefau1t/trojan-go/stat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	log.Debug("API: GetTraffic")
	if req.User == nil {
		return nil, common.NewError("User is unspecified")
	}
	if req.User.Hash == "" {
		req.User.Hash = common.SHA224String(req.User.Password)
	}
	valid, user := s.auth.AuthUser(req.User.Hash)
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

func RunClientAPI(ctx context.Context, config *conf.GlobalConfig, auth stat.Authenticator) error {
	var server *grpc.Server
	if config.API.APITLS {
		creds := credentials.NewTLS(&tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: config.TLS.KeyPair,
			ClientCAs:    config.TLS.ClientCertPool,
		})
		server = grpc.NewServer(grpc.Creds(creds))
	} else {
		server = grpc.NewServer()
	}
	service := &ClientAPI{
		ctx:  ctx,
		auth: auth,
	}
	RegisterTrojanClientServiceServer(server, service)
	listener, err := net.Listen("tcp", config.API.APIAddress.String())
	if err != nil {
		return err
	}
	log.Info("Trojan-Go client-side API service is listening on", config.API.APIAddress)
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
	proxy.RegisterAPI(conf.Client, RunClientAPI)
}
