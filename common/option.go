package common

type OptionHandler interface {
	Name() string
	Handle() error
	Priority() int
}

var handlers map[string]OptionHandler = make(map[string]OptionHandler)

func RegisterOptionHandler(h OptionHandler) {
	handlers[h.Name()] = h
}

func PopOptionHandler() (OptionHandler, error) {
	var maxHandler OptionHandler = nil
	for _, h := range handlers {
		if maxHandler == nil || maxHandler.Priority() < h.Priority() {
			maxHandler = h
		}
	}
	if maxHandler == nil {
		return nil, NewError("no option left")
	}
	delete(handlers, maxHandler.Name())
	return maxHandler, nil
}
