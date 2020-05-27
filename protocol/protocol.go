package protocol

import (
	"bytes"
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
	Command
	*common.Address
}

func (r *Request) Marshal(rr io.Reader) error {
	byteBuf := [1]byte{}
	_, err := rr.Read(byteBuf[:])
	if err != nil {
		return err
	}
	r.Command = Command(byteBuf[0])
	switch r.Command {
	case Connect, Bind, Associate, Mux:
		r.Address = new(common.Address)
		err := r.Address.Marshal(rr)
		if err != nil {
			return common.NewError("Failed to marshal address").Base(err)
		}
	default:
		return common.NewError(fmt.Sprintf("Invalid command %d", r.Command))
	}
	return nil
}

func (r *Request) Unmarshal(w io.Writer) error {
	buf := bytes.NewBuffer(make([]byte, 0, 64))
	buf.WriteByte(byte(r.Command))
	if err := r.Address.Unmarshal(buf); err != nil {
		return err
	}
	//use tcp by default
	r.Address.NetworkType = "tcp"
	_, err := w.Write(buf.Bytes())
	return err
}

func (r *Request) Network() string {
	if r.Address != nil {
		return r.Address.Network()
	}
	return "empty"
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

type ConnSession interface {
	io.ReadWriteCloser
}

type PacketSession interface {
	PacketReadWriter
	io.Closer
}

var timeout time.Duration

func GetRandomTimeoutDuration() time.Duration {
	offset := time.Duration(rand.Intn(3000)) * time.Millisecond
	return timeout + offset
}

func SetRandomizedTimeout(conn net.Conn) {
	conn.SetDeadline(time.Now().Add(GetRandomTimeoutDuration()))
}

func CancelTimeout(conn net.Conn) {
	conn.SetDeadline(time.Time{})
}

func init() {
	timeout = time.Duration(rand.Intn(20))*time.Second + TCPTimeout
}
