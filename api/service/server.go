package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net"

	"github.com/p4gefau1t/trojan-go/api"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/statistic"
	"github.com/p4gefau1t/trojan-go/tunnel/trojan"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
				Info:    "invalid user: " + req.User.Hash,
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
		if req.Status == nil {
			return common.NewError("status is unspecified")
		}
		if req.Status.User.Hash == "" {
			req.Status.User.Hash = common.SHA224String(req.Status.User.Password)
		}
		switch req.Operation {
		case SetUsersRequest_Add:
			if err = s.auth.AddUser(req.Status.User.Hash); err != nil {
				err = common.NewError("failed to add new user").Base(err)
				break
			}
			if req.Status.SpeedLimit != nil {
				valid, user := s.auth.AuthUser(req.Status.User.Hash)
				if !valid {
					err = common.NewError("failed to auth new user").Base(err)
					continue
				}
				if req.Status.SpeedLimit != nil {
					user.SetSpeedLimit(int(req.Status.SpeedLimit.DownloadSpeed), int(req.Status.SpeedLimit.UploadSpeed))
				}
				if req.Status.TrafficTotal != nil {
					user.SetTraffic(req.Status.TrafficTotal.DownloadTraffic, req.Status.TrafficTotal.UploadTraffic)
				}
				user.SetIPLimit(int(req.Status.IpLimit))
			}
		case SetUsersRequest_Delete:
			err = s.auth.DelUser(req.Status.User.Hash)
		case SetUsersRequest_Modify:
			valid, user := s.auth.AuthUser(req.Status.User.Hash)
			if !valid {
				err = common.NewError("invalid user " + req.Status.User.Hash)
			} else {
				if req.Status.SpeedLimit != nil {
					user.SetSpeedLimit(int(req.Status.SpeedLimit.DownloadSpeed), int(req.Status.SpeedLimit.UploadSpeed))
				}
				if req.Status.TrafficTotal != nil {
					user.SetTraffic(req.Status.TrafficTotal.DownloadTraffic, req.Status.TrafficTotal.UploadTraffic)
				}
				user.SetIPLimit(int(req.Status.IpLimit))
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
			Status: &UserStatus{
				User: &User{
					Hash: user.Hash(),
				},
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

func newAPIServer(cfg *Config) (*grpc.Server, error) {
	var server *grpc.Server
	if cfg.API.SSL.Enabled {
		log.Info("api tls enabled")
		keyPair, err := tls.LoadX509KeyPair(cfg.API.SSL.CertPath, cfg.API.SSL.KeyPath)
		if err != nil {
			return nil, common.NewError("failed to load key pair").Base(err)
		}
		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{keyPair},
		}
		if cfg.API.SSL.VerifyClient {
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
			tlsConfig.ClientCAs = x509.NewCertPool()
			for _, path := range cfg.API.SSL.ClientCertPath {
				log.Debug("loading client cert: " + path)
				certBytes, err := ioutil.ReadFile(path)
				if err != nil {
					return nil, common.NewError("failed to load cert file").Base(err)
				}
				ok := tlsConfig.ClientCAs.AppendCertsFromPEM(certBytes)
				if !ok {
					return nil, common.NewError("invalid client cert")
				}
			}
		}
		creds := credentials.NewTLS(tlsConfig)
		server = grpc.NewServer(grpc.Creds(creds))
	} else {
		server = grpc.NewServer()
	}
	return server, nil
}

func RunServerAPI(ctx context.Context, auth statistic.Authenticator) error {
	cfg := config.FromContext(ctx, Name).(*Config)
	if !cfg.API.Enabled {
		return nil
	}
	service := &ServerAPI{
		auth: auth,
	}
	server, err := newAPIServer(cfg)
	if err != nil {
		return err
	}
	RegisterTrojanServerServiceServer(server, service)
	addr, err := net.ResolveIPAddr("ip", cfg.API.APIHost)
	if err != nil {
		return common.NewError("api found invalid addr").Base(err)
	}
	listener, err := net.Listen("tcp", (&net.TCPAddr{
		IP:   addr.IP,
		Port: cfg.API.APIPort,
		Zone: addr.Zone,
	}).String())
	if err != nil {
		return common.NewError("server api failed to listen").Base(err)
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
