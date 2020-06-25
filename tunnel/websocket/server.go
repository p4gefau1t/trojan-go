package websocket

import (
	"bufio"
	"context"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/redirector"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"golang.org/x/net/websocket"
)

// Fake response writer
// Websocket ServeHTTP method uses Hijack method to get the ReadWriter
type fakeHTTPResponseWriter struct {
	http.Hijacker
	http.ResponseWriter

	ReadWriter *bufio.ReadWriter
	Conn       net.Conn
}

func (w *fakeHTTPResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.Conn, w.ReadWriter, nil
}

type Server struct {
	underlay  tunnel.Server
	hostname  string
	path      string
	enabled   bool
	redirAddr net.Addr
	redir     *redirector.Redirector
	ctx       context.Context
	cancel    context.CancelFunc
	timeout   time.Duration
}

func (s *Server) Close() error {
	s.cancel()
	return s.underlay.Close()
}

func (s *Server) AcceptConn(tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := s.underlay.AcceptConn(&Tunnel{})
	if err != nil {
		return nil, common.NewError("websocket failed to accept connection from underlying server")
	}
	if !s.enabled {
		s.redir.Redirect(&redirector.Redirection{
			InboundConn: conn,
			RedirectTo:  s.redirAddr,
		})
		return nil, common.NewError("websocket is disabled. redirecting http request from " + conn.RemoteAddr().String())
	}
	rewindConn := common.NewRewindConn(conn)
	rewindConn.SetBufferSize(512)
	defer rewindConn.StopBuffering()
	rw := bufio.NewReadWriter(bufio.NewReader(rewindConn), bufio.NewWriter(rewindConn))
	req, err := http.ReadRequest(rw.Reader)

	if err != nil {
		log.Debug("invalid http request")
		rewindConn.Rewind()
		rewindConn.StopBuffering()
		s.redir.Redirect(&redirector.Redirection{
			InboundConn: rewindConn,
			RedirectTo:  s.redirAddr,
		})
		return nil, common.NewError("not a valid http request: " + conn.RemoteAddr().String()).Base(err)
	}
	if strings.ToLower(req.Header.Get("Upgrade")) != "websocket" || req.URL.Path != s.path {
		log.Debug("invalid http websocket handshake request")
		rewindConn.Rewind()
		rewindConn.StopBuffering()
		s.redir.Redirect(&redirector.Redirection{
			InboundConn: rewindConn,
			RedirectTo:  s.redirAddr,
		})
		return nil, common.NewError("not a valid websocket handshake request: " + conn.RemoteAddr().String()).Base(err)
	}

	handshake := make(chan struct{})

	url := "wss://" + s.hostname + s.path
	origin := "https://" + s.hostname
	wsConfig, err := websocket.NewConfig(url, origin)
	var wsConn *websocket.Conn
	ctx, cancel := context.WithCancel(s.ctx)

	wsServer := websocket.Server{
		Config: *wsConfig,
		Handler: func(conn *websocket.Conn) {
			wsConn = conn //store the websocket after handshaking
			log.Debug("websocket obtained")
			handshake <- struct{}{}
			// this function SHOULD NOT return unless the connection is ended
			// or the websocket will be closed by ServeHTTP method
			<-ctx.Done()
			log.Debug("websocket closed")
		},
		Handshake: func(wsConfig *websocket.Config, httpRequest *http.Request) error {
			log.Debug("websocket url", httpRequest.URL, "origin", httpRequest.Header.Get("Origin"))
			return nil
		},
	}

	respWriter := &fakeHTTPResponseWriter{
		Conn:       conn,
		ReadWriter: rw,
	}
	go wsServer.ServeHTTP(respWriter, req)

	select {
	case <-handshake:
	case <-time.After(s.timeout):
	}

	if wsConn == nil {
		cancel()
		return nil, common.NewError("websocket failed to handshake")
	}

	return &InboundConn{
		OutboundConn: OutboundConn{
			tcpConn: conn,
			Conn:    wsConn,
		},
		ctx:    ctx,
		cancel: cancel,
	}, nil
}

func (s *Server) AcceptPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	return nil, common.NewError("not supported")
}

func NewServer(ctx context.Context, underlay tunnel.Server) (*Server, error) {
	cfg := config.FromContext(ctx, Name).(*Config)
	if cfg.Websocket.Enabled {
		if !strings.HasPrefix(cfg.Websocket.Path, "/") {
			return nil, common.NewError("websocket path must start with \"/\"")
		}
	}
	if cfg.RemoteHost == "" {
		log.Warn("empty websocket redirection hostname")
		cfg.RemoteHost = cfg.Websocket.Hostname
	}
	if cfg.RemotePort == 0 {
		log.Warn("empty websocket redirection port")
		cfg.RemotePort = 80
	}
	ctx, cancel := context.WithCancel(ctx)
	log.Debug("websocket server created")
	return &Server{
		enabled:   cfg.Websocket.Enabled,
		hostname:  cfg.Websocket.Hostname,
		path:      cfg.Websocket.Path,
		ctx:       ctx,
		cancel:    cancel,
		underlay:  underlay,
		timeout:   time.Second * time.Duration(rand.Intn(10)+5),
		redir:     redirector.NewRedirector(ctx),
		redirAddr: tunnel.NewAddressFromHostPort("tcp", cfg.RemoteHost, cfg.RemotePort),
	}, nil
}
