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
	"github.com/p4gefau1t/trojan-go/shadow"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/net/websocket"
)

//this AES layer is used for obfuscation purpose only
type obfReadWriteCloser struct {
	net.Conn
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

func NewInboundObfReadWriteCloser(password string, conn net.Conn) (*obfReadWriteCloser, error) {
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
	if config.Websocket.Obfuscation {
		log.Debug("ws obfs enabled")
		transport = NewOutboundObfReadWriteCloser(config.Passwords[0], wsConn)
	}
	if !config.Websocket.DoubleTLS {
		return transport, nil
	}
	log.Debug("ws double tls enabled")
	tlsConfig := &tls.Config{
		CipherSuites:           config.TLS.CipherSuites,
		RootCAs:                config.TLS.CertPool,
		ServerName:             config.TLS.SNI,
		SessionTicketsDisabled: !config.TLS.SessionTicket,
		ClientSessionCache:     tlsSessionCache,
		InsecureSkipVerify:     !config.Websocket.DoubleTLSVerify,
	}
	tlsConn := tls.Client(transport, tlsConfig)
	if err := tlsConn.Handshake(); err != nil {
		return nil, err
	}
	if config.LogLevel == 0 {
		state := tlsConn.ConnectionState()
		chain := state.VerifiedChains
		log.Debug("websocket tls handshaked", "cipher:", tls.CipherSuiteName(state.CipherSuite), "resume:", state.DidResume)
		for i := range chain {
			for j := range chain[i] {
				log.Debug("subject:", chain[i][j].Subject, ", issuer:", chain[i][j].Issuer)
			}
		}
	}
	return tlsConn, nil
}

func dialToWebosocketServer(config *conf.GlobalConfig, url, origin string) (*websocket.Conn, error) {
	wsConfig, err := websocket.NewConfig(url, origin)
	if err != nil {
		return nil, err
	}
	conn, err := net.Dial("tcp", config.RemoteAddress.String())
	if err != nil {
		return nil, err
	}
	newWsConn, err := websocket.NewClient(wsConfig, conn)
	if err != nil {
		return nil, err
	}
	return newWsConn, nil
}

func getWebsocketScapegoat(config *conf.GlobalConfig, url, origin, info string, conn net.Conn) (*shadow.Scapegoat, error) {
	shadowConn, err := dialToWebosocketServer(config, url, origin)
	if err != nil {
		return nil, err
	}
	return &shadow.Scapegoat{
		Conn:       conn,
		ShadowConn: shadowConn,
		Info:       info,
	}, nil
}

func NewInboundWebsocket(ctx context.Context, conn net.Conn, config *conf.GlobalConfig, shadowMan *shadow.ShadowManager) (io.ReadWriteCloser, error) {
	rewindConn := common.NewRewindConn(conn)
	rewindConn.R.SetBufferSize(512)
	defer rewindConn.R.StopBuffering()

	bufrw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	httpRequest, obfErr := http.ReadRequest(bufrw.Reader)
	if obfErr != nil {
		log.Debug(common.NewError("not a http request:").Base(obfErr))
		return nil, nil
	}

	//this is a http request
	if (config.Websocket.HostName != "" && httpRequest.Host != config.Websocket.HostName) || //check hostname
		httpRequest.URL.Path != config.Websocket.Path || //check url path
		httpRequest.Header.Get("Upgrade") != "websocket" { //check upgrade field
		//not a valid websocket conn
		rewindConn.R.Rewind()
		shadowMan.CommitScapegoat(&shadow.Scapegoat{
			Conn:          rewindConn,
			ShadowAddress: config.RemoteAddress,
			Info:          "not a valid http upgrade request from " + conn.RemoteAddr().String(),
		})
		return nil, common.NewError("invalid ws url or hostname")
	}

	//this is a websocket upgrade request
	//no need to record the recv content for now
	rewindConn.R.SetBufferSize(0)
	url := "wss://" + config.Websocket.HostName + config.Websocket.Path
	origin := "https://" + config.Websocket.HostName
	wsConfig, obfErr := websocket.NewConfig(url, origin)

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
		//conn has been closed at this point
		return nil, common.NewError("failed to perform websocket handshake")
	}

	var transport net.Conn
	transport = common.NewRewindConn(wsConn)

	//start buffering the websocket payload
	rewindConn.R.SetBufferSize(512)
	defer rewindConn.R.StopBuffering()

	if config.Websocket.Obfuscation {
		log.Debug("ws obfs")

		//deadline for sending the iv and hash
		rewindConn.SetDeadline(time.Now().Add(protocol.TCPTimeout))
		transport, obfErr = NewInboundObfReadWriteCloser(config.Passwords[0], rewindConn)
		rewindConn.SetDeadline(time.Time{})

		if obfErr != nil {
			rewindConn.R.Rewind()
			//proxy this to our own ws server
			obfErr = common.NewError("remote websocket conn:" + conn.RemoteAddr().String() + "didn't send any valid iv/hash").Base(obfErr)
			goat, err := getWebsocketScapegoat(
				config,
				url,
				origin,
				obfErr.Error(),
				rewindConn,
			)
			if err != nil {
				log.Error(common.NewError("failed to obtain websocket scapegoat").Base(err))
				wsConn.WriteClose(500)
			} else {
				shadowMan.CommitScapegoat(goat)
			}
			return nil, obfErr
		}
	}
	if !config.Websocket.DoubleTLS {
		rewindConn.R.SetBufferSize(0)
		return transport, nil
	}
	tlsConfig := &tls.Config{
		Certificates:             config.TLS.KeyPair,
		CipherSuites:             config.TLS.CipherSuites,
		PreferServerCipherSuites: config.TLS.PreferServerCipher,
		SessionTicketsDisabled:   !config.TLS.SessionTicket,
	}
	tlsConn := tls.Server(transport, tlsConfig)
	if tlsErr := tlsConn.Handshake(); tlsErr != nil {
		rewindConn.R.Rewind()
		//proxy this to our own ws server
		tlsErr = common.NewError("invalid double tls handshake from" + conn.RemoteAddr().String()).Base(tlsErr)
		goat, err := getWebsocketScapegoat(
			config,
			url,
			origin,
			tlsErr.Error(),
			rewindConn,
		)
		if err != nil {
			log.Error(common.NewError("failed to obtain websocket scapegoat").Base(err))
			wsConn.WriteClose(500)
		} else {
			shadowMan.CommitScapegoat(goat)
		}
		return nil, tlsErr
	}
	rewindConn.R.SetBufferSize(0)
	return tlsConn, nil
}
