package stat

import (
	"io"
)

type TrafficMeter interface {
	Count(passwordHash string, sent int, recv int)
	io.Closer
}

type Authenticator interface {
	CheckHash(hash string) bool
	io.Closer
}
