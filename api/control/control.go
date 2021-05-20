package control

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"google.golang.org/grpc"

	"github.com/p4gefau1t/trojan-go/api/service"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/option"
)

type apiController struct {
	address *string
	key     *string
	hash    *string
	cert    *string

	cmd                *string
	password           *string
	add                *bool
	delete             *bool
	modify             *bool
	list               *bool
	uploadSpeedLimit   *int
	downloadSpeedLimit *int
	ipLimit            *int
	ctx                context.Context
}

func (apiController) Name() string {
	return "api"
}

func (o *apiController) listUsers(apiClient service.TrojanServerServiceClient) error {
	stream, err := apiClient.ListUsers(o.ctx, &service.ListUsersRequest{})
	if err != nil {
		return err
	}
	defer stream.CloseSend()
	result := []*service.ListUsersResponse{}
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		result = append(result, resp)
	}
	data, err := json.Marshal(result)
	common.Must(err)
	fmt.Println(string(data))
	return nil
}

func (o *apiController) getUsers(apiClient service.TrojanServerServiceClient) error {
	stream, err := apiClient.GetUsers(o.ctx)
	if err != nil {
		return err
	}
	defer stream.CloseSend()
	err = stream.Send(&service.GetUsersRequest{
		User: &service.User{
			Password: *o.password,
			Hash:     *o.hash,
		},
	})
	if err != nil {
		return err
	}
	resp, err := stream.Recv()
	if err != nil {
		return err
	}
	data, err := json.Marshal(resp)
	common.Must(err)
	fmt.Print(string(data))
	return nil
}

func (o *apiController) setUsers(apiClient service.TrojanServerServiceClient) error {
	stream, err := apiClient.SetUsers(o.ctx)
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	req := &service.SetUsersRequest{
		Status: &service.UserStatus{
			User: &service.User{
				Password: *o.password,
				Hash:     *o.hash,
			},
			IpLimit: int32(*o.ipLimit),
			SpeedLimit: &service.Speed{
				UploadSpeed:   uint64(*o.uploadSpeedLimit),
				DownloadSpeed: uint64(*o.downloadSpeedLimit),
			},
		},
	}

	switch {
	case *o.add:
		req.Operation = service.SetUsersRequest_Add
	case *o.modify:
		req.Operation = service.SetUsersRequest_Modify
	case *o.delete:
		req.Operation = service.SetUsersRequest_Delete
	default:
		return common.NewError("Invalid operation")
	}

	err = stream.Send(req)
	if err != nil {
		return err
	}
	resp, err := stream.Recv()
	if err != nil {
		return err
	}
	if resp.Success {
		fmt.Println("Done")
	} else {
		fmt.Println("Failed: " + resp.Info)
	}
	return nil
}

func (o *apiController) Handle() error {
	if *o.cmd == "" {
		return common.NewError("")
	}
	conn, err := grpc.Dial(*o.address, grpc.WithInsecure())
	if err != nil {
		log.Error(err)
		return nil
	}
	defer conn.Close()
	apiClient := service.NewTrojanServerServiceClient(conn)
	switch *o.cmd {
	case "list":
		err := o.listUsers(apiClient)
		if err != nil {
			log.Error(err)
		}
	case "get":
		err := o.getUsers(apiClient)
		if err != nil {
			log.Error(err)
		}
	case "set":
		err := o.setUsers(apiClient)
		if err != nil {
			log.Error(err)
		}
	default:
		log.Error("unknown command " + *o.cmd)
	}
	return nil
}

func (o *apiController) Priority() int {
	return 50
}

func init() {
	option.RegisterHandler(&apiController{
		cmd:                flag.String("api", "", "Connect to a Trojan-Go API service. \"-api add/get/list\""),
		address:            flag.String("api-addr", "127.0.0.1:10000", "Address of Trojan-Go API service"),
		password:           flag.String("target-password", "", "Password of the target user"),
		hash:               flag.String("target-hash", "", "Hash of the target user"),
		add:                flag.Bool("add-profile", false, "Add a new profile with API"),
		delete:             flag.Bool("delete-profile", false, "Delete an existing profile with API"),
		modify:             flag.Bool("modify-profile", false, "Modify an existing profile with API"),
		uploadSpeedLimit:   flag.Int("upload-speed-limit", 0, "Limit the upload speed with API"),
		downloadSpeedLimit: flag.Int("download-speed-limit", 0, "Limit the download speed with API"),
		ipLimit:            flag.Int("ip-limit", 0, "Limit the number of IP with API"),
		ctx:                context.Background(),
	})
}
