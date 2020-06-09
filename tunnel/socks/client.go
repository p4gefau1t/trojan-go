package socks

import "github.com/p4gefau1t/trojan-go/tunnel"

// TODO implement socks5 client

type Client struct {
	underlay tunnel.Client
}
