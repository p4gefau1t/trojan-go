package main

import (
	"flag"
	"os"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"

	//the following modules are optional
	//you can comment some of them if you don't need them
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/p4gefau1t/trojan-go/cert"
	_ "github.com/p4gefau1t/trojan-go/daemon"
	_ "github.com/p4gefau1t/trojan-go/proxy/client"
	_ "github.com/p4gefau1t/trojan-go/proxy/forward"
	_ "github.com/p4gefau1t/trojan-go/proxy/server"
	_ "github.com/p4gefau1t/trojan-go/version"
)

var logger = log.New(os.Stdout)

func main() {
	flag.Parse()
	for {
		h, err := common.PopOptionHandler()
		if err != nil {
			logger.Fatal("invalid options")
		}
		err = h.Handle()
		if err == nil {
			break
		}
	}
}
