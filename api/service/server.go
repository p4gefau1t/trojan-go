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
	"io"
	"net"
)

type ServerAPI struct {
	TrojanServerServiceServer
	auth statistic.Authenticator
}

func (s *ServerAPI) GetUsers(stream TrojanServerService_GetUsersServer) error {
	log.Debug("API: GetUsers")
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
		valid, user := s.auth.AuthUser(req.User.Hash)
		if !valid {
			stream.Send(&GetUsersResponse{
				Success: false,
				Info:    "Invalid user: " + req.User.Hash,
			})
			continue
		}
		downloadTraffic, uploadTraffic := user.GetTraffic()
		downloadSpeed, uploadSpeed := user.GetSpeed()
		downloadSpeedLimit, uploadSpeedLimit := user.GetSpeedLimit()
		ipLimit := user.GetIPLimit()
		ipCurrent := user.GetIP()
		err = stream.Send(&GetUsersResponse{
			Success: true,
			Status: &UserStatus{
				User: req.User,
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
				IpCurrent: int32(ipCurrent),
				IpLimit:   int32(ipLimit),
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
		case SetUsersRequest_Add:
			err = s.auth.AddUser(req.User.Hash)
			if req.SpeedLimit != nil {
				valid, user := s.auth.AuthUser(req.User.Hash)
				if !valid {
					return common.NewError("failed to add new user")
				}
				user.SetSpeedLimit(int(req.SpeedLimit.DownloadSpeed), int(req.SpeedLimit.UploadSpeed))
			}
		case SetUsersRequest_Delete:
			err = s.auth.DelUser(req.User.Hash)
		case SetUsersRequest_Modify:
			valid, user := s.auth.AuthUser(req.User.Hash)
			if !valid {
				err = common.NewError("invalid user " + req.User.Hash)
			} else {
				if req.SpeedLimit.DownloadSpeed > 0 || req.SpeedLimit.UploadSpeed > 0 {
					user.SetSpeedLimit(int(req.SpeedLimit.DownloadSpeed), int(req.SpeedLimit.UploadSpeed))
				}
				if req.IpLimit > 0 {
					user.SetIPLimit(int(req.IpLimit))
				}
				if req.TrafficTotal.DownloadTraffic > 0 || req.TrafficTotal.UploadTraffic > 0 {
					user.SetTrafficTotal(req.TrafficTotal.DownloadTraffic, req.TrafficTotal.UploadTraffic)
				}
			}
		}
		if err != nil {
			stream.Send(&SetUsersResponse{
				Success: false,
				Info:    err.Error(),
			})
			continue
		}
		stream.Send(&SetUsersResponse{
			Success: true,
		})
	}
}

func (s *ServerAPI) ListUsers(req *ListUsersRequest, stream TrojanServerService_ListUsersServer) error {
	log.Debug("API: ListUsers")
	users := s.auth.ListUsers()
	for _, user := range users {
		downloadTraffic, uploadTraffic := user.GetTraffic()
		downloadSpeed, uploadSpeed := user.GetSpeed()
		downloadSpeedLimit, uploadSpeedLimit := user.GetSpeedLimit()
		ipLimit := user.GetIPLimit()
		ipCurrent := user.GetIP()
		err := stream.Send(&ListUsersResponse{
			User: &User{
				Hash: user.Hash(),
			},
			Status: &UserStatus{
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
				IpLimit:   int32(ipLimit),
				IpCurrent: int32(ipCurrent),
			},
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func RunServerAPI(ctx context.Context, auth statistic.Authenticator) error {
	cfg := config.FromContext(ctx, Name).(*Config)
	if !cfg.API.Enabled {
		return nil
	}
	server := grpc.NewServer()
	service := &ServerAPI{
		auth: auth,
	}
	RegisterTrojanServerServiceServer(server, service)
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.API.APIHost, cfg.API.APIPort))
	if err != nil {
		return err
	}
	log.Info("server-side api service is listening on", listener.Addr().String())
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
	api.RegisterHandler(trojan.Name+"_SERVER", RunServerAPI)
}
