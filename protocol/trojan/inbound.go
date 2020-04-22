package trojan

import (
	"context"
	"io"
	"net"

	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/conf"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/protocol"
	"github.com/p4gefau1t/trojan-go/stat"
)

type TrojanInboundConnSession struct {
	protocol.ConnSession
	protocol.NeedAuth
	protocol.NeedMeter

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
	if i.meter != nil {
		i.meter.Count(i.passwordHash, uint64(n), 0)
	}
	i.sent += uint64(n)
	return n, err
}

func (i *TrojanInboundConnSession) Read(p []byte) (int, error) {
	n, err := i.rwc.Read(p)
	if i.meter != nil {
		i.meter.Count(i.passwordHash, 0, uint64(n))
	}
	i.recv += uint64(n)
	return n, err
}

func (i *TrojanInboundConnSession) Close() error {
	log.Info("user", i.passwordHash, "conn to", i.request, "closed", "sent:", common.HumanFriendlyTraffic(i.sent), "recv:", common.HumanFriendlyTraffic(i.recv))
	i.cancel()
	return i.rwc.Close()
}

func (i *TrojanInboundConnSession) parseRequest(r *common.RewindReader) error {
	userHash := [56]byte{}
	n, err := r.Read(userHash[:])
	if err != nil || n != 56 {
		return common.NewError("failed to read hash").Base(err)
	}
	if !i.auth.CheckHash(string(userHash[:])) {
		return common.NewError("invalid hash:" + string(userHash[:]))
	}
	i.passwordHash = string(userHash[:])

	crlf := [2]byte{}
	r.Read(crlf[:])

	cmd, err := r.ReadByte()
	if err != nil {
		return common.NewError("failed to read cmd").Base(err)
	}

	addr, err := protocol.ParseAddress(r, "tcp")
	if err != nil {
		return common.NewError("failed to parse address").Base(err)
	}
	req := &protocol.Request{
		Command: protocol.Command(cmd),
		Address: addr,
	}
	i.request = req
	r.Read(crlf[:])
	return nil
}

func (i *TrojanInboundConnSession) SetMeter(meter stat.TrafficMeter) {
	i.meter = meter
}

func NewInboundConnSession(ctx context.Context, conn net.Conn, config *conf.GlobalConfig, auth stat.Authenticator) (protocol.ConnSession, *protocol.Request, error) {
	ctx, cancel := context.WithCancel(context.Background())

	rwc := common.NewRewindReadWriteCloser(conn)
	i := &TrojanInboundConnSession{
		config:       config,
		auth:         auth,
		passwordHash: "INVALID_HASH",
		ctx:          ctx,
		cancel:       cancel,
		rwc:          rwc,
	}
	//start buffering
	rwc.SetBufferSize(512)
	if i.config.Websocket.Enabled {
		//try to treat it as a websocket connection first
		ws, err := NewInboundWebsocket(i.ctx, conn, rwc.RewindReader, config)
		if err != nil {
			//websocket with wrong url path/origin, no need to continue parsing
			rwc.Rewind()
			rwc.StopBuffering()
			i.request = &protocol.Request{
				Address: config.RemoteAddress,
				Command: protocol.Connect,
			}
			log.Warn("remote", conn.RemoteAddr(), "is a invalid websocket conn")
			return i, i.request, nil
		}
		if ws != nil {
			//a websocket conn, try to verify it
			log.Debug("websocket conn")
			//disable the read buffer, use ws as new transport layer
			rwc.SetBufferSize(0)
			rwc = common.NewRewindReadWriteCloser(ws)
			i.rwc = rwc
			//parse it with trojan protocol format
			if err := i.parseRequest(rwc.RewindReader); err != nil {
				//not valid, just simply close it
				ws.Close()
				return nil, nil, common.NewError("invalid trojan over ws conn").Base(err)
			}
			return i, i.request, nil
		}
		//not a websocket conn, it might be a normal trojan conn
		rwc.Rewind()
	}

	//normal trojan conn
	if err := i.parseRequest(rwc.RewindReader); err != nil {
		//not a valid trojan request, proxy it to the remote_addr
		rwc.Rewind()
		rwc.StopBuffering()
		i.request = &protocol.Request{
			Address: i.config.RemoteAddress,
			Command: protocol.Connect,
		}
		log.Warn(common.NewError("invalid trojan protocol over websocket from " + conn.RemoteAddr().String()).Base(err))
		return i, i.request, nil
	}
	rwc.SetBufferSize(0)
	rwc.StopBuffering()
	return i, i.request, nil
}
