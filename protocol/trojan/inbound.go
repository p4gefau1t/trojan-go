package trojan

import (
	"context"
	"io"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/shadow"
	"github.com/p4gefau1t/trojan-go/stat"
)

type TrojanInboundConnSession struct {
	protocol.ConnSession

	rwc          io.ReadWriteCloser
	ctx          context.Context
	config       *conf.GlobalConfig
	request      *protocol.Request
	auth         stat.Authenticator
	meter        stat.TrafficMeter
	sent         uint64
	recv         uint64
	passwordHash string
	cancel       context.CancelFunc
}

func (i *TrojanInboundConnSession) Write(p []byte) (int, error) {
	n, err := i.rwc.Write(p)
	i.sent += uint64(n)
	i.meter.Count(n, 0)
	return n, err
}

func (i *TrojanInboundConnSession) Read(p []byte) (int, error) {
	n, err := i.rwc.Read(p)
	i.recv += uint64(n)
	i.meter.Count(0, n)
	return n, err
}

func (i *TrojanInboundConnSession) Close() error {
	log.Info("User", i.passwordHash, "to", i.request, "closed", "sent:", common.HumanFriendlyTraffic(i.sent), "recv:", common.HumanFriendlyTraffic(i.recv))
	i.cancel()
	return i.rwc.Close()
}

func (i *TrojanInboundConnSession) parseRequest(r *common.RewindReader) error {
	userHash := [56]byte{}

	n, err := r.Read(userHash[:])
	if err != nil || n != 56 {
		return common.NewError("Failed to read hash").Base(err)
	}

	valid, meter := i.auth.AuthUser(string(userHash[:]))
	if !valid {
		return common.NewError("Invalid hash:" + string(userHash[:]))
	}
	i.passwordHash = string(userHash[:])
	i.meter = meter

	crlf := [2]byte{}
	r.Read(crlf[:])

	i.request = new(protocol.Request)
	if err := i.request.Marshal(r); err != nil {
		return err
	}
	r.Read(crlf[:])
	return nil
}

func (i *TrojanInboundConnSession) SetMeter(meter stat.TrafficMeter) {
	i.meter = meter
}

func NewInboundConnSession(ctx context.Context, conn net.Conn, config *conf.GlobalConfig, auth stat.Authenticator, shadowMan *shadow.ShadowManager) (protocol.ConnSession, *protocol.Request, error) {
	ctx, cancel := context.WithCancel(context.Background())
	rewindConn := common.NewRewindConn(conn)

	i := &TrojanInboundConnSession{
		config:       config,
		auth:         auth,
		passwordHash: "INVALID_HASH",
		ctx:          ctx,
		cancel:       cancel,
		rwc:          rewindConn,
	}

	//start buffering
	rewindConn.R.SetBufferSize(512)
	defer rewindConn.R.StopBuffering()

	if i.config.Websocket.Enabled {
		//try to treat it as a websocket connection first
		ws, err := NewInboundWebsocket(i.ctx, rewindConn, config, shadowMan)
		if err != nil {
			return nil, nil, common.NewError("Invalid websocket request").Base(err)
		}
		if ws != nil {
			//a websocket conn, try to verify it
			log.Debug("websocket conn")
			//disable the current read buffer, use ws as the new transport layer
			rewindConn.R.SetBufferSize(0)
			newTrapsport := common.NewRewindReadWriteCloser(ws)
			i.rwc = newTrapsport
			//parse it with trojan protocol format
			if err := i.parseRequest(newTrapsport.RewindReader); err != nil {
				//invalid ws, just simply close it
				ws.Close()
				return nil, nil, common.NewError("Invalid trojan header over websocket conn").Base(err)
			}
			return i, i.request, nil
		}
		//not a websocket conn, it might be a normal trojan conn
		rewindConn.R.Rewind()
	}

	//normal trojan conn
	if err := i.parseRequest(rewindConn.R); err != nil {
		//not a valid trojan request, proxy it to the remote_addr
		rewindConn.R.Rewind()
		err := common.NewError("Invalid trojan header from " + conn.RemoteAddr().String()).Base(err)
		shadowMan.SubmitScapegoat(&shadow.Scapegoat{
			Conn:          rewindConn,
			ShadowAddress: i.config.RemoteAddress,
			Info:          err.Error(),
		})
		return nil, nil, err
	}
	//release the buffer
	rewindConn.R.SetBufferSize(0)
	return i, i.request, nil
}
