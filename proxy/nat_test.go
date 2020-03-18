package proxy

import (
	"io/ioutil"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
)

func TestNAT(t *testing.T) {
	data, err := ioutil.ReadFile("nat.json")
	common.Must(err)
	config, err := conf.ParseJSON(data)
	common.Must(err)

	nat := NAT{
		config: config,
	}
	err = nat.Run()
	common.Must(err)
}
