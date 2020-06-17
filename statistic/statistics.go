package statistic

import (
	"context"
	"github.com/p4gefau1t/trojan-go/log"
	"io"
	"strings"
	"sync"

	"github.com/p4gefau1t/trojan-go/common"
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
