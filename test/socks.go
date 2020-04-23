package test

// Copyright 2012, Hailiang Wang. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package socks implements a SOCKS (SOCKS4, SOCKS4A and SOCKS5) proxy client.

A complete example using this package:
	package main

	import (
		"h12.io/socks"
		"fmt"
		"net/http"
		"io/ioutil"
	)

	func main() {
		dialSocksProxy := socks.Dial("socks5://127.0.0.1:1080?timeout=5s")
		tr := &http.Transport{Dial: dialSocksProxy}
		httpClient := &http.Client{Transport: tr}

		bodyText, err := TestHttpsGet(httpClient, "https://h12.io/about")
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Print(bodyText)
	}

	func TestHttpsGet(c *http.Client, url string) (bodyText string, err error) {
		resp, err := c.Get(url)
		if err != nil { return }
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil { return }
		bodyText = string(body)
		return
	}
*/

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"time"
)

// Constants to choose which version of SOCKS protocol to use.
const (
	SOCKS4 = iota
	SOCKS4A
	SOCKS5
)

type (
	Config struct {
		Proto   int
		Host    string
		Auth    Auth
		Timeout time.Duration
	}
	Auth struct {
		Username string
		Password string
	}
)

func parse(proxyURI string) (*Config, error) {
	uri, err := url.Parse(proxyURI)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	switch uri.Scheme {
	case "socks4":
		cfg.Proto = SOCKS4
	case "socks4a":
		cfg.Proto = SOCKS4A
	case "socks5":
		cfg.Proto = SOCKS5
	default:
		return nil, fmt.Errorf("unknown SOCKS protocol %s", uri.Scheme)
	}
	cfg.Host = uri.Host
	if uri.User != nil {
		cfg.Auth.Username = uri.User.Username()
		cfg.Auth.Password, _ = uri.User.Password()
	}
	query := uri.Query()
	timeout := query.Get("timeout")
	if timeout != "" {
		var err error
		cfg.Timeout, err = time.ParseDuration(timeout)
		if err != nil {
			return nil, err
		}
	}
	return cfg, nil
}

// Dial returns the dial function to be used in http.Transport object.
// Argument proxyURI should be in the format: "socks5://user:password@127.0.0.1:1080?timeout=5s".
// The protocol could be socks5, socks4 and socks4a.
func Dial(proxyURI string) func(string, string) (net.Conn, error) {
	cfg, err := parse(proxyURI)
	if err != nil {
		return dialError(err)
	}
	return cfg.dialFunc()
}

// DialSocksProxy returns the dial function to be used in http.Transport object.
// Argument socksType should be one of SOCKS4, SOCKS4A and SOCKS5.
// Argument proxy should be in this format "127.0.0.1:1080".
func DialSocksProxy(socksType int, proxy string) func(string, string) (net.Conn, error) {
	return (&Config{Proto: socksType, Host: proxy}).dialFunc()
}

func (c *Config) dialFunc() func(string, string) (net.Conn, error) {
	switch c.Proto {
	case SOCKS5:
		return func(_, targetAddr string) (conn net.Conn, err error) {
			return c.dialSocks5(targetAddr)
		}
	case SOCKS4, SOCKS4A:
		return func(_, targetAddr string) (conn net.Conn, err error) {
			return c.dialSocks4(targetAddr)
		}
	}
	return dialError(fmt.Errorf("unknown SOCKS protocol %v", c.Proto))
}

func (cfg *Config) dialSocks5(targetAddr string) (conn net.Conn, err error) {
	proxy := cfg.Host

	// dial TCP
	conn, err = net.Dial("tcp", proxy)
	if err != nil {
		return
	}

	// version identifier/method selection request
	req := []byte{
		5, // version number
		1, // number of methods
		0, // method 0: no authentication (only anonymous access supported for now)
	}
	resp, err := cfg.sendReceive(conn, req)
	if err != nil {
		return
	} else if len(resp) != 2 {
		err = errors.New("Server does not respond properly.")
		return
	} else if resp[0] != 5 {
		err = errors.New("Server does not support Socks 5.")
		return
	} else if resp[1] != 0 { // no auth
		err = errors.New("socks method negotiation failed.")
		return
	}

	// detail request
	host, port, err := splitHostPort(targetAddr)
	if err != nil {
		return nil, err
	}
	req = []byte{
		5, // version number
		//1,               // connect command
		3,               // associate command
		0,               // reserved, must be zero
		3,               // address type, 3 means domain name
		byte(len(host)), // address length
	}
	req = append(req, []byte(host)...)
	req = append(req, []byte{
		byte(port >> 8), // higher byte of destination port
		byte(port),      // lower byte of destination port (big endian)
	}...)
	resp, err = cfg.sendReceive(conn, req)
	if err != nil {
		return
	} else if len(resp) != 10 {
		err = errors.New("Server does not respond properly.")
	} else if resp[1] != 0 {
		err = errors.New("Can't complete SOCKS5 connection.")
	}

	return
}

func (cfg *Config) dialSocks4(targetAddr string) (conn net.Conn, err error) {
	socksType := cfg.Proto
	proxy := cfg.Host

	// dial TCP
	conn, err = net.Dial("tcp", proxy)
	if err != nil {
		return
	}

	// connection request
	host, port, err := splitHostPort(targetAddr)
	if err != nil {
		return
	}
	ip := net.IPv4(0, 0, 0, 1).To4()
	if socksType == SOCKS4 {
		ip, err = lookupIP(host)
		if err != nil {
			return
		}
	}
	req := []byte{
		4,                          // version number
		1,                          // command CONNECT
		byte(port >> 8),            // higher byte of destination port
		byte(port),                 // lower byte of destination port (big endian)
		ip[0], ip[1], ip[2], ip[3], // special invalid IP address to indicate the host name is provided
		0, // user id is empty, anonymous proxy only
	}
	if socksType == SOCKS4A {
		req = append(req, []byte(host+"\x00")...)
	}

	resp, err := cfg.sendReceive(conn, req)
	if err != nil {
		return
	} else if len(resp) != 8 {
		err = errors.New("Server does not respond properly.")
		return
	}
	switch resp[1] {
	case 90:
		// request granted
	case 91:
		err = errors.New("Socks connection request rejected or failed.")
	case 92:
		err = errors.New("Socks connection request rejected becasue SOCKS server cannot connect to identd on the client.")
	case 93:
		err = errors.New("Socks connection request rejected because the client program and identd report different user-ids.")
	default:
		err = errors.New("Socks connection request failed, unknown error.")
	}
	// clear the deadline before returning
	if err := conn.SetDeadline(time.Time{}); err != nil {
		return nil, err
	}
	return
}

func (cfg *Config) sendReceive(conn net.Conn, req []byte) (resp []byte, err error) {
	if cfg.Timeout > 0 {
		if err := conn.SetWriteDeadline(time.Now().Add(cfg.Timeout)); err != nil {
			return nil, err
		}
	}
	_, err = conn.Write(req)
	if err != nil {
		return
	}
	resp, err = cfg.readAll(conn)
	return
}

func (cfg *Config) readAll(conn net.Conn) (resp []byte, err error) {
	resp = make([]byte, 1024)
	if cfg.Timeout > 0 {
		if err := conn.SetReadDeadline(time.Now().Add(cfg.Timeout)); err != nil {
			return nil, err
		}
	}
	n, err := conn.Read(resp)
	resp = resp[:n]
	return
}

func lookupIP(host string) (ip net.IP, err error) {
	ips, err := net.LookupIP(host)
	if err != nil {
		return
	}
	if len(ips) == 0 {
		err = fmt.Errorf("Cannot resolve host: %s.", host)
		return
	}
	ip = ips[0].To4()
	if len(ip) != net.IPv4len {
		fmt.Println(len(ip), ip)
		err = errors.New("IPv6 is not supported by SOCKS4.")
		return
	}
	return
}

func splitHostPort(addr string) (host string, port uint16, err error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return "", 0, err
	}
	portInt, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return "", 0, err
	}
	port = uint16(portInt)
	return
}

func dialError(err error) func(string, string) (net.Conn, error) {
	return func(_, _ string) (net.Conn, error) {
		return nil, err
	}
}
