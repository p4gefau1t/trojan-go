package api

import "github.com/p4gefau1t/trojan-go/common"

// TODO implement api service client

type apiOption struct {
	common.OptionHandler
}

func (apiOption) Name() string {
	return "api"
}

func (o *apiOption) Handle() error {
	return nil
}

func (o *apiOption) Priority() int {
	return 50
}
