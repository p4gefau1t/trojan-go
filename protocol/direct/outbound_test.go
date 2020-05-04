package direct

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/test"
)

func TestUDPDirectOutbound(t *testing.T) {
	go test.RunMultipleUDPEchoServer(context.Background())
	outbound, _ := NewOutboundPacketSession(context.Background())
	go func() {
		for i := 0; i < 5; i++ {
			req, buf, err := outbound.ReadPacket()
			fmt.Println(req, string(buf), err)
		}
	}()
	for i := 0; i < 5; i++ {
		req := &protocol.Request{
			Address: &common.Address{
				IP:          net.ParseIP("127.0.0.1"),
				Port:        6000 + rand.Intn(10),
				AddressType: common.IPv4,
			},
		}
		req.Port += rand.Intn(10)
		packet := []byte(fmt.Sprintf("hello motherfucker %d, port=%d", i, req.Port))
		_, err := outbound.WritePacket(req, packet)
		common.Must(err)
	}
	time.Sleep(time.Second * 5)
}

func TestDNS(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &conf.GlobalConfig{
		DNS: []string{"114.114.114.114:53"},
	}
	req := &protocol.Request{
		Address: &common.Address{
			DomainName:  "www.baidu.com",
			Port:        80,
			AddressType: common.DomainName,
			NetworkType: "tcp",
		},
	}
	conn, err := NewOutboundConnSession(ctx, req, config)
	common.Must(err)
	httpReq, err := http.NewRequest("GET", "http://www.baidu.com", nil)
	common.Must(err)
	httpReq.Write(conn)
	buf := [128]byte{}
	conn.Read(buf[:])
	fmt.Println(string(buf[:]))
	cancel()
}

func TestDOT(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &conf.GlobalConfig{
		DNS: []string{"dot://223.5.5.5:853"},
	}
	req := &protocol.Request{
		Address: &common.Address{
			DomainName:  "www.baidu.com",
			Port:        80,
			AddressType: common.DomainName,
			NetworkType: "tcp",
		},
	}
	conn, err := NewOutboundConnSession(ctx, req, config)
	common.Must(err)
	httpReq, err := http.NewRequest("GET", "http://www.baidu.com", nil)
	common.Must(err)
	httpReq.Write(conn)
	buf := [128]byte{}
	conn.Read(buf[:])
	fmt.Println(string(buf[:]))
	cancel()
}

func TestDOH(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &conf.GlobalConfig{
		DNS: []string{"https://223.5.5.5:443"},
	}
	req := &protocol.Request{
		Address: &common.Address{
			DomainName:  "www.baidu.com",
			Port:        80,
			AddressType: common.DomainName,
			NetworkType: "tcp",
		},
	}
	conn, err := NewOutboundConnSession(ctx, req, config)
	common.Must(err)
	httpReq, err := http.NewRequest("GET", "http://www.baidu.com", nil)
	common.Must(err)
	httpReq.Write(conn)
	buf := [128]byte{}
	conn.Read(buf[:])
	fmt.Println(string(buf[:]))
	cancel()
}

func TestCache(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	config := &conf.GlobalConfig{
		DNS: []string{"223.5.5.5:53"},
	}
	req := &protocol.Request{
		Address: &common.Address{
			DomainName:  "www.baidu.com",
			Port:        80,
			AddressType: common.DomainName,
			NetworkType: "tcp",
		},
	}
	conn, err := NewOutboundConnSession(ctx, req, config)
	common.Must(err)
	conn.Close()
	conn, err = NewOutboundConnSession(ctx, req, config)
	common.Must(err)
	conn.Close()
	cancel()
}
