package proxy

import (
	"io"
	"log"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
)

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

func proxyConn(a io.ReadWriteCloser, b io.ReadWriteCloser) {
	errChan := make(chan error, 2)
	go copyConn(a, b, errChan)
	go copyConn(b, a, errChan)
	err := <-errChan
	if err != nil {
		log.Println(err)
	}
	time.Sleep(time.Second * 3)
}

func proxyPacket(a protocol.PacketReadWriter, b protocol.PacketReadWriter) {
	errChan := make(chan error, 2)
	go copyPacket(a, b, errChan)
	go copyPacket(b, a, errChan)
	err := <-errChan
	if err != nil {
		log.Println(err)
	}
	time.Sleep(time.Second * 3)
}

func NewProxy(config *conf.GlobalConfig) common.Runnable {
	switch config.RunType {
	case conf.ClientRunType:
		client := &Client{
			config: config,
		}
		return client
	case conf.ServerRunType:
		server := &Server{
			config: config,
		}
		return server
	case conf.ForwardRunType:
		forward := &Forward{
			config: config,
		}
		return forward
	default:
		panic("invalid run type")
	}
}
