package main

import (
	"flag"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"

	//the following modules are optional
	//you can comment some of them if you don't need them
	_ "github.com/p4gefau1t/trojan-go/api"
	_ "github.com/p4gefau1t/trojan-go/cert"
	_ "github.com/p4gefau1t/trojan-go/daemon"
	_ "github.com/p4gefau1t/trojan-go/easy"
	_ "github.com/p4gefau1t/trojan-go/log/golog"
	_ "github.com/p4gefau1t/trojan-go/proxy/client"
	_ "github.com/p4gefau1t/trojan-go/proxy/relay"
	_ "github.com/p4gefau1t/trojan-go/proxy/server"
	_ "github.com/p4gefau1t/trojan-go/router/mixed"
	_ "github.com/p4gefau1t/trojan-go/stat/db"
	_ "github.com/p4gefau1t/trojan-go/stat/memory"
	_ "github.com/p4gefau1t/trojan-go/version"
	//_ "github.com/p4gefau1t/trojan-go/log/simplelog"
)

func main() {
	flag.Parse()
	for {
		h, err := common.PopOptionHandler()
		if err != nil {
			log.Fatal("invalid options")
		}
		err = h.Handle()
		if err == nil {
			break
		}
	}
}
