# Trojan-Go [![Go Report Card](https://goreportcard.com/badge/github.com/p4gefau1t/trojan-go)](https://goreportcard.com/report/github.com/p4gefau1t/trojan-go) [![Downloads](https://img.shields.io/github/downloads/p4gefau1t/trojan-go/total?label=downloads&logo=github&style=flat-square)](https://img.shields.io/github/downloads/p4gefau1t/trojan-go/total?label=downloads&logo=github&style=flat-square)

基于原仓库的二次开发，增加了一个完全匹配hysteria2的http验证和流量统计api

```json
{
	"http": {
		"enabled": true,
		"api_url": "POST请求地址",
		"timeout": "超时时间"
	},
	"api": {
		"enabled": true,
		"api_key": "同HY2",
		"api_addr": "监听IP",
		"api_port": "监听端口"
	}
}


```
