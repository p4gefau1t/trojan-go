package common

import (
	"fmt"
	"net"
)

type AddressType byte

const (
	IPv4       AddressType = 1
	DomainName AddressType = 3
	IPv6       AddressType = 4
)

type Address struct {
	net.Addr

	DomainName  string
	IP          net.IP
	Port        int
	NetworkType string
	AddressType
}

func (a *Address) String() string {
	switch a.AddressType {
	case IPv4:
		return fmt.Sprintf("%s:%d", a.IP.String(), a.Port)
	case IPv6:
		return fmt.Sprintf("[%s]:%d", a.IP.String(), a.Port)
	case DomainName:
		return fmt.Sprintf("%s:%d", a.DomainName, a.Port)
	default:
		return ""
	}
}

func (a *Address) Network() string {
	return a.NetworkType
}

func (a *Address) ResolveIP(preferV4 bool) (net.IP, error) {
	if a.AddressType == IPv4 || a.AddressType == IPv6 {
		return a.IP, nil
	}
	network := "ip"
	if preferV4 {
		network = "ip4"
	}
	addr, err := net.ResolveIPAddr(network, a.DomainName)
	if err != nil {
		return nil, err
	}
	return addr.IP, nil
}

func NewAddress(host string, port int, network string) *Address {
	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() != nil {
			return &Address{
				IP:          ip,
				Port:        port,
				AddressType: IPv4,
				NetworkType: network,
			}
		}
		return &Address{
			IP:          ip,
			Port:        port,
			AddressType: IPv6,
			NetworkType: network,
		}
	}
	return &Address{
		DomainName:  host,
		Port:        port,
		AddressType: DomainName,
		NetworkType: network,
	}
}

const (
	KiB = 1024
	MiB = KiB * 1024
	GiB = MiB * 1024
)

func HumanFriendlyTraffic(bytes int) string {
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
