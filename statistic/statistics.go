package statistic

import (
	"context"
	"io"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
)

const (
	TrafficMeterKey  = "TRAFFIC_METER"
	AuthenticatorKey = "AUTHENTICATOR"
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

type Creator func(ctx context.Context) (Authenticator, error)

var authCreators = map[string]Creator{}

func RegisterAuthenticatorCreator(name string, creator Creator) {
	authCreators[name] = creator
}

func NewAuthenticator(ctx context.Context, name string) (Authenticator, error) {
	creator, found := authCreators[strings.ToUpper(name)]
	if !found {
		return nil, common.NewError("Auth driver name " + name + " not found")
	}
	return creator(ctx)
}
