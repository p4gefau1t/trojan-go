package stat

import (
	"context"
	"io"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
)

type TrafficMeter interface {
	io.Closer
	Hash() string
	AddTraffic(sent, recv int)
	GetTraffic() (sent, recv uint64)
	ResetTraffic()
	GetAndResetTraffic() (sent, recv uint64)
	GetSpeed() (sent, recv uint64)
	SetSpeedLimit(send, recv int)
	GetSpeedLimit() (send, recv int)
}

type IPRecorder interface {
	AddIP(string) bool
	DelIP(string) bool
	GetIP() int
	SetIPLimit(int)
	GetIPLimit() int
}

type User interface {
	TrafficMeter
	IPRecorder
}

type Authenticator interface {
	io.Closer
	AuthUser(hash string) (valid bool, user User)
	AddUser(hash string) error
	DelUser(hash string) error
	ListUsers() []User
}

type AuthCreator func(ctx context.Context, config *conf.GlobalConfig) (Authenticator, error)

var authCreators = map[string]AuthCreator{}

func RegisterAuthCreator(name string, creator AuthCreator) {
	authCreators[name] = creator
}

func NewAuth(ctx context.Context, name string, config *conf.GlobalConfig) (Authenticator, error) {
	creator, found := authCreators[name]
	if !found {
		return nil, common.NewError("Auth driver name " + name + " not found")
	}
	return creator(ctx, config)
}
