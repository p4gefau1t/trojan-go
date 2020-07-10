package tunnel

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/p4gefau1t/trojan-go/common"
)

type Command byte

type Metadata struct {
	Command
	*Address
}

func (r *Metadata) ReadFrom(rr io.Reader) error {
	byteBuf := [1]byte{}
	_, err := io.ReadFull(rr, byteBuf[:])
	if err != nil {
		return err
	}
	r.Command = Command(byteBuf[0])
	r.Address = new(Address)
	err = r.Address.ReadFrom(rr)
	if err != nil {
		return common.NewError("failed to marshal address").Base(err)
	}
	return nil
}

func (r *Metadata) WriteTo(w io.Writer) error {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	buf.WriteByte(byte(r.Command))
	if err := r.Address.WriteTo(buf); err != nil {
		return err
	}
	//use tcp by default
	r.Address.NetworkType = "tcp"
	_, err := w.Write(buf.Bytes())
	return err
}

func (r *Metadata) Network() string {
	return r.Address.Network()
}

func (r *Metadata) String() string {
	return r.Address.String()
}

type AddressType byte

const (
	IPv4       AddressType = 1
	DomainName AddressType = 3
	IPv6       AddressType = 4
)

type Address struct {
	DomainName  string
	Port        int
	NetworkType string
	net.IP
	AddressType
}

func (a *Address) String() string {
	switch a.AddressType {
	case IPv4:
		return fmt.Sprintf("%s:%d", a.IP.String(), a.Port)
	case IPv6:
		return fmt.Sprintf("[%s]:%d", a.IP.String(), a.Port)
	case DomainName:
		return fmt.Sprintf("%s:%d", a.DomainName, a.Port)
	default:
		return "INVALID_ADDRESS_TYPE"
	}
}

func (a *Address) Network() string {
	return a.NetworkType
}

func (a *Address) ResolveIP() (net.IP, error) {
	if a.AddressType == IPv4 || a.AddressType == IPv6 {
		return a.IP, nil
	}
	if a.IP != nil {
		return a.IP, nil
	}
	addr, err := net.ResolveIPAddr("ip", a.DomainName)
	if err != nil {
		return nil, err
	}
	a.IP = addr.IP
	return addr.IP, nil
}

func NewAddressFromAddr(network string, addr string) (*Address, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.ParseInt(portStr, 10, 32)
	common.Must(err)
	return NewAddressFromHostPort(network, host, int(port)), nil
}

func NewAddressFromHostPort(network string, host string, port int) *Address {
	if ip := net.ParseIP(host); ip != nil {
		if ip.To4() != nil {
			return &Address{
				IP:          ip,
				Port:        int(port),
				AddressType: IPv4,
				NetworkType: network,
			}
		}
		return &Address{
			IP:          ip,
			Port:        int(port),
			AddressType: IPv6,
			NetworkType: network,
		}
	}
	return &Address{
		DomainName:  host,
		Port:        int(port),
		AddressType: DomainName,
		NetworkType: network,
	}
}

func (a *Address) ReadFrom(r io.Reader) error {
	byteBuf := [1]byte{}
	_, err := io.ReadFull(r, byteBuf[:])
	if err != nil {
		return common.NewError("unable to read ATYP").Base(err)
	}
	a.AddressType = AddressType(byteBuf[0])
	switch a.AddressType {
	case IPv4:
		var buf [6]byte
		_, err := io.ReadFull(r, buf[:])
		if err != nil {
			return common.NewError("failed to read IPv4").Base(err)
		}
		a.IP = buf[0:4]
		a.Port = int(binary.BigEndian.Uint16(buf[4:6]))
	case IPv6:
		var buf [18]byte
		_, err := io.ReadFull(r, buf[:])
		if err != nil {
			return common.NewError("failed to read IPv6").Base(err)
		}
		a.IP = buf[0:16]
		a.Port = int(binary.BigEndian.Uint16(buf[16:18]))
	case DomainName:
		_, err := io.ReadFull(r, byteBuf[:])
		length := byteBuf[0]
		if err != nil {
			return common.NewError("failed to read domain name length")
		}
		buf := make([]byte, length+2)
		_, err = io.ReadFull(r, buf)
		if err != nil {
			return common.NewError("failed to read domain name")
		}
		//the fucking browser uses IP as a domain name sometimes
		host := buf[0:length]
		if ip := net.ParseIP(string(host)); ip != nil {
			a.IP = ip
			if ip.To4() != nil {
				a.AddressType = IPv4
			} else {
				a.AddressType = IPv6
			}
		} else {
			a.DomainName = string(host)
		}
		a.Port = int(binary.BigEndian.Uint16(buf[length : length+2]))
	default:
		return common.NewError("invalid ATYP " + strconv.FormatInt(int64(a.AddressType), 10))
	}
	return nil
}

func (a *Address) WriteTo(w io.Writer) error {
	_, err := w.Write([]byte{byte(a.AddressType)})
	switch a.AddressType {
	case DomainName:
		w.Write([]byte{byte(len(a.DomainName))})
		_, err = w.Write([]byte(a.DomainName))
	case IPv4:
		_, err = w.Write(a.IP.To4())
	case IPv6:
		_, err = w.Write(a.IP.To16())
	default:
		return common.NewError("invalid ATYP " + strconv.FormatInt(int64(a.AddressType), 10))
	}
	if err != nil {
		return err
	}
	port := [2]byte{}
	binary.BigEndian.PutUint16(port[:], uint16(a.Port))
	_, err = w.Write(port[:])
	return err
}
