package proxy

import (
	"context"
	"io"
	"net"
	"strings"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

const Name = "PROXY"

const (
	MaxPacketSize = 1024 * 8
)

// Proxy relay connections and packets
type Proxy struct {
	sources []tunnel.Server
	sink    tunnel.Client
	errChan chan error
	ctx     context.Context
	cancel  context.CancelFunc
}

func (p *Proxy) Run() error {
	p.relayConnLoop()
	p.relayPacketLoop()
	return <-p.errChan
}

func (p *Proxy) Close() error {
	p.cancel()
	for _, source := range p.sources {
		source.Close()
	}
	return p.sink.Close()
}

func (p *Proxy) relayConnLoop() {
	for _, source := range p.sources {
		go func(source tunnel.Server) {
			for {
				inbound, err := source.AcceptConn(nil)
				if err != nil {
					select {
					case <-p.ctx.Done():
						log.Debug("exiting")
						return
					default:
					}
					log.Error(common.NewError("failed to accept connection").Base(err))
					continue
				}
				go func(inbound tunnel.Conn) {
					defer inbound.Close()
					outbound, err := p.sink.DialConn(inbound.Metadata().Address, nil)
					if err != nil {
						log.Error(err)
						return
					}
					defer outbound.Close()
					errChan := make(chan error, 2)
					copyConn := func(a, b net.Conn) {
						_, err := io.Copy(a, b)
						errChan <- err
						return
					}
					go copyConn(inbound, outbound)
					go copyConn(outbound, inbound)
					err = <-errChan
					if err != nil {
						log.Error(err)
					}
					log.Debug("conn relay ends")
				}(inbound)
			}
		}(source)
	}
}

func (p *Proxy) relayPacketLoop() {
	for _, source := range p.sources {
		go func(source tunnel.Server) {
			for {
				inbound, err := source.AcceptPacket(nil)
				if err != nil {
					select {
					case <-p.ctx.Done():
						log.Debug("exiting")
						return
					default:
					}
					log.Error(common.NewError("failed to accept packet").Base(err))
					continue
				}
				go func(inbound tunnel.PacketConn) {
					defer inbound.Close()
					outbound, err := p.sink.DialPacket(nil)
					if err != nil {
						log.Error(err)
						return
					}
					defer outbound.Close()
					errChan := make(chan error, 2)
					copyPacket := func(a, b tunnel.PacketConn) {
						buf := make([]byte, MaxPacketSize)
						n, metadata, err := a.ReadWithMetadata(buf)
						if err != nil {
							errChan <- err
							return
						}
						n, err = b.WriteWithMetadata(buf[:n], metadata)
						if err != nil {
							errChan <- err
							return
						}
					}
					go copyPacket(inbound, outbound)
					go copyPacket(outbound, inbound)
					err = <-errChan
					if err != nil {
						log.Error(err)
					}
					log.Debug("packet relay ends")
				}(inbound)
			}
		}(source)
	}
}

func NewProxy(ctx context.Context, sources []tunnel.Server, sink tunnel.Client) *Proxy {
	ctx, cancel := context.WithCancel(ctx)
	return &Proxy{
		sources: sources,
		sink:    sink,
		errChan: make(chan error, 32),
		ctx:     ctx,
		cancel:  cancel,
	}
}

type Creator func(ctx context.Context) (*Proxy, error)

var creators = make(map[string]Creator)

func RegisterProxyCreator(name string, creator Creator) {
	creators[name] = creator
}

func NewProxyFromConfigData(data []byte, isJSON bool) (*Proxy, error) {
	ctx := context.Background()
	var err error
	if isJSON {
		ctx, err = config.WithJSONConfig(context.Background(), data)
		if err != nil {
			return nil, err
		}
	} else {
		ctx, err = config.WithYAMLConfig(context.Background(), data)
		if err != nil {
			return nil, err
		}
	}
	cfg := config.FromContext(ctx, Name).(*Config)
	create, ok := creators[strings.ToUpper(cfg.RunType)]
	if !ok {
		return nil, common.NewError("unknown proxy type: " + cfg.RunType)
	}
	log.SetLogLevel(log.LogLevel(cfg.LogLevel))
	if cfg.LogFile != "" {

	}
	return create(ctx)
}
