package stat

import (
	"io"
	"os"

	"github.com/withmandala/go-log"
)

var logger = log.New(os.Stdout).WithColor()

type TrafficCounter interface {
	Count(passwordHash string, upload int, download int)
	io.Closer
}

type Authenticator interface {
	IsValid(passwordHash string) bool
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

func NewEmptyTrafficCounter() TrafficCounter {
	return &EmptyTrafficCounter{}
}
