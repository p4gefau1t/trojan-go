package control

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/p4gefau1t/trojan-go/api/service"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"google.golang.org/grpc"
)

type apiOption struct {
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
	iplimit            *int
	ctx                context.Context
}

func (apiOption) Name() string {
	return "api"
}

func (o *apiOption) listUsers(apiClient service.TrojanServerServiceClient) error {
	stream, err := apiClient.ListUsers(o.ctx, &service.ListUsersRequest{})
	if err != nil {
		return err
	}
	defer stream.CloseSend()
	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		data, err := json.Marshal(resp)
		common.Must(err)
		fmt.Print(string(data))
	}
}

func (o *apiOption) getUsers(apiClient service.TrojanServerServiceClient) error {
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

func (o *apiOption) setUsers(apiClient service.TrojanServerServiceClient) error {
	stream, err := apiClient.SetUsers(o.ctx)
	if err != nil {
		return err
	}
	defer stream.CloseSend()

	req := &service.SetUsersRequest{
		User: &service.User{
			Password: *o.password,
			Hash:     *o.hash,
		},
		IpLimit: int32(*o.iplimit),
		SpeedLimit: &service.Speed{
			UploadSpeed:   uint64(*o.uploadSpeedLimit),
			DownloadSpeed: uint64(*o.downloadSpeedLimit),
		},
	}
	if *o.add {
		req.Operation = service.SetUsersRequest_Add
	} else if *o.modify {
		req.Operation = service.SetUsersRequest_Modify
	} else if *o.delete {
		req.Operation = service.SetUsersRequest_Delete
	} else {
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

func (o *apiOption) Handle() error {
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
		log.Error("Unknown command " + *o.cmd)
	}
	return nil
}

func (o *apiOption) Priority() int {
	return 50
}

func init() {
	common.RegisterOptionHandler(&apiOption{
		cmd:                flag.String("api", "", "Connect to a Trojan-Go API service. \"-api add/get/list\""),
		address:            flag.String("api-addr", "127.0.0.1:10000", "Address of Trojan-Go API service"),
		password:           flag.String("target-password", "", "Password of the target user"),
		hash:               flag.String("target-hash", "", "Hash of the target user"),
		add:                flag.Bool("add-profile", false, "Add a new profile with API"),
		delete:             flag.Bool("delete-profile", false, "Delete an existing profile with API"),
		modify:             flag.Bool("modify-profile", false, "Modify an existing profile with API"),
		uploadSpeedLimit:   flag.Int("upload-speed-limit", 0, "Limit the upload speed with API"),
		downloadSpeedLimit: flag.Int("download-speed-limit", 0, "Limit the download speed with API"),
		iplimit:            flag.Int("ip-limit", 0, "Limit the number of IP with API"),
		ctx:                context.Background(),
	})
}
