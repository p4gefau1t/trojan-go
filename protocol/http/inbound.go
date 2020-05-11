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
		n, err := strconv.ParseUint(port, 10, 16)
		common.Must(err)
		request.Port = int(n)
	}
	return request
}

type HTTPInboundTunnelConnSession struct {
	protocol.ConnSession
	protocol.NeedRespond

	request     *protocol.Request
	httpRequest *http.Request
	bufReader   *bufio.Reader
	rwc         io.ReadWriteCloser
	bodyReader  io.Reader
}

func (i *HTTPInboundTunnelConnSession) Read(p []byte) (int, error) {
	return i.bufReader.Read(p)
}

func (i *HTTPInboundTunnelConnSession) Write(p []byte) (int, error) {
	return i.rwc.Write(p)
}

func (i *HTTPInboundTunnelConnSession) Close() error {
	return i.rwc.Close()
}

func (i *HTTPInboundTunnelConnSession) Respond() error {
	payload := fmt.Sprintf("HTTP/%d.%d 200 Connection established\r\n\r\n", i.httpRequest.ProtoMajor, i.httpRequest.ProtoMinor)
	_, err := i.Write([]byte(payload))
	return err
}

func (i *HTTPInboundTunnelConnSession) parseRequest() (bool, error) {
	httpRequest, err := http.ReadRequest(i.bufReader)
	if err != nil {
		return false, err
	}
	if httpRequest.Method != "CONNECT" {
		return true, common.NewError("Not a connection")
	}
	i.bodyReader = httpRequest.Body
	i.httpRequest = httpRequest
	i.request = parseHTTPRequest(httpRequest)
	return true, nil
}

type HTTPInboundPacketSession struct {
	protocol.PacketSession

	rwc         io.ReadWriteCloser
	bufReader   *bufio.Reader
	request     *protocol.Request
	httpRequest *http.Request
}

func (i *HTTPInboundPacketSession) Close() error {
	return i.rwc.Close()
}

func (i *HTTPInboundPacketSession) ReadPacket() (*protocol.Request, []byte, error) {
	httpRequest, err := http.ReadRequest(i.bufReader)
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
	n, err := i.rwc.Write(packet)
	return n, err
}

func NewHTTPInbound(rwc *common.RewindReadWriteCloser) (protocol.ConnSession, *protocol.Request, protocol.PacketSession, error) {
	connSession := &HTTPInboundTunnelConnSession{
		rwc:       rwc,
		bufReader: bufio.NewReader(rwc),
	}
	rwc.SetBufferSize(512)
	defer rwc.StopBuffering()
	isHTTP, err := connSession.parseRequest()
	if !isHTTP {
		//invalid http format
		rwc.SetBufferSize(0)
		return nil, nil, nil, common.NewError("Failed to parse http header").Base(err)
	}
	if err == nil {
		//http tunnel
		rwc.SetBufferSize(0)
		return connSession, connSession.request, nil, nil
	}
	rwc.Rewind()
	packetSession := &HTTPInboundPacketSession{
		rwc:       rwc,
		bufReader: bufio.NewReader(rwc),
	}
	// TODO release the read buffer
	return nil, nil, packetSession, nil
}
