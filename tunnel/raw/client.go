package raw

import (
	"context"
	"crypto/tls"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/p4gefau1t/trojan-go/tunnel"
)

type Client struct {
	preferIPv4 bool
	noDelay    bool
	keepAlive  bool
	dns        []string
	ctx        context.Context
}

func (c *Client) resolveIP(addr *tunnel.Address) ([]net.IPAddr, error) {
	for _, s := range c.dns {
		var dnsAddr string
		var dnsHost, dnsType string
		var err error

		dnsURL, err := url.Parse(s)
		if err != nil || dnsURL.Scheme == "" {
			dnsType = "udp"
			dnsAddr = s
		} else {
			dnsType = dnsURL.Scheme
			dnsAddr = dnsURL.Host
		}

		dnsHost, tmp, err := net.SplitHostPort(dnsAddr)
		dnsPort, err := strconv.ParseInt(tmp, 10, 32)
		common.Must(err)

		if err != nil {
			dnsHost = dnsAddr
			switch dnsType {
			case "dot":
				dnsPort = 853
			case "tcp", "udp":
				dnsPort = 53
			}
		}

		resolver := &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				dnsAddress := tunnel.NewAddressFromHostPort("tcp", dnsHost, int(dnsPort))
				switch dnsType {
				case "udp", "tcp":
					d := net.Dialer{
						Timeout: time.Second * 5,
					}
					conn, err := d.DialContext(ctx, dnsType, dnsAddress.String())
					if err != nil {
						return nil, err
					}
					return conn, nil
				case "dot":
					tlsConn, err := tls.Dial("tcp", dnsAddress.String(), nil)
					if err != nil {
						return nil, err
					}
					return tlsConn, nil
				}
				return nil, common.NewError("invalid dns type:" + dnsType)
			},
		}
		ip, err := resolver.LookupIPAddr(c.ctx, addr.String())
		if err != nil {
			log.Error(err)
			continue
		}
		return ip, nil
	}
	return nil, common.NewError("address not found")
}

func (c *Client) DialConn(addr *tunnel.Address, t tunnel.Tunnel) (tunnel.Conn, error) {
	network := "tcp"
	if c.preferIPv4 {
		network = "tcp4"
	}
	tcpConn, err := net.Dial(network, addr.String())
	if err != nil {
		return nil, err
	}

	tcpConn.(*net.TCPConn).SetKeepAlive(c.keepAlive)
	tcpConn.(*net.TCPConn).SetNoDelay(c.noDelay)
	return &Conn{
		TCPConn: tcpConn.(*net.TCPConn),
	}, nil
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	network := "udp"
	if c.preferIPv4 {
		network = "udp4"
	}
	udpConn, err := net.ListenPacket(network, "")
	if err != nil {
		return nil, err
	}
	return &PacketConn{
		UDPConn: udpConn.(*net.UDPConn),
	}, nil
}

func (c *Client) Close() error {
	return nil
}

func NewClient(ctx context.Context, client tunnel.Client) (*Client, error) {
	// TODO implement dns
	cfg := config.FromContext(ctx, Name).(*Config)
	return &Client{
		dns:        cfg.DNS,
		noDelay:    cfg.TCP.NoDelay,
		keepAlive:  cfg.TCP.KeepAlive,
		preferIPv4: cfg.TCP.PreferIPV4,
	}, nil
}

// FixedClient will always dial to the FixedAddr
type FixedClient struct {
	FixedAddr *tunnel.Address
	Client
}

func (c *FixedClient) DialConn(addr *tunnel.Address, t tunnel.Tunnel) (tunnel.Conn, error) {
	return c.Client.DialConn(c.FixedAddr, t)
}
