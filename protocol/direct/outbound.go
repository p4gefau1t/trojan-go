package direct

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/patrickmn/go-cache"
)

var dnsCache = cache.New(5*time.Minute, 1*time.Minute)

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
	var err error
	//look up the domain name in cache first
	if req.AddressType == common.DomainName && len(config.DNS) != 0 { //customized dns server
		ip, found := dnsCache.Get(req.DomainName)
		if found {
			log.Trace("DNS cache hit:", req.DomainName, "->", ip.(net.IP).String())
			newConn, err = net.DialTCP("tcp", nil, &net.TCPAddr{
				IP:   ip.(net.IP),
				Port: req.Port,
			})
			if err != nil {
				return nil, err
			}
			goto done
		}
		log.Trace("DNS cache missed:", req.DomainName)
		//find a avaliable dns server
		for _, s := range config.DNS {
			var dnsType conf.DNSType
			var dnsAddr string
			var dnsHost, dnsPort string
			var err error
			dnsURL, err := url.Parse(s)
			if err != nil || dnsURL.Scheme == "" {
				dnsType = conf.UDP
				dnsAddr = s
			} else {
				dnsType = conf.DNSType(dnsURL.Scheme)
				dnsAddr = dnsURL.Host
			}

			dnsHost, dnsPort, err = net.SplitHostPort(dnsAddr)
			if err != nil { //port not specifiet
				dnsHost = dnsAddr
				switch dnsType {
				case conf.DOT:
					dnsPort = "853"
				case conf.TCP, conf.UDP:
					dnsPort = "53"
				}
			}

			resolver := &net.Resolver{
				PreferGo: true,
				Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
					switch dnsType {
					case conf.UDP, conf.TCP:
						d := net.Dialer{
							Timeout: time.Second * time.Duration(protocol.UDPTimeout),
						}
						conn, err := d.DialContext(ctx, string(dnsType), dnsHost+":"+dnsPort)
						if err != nil {
							return nil, err
						}
						return conn, nil
					case conf.DOT:
						tlsConn, err := tls.Dial("tcp", dnsHost+":"+dnsPort, nil)
						if err != nil {
							return nil, err
						}
						return tlsConn, nil
					}
					return nil, common.NewError("Invalid dns type :" + string(dnsType))
				},
			}
			d := net.Dialer{
				Resolver: resolver,
			}
			newConn, err = d.Dial("tcp", req.DomainName+":"+strconv.FormatInt(int64(req.Port), 10))
			if err != nil {
				log.Error(err)
				continue
			}
			addr, _, err := net.SplitHostPort(newConn.RemoteAddr().String())
			if err != nil {
				log.Warn(err)
			} else {
				if ip := net.ParseIP(addr); ip != nil {
					log.Trace("DNS cache set", req.DomainName, "->", addr)
					dnsCache.Set(req.DomainName, ip, cache.DefaultExpiration)
				} else {
					log.Warn("Invalid resolved addr", addr)
				}
			}
			break
		}
		if newConn == nil {
			return nil, common.NewError("All dns servers down")
		}
	} else {
		//default resolver
		var err error
		newConn, err = net.Dial("tcp", req.String())
		if err != nil {
			return nil, err
		}
	}
done:
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
			log.Debug(common.NewError("Packet session ends").Base(err))
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
		return nil, nil, common.NewError("Session closed")
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
		return 0, common.NewError("Failed to dial udp").Base(err)
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
