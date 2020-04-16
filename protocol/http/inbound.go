package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
)

func parseHTTPRequest(httpRequest *http.Request) *protocol.Request {
	request := &protocol.Request{
		Address: &common.Address{
			NetworkType: "tcp",
			Port:        80,
		},
		Command: protocol.Connect,
	}
	host, port, err := net.SplitHostPort(httpRequest.Host)
	if err != nil {
		if ip := net.ParseIP(httpRequest.Host); ip != nil {
			request.IP = ip
			if ip.To4() != nil {
				request.AddressType = common.IPv4
			} else {
				request.AddressType = common.IPv6
			}
		} else {
			request.DomainName = httpRequest.Host
			request.AddressType = common.DomainName
		}
	} else {
		request.DomainName = host
		request.AddressType = common.DomainName
		n, err := strconv.ParseInt(port, 10, 16)
		common.Must(err)
		request.Port = int(n)
	}
	return request
}

type HTTPInboundTunnelConnSession struct {
	protocol.ConnSession
	protocol.NeedRespond

	request       *protocol.Request
	httpRequest   *http.Request
	conn          io.ReadWriteCloser
	bufReadWriter *bufio.ReadWriter
	bodyReader    io.Reader
}

func (i *HTTPInboundTunnelConnSession) GetRequest() *protocol.Request {
	return i.request
}

func (i *HTTPInboundTunnelConnSession) Read(p []byte) (int, error) {
	if n, err := i.bodyReader.Read(p); err == nil {
		return n, err
	}
	return i.bufReadWriter.Read(p)
}

func (i *HTTPInboundTunnelConnSession) Write(p []byte) (int, error) {
	n, err := i.bufReadWriter.Write(p)
	i.bufReadWriter.Flush()
	return n, err
}

func (i *HTTPInboundTunnelConnSession) Close() error {
	return i.conn.Close()
}

func (i *HTTPInboundTunnelConnSession) Respond() error {
	payload := fmt.Sprintf("HTTP/%d.%d 200 Connection established\r\n\r\n", i.httpRequest.ProtoMajor, i.httpRequest.ProtoMinor)
	_, err := i.Write([]byte(payload))
	return err
}

func (i *HTTPInboundTunnelConnSession) parseRequest() error {
	httpRequest, err := http.ReadRequest(i.bufReadWriter.Reader)
	if err != nil {
		return err
	}
	if httpRequest.Method != "CONNECT" {
		return common.NewError("not a connection")
	}
	i.bodyReader = httpRequest.Body
	i.httpRequest = httpRequest
	i.request = parseHTTPRequest(httpRequest)
	return nil
}

type HTTPInboundPacketSession struct {
	protocol.PacketSession

	conn          io.ReadWriteCloser
	bufReadWriter *bufio.ReadWriter
	request       *protocol.Request
	httpRequest   *http.Request
}

func (i *HTTPInboundPacketSession) Close() error {
	return i.conn.Close()
}

func (i *HTTPInboundPacketSession) GetRequest() *protocol.Request {
	httpRequest, err := http.ReadRequest(i.bufReadWriter.Reader)
	if err != nil {
		return nil
	}
	i.httpRequest = httpRequest
	i.request = parseHTTPRequest(httpRequest)
	return i.request
}

func (i *HTTPInboundPacketSession) ReadPacket() (*protocol.Request, []byte, error) {
	httpRequest, err := http.ReadRequest(i.bufReadWriter.Reader)
	if err != nil {
		return nil, nil, err
	}
	request := parseHTTPRequest(httpRequest)
	buf := bytes.NewBuffer([]byte{})
	err = httpRequest.Write(buf)
	common.Must(err)
	return request, buf.Bytes(), nil
}

func (i *HTTPInboundPacketSession) WritePacket(req *protocol.Request, packet []byte) (int, error) {
	n, err := i.bufReadWriter.Write(packet)
	i.bufReadWriter.Flush()
	return n, err
}

func NewHTTPInbound(conn io.ReadWriteCloser, rw *bufio.ReadWriter) (protocol.ConnSession, protocol.PacketSession, error) {
	var bufReadWriter *bufio.ReadWriter
	if rw == nil {
		bufReadWriter = common.NewBufReadWriter(conn)
	} else {
		bufReadWriter = rw
	}
	method, err := bufReadWriter.Peek(7)
	if err != nil {
		return nil, nil, err
	}
	if bytes.Equal(method, []byte("CONNECT")) {
		i := &HTTPInboundTunnelConnSession{
			bufReadWriter: bufReadWriter,
			conn:          conn,
		}
		if err := i.parseRequest(); err != nil {
			return nil, nil, common.NewError("failed to parse http header").Base(err)
		}
		return i, nil, nil
	}
	i := &HTTPInboundPacketSession{
		bufReadWriter: bufReadWriter,
		conn:          conn,
	}
	return nil, i, nil
}
