package api

import (
	"sync"

	"github.com/p4gefau1t/trojan-go/stat"
)

type MemoryTraffic struct {
	downloadTraffic uint64
	uploadTraffic   uint64
}

type MemoryUser struct {
	password     string
	hash         string
	trafficTotal MemoryTraffic
	trafficQuota MemoryTraffic
}

type APIAuth struct {
	stat.Authenticator
	users sync.Map
}
