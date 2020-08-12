package statistic

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/p4gefau1t/trojan-go/log"

	"github.com/p4gefau1t/trojan-go/config"

	"github.com/p4gefau1t/trojan-go/common"
)

const Name = "STATISTICS"

type TrafficMeter interface {
	io.Closer
	Hash() string
	AddTraffic(sent, recv int)
	GetTraffic() (sent, recv uint64)
	ResetTraffic() (sent, recv uint64)
	GetSpeed() (sent, recv uint64)
	SetSpeedLimit(sent, recv int)
	GetSpeedLimit() (sent, recv int)
	SetTraffic(sent, recv uint64)
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
var createdAuth = map[context.Context]Authenticator{}
var createdAuthLock = sync.Mutex{}

func RegisterAuthenticatorCreator(name string, creator Creator) {
	authCreators[name] = creator
}

func NewAuthenticator(ctx context.Context, name string) (Authenticator, error) {
	// allocate a unique authenticator for each context
	createdAuthLock.Lock() // avoid concurrent map read/write
	defer createdAuthLock.Unlock()
	if auth, found := createdAuth[ctx]; found {
		log.Debug("authenticator has been created:", name)
		return auth, nil
	}
	creator, found := authCreators[strings.ToUpper(name)]
	if !found {
		return nil, common.NewError("auth driver name " + name + " not found")
	}
	auth, err := creator(ctx)
	if err != nil {
		return nil, err
	}
	createdAuth[ctx] = auth
	return auth, err
}

func ShouldTrackUserIp(ctx context.Context) bool {
	cfg := config.FromContext(ctx, Name).(*Config)
	return cfg.Statistics.TrackUserIp
}
