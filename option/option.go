package option

import "github.com/p4gefau1t/trojan-go/common"

type Handler interface {
	Name() string
	Handle() error
	Priority() int
}

var handlers = make(map[string]Handler)

func RegisterHandler(h Handler) {
	handlers[h.Name()] = h
}

func PopOptionHandler() (Handler, error) {
	var maxHandler Handler = nil
	for _, h := range handlers {
		if maxHandler == nil || maxHandler.Priority() < h.Priority() {
			maxHandler = h
		}
	}
	if maxHandler == nil {
		return nil, common.NewError("no option left")
	}
	delete(handlers, maxHandler.Name())
	return maxHandler, nil
}
