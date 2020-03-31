package http

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/protocol"
)

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
	defer i.bufReadWriter.Flush()
	return i.bufReadWriter.Write(p)
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
	i.request = &protocol.Request{
		NetworkType: "tcp",
		Port:        80,
		Command:     protocol.Connect,
	}
	host, port, err := net.SplitHostPort(httpRequest.Host)
	if err != nil {
		if ip := net.ParseIP(httpRequest.Host); ip != nil {
			i.request.IP = ip
			if ip.To16() != nil {
				i.request.AddressType = protocol.IPv6
			} else {
				i.request.AddressType = protocol.IPv4
			}
		} else {
			i.request.DomainName = []byte(httpRequest.Host)
			i.request.AddressType = protocol.DomainName
		}
	} else {
		i.request.DomainName = []byte(host)
		i.request.AddressType = protocol.DomainName
		fmt.Sscanf(port, "%d", &i.request.Port)
	}
	return nil
}

func parseHTTPRequest(httpRequest *http.Request) *protocol.Request {
	request := &protocol.Request{
		NetworkType: "tcp",
		Port:        80,
		Command:     protocol.Connect,
	}
	host, port, err := net.SplitHostPort(httpRequest.Host)
	if err != nil {
		if ip := net.ParseIP(httpRequest.Host); ip != nil {
			request.IP = ip
			if ip.To16() != nil {
				request.AddressType = protocol.IPv6
			} else {
				request.AddressType = protocol.IPv4
			}
		} else {
			request.DomainName = []byte(httpRequest.Host)
			request.AddressType = protocol.DomainName
		}
	} else {
		request.DomainName = []byte(host)
		request.AddressType = protocol.DomainName
		fmt.Sscanf(port, "%d", &request.Port)
	}
	return request
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

func (i *HTTPInboundPacketSession) Respond(r io.Reader) error {
	return nil
}

func (i *HTTPInboundPacketSession) ReadPacket() (*protocol.Request, []byte, error) {
	httpRequest, err := http.ReadRequest(i.bufReadWriter.Reader)
	if err != nil {
		return nil, nil, err
	}
	request := &protocol.Request{
		NetworkType: "tcp",
		Port:        80,
		Command:     protocol.Connect,
	}
	host, port, err := net.SplitHostPort(httpRequest.Host)
	if err != nil {
		if ip := net.ParseIP(httpRequest.Host); ip != nil {
			request.IP = ip
			if ip.To16() != nil {
				request.AddressType = protocol.IPv6
			} else {
				request.AddressType = protocol.IPv4
			}
		} else {
			request.DomainName = []byte(httpRequest.Host)
			request.AddressType = protocol.DomainName
		}
	} else {
		request.DomainName = []byte(host)
		request.AddressType = protocol.DomainName
		fmt.Sscanf(port, "%d", &request.Port)
	}
	//packet, err := httputil.DumpRequest(httpRequest, true)
	buf := bytes.NewBuffer([]byte{})
	err = httpRequest.Write(buf)
	common.Must(err)
	return request, buf.Bytes(), nil
}

func (i *HTTPInboundPacketSession) WritePacket(req *protocol.Request, packet []byte) (int, error) {
	defer i.bufReadWriter.Flush()
	return i.bufReadWriter.Write(packet)
}
