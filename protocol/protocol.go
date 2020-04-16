package protocol

import (
	"encoding/binary"
	"fmt"
	"io"
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

type HasRequest interface {
	GetRequest() *Request
}

type HasHash interface {
	GetHash() string
}

type NeedRespond interface {
	Respond() error
}

type PacketReader interface {
	ReadPacket() (*Request, []byte, error)
}

type PacketWriter interface {
	WritePacket(req *Request, packet []byte) (int, error)
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
	HasRequest
}

type PacketSession interface {
	PacketReadWriter
	io.Closer
}

func ParseAddress(r io.Reader) (*Request, error) {
	var buf1 [1]byte
	_, err := io.ReadFull(r, buf1[:])
	if err != nil {
		return nil, common.NewError("cannot read atype").Base(err)
	}
	atype := common.AddressType(buf1[0])
	req := &Request{
		Address: &common.Address{
			AddressType: atype,
		},
	}
	switch atype {
	case common.IPv4:
		var buf [6]byte
		_, err := io.ReadFull(r, buf[:])
		if err != nil {
			return nil, common.NewError("failed to read ipv4").Base(err)
		}
		req.IP = buf[0:4]
		req.Port = int(binary.BigEndian.Uint16(buf[4:6]))
	case common.IPv6:
		var buf [18]byte
		_, err := io.ReadFull(r, buf[:])
		if err != nil {
			return nil, common.NewError("failed to read ipv6").Base(err)
		}
		req.IP = buf[0:16]
		req.Port = int(binary.BigEndian.Uint16(buf[16:18]))
	case common.DomainName:
		_, err := io.ReadFull(r, buf1[:])
		if err != nil {
			return nil, common.NewError("failed to read length")
		}
		length := buf1[0]
		buf := make([]byte, length+2)
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return nil, common.NewError("failed to read domain")
		}
		//the fucking browser uses ip as a domain name sometimes
		host := buf[0:length]
		if ip := net.ParseIP(string(host)); ip != nil {
			req.IP = ip
			if ip.To4() != nil {
				req.AddressType = common.IPv4
			} else {
				req.AddressType = common.IPv6
			}
		} else {
			req.DomainName = string(host)
		}
		req.Port = int(binary.BigEndian.Uint16(buf[length : length+2]))
	default:
		return nil, common.NewError("invalid dest type")
	}
	return req, nil
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
