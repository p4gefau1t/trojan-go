package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
	"math/rand"
	"net"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/stat"
)

type Command byte

const (
	Connect   Command = 1
	Bind      Command = 2
	Associate Command = 3
	Mux       Command = 0x7f
)

const (
	MaxUDPPacketSize = 1024 * 4
	UDPTimeout       = time.Second * 5
	TCPTimeout       = time.Second * 5
)

type Request struct {
	net.Addr

	*common.Address
	Command Command
}

func (r *Request) Network() string {
	return r.Address.Network()
}

func (r *Request) String() string {
	return r.Address.String()
}

type HasHash interface {
	GetHash() string
}

type NeedRespond interface {
	Respond() error
}

type PacketReader interface {
	ReadPacket() (req *Request, payload []byte, err error)
}

type PacketWriter interface {
	WritePacket(req *Request, payload []byte) (n int, err error)
}

type PacketReadWriter interface {
	PacketReader
	PacketWriter
}

type NeedConfig interface {
	SetConfig(config *conf.GlobalConfig)
}

type NeedAuth interface {
	SetAuth(auth stat.Authenticator)
}

type NeedMeter interface {
	SetMeter(meter stat.TrafficMeter)
}

type ConnSession interface {
	io.ReadWriteCloser
}

type PacketSession interface {
	PacketReadWriter
	io.Closer
}

func ParseAddress(conn io.Reader, network string) (*common.Address, error) {
	byteBuf := [1]byte{}
	_, err := conn.Read(byteBuf[:])
	if err != nil {
		return nil, common.NewError("cannot read atype").Base(err)
	}
	addr := &common.Address{
		AddressType: common.AddressType(byteBuf[0]),
	}
	switch addr.AddressType {
	case common.IPv4:
		var buf [6]byte
		_, err := conn.Read(buf[:])
		if err != nil {
			return nil, common.NewError("failed to read ipv4").Base(err)
		}
		addr.IP = buf[0:4]
		addr.Port = int(binary.BigEndian.Uint16(buf[4:6]))
	case common.IPv6:
		var buf [18]byte
		conn.Read(buf[:])
		if err != nil {
			return nil, common.NewError("failed to read ipv6").Base(err)
		}
		addr.IP = buf[0:16]
		addr.Port = int(binary.BigEndian.Uint16(buf[16:18]))
	case common.DomainName:
		_, err := conn.Read(byteBuf[:])
		length := byteBuf[0]
		if err != nil {
			return nil, common.NewError("failed to read length")
		}
		buf := make([]byte, length+2)
		_, err = conn.Read(buf)
		if err != nil {
			return nil, common.NewError("failed to read domain")
		}
		//the fucking browser uses ip as a domain name sometimes
		host := buf[0:length]
		if ip := net.ParseIP(string(host)); ip != nil {
			addr.IP = ip
			if ip.To4() != nil {
				addr.AddressType = common.IPv4
			} else {
				addr.AddressType = common.IPv6
			}
		} else {
			addr.DomainName = string(host)
		}
		addr.Port = int(binary.BigEndian.Uint16(buf[length : length+2]))
	default:
		return nil, common.NewError("invalid dest type")
	}
	addr.NetworkType = network
	return addr, nil
}

func WriteAddress(w io.Writer, request *Request) error {
	_, err := w.Write([]byte{byte(request.AddressType)})
	switch request.AddressType {
	case common.DomainName:
		w.Write([]byte{byte((len(request.DomainName)))})
		_, err = w.Write([]byte(request.DomainName))
	case common.IPv4:
		_, err = w.Write(request.IP.To4())
	case common.IPv6:
		_, err = w.Write(request.IP.To16())
	default:
		return common.NewError("invalid address type")
	}
	if err != nil {
		return err
	}
	port := [2]byte{}
	binary.BigEndian.PutUint16(port[:], uint16(request.Port))
	_, err = w.Write(port[:])
	return err
}

func ParsePort(addr net.Addr) (uint16, error) {
	_, portStr, err := net.SplitHostPort(addr.String())
	if err != nil {
		return 0, err
	}
	var port uint16
	_, err = fmt.Sscanf(portStr, "%d", &port)
	return port, err
}

var timeout time.Duration

func RandomizedTimeout(conn net.Conn) {
	conn.SetDeadline(time.Now().Add(timeout))
}

func CancelTimeout(conn net.Conn) {
	conn.SetDeadline(time.Time{})
}

func init() {
	timeout = time.Duration(rand.Intn(20))*time.Second + TCPTimeout
}
