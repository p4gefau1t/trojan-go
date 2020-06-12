package router

import (
	"context"
	"github.com/golang/protobuf/proto"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"github.com/p4gefau1t/trojan-go/tunnel/raw"
	"github.com/p4gefau1t/trojan-go/tunnel/transport"
	"io/ioutil"
	"net"
	"regexp"
	"strconv"
	"strings"
	v2router "v2ray.com/core/app/router"
)

const (
	Proxy  = 0
	Bypass = 1
	Block  = 2
)

const (
	AsIs         = 0
	IPIfNonMatch = 1
	IPOnDemand   = 2
)

const MaxPacketSize = 1024 * 8

func matchDomain(list []*v2router.Domain, target string) bool {
	for _, d := range list {
		switch d.GetType() {
		case v2router.Domain_Full:
			domain := d.GetValue()
			if domain == target {
				log.Trace("domain:", target, "hit domain(full) rule:", domain)
				return true
			}
		case v2router.Domain_Domain:
			domain := d.GetValue()
			if strings.HasSuffix(target, domain) {
				idx := strings.Index(target, domain)
				if idx == 0 || target[idx-1] == '.' {
					log.Trace("domain:", target, "hit domain rule:", domain)
					return true
				}
			}
		case v2router.Domain_Plain:
			//keyword
			if strings.Contains(target, d.GetValue()) {
				log.Trace("domain:", target, "hit keyword rule:", d.GetValue())
				return true
			}
		case v2router.Domain_Regex:
			matched, err := regexp.Match(d.GetValue(), []byte(target))
			if err != nil {
				log.Error("invalid regex", d.GetValue())
				return false
			}
			if matched {
				log.Trace("domain:", target, "hit regex rule:", d.GetValue())
				return true
			}
		default:
			log.Debug("unknown rule type:" + d.GetType().String())
		}
	}
	return false
}

func matchIP(list []*v2router.CIDR, target net.IP) bool {
	isIPv6 := true
	len := net.IPv6len
	if target.To4() != nil {
		len = net.IPv4len
		isIPv6 = false
	}
	for _, c := range list {
		n := int(c.GetPrefix())
		mask := net.CIDRMask(n, 8*len)
		cidrIP := net.IP(c.GetIp())
		if cidrIP.To4() != nil { //IPv4 CIDR
			if isIPv6 {
				continue
			}
		} else { //IPv6 CIDR
			if !isIPv6 {
				continue
			}
		}
		subnet := &net.IPNet{IP: cidrIP.Mask(mask), Mask: mask}
		if subnet.Contains(target) {
			return true
		}
	}
	return false
}

func newIPAddress(address *tunnel.Address) (*tunnel.Address, error) {
	ip, err := address.ResolveIP()
	if err != nil {
		return nil, common.NewError("router failed to resolve ip").Base(err)
	}
	newAddress := &tunnel.Address{
		IP:   ip,
		Port: address.Port,
	}
	if ip.To4() != nil {
		newAddress.AddressType = tunnel.IPv4
	} else {
		newAddress.AddressType = tunnel.IPv6
	}
	return newAddress, nil
}

type Client struct {
	domains        [3][]*v2router.Domain
	cidrs          [3][]*v2router.CIDR
	defaultPolicy  int
	domainStrategy int
	underlay       tunnel.Client
	direct         *raw.Client
	ctx            context.Context
	cancel         context.CancelFunc
}

func (c *Client) Route(address *tunnel.Address) int {
	policy := -1
	var err error
	if c.defaultPolicy == IPOnDemand {
		address, err = newIPAddress(address)
		if err != nil {
			return c.defaultPolicy
		}
	}
	if address.AddressType == tunnel.DomainName {
		for i := 0; i < 3; i++ {
			if matchDomain(c.domains[i], address.DomainName) {
				policy = i
				break
			}
		}
	} else {
		for i := 0; i < 3; i++ {
			if matchIP(c.cidrs[i], address.IP) {
				policy = i
				break
			}
		}
	}
	if policy == -1 && c.domainStrategy == IPIfNonMatch {
		address, err = newIPAddress(address)
		if err != nil {
			return c.defaultPolicy
		}
		for i := 0; i < 3; i++ {
			if matchIP(c.cidrs[i], address.IP) {
				policy = i
				break
			}
		}
	}
	if policy == -1 {
		policy = c.defaultPolicy
	}
	return policy
}

func (c *Client) DialConn(address *tunnel.Address, overlay tunnel.Tunnel) (tunnel.Conn, error) {
	policy := c.Route(address)
	switch policy {
	case Proxy:
		return c.underlay.DialConn(address, overlay)
	case Block:
		return nil, common.NewError("router blocked address: " + address.String())
	case Bypass:
		conn, err := c.direct.DialConn(address, &Tunnel{})
		if err != nil {
			return nil, common.NewError("router dial error").Base(err)
		}
		return &transport.Conn{
			Conn: conn,
		}, nil
	}
	panic("unknown policy")
}

func (c *Client) DialPacket(overlay tunnel.Tunnel) (tunnel.PacketConn, error) {
	directConn, err := net.ListenPacket("udp", "")
	if err != nil {
		return nil, common.NewError("router failed to dial udp (direct)").Base(err)
	}
	proxy, err := c.underlay.DialPacket(overlay)
	if err != nil {
		return nil, common.NewError("router failed to dial udp (proxy)").Base(err)
	}
	ctx, cancel := context.WithCancel(c.ctx)
	return &PacketConn{
		Client:     c,
		PacketConn: directConn,
		proxy:      proxy,
		cancel:     cancel,
		ctx:        ctx,
		packetChan: make(chan *packetInfo, 16),
	}, nil
}

func (c *Client) Close() error {
	c.cancel()
	return c.underlay.Close()
}

type codeInfo struct {
	code     string
	strategy int
}

func loadCode(cfg *Config, prefix string) []codeInfo {
	codes := []codeInfo{}
	for _, s := range cfg.Router.Proxy {
		if strings.HasPrefix(s, prefix) {
			codes = append(codes, codeInfo{
				code:     s[len(prefix):],
				strategy: Proxy,
			})
		}
	}
	for _, s := range cfg.Router.Bypass {
		if strings.HasPrefix(s, prefix) {
			codes = append(codes, codeInfo{
				code:     s[len(prefix):],
				strategy: Bypass,
			})
		}
	}
	for _, s := range cfg.Router.Block {
		if strings.HasPrefix(s, prefix) {
			codes = append(codes, codeInfo{
				code:     s[len(prefix):],
				strategy: Block,
			})
		}
	}
	return codes
}

func NewClient(ctx context.Context, underlay tunnel.Client) (*Client, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	ctx, cancel := context.WithCancel(ctx)

	direct, err := raw.NewClient(ctx, nil)
	if err != nil {
		return nil, common.NewError("failed to initialize raw client").Base(err)
	}
	client := &Client{
		domains:  [3][]*v2router.Domain{},
		cidrs:    [3][]*v2router.CIDR{},
		underlay: underlay,
		direct:   direct,
		ctx:      ctx,
		cancel:   cancel,
	}
	switch cfg.Router.DomainStrategy {
	case "as_is", "as-is":
		client.domainStrategy = AsIs
	case "ip_if_non_match", "ip-if-non-match":
		client.domainStrategy = IPIfNonMatch
	case "ip_on_demand", "ip-on-demand":
		client.domainStrategy = IPOnDemand
	default:
		return nil, common.NewError("unknown strategy: " + cfg.Router.DomainStrategy)
	}

	switch cfg.Router.DefaultPolicy {
	case "proxy":
		client.defaultPolicy = Proxy
	case "bypass":
		client.defaultPolicy = Bypass
	case "block":
		client.defaultPolicy = Block
		return nil, common.NewError("unknown strategy: " + cfg.Router.DomainStrategy)
	}

	geoipData, err := ioutil.ReadFile(cfg.Router.GeoIPFilename)
	if err != nil {
		log.Warn(err)
	} else {
		geoip := new(v2router.GeoIPList)
		if err := proto.Unmarshal(geoipData, geoip); err != nil {
			return nil, err
		}
		ipCode := loadCode(cfg, "geoip:")
		for _, c := range ipCode {
			c.code = strings.ToUpper(c.code)
			found := false
			for _, e := range geoip.GetEntry() {
				code := e.GetCountryCode()
				if c.code == code {
					client.cidrs[c.strategy] = append(client.cidrs[c.strategy], e.GetCidr()...)
					found = true
					break
				}
			}
			if found {
				log.Info("geoip info", c, "loaded")
			} else {
				log.Warn("geoip info", c, "not found")
			}
		}
	}

	geositeData, err := ioutil.ReadFile(cfg.Router.GeoSiteFilename)
	if err != nil {
		log.Warn(err)
	} else {
		geosite := new(v2router.GeoSiteList)
		if err := proto.Unmarshal(geositeData, geosite); err != nil {
			return nil, err
		}
		siteCode := loadCode(cfg, "geosite:")
		for _, c := range siteCode {
			c.code = strings.ToUpper(c.code)
			found := false
			for _, e := range geosite.GetEntry() {
				code := e.GetCountryCode()
				if c.code == code {
					client.domains[c.strategy] = append(client.domains[c.strategy], e.GetDomain()...)
					found = true
					break
				}
			}
			if found {
				log.Info("geosite info", c, "loaded")
			} else {
				log.Warn("geosite info", c, "not found")
			}
		}
	}

	domainInfo := loadCode(cfg, "domain:")
	for _, info := range domainInfo {
		client.domains[info.strategy] = append(client.domains[info.strategy], &v2router.Domain{
			Type:      v2router.Domain_Domain,
			Value:     info.code,
			Attribute: nil,
		})
	}

	regexInfo := loadCode(cfg, "regex:")
	for _, info := range regexInfo {
		client.domains[info.strategy] = append(client.domains[info.strategy], &v2router.Domain{
			Type:      v2router.Domain_Regex,
			Value:     info.code,
			Attribute: nil,
		})
	}

	fullInfo := loadCode(cfg, "full:")
	for _, info := range fullInfo {
		client.domains[info.strategy] = append(client.domains[info.strategy], &v2router.Domain{
			Type:      v2router.Domain_Regex,
			Value:     info.code,
			Attribute: nil,
		})
	}

	cidrInfo := loadCode(cfg, "cidr:")
	for _, info := range cidrInfo {
		tmp := strings.Split(info.code, "/")
		if len(tmp) != 2 {
			return nil, common.NewError("invalid cidr:" + info.code)
		}
		ip := net.ParseIP(tmp[0])
		if ip == nil {
			return nil, common.NewError("invalid cidr ip:" + info.code)
		}
		prefix, err := strconv.ParseInt(tmp[1], 10, 32)
		if err != nil {
			return nil, common.NewError("invalid prefix").Base(err)
		}
		client.cidrs[info.strategy] = append(client.cidrs[info.strategy], &v2router.CIDR{
			Ip:     ip,
			Prefix: uint32(prefix),
		})
	}

	log.Info("router client created")
	return client, nil
}
