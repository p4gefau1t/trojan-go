package trojan

import (
	"bufio"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha1"
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/net/websocket"
)

//this AES layer is used for obfuscation purpose only
type obfReadWriteCloser struct {
	*websocket.Conn
	r     cipher.StreamReader
	w     cipher.StreamWriter
	bufrw *bufio.ReadWriter
}

func (rwc *obfReadWriteCloser) Read(p []byte) (int, error) {
	return rwc.r.Read(p)
}

func (rwc *obfReadWriteCloser) Write(p []byte) (int, error) {
	n, err := rwc.w.Write(p)
	rwc.bufrw.Flush()
	return n, err
}

func (rwc *obfReadWriteCloser) Close() error {
	return rwc.Conn.Close()
}

func NewOutboundObfReadWriteCloser(password string, conn *websocket.Conn) *obfReadWriteCloser {
	//use bufio to avoid fixed ws packet length
	bufrw := common.NewBufioReadWriter(conn)
	randomBytes := [aes.BlockSize + 8]byte{}
	common.Must2(io.ReadFull(rand.Reader, randomBytes[:]))
	bufrw.Write(randomBytes[:])

	iv := randomBytes[:aes.BlockSize]
	salt := randomBytes[aes.BlockSize:]
	key := pbkdf2.Key([]byte(password), salt, 32, aes.BlockSize, sha1.New)
	block, err := aes.NewCipher(key)
	common.Must(err)

	return &obfReadWriteCloser{
		r: cipher.StreamReader{
			S: cipher.NewCTR(block, iv[:]),
			R: bufrw,
		},
		w: cipher.StreamWriter{
			S: cipher.NewCTR(block, iv[:]),
			W: bufrw,
		},
		Conn:  conn,
		bufrw: bufrw,
	}
}

func NewInboundObfReadWriteCloser(password string, conn *websocket.Conn) (*obfReadWriteCloser, error) {
	bufrw := common.NewBufioReadWriter(conn)
	randomBytes := [aes.BlockSize + 8]byte{}
	_, err := bufrw.Read(randomBytes[:])
	if err != nil {
		return nil, err
	}

	iv := randomBytes[:aes.BlockSize]
	salt := randomBytes[aes.BlockSize:]
	key := pbkdf2.Key([]byte(password), salt, 32, aes.BlockSize, sha1.New)
	block, err := aes.NewCipher(key)
	common.Must(err)

	return &obfReadWriteCloser{
		r: cipher.StreamReader{
			S: cipher.NewCTR(block, iv[:]),
			R: bufrw,
		},
		w: cipher.StreamWriter{
			S: cipher.NewCTR(block, iv[:]),
			W: bufrw,
		},
		Conn:  conn,
		bufrw: bufrw,
	}, nil
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

// TODO wrap this with a struct
var tlsSessionCache = tls.NewLRUClientSessionCache(-1)

func NewOutboundWebosocket(conn net.Conn, config *conf.GlobalConfig) (io.ReadWriteCloser, error) {
	url := "wss://" + config.Websocket.HostName + config.Websocket.Path
	origin := "https://" + config.Websocket.HostName
	wsConfig, err := websocket.NewConfig(url, origin)
	if err != nil {
		return nil, err
	}
	wsConn, err := websocket.NewClient(wsConfig, conn)
	if err != nil {
		return nil, err
	}
	var transport net.Conn = wsConn
	if config.Websocket.Obfsucation {
		transport = NewOutboundObfReadWriteCloser(config.Passwords[0], wsConn)
	}
	if !config.Websocket.DoubleTLS {
		return transport, nil
	}
	tlsConfig := &tls.Config{
		CipherSuites:           config.TLS.CipherSuites,
		RootCAs:                config.TLS.CertPool,
		ServerName:             config.TLS.SNI,
		SessionTicketsDisabled: !config.TLS.SessionTicket,
		ClientSessionCache:     tlsSessionCache,
		//InsecureSkipVerify:     !config.TLS.Verify, //must verify it
	}
	tlsConn := tls.Client(transport, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		return nil, err
	}
	if config.LogLevel == 0 {
		state := tlsConn.ConnectionState()
		chain := state.VerifiedChains
		log.Debug("websocket TLS handshaked", "cipher:", tls.CipherSuiteName(state.CipherSuite), "resume:", state.DidResume)
		for i := range chain {
			for j := range chain[i] {
				log.Debug("subject:", chain[i][j].Subject, ", issuer:", chain[i][j].Issuer)
			}
		}
	}
	return tlsConn, nil
}

func NewInboundWebsocket(ctx context.Context, conn net.Conn, r *common.RewindReader, config *conf.GlobalConfig) (io.ReadWriteCloser, error) {
	bufrw := bufio.NewReadWriter(bufio.NewReader(r), bufio.NewWriter(conn))
	httpRequest, err := http.ReadRequest(bufrw.Reader)
	if err != nil {
		return nil, nil
	}

	if httpRequest.Host != config.Websocket.HostName || httpRequest.URL.Path != config.Websocket.Path || httpRequest.Header.Get("Upgrade") != "websocket" {
		return nil, common.NewError("invalid ws url or hostname")
	}

	url := "wss://" + config.Websocket.HostName + config.Websocket.Path
	origin := "https://" + config.Websocket.HostName
	wsConfig, err := websocket.NewConfig(url, origin)

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
			<-ctx.Done()
			log.Debug("websocket closed")
		},
		Handshake: func(wsConfig *websocket.Config, httpRequest *http.Request) error {
			log.Debug("websocket url", httpRequest.URL, "origin", httpRequest.Header.Get("Origin"))
			return nil
		},
	}

	responseWriter := &wsHttpResponseWriter{
		Conn:       conn,
		ReadWriter: bufrw,
	}
	go wsServer.ServeHTTP(responseWriter, httpRequest)

	select {
	case <-handshaked:
	case <-time.After(protocol.TCPTimeout):
	}

	if wsConn == nil {
		return nil, common.NewError("failed to perform websocket handshake")
	}

	var transport net.Conn = wsConn
	if config.Websocket.Obfsucation {
		transport, err = NewInboundObfReadWriteCloser(config.Passwords[0], wsConn)
		if err != nil {
			return nil, common.NewError("failed to init obf layer").Base(err)
		}
	}
	if !config.Websocket.DoubleTLS {
		return transport, nil
	}
	tlsConfig := &tls.Config{
		Certificates:             config.TLS.KeyPair,
		CipherSuites:             config.TLS.CipherSuites,
		PreferServerCipherSuites: config.TLS.PreferServerCipher,
		SessionTicketsDisabled:   !config.TLS.SessionTicket,
	}
	tlsConn := tls.Server(transport, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		return nil, err
	}
	return tlsConn, nil
}
