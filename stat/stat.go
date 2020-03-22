package stat

import (
	"io"
	"os"

	"github.com/p4gefau1t/trojan-go/log"
)

var logger = log.New(os.Stdout)

type TrafficMeter interface {
	Count(passwordHash string, sent int, recv int)
	io.Closer
}

type Authenticator interface {
	CheckHash(hash string) bool
	io.Closer
}
