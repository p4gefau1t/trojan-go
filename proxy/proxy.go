package proxy

import (
	"io"
	"os"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
)

var logger = log.New(os.Stdout)

type Buildable interface {
	Build(config *conf.GlobalConfig) (common.Runnable, error)
}

func copyConn(dst io.Writer, src io.Reader, errChan chan error) {
	_, err := io.Copy(dst, src)
	errChan <- err
}

func copyPacket(dst protocol.PacketWriter, src protocol.PacketReader, errChan chan error) {
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

func ProxyConn(a io.ReadWriteCloser, b io.ReadWriteCloser) {
	errChan := make(chan error, 2)
	go copyConn(a, b, errChan)
	go copyConn(b, a, errChan)
	err := <-errChan
	if err != nil {
		logger.Debug(common.NewError("conn proxy ends").Base(err))
	}
}

func ProxyPacket(a protocol.PacketReadWriter, b protocol.PacketReadWriter) {
	errChan := make(chan error, 2)
	go copyPacket(a, b, errChan)
	go copyPacket(b, a, errChan)
	err := <-errChan
	if err != nil {
		logger.Debug(common.NewError("packet proxy ends").Base(err))
	}
}

func copyPacketWithAliveChan(dst protocol.PacketWriter, src protocol.PacketReader, errChan chan error, aliveChan chan int) {
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
		aliveChan <- 1
	}
}

func ProxyPacketWithAliveChan(a protocol.PacketReadWriter, b protocol.PacketReadWriter, aliveChan chan int) {
	errChan := make(chan error, 2)
	go copyPacket(a, b, errChan)
	go copyPacket(b, a, errChan)
	err := <-errChan
	if err != nil {
		logger.Debug(common.NewError("packet proxy ends").Base(err))
	}
}

var buildableMap map[conf.RunType]Buildable = make(map[conf.RunType]Buildable)

func NewProxy(config *conf.GlobalConfig) (common.Runnable, error) {
	runType := config.RunType
	if buildable, found := buildableMap[runType]; found {
		return buildable.Build(config)
	}
	return nil, common.NewError("invalid run_type")
}

func RegisterProxy(t conf.RunType, b Buildable) {
	buildableMap[t] = b
}
