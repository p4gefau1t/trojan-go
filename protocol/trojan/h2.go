package trojan

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"net/http"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/posener/h2conn"
	"golang.org/x/net/http2"
)

//just for fun

func NewH2InboundConn(ctx context.Context, conn net.Conn) (io.ReadWriteCloser, error) {
	rewindConn := common.NewRewindConn(conn)
	rewindConn.R.SetBufferSize(512)
	defer rewindConn.R.StopBuffering()
	framer := http2.NewFramer(nil, rewindConn)
	frame, err := framer.ReadFrame()
	if err != nil {
		return nil, err
	}
	log.Debug(frame.Header())
	rewindConn.R.Rewind()
	var newConn *h2conn.Conn
	errChan := make(chan error)

	h2Server := http2.Server{}
	go h2Server.ServeConn(rewindConn, &http2.ServeConnOpts{
		Context: ctx,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var err error
			newConn, err = h2conn.Accept(w, r)
			errChan <- err
			if err != nil {
				return
			}
			<-ctx.Done()
		}),
	})
	err = <-errChan
	if err != nil {
		return nil, err
	}
	return newConn, nil
}

func NewH2OutboundConn(ctx context.Context, conn net.Conn) (io.ReadWriteCloser, error) {
	httpClient := &http.Client{
		Transport: &http2.Transport{
			DialTLS: func(string, string, *tls.Config) (net.Conn, error) {
				return conn, nil
			},
		},
	}
	h2ConnClient := h2conn.Client{
		Client: httpClient,
	}
	newConn, resp, err := h2ConnClient.Connect(ctx, "https://trojan.server/testpath")
	log.Debug(resp)
	if err != nil {
		return nil, err
	}
	return newConn, nil
}
