package socks

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
)

const (
	Connect   tunnel.Command = 1
	Associate tunnel.Command = 3
)

const (
	MaxPacketSize = 1024 * 8
)

// Server is a socks5 server
type Server struct {
	connChan   chan tunnel.Conn
	packetChan chan tunnel.PacketConn
	underlay   tunnel.Server
	localHost  string
	timeout    time.Duration
	ctx        context.Context
	cancel     context.CancelFunc
}

func (s *Server) AcceptConn(tunnel.Tunnel) (tunnel.Conn, error) {
	select {
	case conn := <-s.connChan:
		return conn, nil
	case <-s.ctx.Done():
		return nil, common.NewError("socks server closed")
	}
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	select {
	case conn := <-s.packetChan:
		return conn, nil
	case <-s.ctx.Done():
		return nil, common.NewError("socks server closed")
	}
}

func (s *Server) Close() error {
	s.cancel()
	return s.underlay.Close()
}

func (s *Server) handshake(conn net.Conn) (*Conn, error) {
	version := [1]byte{}
	if _, err := conn.Read(version[:]); err != nil {
		return nil, common.NewError("failed to read socks version").Base(err)
	}
	if version[0] != 5 {
		return nil, common.NewError(fmt.Sprintf("invalid socks version %d", version[0]))
	}
	nmethods := [1]byte{}
	if _, err := conn.Read(nmethods[:]); err != nil {
		return nil, common.NewError("failed to read NMETHODS")
	}
	if _, err := io.CopyN(ioutil.Discard, conn, int64(nmethods[0])); err != nil {
		return nil, common.NewError("socks5 failed to read methods").Base(err)
	}
	if _, err := conn.Write([]byte{0x5, 0x0}); err != nil {
		return nil, common.NewError("failed to respond auth").Base(err)
	}

	buf := [3]byte{}
	if _, err := conn.Read(buf[:]); err != nil {
		return nil, common.NewError("failed to read command")
	}

	addr := new(tunnel.Address)
	if err := addr.ReadFrom(conn); err != nil {
		return nil, err
	}

	return &Conn{
		metadata: &tunnel.Metadata{
			Command: tunnel.Command(buf[1]),
			Address: addr,
		},
		Conn: conn,
	}, nil
}

func (s *Server) connect(conn net.Conn) error {
	_, err := conn.Write([]byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	return err
}

func (s *Server) associate(conn net.Conn, addr *tunnel.Address) error {
	buf := bytes.NewBuffer([]byte{0x05, 0x00, 0x00})
	common.Must(addr.WriteTo(buf))
	_, err := conn.Write(buf.Bytes())
	return err
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.underlay.AcceptConn(&Tunnel{})
		if err != nil {
			log.Error(common.NewError("socks5 accept err").Base(err))
			return
		}
		go func(conn net.Conn) {
			newConn, err := s.handshake(conn)
			if err != nil {
				log.Error(common.NewError("socks5 failed to handshake with client").Base(err))
				return
			}

			metadata := newConn.metadata.String()
			log.Info("socks5 connection from", conn.RemoteAddr(), "metadata", metadata)
			if strings.HasPrefix(metadata, "0.") {
				return
			}

			switch newConn.metadata.Command {
			case Connect:
				if err := s.connect(newConn); err != nil {
					log.Error(common.NewError("socks5 failed to respond CONNECT").Base(err))
					newConn.Close()
					return
				}
				s.connChan <- newConn
				return
			case Associate:
				defer newConn.Close()
				port := common.PickPort("udp", s.localHost)
				// TODO use underlying server
				associateAddr := tunnel.NewAddressFromHostPort("udp", s.localHost, port)
				l, err := net.ListenPacket("udp", associateAddr.String())
				if err != nil {
					log.Error(common.NewError("socks5 failed to bind udp").Base(err))
					return
				}
				s.packetChan <- NewPacketConn(l, s.timeout)
				log.Info("socks5 udp session")
				if err := s.associate(newConn, associateAddr); err != nil {
					log.Error(common.NewError("socks5 failed to respond to associate request").Base(err))
					return
				}
				buf := [1]byte{}
				newConn.Read(buf[:])
				log.Debug("socks5 udp session ends")
			default:
				log.Error(common.NewError(fmt.Sprintf("unknown socks5 command %d", newConn.metadata.Command)))
				newConn.Close()
			}
		}(conn)
	}
}

// NewServer create a socks server
func NewServer(ctx context.Context, underlay tunnel.Server) (tunnel.Server, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	ctx, cancel := context.WithCancel(ctx)
	server := &Server{
		underlay:   underlay,
		ctx:        ctx,
		cancel:     cancel,
		connChan:   make(chan tunnel.Conn, 32),
		packetChan: make(chan tunnel.PacketConn, 32),
		timeout:    time.Duration(cfg.UDPTimeout) * time.Second,
	}
	go server.acceptLoop()
	log.Debug("socks server created")
	return server, nil
}
