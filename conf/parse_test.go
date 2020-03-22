package conf

import (
	"testing"

	"github.com/p4gefau1t/trojan-go/common"
)

func TestParseJSON(t *testing.T) {
	data := `
	{
    "run_type": "client",
    "local_addr": "127.0.0.1",
    "local_port": 1080,
    "remote_addr": "baidu.com",
    "remote_port": 443,
    "password": [
        "password1"
    ],
    "log_level": 1,
    "ssl": {
        "verify": true,
        "verify_hostname": true,
        "cert": "server.crt",
        "cipher": "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-AES256-SHA:ECDHE-ECDSA-AES128-SHA:ECDHE-RSA-AES128-SHA:ECDHE-RSA-AES256-SHA:DHE-RSA-AES128-SHA:DHE-RSA-AES256-SHA:AES128-SHA:AES256-SHA:DES-CBC3-SHA",
        "cipher_tls13": "TLS_AES_128_GCM_SHA256:TLS_CHACHA20_POLY1305_SHA256:TLS_AES_256_GCM_SHA384",
        "sni": "",
        "alpn": [
            "h2",
            "http/1.1"
        ],
        "reuse_session": true,
        "session_ticket": false,
        "curves": ""
    },
    "tcp": {
        "no_delay": true,
        "keep_alive": true,
        "reuse_port": false,
        "fast_open": false,
        "fast_open_qlen": 20
    }
}
	`
	_, err := ParseJSON([]byte(data))
	common.Must(err)

	data = `
{
	"run_type": "server",
	"local_addr": "0.0.0.0",
	"local_port": 4445,
	"remote_addr": "127.0.0.1",
	"remote_port": 80,
	"password": [
		"pass123123"
	],
	"log_level": 2,
	"ssl": {
		"verify": false,
		"verify_hostname": false,
		"cert": "pass.crt",
		"key": "pass.key",
		"key_password": "",
		"cipher_tls13":"TLS_AES_128_GCM_SHA256:TLS_CHACHA20_POLY1305_SHA256:TLS_AES_256_GCM_SHA384",
		"prefer_server_cipher": true,
		"alpn": [
			"h2",
			"http/1.1"
		],
		"reuse_session": true,
		"session_ticket": false,
		"session_timeout": 600,
		"plain_http_response": "",
		"curves": "",
		"dhparam": ""
	},
	"tcp": {
		"no_delay": true,
		"keep_alive": true,
		"fast_open": false,
		"fast_open_qlen": 20
	},
	"mysql": {
		"enabled": true,
		"server_addr": "127.0.0.1",
		"server_port": 3306,
		"database": "trojan",
		"username": "root",
		"password": "password"
	}
}
    `
	_, err = ParseJSON([]byte(data))
	common.Must(err)
}
