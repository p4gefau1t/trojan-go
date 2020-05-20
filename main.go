package main

import (
	"flag"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"

	_ "github.com/p4gefau1t/trojan-go/build"
)

func main() {
	log.Info("Trojan-Go", common.Version)
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
