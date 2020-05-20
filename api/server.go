package api

import (
	"context"
	"crypto/tls"
	"io"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/proxy"
	"github.com/p4gefau1t/trojan-go/stat"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type ServerAPI struct {
	TrojanServerServiceServer
	auth stat.Authenticator
}

func (s *ServerAPI) GetTraffic(stream TrojanServerService_GetTrafficServer) error {
	log.Debug("API: GetTraffic")
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if req.User == nil {
			return common.NewError("User is unspecified")
		}
		if req.User.Hash == "" {
			req.User.Hash = common.SHA224String(req.User.Password)
		}
		valid, meter := s.auth.AuthUser(req.User.Hash)
		if !valid {
			stream.Send(&GetTrafficResponse{
				Success: false,
				Info:    "Invalid user " + req.User.Hash,
			})
			continue
		}
		downloadTraffic, uploadTraffic := meter.Get()
		downloadSpeed, uploadSpeed := meter.GetSpeed()
		downloadSpeedLimit, uploadSpeedLimit := meter.GetSpeedLimit()
		err = stream.Send(&GetTrafficResponse{
			Success: true,
			TrafficTotal: &Traffic{
				UploadTraffic:   uploadTraffic,
				DownloadTraffic: downloadTraffic,
			},
			SpeedCurrent: &Speed{
				DownloadSpeed: downloadSpeed,
				UploadSpeed:   uploadSpeed,
			},
			SpeedLimit: &Speed{
				DownloadSpeed: uint64(downloadSpeedLimit),
				UploadSpeed:   uint64(uploadSpeedLimit),
			},
		})
		if err != nil {
			return err
		}
	}
}

func (s *ServerAPI) SetUsers(stream TrojanServerService_SetUsersServer) error {
	log.Debug("API: SetUsers")
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if req.User == nil {
			return common.NewError("User is unspecified")
		}
		if req.User.Hash == "" {
			req.User.Hash = common.SHA224String(req.User.Password)
		}
		switch req.Operation {
		case SetUserRequest_Add:
			err = s.auth.AddUser(req.User.Hash)
			if req.SpeedLimit != nil {
				valid, meter := s.auth.AuthUser(req.User.Hash)
				if !valid {
					return common.NewError("Failed to add new user")
				}
				meter.LimitSpeed(int(req.SpeedLimit.DownloadSpeed), int(req.SpeedLimit.UploadSpeed))
			}
		case SetUserRequest_Delete:
			err = s.auth.DelUser(req.User.Hash)
		case SetUserRequest_Modify:
			valid, meter := s.auth.AuthUser(req.User.Hash)
			if !valid {
				err = common.NewError("Invalid user " + req.User.Hash)
			} else {
				meter.LimitSpeed(int(req.SpeedLimit.DownloadSpeed), int(req.SpeedLimit.UploadSpeed))
			}
		}
		if err != nil {
			stream.Send(&SetUserResponse{
				Success: false,
				Info:    err.Error(),
			})
			continue
		}
		stream.Send(&SetUserResponse{
			Success: true,
		})
	}
}

func (s *ServerAPI) ListUsers(req *ListUserRequest, stream TrojanServerService_ListUsersServer) error {
	log.Debug("API: ListUsers")
	users := s.auth.ListUsers()
	for _, meter := range users {
		downloadTraffic, uploadTraffic := meter.Get()
		downloadSpeed, uploadSpeed := meter.GetSpeed()
		downloadSpeedLimit, uploadSpeedLimit := meter.GetSpeedLimit()
		online := false
		if downloadSpeed > 0 || uploadSpeed > 0 {
			online = true
		}
		err := stream.Send(&ListUserResponse{
			User: &User{
				Hash: meter.Hash(),
			},
			Online: online,
			TrafficTotal: &Traffic{
				DownloadTraffic: downloadTraffic,
				UploadTraffic:   uploadTraffic,
			},
			SpeedCurrent: &Speed{
				DownloadSpeed: downloadSpeed,
				UploadSpeed:   uploadSpeed,
			},
			SpeedLimit: &Speed{
				DownloadSpeed: uint64(downloadSpeedLimit),
				UploadSpeed:   uint64(uploadSpeedLimit),
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func RunServerAPI(ctx context.Context, config *conf.GlobalConfig, auth stat.Authenticator) error {
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
		log.Warn("Using insecure API service. Please set \"api_tls\" to enable TLS-based gRPC service.")
	}
	service := &ServerAPI{
		auth: auth,
	}
	RegisterTrojanServerServiceServer(server, service)
	listener, err := net.Listen("tcp", config.API.APIAddress.String())
	if err != nil {
		return err
	}
	log.Info("Trojan-Go server-side API service is listening on", config.API.APIAddress)
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
	proxy.RegisterAPI(conf.Server, RunServerAPI)
}
