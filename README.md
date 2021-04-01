# Trojan-Go [![Go Report Card](https://goreportcard.com/badge/github.com/p4gefau1t/trojan-go)](https://goreportcard.com/report/github.com/p4gefau1t/trojan-go) [![Downloads](https://img.shields.io/github/downloads/p4gefau1t/trojan-go/total?label=downloads&logo=github&style=flat-square)](https://img.shields.io/github/downloads/p4gefau1t/trojan-go/total?label=downloads&logo=github&style=flat-square)

使用 Go 实现的完整 Trojan 代理，兼容原版 Trojan 协议及配置文件格式。安全、高效、轻巧、易用。

Trojan-Go 支持[多路复用](#多路复用)提升并发性能；使用[路由模块](#路由模块)实现国内外分流；支持 [CDN 流量中转](#Websocket)(基于 WebSocket over TLS)；支持使用 AEAD 对 Trojan 流量进行[二次加密](#aead-加密)(基于 Shadowsocks AEAD)；支持可插拔的[传输层插件](#传输层插件)，允许替换 TLS，使用其他加密隧道传输 Trojan 协议流量。

预编译二进制可执行文件可在 [Release 页面](https://github.com/p4gefau1t/trojan-go/releases)下载。解压后即可直接运行，无其他组件依赖。

如遇到配置和使用问题、发现 bug，或是有更好的想法，欢迎加入 [Telegram 交流反馈群](https://t.me/trojan_go_chat)。

## 简介

**完整介绍和配置教程，参见 [Trojan-Go 文档](https://p4gefau1t.github.io/trojan-go)。**

Trojan-Go 兼容原版 Trojan 的绝大多数功能，包括但不限于：

- TLS 隧道传输
- UDP 代理
- 透明代理 (NAT 模式，iptables 设置参考[这里](https://github.com/shadowsocks/shadowsocks-libev/tree/v3.3.1#transparent-proxy))
- 对抗 GFW 被动检测 / 主动检测的机制
- MySQL 数据持久化方案
- MySQL 用户权限认证
- 用户流量统计和配额限制

同时，Trojan-Go 还扩展实现了更多高效易用的功能特性：

- 便于快速部署的「简易模式」
- Socks5 / HTTP 代理自动适配
- 基于 TProxy 的透明代理（TCP / UDP）
- 全平台支持，无特殊依赖
- 基于多路复用（smux）降低延迟，提升并发性能
- 自定义路由模块，可实现国内外分流 / 广告屏蔽等功能
- Websocket 传输支持，以实现 CDN 流量中转（基于 WebSocket over TLS）和对抗 GFW 中间人攻击
- TLS 指纹伪造，以对抗 GFW 针对 TLS Client Hello 的特征识别
- 基于 gRPC 的 API 支持，以实现用户管理和速度限制等
- 可插拔传输层，可将 TLS 替换为其他协议或明文传输，同时有完整的 Shadowsocks 混淆插件支持
- 支持对用户更友好的 YAML 配置文件格式

## 图形界面客户端

Trojan-Go 服务端兼容所有原 Trojan 客户端，如 Igniter、ShadowRocket 等。以下是支持 Trojan-Go 扩展特性（Websocket / Mux 等）的客户端：

- [Trojan-Qt5](https://github.com/Trojan-Qt5/Trojan-Qt5/)：跨平台客户端，支持 Windows / macOS / Linux，使用 Trojan-Go 核心，支持所有 Trojan-Go 扩展特性。
- [Qv2ray](https://github.com/Qv2ray/Qv2ray)：跨平台客户端，支持 Windows / macOS / Linux，使用 Trojan-Go 核心，支持所有 Trojan-Go 扩展特性。
- [Igniter-Go](https://github.com/p4gefau1t/trojan-go-android)：Android 客户端，Fork 自 Igniter，将 Igniter 核心替换为 Trojan-Go 并做了一定修改，支持所有 Trojan-Go 扩展特性。

## 使用方法

1. 快速启动服务端和客户端（简易模式）

    - 服务端

        ```shell
        sudo ./trojan-go -server -remote 127.0.0.1:80 -local 0.0.0.0:443 -key ./your_key.key -cert ./your_cert.crt -password your_password
        ```

    - 客户端

        ```shell
        ./trojan-go -client -remote example.com:443 -local 127.0.0.1:1080 -password your_password
        ```

2. 使用配置文件启动客户端 / 服务端 / 透明代理 / 中继（一般模式）

    ```shell
    ./trojan-go -config config.json
    ```

3. 使用 URL 启动客户端（格式参见文档）

    ```shell
    ./trojan-go -url 'trojan-go://password@cloudflare.com/?type=ws&path=%2Fpath&host=your-site.com'
    ```

4. 使用 Docker 部署

    ```shell
    docker run \
        --name trojan-go \
        -d \
        -v /etc/trojan-go/:/etc/trojan-go \
        --network host \
        p4gefau1t/trojan-go
    ```

   或者

    ```shell
    docker run \
        --name trojan-go \
        -d \
        -v /path/to/host/config:/path/in/container \
        --network host \
        p4gefau1t/trojan-go \
        /path/in/container/config.json
    ```

## 特性

一般情况下，Trojan-Go 和 Trojan 是互相兼容的，但一旦使用下面介绍的扩展特性（如多路复用、Websocket 等），则无法兼容。

### 移植性

编译得到的 Trojan-Go 单个可执行文件不依赖其他组件。同时，你可以很方便地编译（或交叉编译） Trojan-Go，然后在你的服务器、PC、树莓派，甚至路由器上部署；可以方便地使用 build tag 删减模块，以缩小可执行文件体积。

例如，交叉编译一个可在 mips 处理器、Linux 操作系统上运行的、只有客户端功能的 Trojan-Go，只需执行下面的命令，得到的可执行文件可以直接在目标平台运行：

```shell
CGO_ENABLED=0 GOOS=linux GOARCH=mips go build -tags "client" -trimpath -ldflags "-s -w -buildid="
```

完整的 tag 说明参见 [Trojan-Go 文档](https://p4gefau1t.github.io/trojan-go)。

### 易用

配置文件格式与原版 Trojan 兼容，但做了大幅简化，未指定的字段会被赋予默认值，由此可以更方便地部署服务端和客户端。以下是一个简单例子，完整的配置文件可以参见[这里](https://p4gefau1t.github.io/trojan-go)。

服务端配置文件 `server.json`：

```json
{
  "run_type": "server",
  "local_addr": "0.0.0.0",
  "local_port": 443,
  "remote_addr": "127.0.0.1",
  "remote_port": 80,
  "password": ["your_awesome_password"],
  "ssl": {
    "cert": "your_cert.crt",
    "key": "your_key.key",
    "sni": "www.your-awesome-domain-name.com"
  }
}
```

客户端配置文件 `client.json`：

```json
{
  "run_type": "client",
  "local_addr": "127.0.0.1",
  "local_port": 1080,
  "remote_addr": "www.your-awesome-domain-name.com",
  "remote_port": 443,
  "password": ["your_awesome_password"]
}
```

可以使用更简明易读的 YAML 语法进行配置。以下是一个客户端的例子，与上面的 `client.json` 等价：

客户端配置文件 `client.yaml`：

```yaml
run-type: client
local-addr: 127.0.0.1
local-port: 1080
remote-addr: www.your-awesome-domain_name.com
remote-port: 443
password:
  - your_awesome_password
```

### WebSocket

Trojan-Go 支持使用 TLS + Websocket 承载 Trojan 协议，使得利用 CDN 进行流量中转成为可能。

服务端和客户端配置文件中同时添加 `websocket` 选项即可启用 Websocket 支持，例如

```json
"websocket": {
    "enabled": true,
    "path": "/your-websocket-path",
    "hostname": "www.your-awesome-domain-name.com"
}
```

完整的选项说明参见 [Trojan-Go 文档](https://p4gefau1t.github.io/trojan-go)。

可以省略 `hostname`, 但服务端和客户端的 `path` 必须一致。服务端开启 Websocket 支持后，可以同时支持 Websocket 和一般 Trojan 流量。未配置 Websocket 选项的客户端依然可以正常使用。

由于 Trojan 并不支持 Websocket，因此，虽然开启了 Websocket 支持的 Trojan-Go 服务端可以兼容所有客户端，但如果要使用 Websocket 承载流量，请确保双方都使用 Trojan-Go。

### 多路复用

在很差的网络条件下，一次 TLS 握手可能会花费很多时间。Trojan-Go 支持多路复用（基于 [smux](https://github.com/xtaci/smux)），通过一条 TLS 隧道连接承载多条 TCP 连接的方式，减少 TCP 和 TLS 握手带来的延迟，以期提升高并发情景下的性能。

> 启用多路复用并不能提高测速得到的链路速度，但能降低延迟、提升大量并发请求时的网络体验，例如浏览含有大量图片的网页等。

你可以通过设置客户端的 `mux` 选项 `enabled` 字段启用它：

```json
"mux": {
    "enabled": true
}
```

只需开启客户端 mux 配置即可，服务端会自动检测是否启用多路复用并提供支持。完整的选项说明参见 [Trojan-Go 文档](https://p4gefau1t.github.io/trojan-go)。

### 路由模块

Trojan-Go 客户端内建一个简单实用的路由模块，以方便实现国内直连、海外代理等自定义路由功能。

路由策略有三种：

- `Proxy` 代理：将请求通过 TLS 隧道进行代理，由 Trojan 服务端与目的地址进行连接。
- `Bypass` 绕过：直接使用本地设备与目的地址进行连接。
- `Block` 封锁：不发送请求，直接关闭连接。

要激活路由模块，请在配置文件中添加 `router` 选项，并设置 `enabled` 字段为 `true`：

```json
"router": {
    "enabled": true,
    "bypass": [
        "geoip:cn",
        "geoip:private",
        "full:localhost"
    ],
    "block": [
        "cidr:192.168.1.1/24",
    ],
    "proxy": [
        "domain:google.com",
    ],
    "default_policy": "proxy"
}
```

完整的选项说明参见 [Trojan-Go 文档](https://p4gefau1t.github.io/trojan-go)。

### AEAD 加密

Trojan-Go 支持基于 Shadowsocks AEAD 对 Trojan 协议流量进行二次加密，以保证 Websocket 传输流量无法被不可信的 CDN 识别和审查：

```json
"shadowsocks": {
    "enabled": true,
    "password": "my-password"
}
```

如需开启，服务端和客户端必须同时开启并保证密码一致。

### 传输层插件

Trojan-Go 支持可插拔的传输层插件，并支持 Shadowsocks [SIP003](https://shadowsocks.org/en/wiki/Plugin.html) 标准的混淆插件。下面是使用 `v2ray-plugin` 的一个例子：

> **此配置并不安全，仅作为演示**

服务端配置：

```json
"transport_plugin": {
    "enabled": true,
    "type": "shadowsocks",
    "command": "./v2ray-plugin",
    "arg": ["-server", "-host", "www.baidu.com"]
}
```

客户端配置：

```json
"transport_plugin": {
    "enabled": true,
    "type": "shadowsocks",
    "command": "./v2ray-plugin",
    "arg": ["-host", "www.baidu.com"]
}
```

完整的选项说明参见 [Trojan-Go 文档](https://p4gefau1t.github.io/trojan-go)。

## 构建

> 请确保 Go 版本 >= 1.14

使用 `make` 进行编译：

```shell
git clone https://github.com/p4gefau1t/trojan-go.git
cd trojan-go
make
make install #安装systemd服务等，可选
```

或者使用 Go 自行编译：

```shell
go build -tags "full"
```

Go 支持通过设置环境变量进行交叉编译，例如：

编译适用于 64 位 Windows 操作系统的可执行文件：

```shell
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags "full"
```

编译适用于 Apple Silicon 的可执行文件：

```shell
CGO_ENABLED=0 GOOS=macos GOARCH=arm64 go build -tags "full"
```

编译适用于 64 位 Linux 操作系统的可执行文件：

```shell
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags "full"
```

## 致谢

- [Trojan](https://github.com/trojan-gfw/trojan)
- [V2Fly](https://github.com/v2fly)
- [utls](https://github.com/refraction-networking/utls)
- [smux](https://github.com/xtaci/smux)
- [go-tproxy](https://github.com/LiamHaworth/go-tproxy)

## Stargazers over time

[![Stargazers over time](https://starchart.cc/p4gefau1t/trojan-go.svg)](https://starchart.cc/p4gefau1t/trojan-go)
