package main

import (
	"flag"
	"os"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/p4gefau1t/trojan-go/cert"
	_ "github.com/p4gefau1t/trojan-go/proxy/client"
	_ "github.com/p4gefau1t/trojan-go/proxy/forward"
	_ "github.com/p4gefau1t/trojan-go/proxy/server"
)

var logger = log.New(os.Stdout)

func main() {
	logger.Info("Trojan-Go initializing...")
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
	logger.Info("Trojan-Go exited")
}
