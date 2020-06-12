package dokodemo

import (
	"context"
	"fmt"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/test/util"
	"net"
	"sync"
	"testing"
)

func TestDokodemo(t *testing.T) {
	cfg := &Config{
		LocalHost:  "127.0.0.1",
		LocalPort:  common.PickPort("tcp", "127.0.0.1"),
		TargetHost: "127.0.0.1",
		TargetPort: common.PickPort("tcp", "127.0.0.1"),
		UDPTimeout: 30,
	}
	ctx := config.WithConfig(context.Background(), Name, cfg)
	s, err := NewServer(ctx, nil)
	common.Must(err)
	conn1, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", cfg.LocalPort))
	common.Must(err)
	conn2, err := s.AcceptConn(nil)
	common.Must(err)
	if !util.CheckConn(conn1, conn2) {
		t.Fail()
	}
	conn1.Close()
	conn2.Close()

	wg := sync.WaitGroup{}
	wg.Add(1)

	packet1, err := net.ListenPacket("udp", "")
	common.Must(err)
	common.Must2(packet1.(*net.UDPConn).WriteToUDP([]byte("hello1"), &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: cfg.LocalPort,
	}))
	packet2, err := s.AcceptPacket(nil)
	buf := [100]byte{}
	n, m, err := packet2.ReadWithMetadata(buf[:])
	if m.Address.Port != cfg.TargetPort {
		t.Fail()
	}
	if string(buf[:n]) != "hello1" {
		t.Fail()
	}
	fmt.Println(n, m, string(buf[:n]))

	if !util.CheckPacket(packet1, packet2) {
		t.Fail()
	}

	packet3, err := net.ListenPacket("udp", "")
	common.Must(err)
	common.Must2(packet3.(*net.UDPConn).WriteToUDP([]byte("hello2"), &net.UDPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: cfg.LocalPort,
	}))
	packet4, err := s.AcceptPacket(nil)
	n, m, err = packet4.ReadWithMetadata(buf[:])
	if m.Address.Port != cfg.TargetPort {
		t.Fail()
	}
	if string(buf[:n]) != "hello2" {
		t.Fail()
	}
	fmt.Println(n, m, string(buf[:n]))

	wg = sync.WaitGroup{}
	wg.Add(2)
	go func() {
		if !util.CheckPacket(packet3, packet4) {
			t.Fail()
		}
		wg.Done()
	}()
	go func() {
		if !util.CheckPacket(packet1, packet2) {
			t.Fail()
		}
		wg.Done()
	}()
	wg.Wait()
}
