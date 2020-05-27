package proxy

import (
	"context"
	"io"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/router"
	"github.com/p4gefau1t/trojan-go/stat"
)

type Buildable interface {
	Build(config *conf.GlobalConfig) (common.Runnable, error)
}

func RelayConn(ctx context.Context, a, b io.ReadWriter, bufferSize int) {
	if a == nil || b == nil {
		log.Debug("Empty RW")
		return
	}
	errChan := make(chan error, 2)
	copyConn := func(dst io.Writer, src io.Reader) {
		buf := make([]byte, bufferSize)
		_, err := io.CopyBuffer(dst, src, buf)
		errChan <- err
	}
	go copyConn(a, b)
	go copyConn(b, a)
	select {
	case err := <-errChan:
		if err != nil {
			log.Debug(common.NewError("Conn relaying ends").Base(err))
		}
	case <-ctx.Done():
		return
	}
}

func RelayPacket(ctx context.Context, a, b protocol.PacketReadWriter) {
	if a == nil || b == nil {
		log.Debug("Empty RW")
		return
	}
	errChan := make(chan error, 2)
	copyPacket := func(dst protocol.PacketWriter, src protocol.PacketReader) {
		for {
			req, packet, err := src.ReadPacket()
			if err != nil {
				errChan <- err
				return
			}
			_, err = dst.WritePacket(req, packet)
			if err != nil {
				errChan <- err
				return
			}
		}
	}
	go copyPacket(a, b)
	go copyPacket(b, a)
	select {
	case err := <-errChan:
		log.Debug(common.NewError("Packet relaying ends").Base(err))
	case <-ctx.Done():
		return
	}
}

func RelayPacketWithRouter(ctx context.Context, from protocol.PacketReadWriter, table map[router.Policy]protocol.PacketReadWriter, router router.Router) {
	errChan := make(chan error, 1+len(table))
	copyPacket := func(dst protocol.PacketWriter, src protocol.PacketReader) {
		for {
			req, packet, err := src.ReadPacket()
			if err != nil {
				errChan <- err
				return
			}
			_, err = dst.WritePacket(req, packet)
			if err != nil {
				errChan <- err
				return
			}
		}
	}
	copyToDst := func() {
		for {
			req, packet, err := from.ReadPacket()
			if err != nil {
				errChan <- err
				return
			}
			policy, err := router.RouteRequest(req)
			if err != nil {
				errChan <- err
				return
			}
			to, found := table[policy]
			if !found {
				log.Debug("policy not found, skiped:", policy)
				continue
			}
			log.Debug("udp packet ", req, "routing policy:", policy)
			_, err = to.WritePacket(req, packet)
			if err != nil {
				errChan <- err
				return
			}
		}
	}

	for _, to := range table {
		go copyPacket(from, to)
	}
	go copyToDst()
	select {
	case err := <-errChan:
		log.Debug(common.NewError("Packet relaying with router ends").Base(err))
	case <-ctx.Done():
		return
	}
}

var proxys = make(map[conf.RunType]Buildable)

func NewProxy(config *conf.GlobalConfig) (common.Runnable, error) {
	runType := config.RunType
	if buildable, found := proxys[runType]; found {
		return buildable.Build(config)
	}
	return nil, common.NewError("Invalid run_type " + string(runType))
}

func RegisterProxy(t conf.RunType, b Buildable) {
	proxys[t] = b
}

type APIRunner func(context.Context, *conf.GlobalConfig, stat.Authenticator) error

var apis = make(map[conf.RunType]APIRunner)

func RegisterAPI(t conf.RunType, r APIRunner) {
	apis[t] = r
}

func RunAPIService(t conf.RunType, ctx context.Context, config *conf.GlobalConfig, auth stat.Authenticator) error {
	r, ok := apis[t]
	if !ok {
		return common.NewError("API module for" + string(t) + "not found")
	}
	return r(ctx, config, auth)
}
