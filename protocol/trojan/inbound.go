package trojan

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/stat"
	"golang.org/x/net/websocket"
)

type TrojanInboundConnSession struct {
	protocol.ConnSession
	protocol.NeedAuth
	protocol.NeedMeter
	protocol.HasHash

	config        *conf.GlobalConfig
	request       *protocol.Request
	bufReadWriter *bufio.ReadWriter
	conn          io.ReadWriteCloser
	auth          stat.Authenticator
	meter         stat.TrafficMeter
	sent          int
	recv          int
	passwordHash  string
	ctx           context.Context
	cancel        context.CancelFunc
	readBytes     *bytes.Buffer
}

func (i *TrojanInboundConnSession) Write(p []byte) (int, error) {
	n, err := i.bufReadWriter.Write(p)
	i.bufReadWriter.Flush()
	i.sent += n
	return n, err
}

func (i *TrojanInboundConnSession) Read(p []byte) (int, error) {
	if i.readBytes != nil {
		n, err := i.readBytes.Read(p)
		if err == io.EOF {
			i.readBytes = nil
		}
		return n, err
	}
	n, err := i.bufReadWriter.Read(p)
	i.recv += n
	return n, err
}

func (i *TrojanInboundConnSession) Close() error {
	log.Info("user", i.passwordHash, "conn to", i.request, "closed", "sent:", common.HumanFriendlyTraffic(i.sent), "recv:", common.HumanFriendlyTraffic(i.recv))
	i.meter.Count(i.passwordHash, i.sent, i.recv)
	i.cancel()
	return i.conn.Close()
}

func (i *TrojanInboundConnSession) GetRequest() *protocol.Request {
	return i.request
}

func (i *TrojanInboundConnSession) GetHash() string {
	return i.passwordHash
}

func (i *TrojanInboundConnSession) parseRequest() error {
	userHash, err := i.bufReadWriter.Peek(56)
	if err != nil {
		return common.NewError("failed to read hash").Base(err)
	}
	if !i.auth.CheckHash(string(userHash)) {
		return common.NewError("invalid hash")
	}
	i.passwordHash = string(userHash)
	i.bufReadWriter.Discard(56 + 2)

	cmd, err := i.bufReadWriter.ReadByte()
	network := "tcp"
	switch protocol.Command(cmd) {
	case protocol.Connect, protocol.Mux:
		network = "tcp"
	case protocol.Associate:
		network = "udp"
	default:
		return common.NewError("invalid command")
	}
	if err != nil {
		return common.NewError("failed to read cmd").Base(err)
	}

	req, err := protocol.ParseAddress(i.bufReadWriter)
	if err != nil {
		return common.NewError("failed to parse address").Base(err)
	}
	req.Command = protocol.Command(cmd)
	req.NetworkType = network
	i.request = req

	i.bufReadWriter.Discard(2)
	return nil
}

//Fake response writer
//Websocket ServeHTTP method uses its Hijack method to get the Readwriter
type wsHttpResponseWriter struct {
	http.Hijacker
	http.ResponseWriter

	ReadWriter *bufio.ReadWriter
	Conn       net.Conn
}

func (w *wsHttpResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.Conn, w.ReadWriter, nil
}

func (i *TrojanInboundConnSession) parseWebsocket() (bool, error) {
	correct := "GET " + i.config.Websocket.Path
	first, err := i.bufReadWriter.Peek(len(correct))
	if err != nil {
		return false, err
	}
	if !bytes.Equal([]byte(correct), first) {
		//it may be a normal trojan conn
		log.Debug("not a ws conn", string(first))
		return true, common.NewError("invalid header")
	}

	httpRequest, err := http.ReadRequest(i.bufReadWriter.Reader)
	if err != nil {
		//malformed http request
		return false, err
	}

	url := "wss://" + i.config.Websocket.HostName + i.config.Websocket.Path
	origin := "https://" + i.config.Websocket.HostName
	wsConfig, err := websocket.NewConfig(url, origin)

	if httpRequest.URL.String() != i.config.Websocket.Path {
		log.Error("invalid websocket path, url", httpRequest.URL, "origin", httpRequest.Header.Get("Origin"))
		i.readBytes = bytes.NewBuffer([]byte{})
		httpRequest.Write(i.readBytes)
		return false, common.NewError("invalid url")
	}

	handshaked := make(chan struct{})

	var wsConn *websocket.Conn
	wsServer := websocket.Server{
		Config: *wsConfig,
		Handler: func(conn *websocket.Conn) {
			wsConn = conn //store the websocket after handshaking
			log.Debug("websocket obtained")
			handshaked <- struct{}{}
			//this function will NOT return unless the connection is ended
			//or the websocket will be closed by ServeHTTP method
			<-i.ctx.Done()
		},
		Handshake: func(wsConfig *websocket.Config, httpRequest *http.Request) error {
			log.Debug("websocket url", httpRequest.URL, "origin", httpRequest.Header.Get("Origin"))
			return nil
		},
	}

	responseWriter := &wsHttpResponseWriter{
		Conn:       i.conn.(net.Conn),
		ReadWriter: i.bufReadWriter,
	}
	go wsServer.ServeHTTP(responseWriter, httpRequest)

	select {
	case <-handshaked:
	case <-time.After(protocol.TCPTimeout):
	}

	if wsConn == nil {
		return false, common.NewError("failed to perform websocket handshake")
	}
	//setup new readwriter
	i.conn = wsConn
	i.bufReadWriter = common.NewBufReadWriter(wsConn)
	return true, nil
}

func (i *TrojanInboundConnSession) SetAuth(auth stat.Authenticator) {
	i.auth = auth
}

func (i *TrojanInboundConnSession) SetMeter(meter stat.TrafficMeter) {
	i.meter = meter
}

func NewInboundConnSession(conn net.Conn, config *conf.GlobalConfig, auth stat.Authenticator) (protocol.ConnSession, error) {
	ctx, cancel := context.WithCancel(context.Background())
	i := &TrojanInboundConnSession{
		config:        config,
		conn:          conn,
		bufReadWriter: common.NewBufReadWriter(conn),
		meter:         &stat.EmptyTrafficMeter{},
		auth:          auth,
		passwordHash:  "INVALID_HASH",
		ctx:           ctx,
		cancel:        cancel,
	}
	if i.config.Websocket.Enabled {
		validConn, err := i.parseWebsocket()
		if err == nil {
			log.Debug("websocket conn")
		}
		if !validConn {
			//no need to continue parsing
			i.request = &protocol.Request{
				IP:          i.config.RemoteIP,
				Port:        i.config.RemotePort,
				NetworkType: "tcp",
			}
			log.Warn("remote", conn.RemoteAddr(), "invalid websocket conn")
			return i, nil
		}
	}
	if err := i.parseRequest(); err != nil {
		i.request = &protocol.Request{
			IP:          i.config.RemoteIP,
			Port:        i.config.RemotePort,
			NetworkType: "tcp",
		}
		log.Warn("remote", conn.RemoteAddr(), "invalid hash or other protocol")
	}
	return i, nil
}
