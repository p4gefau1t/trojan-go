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
	Count(sent int, recv int)
	Get() (sent uint64, recv uint64)
	Reset()
	GetAndReset() (sent uint64, recv uint64)
	GetSpeed() (sent uint64, recv uint64)
	LimitSpeed(sent int, recv int)
}

type Authenticator interface {
	io.Closer
	AuthUser(hash string) (valid bool, meter TrafficMeter)
	AddUser(hash string) error
	DelUser(hash string) error
	ListUsers() []TrafficMeter
}

type AuthCreator func(ctx context.Context, config *conf.GlobalConfig) (Authenticator, error)

var authCreators = map[string]AuthCreator{}

func RegisterAuthCreator(name string, creator AuthCreator) {
	authCreators[name] = creator
}

func NewAuth(ctx context.Context, name string, config *conf.GlobalConfig) (Authenticator, error) {
	creator, found := authCreators[name]
	if !found {
		return nil, common.NewError("driver name " + name + " not found")
	}
	return creator(ctx, config)
}
