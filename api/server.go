package api

import (
	"context"
	"io"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/stat"
	grpc "google.golang.org/grpc"
)

type ServerAPI struct {
	TrojanServerServiceServer
	auth stat.Authenticator
}

func (s *ServerAPI) GetTraffic(stream TrojanServerService_GetTrafficServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if req.User == nil {
			return common.NewError("user is unspecified")
		}
		if req.User.Hash == "" {
			req.User.Hash = common.SHA224String(req.User.Password)
		}
		valid, meter := s.auth.AuthUser(req.User.Hash)
		if !valid {
			stream.Send(&GetTrafficResponse{
				Success: false,
				Info:    "invalid user",
			})
			continue
		}
		downloadTraffic, uploadTraffic := meter.Get()
		downloadSpeed, uploadSpeed := meter.GetSpeed()
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
		})
		if err != nil {
			return err
		}
	}
}

func (s *ServerAPI) SetUsers(stream TrojanServerService_SetUsersServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if req.User == nil {
			return common.NewError("user is unspecified")
		}
		if req.User.Hash == "" {
			req.User.Hash = common.SHA224String(req.User.Password)
		}
		switch req.Operation {
		case SetUserRequest_Add:
			err = s.auth.AddUser(req.User.Hash)
		case SetUserRequest_Delete:
			err = s.auth.DelUser(req.User.Hash)
		case SetUserRequest_Modify:
			err = common.NewError("not support yet")
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
	users := s.auth.ListUsers()
	for _, meter := range users {
		downloadTraffic, uploadTraffic := meter.Get()
		downloadSpeed, uploadSpeed := meter.GetSpeed()
		err := stream.Send(&ListUserResponse{
			User: &User{
				Hash: meter.Hash(),
			},
			TrafficTotal: &Traffic{
				DownloadTraffic: downloadTraffic,
				UploadTraffic:   uploadTraffic,
			},
			SpeedCurrent: &Speed{
				DownloadSpeed: downloadSpeed,
				UploadSpeed:   uploadSpeed,
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func RunServerAPI(ctx context.Context, config *conf.GlobalConfig, auth stat.Authenticator) error {
	server := grpc.NewServer()
	service := &ServerAPI{
		auth: auth,
	}
	RegisterTrojanServerServiceServer(server, service)
	listener, err := net.Listen("tcp", config.API.APIAddress.String())
	if err != nil {
		return err
	}
	log.Info("server api service is running at", config.API.APIAddress)
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
