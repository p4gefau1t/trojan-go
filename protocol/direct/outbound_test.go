package direct

import (
	"fmt"
	"math/rand"
	"net"
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/test"
)

func TestDirectOutbound(t *testing.T) {
	for i := 0; i < 10; i++ {
		go test.RunEchoUDPServer(6543 + i)
	}
	outbound, _ := NewOutboundPacketSession()
	for i := 0; i < 30; i++ {
		req := &protocol.Request{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 6543,
		}
		req.Port += rand.Intn(10)
		packet := []byte(fmt.Sprintf("hello motherfucker %d, port=%d", i, req.Port))
		_, err := outbound.WritePacket(req, packet)
		common.Must(err)
	}
	for i := 0; i < 30; i++ {
		req, buf, err := outbound.ReadPacket()
		logger.Info(req, string(buf), err)
	}
}
