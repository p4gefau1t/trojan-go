package common

import (
	"fmt"
	"net"
	"strconv"
)

const (
	KiB = 1024
	MiB = KiB * 1024
	GiB = MiB * 1024
)

func HumanFriendlyTraffic(bytes uint64) string {
	if bytes <= KiB {
		return fmt.Sprintf("%d B", bytes)
	}
	if bytes <= MiB {
		return fmt.Sprintf("%.2f KiB", float32(bytes)/KiB)
	}
	if bytes <= GiB {
		return fmt.Sprintf("%.2f MiB", float32(bytes)/MiB)
	}
	return fmt.Sprintf("%.2f GiB", float32(bytes)/GiB)
}

func PickPort(network string, host string) int {
	switch network {
	case "tcp":
		l, err := net.Listen("tcp", host+":0")
		Must(err)
		defer l.Close()
		_, port, err := net.SplitHostPort(l.Addr().String())
		Must(err)
		p, err := strconv.ParseInt(port, 10, 32)
		Must(err)
		return int(p)
	case "udp":
		conn, err := net.ListenPacket("udp", host+":0")
		Must(err)
		defer conn.Close()
		_, port, err := net.SplitHostPort(conn.LocalAddr().String())
		Must(err)
		p, err := strconv.ParseInt(port, 10, 32)
		Must(err)
		return int(p)
	default:
		return 0
	}
}
