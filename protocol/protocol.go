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
type AddressType byte

const (
	Connect   Command = 1
	Bind      Command = 2
	Associate Command = 3
	Mux       Command = 0x7f
)

const (
	IPv4       AddressType = 1
	DomainName AddressType = 3
	IPv6       AddressType = 4
)

const (
	MaxUDPPacketSize = 1024 * 4
	UDPTimeout       = time.Second * 6
	TCPTimeout       = time.Second * 6
)

type Request struct {
	DomainName  []byte
	Port        uint16
	IP          net.IP
	AddressType AddressType
	NetworkType string
	Command     Command
	net.Addr
}

func (r *Request) Network() string {
	return r.NetworkType
}

func (r *Request) String() string {
	if r.DomainName == nil || len(r.DomainName) == 0 {
		if r.IP.To4() != nil {
			return fmt.Sprintf("%s:%d", r.IP.String(), r.Port)
		} else {
			return fmt.Sprintf("[%s]:%d", r.IP.String(), r.Port)
		}
	} else {
		return fmt.Sprintf("%s:%d", r.DomainName, r.Port)
	}
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
	atype := AddressType(buf1[0])
	req := &Request{
		AddressType: atype,
	}
	switch atype {
	case IPv4:
		var buf [6]byte
		_, err := io.ReadFull(r, buf[:])
		if err != nil {
			return nil, common.NewError("failed to read ipv4").Base(err)
		}
		req.IP = buf[0:4]
		req.Port = binary.BigEndian.Uint16(buf[4:6])
	case IPv6:
		var buf [18]byte
		_, err := io.ReadFull(r, buf[:])
		if err != nil {
			return nil, common.NewError("failed to read ipv6").Base(err)
		}
		req.IP = buf[0:16]
		req.Port = binary.BigEndian.Uint16(buf[16:18])
	case DomainName:
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
				req.AddressType = IPv4
			} else {
				req.AddressType = IPv6
			}
		} else {
			req.DomainName = host
		}
		req.Port = binary.BigEndian.Uint16(buf[length : length+2])
	default:
		return nil, common.NewError("invalid dest type")
	}
	return req, nil
}

func WriteAddress(w io.Writer, request *Request) error {
	_, err := w.Write([]byte{byte(request.AddressType)})
	switch request.AddressType {
	case DomainName:
		w.Write([]byte{byte((len(request.DomainName)))})
		_, err = w.Write(request.DomainName)
	case IPv4:
		_, err = w.Write(request.IP.To4())
	case IPv6:
		_, err = w.Write(request.IP.To16())
	default:
		return common.NewError("invalid address type")
	}
	port := [2]byte{}
	binary.BigEndian.PutUint16(port[:], request.Port)
	w.Write(port[:])
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
