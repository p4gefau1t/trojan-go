package url

import (
	"errors"
	"fmt"
	neturl "net/url"
	"strconv"
	"strings"
)

const (
	ShareInfoTypeOriginal  = "original"
	ShareInfoTypeWebSocket = "ws"
)

var validTypes = map[string]struct{}{
	ShareInfoTypeOriginal:  {},
	ShareInfoTypeWebSocket: {},
}

var validEncryptionProviders = map[string]struct{}{
	"ss": {},
}

var validSSEncryptionMap = map[string]struct{}{
	"aes-128-gcm":            {},
	"aes-256-gcm":            {},
	"chacha20-ietf-poly1305": {},
}

type ShareInfo struct {
	TrojanHost     string // 节点 IP / 域名
	Port           uint16 // 节点端口
	TrojanPassword string // Trojan 密码

	SNI  string // SNI
	Type string // 类型
	Host string // HTTP Host Header

	Path       string // WebSocket / H2 Path
	Encryption string // 额外加密
	Plugin     string // 插件设定

	Description string // 节点说明
}

func NewShareInfoFromURL(shareLink string) (info ShareInfo, e error) {
	// share link must be valid url
	parse, e := neturl.Parse(shareLink)
	if e != nil {
		e = errors.New(fmt.Sprintf("invalid url: %s", e.Error()))
		return
	}

	// share link must have `trojan-go://` scheme
	if parse.Scheme != "trojan-go" {
		e = errors.New("url does not have a trojan-go:// scheme")
		return
	}

	// password
	if info.TrojanPassword = parse.User.Username(); info.TrojanPassword == "" {
		e = errors.New("no password specified")
		return
	} else if _, hasPassword := parse.User.Password(); hasPassword {
		e = errors.New("password possibly missing percentage encoding for colon")
		return
	}

	// trojanHost: not empty & strip [] from IPv6 addresses
	if info.TrojanHost = parse.Hostname(); info.TrojanHost == "" {
		e = errors.New("host is empty")
		return
	}

	// port
	if info.Port, e = handleTrojanPort(parse.Port()); e != nil {
		return
	}

	// strictly parse the query
	query, e := neturl.ParseQuery(parse.RawQuery)
	if e != nil {
		return
	}

	// sni
	if SNIs, ok := query["sni"]; !ok {
		info.SNI = info.TrojanHost
	} else if len(SNIs) > 1 {
		e = errors.New("multiple SNIs")
		return
	} else if info.SNI = SNIs[0]; info.SNI == "" {
		e = errors.New("empty SNI")
		return
	}

	// type
	if types, ok := query["type"]; !ok {
		info.Type = ShareInfoTypeOriginal
	} else if len(types) > 1 {
		e = errors.New("multiple transport types")
		return
	} else if info.Type = types[0]; info.Type == "" {
		e = errors.New("empty transport type")
		return
	} else if _, ok := validTypes[info.Type]; !ok {
		e = errors.New(fmt.Sprintf("unknown transport type: %s", info.Type))
		return
	}

	// host
	if hosts, ok := query["host"]; !ok {
		info.Host = info.TrojanHost
	} else if len(hosts) > 1 {
		e = errors.New("multiple hosts")
		return
	} else if info.Host = hosts[0]; info.Host == "" {
		e = errors.New("empty host")
		return
	}

	// path
	if info.Type == ShareInfoTypeWebSocket {
		if paths, ok := query["path"]; !ok {
			e = errors.New("path is required in websocket")
			return
		} else if len(paths) > 1 {
			e = errors.New("multiple paths")
			return
		} else if info.Path = paths[0]; info.Path == "" {
			e = errors.New("empty path")
			return
		}

		if !strings.HasPrefix(info.Path, "/") {
			e = errors.New("path must start with /")
			return
		}
	}

	// encryption
	if encryptionArr, ok := query["encryption"]; !ok {
		// no encryption. that's okay.
	} else if len(encryptionArr) > 1 {
		e = errors.New("multiple encryption fields")
		return
	} else if info.Encryption = encryptionArr[0]; info.Encryption == "" {
		e = errors.New("empty encryption")
		return
	} else {
		encryptionParts := strings.SplitN(info.Encryption, ";", 2)
		encryptionProviderName := encryptionParts[0]

		if _, ok := validEncryptionProviders[encryptionProviderName]; !ok {
			e = errors.New(fmt.Sprintf("unsupported encryption provider name: %s", encryptionProviderName))
			return
		}

		var encryptionParams string
		if len(encryptionParts) >= 2 {
			encryptionParams = encryptionParts[1]
		}

		if encryptionProviderName == "ss" {
			ssParams := strings.SplitN(encryptionParams, ":", 2)
			if len(ssParams) < 2 {
				e = errors.New("missing ss password")
				return
			}

			ssMethod, ssPassword := ssParams[0], ssParams[1]
			if _, ok := validSSEncryptionMap[ssMethod]; !ok {
				e = errors.New(fmt.Sprintf("unsupported ss method: %s", ssMethod))
				return
			}

			if ssPassword == "" {
				e = errors.New("ss password cannot be empty")
				return
			}
		} else {
			e = errors.New(fmt.Sprintf("encryption param %s is not supported", info.Encryption))
			return
		}
	}

	// plugin
	if plugins, ok := query["plugin"]; !ok {
		// no plugin. that's okay.
	} else if len(plugins) > 1 {
		e = errors.New("multiple plugins")
		return
	} else if info.Plugin = plugins[0]; info.Plugin == "" {
		e = errors.New("empty plugin")
		return
	}

	// description
	info.Description = parse.Fragment

	return
}

func handleTrojanPort(p string) (port uint16, e error) {
	if p == "" {
		return 443, nil
	}

	portParsed, e := strconv.Atoi(p)
	if e != nil {
		return
	}

	if portParsed < 1 || portParsed > 65535 {
		e = errors.New(fmt.Sprintf("invalid port %d", portParsed))
		return
	}

	port = uint16(portParsed)
	return
}
