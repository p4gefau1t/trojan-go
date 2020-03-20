package stat

import (
	"io"
	"os"

	"github.com/withmandala/go-log"
)

var logger = log.New(os.Stdout).WithColor()

type TrafficMeter interface {
	Count(passwordHash string, upload int, download int)
	io.Closer
}

type Authenticator interface {
	CheckHash(hash string) bool
	io.Closer
}
