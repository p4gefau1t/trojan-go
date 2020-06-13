package util

import (
	"crypto/rand"
	"fmt"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"golang.org/x/net/websocket"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

var HTTPAddr string
var HTTPPort string

func runHelloHTTPServer() {
	httpHello := func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("HelloWorld"))
	}

	wsConfig, err := websocket.NewConfig("wss://127.0.0.1/websocket", "https://127.0.0.1")
	common.Must(err)
	wsServer := websocket.Server{
		Config: *wsConfig,
		Handler: func(conn *websocket.Conn) {
			conn.Write([]byte("HelloWorld"))
		},
		Handshake: func(wsConfig *websocket.Config, httpRequest *http.Request) error {
			log.Debug("websocket url", httpRequest.URL, "origin", httpRequest.Header.Get("Origin"))
			return nil
		},
	}
	mux := &http.ServeMux{}
	mux.HandleFunc("/", httpHello)
	mux.HandleFunc("/websocket", wsServer.ServeHTTP)
	HTTPAddr = GetTestAddr()
	_, HTTPPort, _ = net.SplitHostPort(HTTPAddr)
	server := http.Server{
		Addr:    HTTPAddr,
		Handler: mux,
	}
	go server.ListenAndServe()
	time.Sleep(time.Second * 3) // wait for http server
	fmt.Println("http test server listening on", HTTPAddr)
	wg.Done()
}

var EchoAddr string
var EchoPort int

func runTCPEchoServer() {
	listener, err := net.Listen("tcp", EchoAddr)
	common.Must(err)
	wg.Done()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go func(conn net.Conn) {
				for {
					conn.SetDeadline(time.Now().Add(time.Second))
					buf := make([]byte, 2048)
					n, err := conn.Read(buf)
					if err != nil {
						return
					}
					_, err = conn.Write(buf[0:n])
					if err != nil {
						return
					}
				}
			}(conn)
		}
	}()
}

func runUDPEchoServer() {
	conn, err := net.ListenPacket("udp", EchoAddr)
	common.Must(err)
	wg.Done()
	go func() {
		for {
			buf := make([]byte, 1024*8)
			n, addr, err := conn.ReadFrom(buf[:])
			if err != nil {
				return
			}
			log.Info("Echo from", addr)
			conn.WriteTo(buf[0:n], addr)
		}
	}()
}

func GeneratePayload(length int) []byte {
	buf := make([]byte, length)
	io.ReadFull(rand.Reader, buf)
	return buf
}

var wg = sync.WaitGroup{}

func init() {
	wg.Add(3)
	runHelloHTTPServer()
	EchoPort = common.PickPort("tcp", "127.0.0.1")
	EchoAddr = fmt.Sprintf("127.0.0.1:%d", EchoPort)
	runTCPEchoServer()
	runUDPEchoServer()
	wg.Wait()
}
