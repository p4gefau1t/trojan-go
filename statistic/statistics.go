package statistic

import (
	"context"
	"io"
	"strings"
	"sync"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
)

const Name = "STATISTICS"

type Metadata interface {
	GetHash() string
	GetTraffic() (sent, recv uint64)
	GetSpeedLimit() (sent, recv int)
	GetIPLimit() int
}

type TrafficMeter interface {
	io.Closer
	AddTraffic(sent, recv int)
	ResetTraffic() (sent, recv uint64)
	GetSpeed() (sent, recv uint64)
}

type IPRecorder interface {
	AddIP(string) bool
	DelIP(string) bool
	GetIP() int
}

type User interface {
	Metadata
	TrafficMeter
	IPRecorder
}

type Persistencer interface {
	SaveUser(Metadata) error
	LoadUser(hash string) (Metadata, error)
	DeleteUser(hash string) error
	ListUser(func(hash string, u Metadata) bool) error
	UpdateUserTraffic(hash string, sent, recv uint64) error
}

type Authenticator interface {
	io.Closer
	AuthUser(hash string) (valid bool, user User)
	AddUser(hash string) error
	DelUser(hash string) error
	SetUserTraffic(hash string, sent, recv uint64) error
	SetUserSpeedLimit(hash string, send, recv int) error
	SetUserIPLimit(hash string, limit int) error
	ListUsers() []User
}

type Creator func(ctx context.Context) (Authenticator, error)

var (
	createdAuthLock sync.Mutex
	authCreators    = make(map[string]Creator)
	createdAuth     = make(map[context.Context]Authenticator)
)

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
