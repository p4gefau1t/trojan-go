package stat

import (
	"io"
)

type TrafficMeter interface {
	io.Closer
	Count(passwordHash string, sent uint64, recv uint64)
	Query(passwordHash string) (sent uint64, recv uint64)
}

type Authenticator interface {
	io.Closer

	CheckHash(hash string) bool
}
