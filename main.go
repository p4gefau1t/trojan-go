package main

import (
	"flag"
	"github.com/p4gefau1t/trojan-go/option"

	_ "github.com/p4gefau1t/trojan-go/build"
	"github.com/p4gefau1t/trojan-go/log"
)

func main() {
	flag.Parse()
	for {
		h, err := option.PopOptionHandler()
		if err != nil {
			log.Fatal("invalid options")
		}
		err = h.Handle()
		if err == nil {
			break
		}
	}
}
