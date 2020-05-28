package test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
)

func TestRealProxy(t *testing.T) {
	if os.Getenv("real_test") == "" {
		t.Skip("skipping real proxy test")
	}
	clientConfig := addMuxConfig(getBasicClientConfig())
	serverConfig := getBasicServerConfig()
	go RunClient(context.Background(), clientConfig)
	go RunHelloHTTPServer(context.Background())
	RunServer(context.Background(), serverConfig)
}

func TestRealClient(t *testing.T) {
	if os.Getenv("real_test") == "" {
		t.Skip("skipping real proxy test")
	}
	b, err := ioutil.ReadFile("client.json")
	common.Must(err)
	config, err := conf.ParseJSON(b)
	common.Must(err)
	RunClient(context.Background(), config)
}

func TestRealServer(t *testing.T) {
	if os.Getenv("real_test") == "" {
		t.Skip("skipping real proxy test")
	}
	b, err := ioutil.ReadFile("server.json")
	common.Must(err)
	config, err := conf.ParseJSON(b)
	common.Must(err)
	RunServer(context.Background(), config)
}
