// +build !linux
// +build !windows
// +build !darwin

package sockopt

import (
	"runtime"

	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
)

func ApplySocketOption(fd uintptr, config *conf.TCPConfig, isInbound bool) error {
	log.Warn("tcp options is ignored in this os:", runtime.GOOS)
	return nil
}
