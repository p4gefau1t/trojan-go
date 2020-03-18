// +build !linux

package proxy

import (
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
)

type NAT struct {
	common.Runnable
	config *conf.GlobalConfig
}

func (n *NAT) Run() error {
	return common.NewError("not supported os")
}
