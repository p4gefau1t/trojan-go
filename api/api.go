package api

import (
	"context"

	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/statistic"
)

type Handler func(ctx context.Context, auth statistic.Authenticator) error

var handlers = map[string]Handler{}

func RegisterHandler(name string, handler Handler) {
	handlers[name] = handler
}

func RunService(ctx context.Context, name string, auth statistic.Authenticator) error {
	if h, ok := handlers[name]; ok {
		log.Debug("api handler found", name)
		return h(ctx, auth)
	}
	log.Debug("api handler not found", name)
	return nil
}
