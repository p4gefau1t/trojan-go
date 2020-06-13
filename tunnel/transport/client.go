package transport

import (
	"context"
	"github.com/p4gefau1t/trojan-go/common"
	"github.com/p4gefau1t/trojan-go/config"
	"github.com/p4gefau1t/trojan-go/log"
	"github.com/p4gefau1t/trojan-go/tunnel"
	"net"
	"os"
	"os/exec"
	"strconv"
)

// Client implements tunnel.Client
type Client struct {
	serverAddress *tunnel.Address
	cmd           *exec.Cmd
}

func (c *Client) Close() error {
	if c.cmd != nil {
		c.cmd.Process.Kill()
	}
	return nil
}

func (c *Client) DialPacket(tunnel.Tunnel) (tunnel.PacketConn, error) {
	panic("not supported")
}

// DialConn implements tunnel.Client. It will ignore the params and directly dial to the remote server
func (c *Client) DialConn(*tunnel.Address, tunnel.Tunnel) (tunnel.Conn, error) {
	conn, err := net.Dial("tcp", c.serverAddress.String())
	if err != nil {
		return nil, common.NewError("transport failed to connect to remote server")
	}
	return &Conn{
		Conn: conn,
	}, nil
}

// NewClient creates a transport layer client
func NewClient(ctx context.Context, c tunnel.Client) (*Client, error) {
	cfg := config.FromContext(ctx, Name).(*Config)

	var cmd *exec.Cmd
	serverAddress := tunnel.NewAddressFromHostPort("tcp", cfg.RemoteHost, cfg.RemotePort)

	if cfg.TransportPlugin.Enabled {
		log.Warn("trojan-go will use transport plugin and work in plain text mode")
		switch cfg.TransportPlugin.Type {
		case "shadowsocks":
			pluginHost := "127.0.0.1"
			pluginPort := common.PickPort("tcp", pluginHost)
			cfg.TransportPlugin.Env = append(
				cfg.TransportPlugin.Env,
				"SS_LOCAL_HOST="+pluginHost,
				"SS_LOCAL_PORT="+strconv.FormatInt(int64(pluginPort), 10),
				"SS_REMOTE_HOST="+cfg.RemoteHost,
				"SS_REMOTE_PORT="+strconv.FormatInt(int64(cfg.RemotePort), 10),
				"SS_PLUGIN_OPTIONS="+cfg.TransportPlugin.PluginOption,
			)
			cfg.RemoteHost = pluginHost
			cfg.RemotePort = pluginPort
			serverAddress = tunnel.NewAddressFromHostPort("tcp", cfg.RemoteHost, cfg.RemotePort)
			log.Debug("plugin address", serverAddress.String())
			log.Debug("plugin env", cfg.TransportPlugin.Env)

			cmd = exec.Command(cfg.TransportPlugin.Command, cfg.TransportPlugin.Arg...)
			cmd.Env = append(cmd.Env, cfg.TransportPlugin.Env...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			cmd.Start()
		case "other":
			cmd = exec.Command(cfg.TransportPlugin.Command, cfg.TransportPlugin.Arg...)
			cmd.Env = append(cmd.Env, cfg.TransportPlugin.Env...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stdout
			cmd.Start()
		case "plaintext":
			// do nothing
		default:
			return nil, common.NewError("invalid plugin type: " + cfg.TransportPlugin.Type)
		}
	}
	client := &Client{
		serverAddress: serverAddress,
		cmd:           cmd,
	}
	return client, nil
}
