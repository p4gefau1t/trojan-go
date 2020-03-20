package stat

import (
	"io"
	"os"

	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/withmandala/go-log"
)

var logger = log.New(os.Stdout).WithColor()

type TrafficCounter interface {
	Count(passwordHash string, upload int, download int)
	io.Closer
}

type Authenticator interface {
	CheckHash(hash string) bool
	io.Closer
}

type EmptyTrafficCounter struct {
	TrafficCounter
}

func (t *EmptyTrafficCounter) Count(string, int, int) {
	//do nothing
}

func (t *EmptyTrafficCounter) Close() error {
	//do nothing
	return nil
}

type EmptyAuthenticator struct {
	Authenticator
}

func (a *EmptyAuthenticator) CheckHash(hash string) bool {
	return true
}

func (a *EmptyAuthenticator) Close() error {
	return nil
}

type ConfigUserAuthenticator struct {
	Config *conf.GlobalConfig
	Authenticator
}

func (a *ConfigUserAuthenticator) CheckHash(hash string) bool {
	_, found := a.Config.Hash[hash]
	return found
}

func (a *ConfigUserAuthenticator) Close() error {
	return nil
}
