package direct

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/url"
	"time"

	"github.com/babolivier/go-doh-client"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
)

type DirectOutboundConnSession struct {
	protocol.ConnSession
	conn    io.ReadWriteCloser
	request *protocol.Request
}

func (o *DirectOutboundConnSession) Read(p []byte) (int, error) {
	return o.conn.Read(p)
}

func (o *DirectOutboundConnSession) Write(p []byte) (int, error) {
	return o.conn.Write(p)
}

func (o *DirectOutboundConnSession) Close() error {
	return o.conn.Close()
}

func NewOutboundConnSession(ctx context.Context, req *protocol.Request, config *conf.GlobalConfig) (protocol.ConnSession, error) {
	var newConn net.Conn
	//custom dns server
	if req.AddressType == common.DomainName && len(config.DNS) != 0 {
		//find a avaliable dns server
		for _, s := range config.DNS {
			var dnsType conf.DNSType
			var dnsAddr string
			dnsURL, err := url.Parse(s)
			if err != nil {
				dnsType = conf.UDP
				dnsAddr = s
			} else {
				dnsType = conf.DNSType(dnsURL.Scheme)
				dnsAddr = dnsURL.Host
			}

			if dnsType == conf.DOH {
				resolver := doh.Resolver{
					Host:  dnsURL.Host,
					Class: doh.IN,
				}
				result := []string{}
				a, _, err := resolver.LookupA(req.DomainName)
				if err != nil {
					log.Error(err)
					continue
				}
				if !config.TCP.PreferIPV4 {
					aaaa, _, err := resolver.LookupAAAA(req.DomainName)
					if err != nil {
						log.Error(err)
						continue
					}
					for _, record := range aaaa {
						result = append(result, record.IP6)
					}
				}
				for _, record := range a {
					result = append(result, record.IP4)
				}
				if len(result) == 0 {
					log.Error("a record not found for" + req.DomainName)
					continue
				}
				for _, ip := range result {
					newConn, err = net.DialTCP("tcp", nil, &net.TCPAddr{
						IP:   net.ParseIP(ip),
						Port: req.Port,
					})
					if err != nil {
						return nil, err
					}
					break
				}
			} else {
				resolver := &net.Resolver{
					PreferGo: true,
					Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
						switch dnsType {
						case conf.UDP, conf.TCP:
							d := net.Dialer{
								Timeout: time.Second * time.Duration(protocol.UDPTimeout),
							}
							conn, err := d.DialContext(ctx, string(dnsType), dnsAddr)
							if err != nil {
								return nil, err
							}
							return conn, nil
						case conf.DOT:
							tlsConn, err := tls.Dial("tcp", dnsAddr, nil)
							if err != nil {
								return nil, err
							}
							return tlsConn, nil
						}
						return nil, common.NewError("invalid dns type" + string(dnsType))
					},
				}
				ips, err := resolver.LookupIPAddr(ctx, req.DomainName)
				if err != nil {
					log.Debug("dns server " + s + " sucks")
					continue
				}
				log.Debug("dns connected:" + s)
				if len(ips) == 0 {
					return nil, common.NewError("record of " + req.DomainName + " not found in dns server " + s)
				}
				for _, ip := range ips {
					newConn, err = net.DialTCP("tcp", nil, &net.TCPAddr{
						IP:   ip.IP,
						Port: req.Port,
					})
					if err != nil {
						return nil, err
					}
					break
				}
			}
			break
		}
		if newConn == nil {
			return nil, common.NewError("all dns servers are down")
		}
	} else {
		//default resolver
		var err error
		newConn, err = net.Dial(req.Network(), req.String())
		if err != nil {
			return nil, err
		}
	}
	o := &DirectOutboundConnSession{
		request: req,
		conn:    newConn,
	}
	return o, nil
}

type packetInfo struct {
	request *protocol.Request
	packet  []byte
}

type DirectOutboundPacketSession struct {
	protocol.PacketSession
	packetChan chan *packetInfo
	ctx        context.Context
	cancel     context.CancelFunc
}

func (o *DirectOutboundPacketSession) listenConn(req *protocol.Request, conn *net.UDPConn) {
	defer conn.Close()
	for {
		buf := make([]byte, protocol.MaxUDPPacketSize)
		conn.SetReadDeadline(time.Now().Add(protocol.UDPTimeout))
		n, addr, err := conn.ReadFromUDP(buf)
		conn.SetReadDeadline(time.Time{})
		if err != nil {
			log.Info(err)
			return
		}
		if addr.String() != req.String() {
			panic("addr != req, something went wrong")
		}
		info := &packetInfo{
			request: req,
			packet:  buf[0:n],
		}
		o.packetChan <- info
	}
}

func (o *DirectOutboundPacketSession) Close() error {
	o.cancel()
	return nil
}

func (o *DirectOutboundPacketSession) ReadPacket() (*protocol.Request, []byte, error) {
	select {
	case info := <-o.packetChan:
		return info.request, info.packet, nil
	case <-o.ctx.Done():
		return nil, nil, common.NewError("session closed")
	}
}

func (o *DirectOutboundPacketSession) WritePacket(req *protocol.Request, packet []byte) (int, error) {
	var remote *net.UDPAddr
	if req.AddressType == common.DomainName {
		remote, err := net.ResolveUDPAddr("", string(req.DomainName))
		if err != nil {
			return 0, err
		}
		remote.Port = req.Port
	} else {
		remote = &net.UDPAddr{
			IP:   req.IP,
			Port: req.Port,
		}
	}
	conn, err := net.DialUDP("udp", nil, remote)
	if err != nil {
		return 0, common.NewError("cannot dial udp").Base(err)
	}
	log.Debug("udp directly dialing to", remote)
	go o.listenConn(req, conn)
	n, err := conn.Write(packet)
	return n, err
}

func NewOutboundPacketSession(ctx context.Context) (protocol.PacketSession, error) {
	ctx, cancel := context.WithCancel(ctx)
	return &DirectOutboundPacketSession{
		ctx:        ctx,
		cancel:     cancel,
		packetChan: make(chan *packetInfo, 256),
	}, nil
}
