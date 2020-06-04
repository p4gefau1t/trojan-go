# Trojan-Go

[![License](https://img.shields.io/github/license/p4gefau1t/trojan-go)](https://img.shields.io/github/license/p4gefau1t/trojan-go)
[![Downloads](https://img.shields.io/github/downloads/p4gefau1t/trojan-go/total?label=downloads&logo=github&style=flat-square)](https://img.shields.io/github/downloads/p4gefau1t/trojan-go/total?label=downloads&logo=github&style=flat-square)
[![HitCounts](http://hits.dwyl.io/p4gefau1t/trojan-go.svg)](http://hits.dwyl.io/p4gefau1t/trojan-go)
[![Release](https://img.shields.io/github/v/release/p4gefau1t/trojan-go?include_prereleases)](https://img.shields.io/github/v/release/p4gefau1t/trojan-go?include_prereleases)
[![Release Date](https://img.shields.io/github/release-date-pre/p4gefau1t/trojan-go)](https://img.shields.io/github/release-date-pre/p4gefau1t/trojan-go)
[![Docker Image](https://images.microbadger.com/badges/image/p4gefau1t/trojan-go.svg)](https://hub.docker.com/r/p4gefau1t/trojan-go)

[![Commit](https://img.shields.io/github/last-commit/p4gefau1t/trojan-go)](https://img.shields.io/github/last-commit/p4gefau1t/trojan-go)
[![Commit Activity](https://img.shields.io/github/commit-activity/m/p4gefau1t/trojan-go)](https://img.shields.io/github/commit-activity/m/p4gefau1t/trojan-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/p4gefau1t/trojan-go)](https://goreportcard.com/report/github.com/p4gefau1t/trojan-go)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg?style=flat-square)](https://github.com/p4gefau1t/trojan-go/pulls)

使用Go实现的完整Trojan代理，与Trojan协议以及Trojan-GFW版本的配置文件格式兼容。安全，高效，轻巧，易用。

支持使用[多路复用](#多路复用)提升并发性能，使用[路由模块](#路由模块)实现国内直连。

支持[CDN流量中转](#Websocket)(基于WebSocket over TLS/SSL)。

支持基于ACME协议从Let's Encrypt[自动申请和更新](#证书申请)HTTPS证书，只需提供域名和邮箱。

预编译的版本可在 [Release 页面](https://github.com/p4gefau1t/trojan-go/releases)下载。直接运行解压得到的执行文件即可，无其他组件依赖。

如果你遇到配置和使用方面的问题，发现了软件Bug，或是有更好的想法，欢迎加入Trojan-Go的[Telegram交流反馈群](https://t.me/trojan_go_chat)。

## **下面的说明为简单介绍，完整配置教程和配置介绍参见[Trojan-Go文档](https://p4gefau1t.github.io/trojan-go)。**

Trojan-Go支持并且兼容Trojan-GFW的绝大多数功能，包括但不限于：

- TLS/SSL隧道传输

- 透明代理 (NAT模式，iptables设置参见[这里](https://github.com/shadowsocks/shadowsocks-libev/tree/v3.3.1#transparent-proxy))

- UDP代理

- 对抗GFW被动/主动检测的机制

- MySQL数据库支持

- 流量统计，用户流量配额限制

- 从数据库中的用户列表进行认证

- TCP Keep Alive，TCP Fast Open，端口复用等TCP选项

同时，Trojan-Go还扩展了更多高效易用的功能特性：

- 简易模式，快速部署使用

- Socks5/HTTP代理自动适配

- 多平台和多操作系统支持，无特殊依赖

- 多路复用，降低延迟，提升并发性能

- 自定义路由模块，可实现国内直连/广告屏蔽等功能

- Websocket传输支持，用于实现CDN流量中转(基于WebSocket over TLS/SSL)和对抗GFW中间人攻击

- 自动化HTTPS证书申请，从Let's Encrypt自动申请和更新HTTPS证书

- TLS指纹伪造，绕过GFW针对TLS Client Hello的特征识别

- 基于gRPC的API支持，支持动态用户管理和流量速度限制

- 可插拔传输层，可将TLS替换为其他协议或明文传输。同时有完整的Shadowsocks混淆插件支持。

## 图形界面客户端

Trojan-Go服务端可以兼容所有Trojan-GFW的客户端，如Igniter，ShadowRocket等。

下面是支持Trojan-Go扩展特性（Websocket/Mux等）的客户端。

[Trojan-Qt5](https://github.com/Trojan-Qt5/Trojan-Qt5/)，跨平台客户端，支持Windows/MacOS/Linux，使用Trojan-Go核心，支持所有Trojan-Go扩展特性。

[Igniter-Go](https://github.com/p4gefau1t/trojan-go-android)，安卓客户端，Fork自Igniter，将Igniter核心替换为Trojan-Go并做了一定修改，支持所有Trojan-Go扩展特性。

## 使用方法

1. 快速证书配置

    - 自动申请证书

        ```shell
        sudo ./trojan-go -autocert request
        ```

        (**注意备份生成的证书和密钥，并确保其安全**)

    - 为证书续期

        ```shell
        sudo ./trojan-go -autocert renew
        ```

    关于证书申请[更详细的说明](#证书申请)。

2. 快速启动服务器和客户端（简易模式）

    - 服务端

        ```shell
        sudo ./trojan-go -server -remote 127.0.0.1:80 -local 0.0.0.0:443 -key ./your_key.key -cert ./your_cert.crt -password your_password
        ```

    - 客户端

        ```shell
        ./trojan-go -client -remote example.com:443 -local 127.0.0.1:1080 -password your_password
        ```

3. 使用配置文件启动客户端/服务端/透明代理/中继（一般模式）

    ```shell
    ./trojan-go -config config.json
    ```

4. 使用Docker部署

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
    
    镜像的latest标签对应master分支，nightly标签对应dev分支。

## 特性

### 移植性

运行Trojan-Go的可执行文件不依赖其他组件。

可以将编译得到的单个可执行文件在目标机器上直接执行而不需要考虑依赖的问题。同时，你可以很方便地编译（或者交叉编译）它，然后在你的服务器，PC，树莓派，甚至路由器上部署。可以方便地使用build tag删减各模块，以缩小可执行文件体积。

例如，交叉编译一个在mips处理器，linux操作系统上运行的，只有客户端功能的Trojan-Go，只需执行下面的命令即可

```shell
CGO_ENABLE=0 GOOS=linux GOARCH=mips go build -tags "client" -ldflags "-s -w"
```

得到的可执行文件可以直接在目标平台运行。

完整的tag说明参见[Trojan-Go 文档](https://p4gefau1t.github.io/trojan-go)。

### 易用

配置文件格式与Trojan-GFW兼容，但做了大幅简化，未指定的字段会被赋给一个默认值。你可以更方便地部署你的服务器和客户端。下面是一个简单的例子，完整的配置文件可以参见[这里](https://p4gefau1t.github.io/trojan-go)。

服务器配置文件

server.json

```json
{
    "run_type": "server",
    "local_addr": "0.0.0.0",
    "local_port": 443,
    "remote_addr": "127.0.0.1",
    "remote_port": 80,
    "password": [
        "your_awesome_password"
    ],
    "ssl": {
        "cert": "your_cert.crt",
        "key": "your_key.key"
    }
}
```

客户端配置文件

client.json

```json
{
    "run_type": "client",
    "local_addr": "127.0.0.1",
    "local_port": 1080,
    "remote_addr": "www.your_awesome_domain_name.com",
    "remote_port": 443,
    "password": [
        "your_awesome_password"
    ]
}
```

<a name="证书申请"></a>

### 自动证书申请

使用

```shell
sudo ./trojan-go -autocert request
```

向Let's Encrypt申请证书。

申请过程中，按照ACME协议要求，trojan-go需要和letsencrypt服务器交互，因此需要暂时占用本地443和80端口，此时请暂时关闭nginx，apache，或者trojan等服务。

Linux下，绑定80和443端口需要root权限，因此你需要使用sudo执行trojan-go才能正常证书申请流程。

你也可以指定自定义端口，然后使用nginx等web服务器进行443和80分流，将ACME协议流量代理到自定义端口上。

如果申请成功，本目录下会得到

- server.key 服务器私钥

- server.crt 经过Let's Encrypt签名的服务器证书

- user.key 用户Email对应的私钥

- domain_info.json 域名和用户Email信息

请备份这几个文件并且妥善保管。接下来你可以将服务器私钥和证书文件名填入你的配置文件，开启你的trojan-go服务器即可。

如果证书过期了，使用

```shell
sudo ./trojan-go -autocert renew
```

更新证书，确保上面提到的四个文件在trojan-go所在目录，运行后trojan-go将自动更新证书文件。

### WebSocket

<a name="WebSocket"></a>

Trojan-Go支持使用TLS+Websocket承载Trojan协议，使得利用CDN进行流量中转成为可能。

服务器和客户端配置文件中同时添加```websocket```选项即可启用Websocket支持，例如

```json
"websocket": {
    "enabled": true,
    "path": "/im_a_url_path",
    "hostname": "www.your_awesome_domain_name.com"
}
```

完整的选项说明参见[Trojan-Go 文档](https://p4gefau1t.github.io/trojan-go)。

可以省略```hostname```, 但是服务器和客户端的```path```必须一致。服务器开启Websocket支持后可以同时支持Websocket和一般Trojan流量，未配置Websocket选项的客户端依然可以正常使用。

由于Trojan-GFW版本并不支持Websocket，因此，虽然开启了Websocket支持的Trojan-Go服务端可以兼容所有客户端，但是如果要使用Websocket承载流量，请确保双方都使用Trojan-Go。

### 多路复用

<a name="多路复用"></a>

在很差的网络条件下，一次TLS握手可能会花费很多时间。

Trojan-Go支持多路复用(基于[smux](https://github.com/xtaci/smux))。通过一个TLS隧道连接承载多个TCP连接的方式，减少TCP和TLS握手带来的延迟，以期提升高并发情景下的性能。

启用多路复用并不会增加你测速得到的链路速度，但会降低延迟，提升大量并发请求时的网络体验，例如浏览含有大量图片的网页等。

注意，这个特性和Trojan-GFW**不兼容**，出于兼容性考虑，这个特性是默认关闭的。你可以通过设置客户端的mux选项"enabled"字段启用它。如下

```json
"mux": {
    "enabled": true
}
```

完整的选项说明参见[Trojan-Go 文档](https://p4gefau1t.github.io/trojan-go)。

你只需要设置客户端的配置文件即可，服务端会自动检测是否启用多路复用并提供支持。

### 路由模块

<a name="路由模块"></a>

Trojan-Go的客户端内建一个简单实用的路由模块用以方便实现国内直连等自定义路由功能。

路由策略有三种

- Proxy 代理。将请求通过TLS隧道进行代理，由trojan服务器和目的地址进行连接。

- Bypass 绕过。直接在本地和目的地址进行连接。

- Block 封锁。不代理请求，直接关闭连接。

要激活模块，在你的配置文件中添加router选项，并且设置enabled为true，例如

```json
"router": {
    "enabled": true,
    "bypass": [
        "geoip:tag1",
        "geosite:tag2",
        "bypass_list1.txt",
        "bypass_list2.txt"
    ],
    "block": [
        "block_list.txt"
    ],
    "proxy": [
        "proxy_list.txt"
    ],
    "default_policy": "proxy"
}
```

其中```bypass```,```block```, ```proxy```字段中填入相应的列表文件或者geo数据库tag。列表文件每行是一个域名或者IP地址段(CIDR)。geo数据库geoip和geosite为IP数据库和域名数据库。一旦匹配，则执行相应策略。

完整的选项说明参见[Trojan-Go 文档](https://p4gefau1t.github.io/trojan-go)。

下面是一个实现国内直连的选项，它将绕过中国大陆IP地址，中国大陆域名，以及内网IP等保留的私有IP地址，直接连接远端而不通过隧道代理。

```json
"router": {
    "enabled": true,
    "bypass": [
        "geoip:cn",
        "geoip:private",
        "geosite:cn"
    ]
}
```

所需要的geoip.dat和geosite.dat已经包含在release的压缩包中。它们来自v2ray的[domain-list-community](https://github.com/v2ray/domain-list-community)和[geoip](https://github.com/v2ray/geoip)。

完整的选项说明参见[Trojan-Go 文档](https://p4gefau1t.github.io/trojan-go)。

## 混淆插件

Trojan-Go支持可插拔的传输层插件，并支持Shadowsocks SIP003标准的混淆插件。下面是使用v2ray-plugin的一个例子：

**此配置并不安全，仅用作演示**

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

完整的选项说明参见[Trojan-Go 文档](https://p4gefau1t.github.io/trojan-go)。

## 构建

确保你的Go版本 >= 1.14，推荐使用snap安装Go保持与上游同步。

```shell
git clone https://github.com/p4gefau1t/trojan-go.git
cd trojan-go
go build -tags "full"
```

Go支持通过设置环境变量进行交叉编译，例如

```shell
CGO_ENABLE=0 GOOS=windows GOARCH=amd64 go build -tags "full"
```

以及

```shell
CGO_ENABLE=0 GOOS=linux GOARCH=arm go build -tags "full"
```

## 致谢

[trojan](https://github.com/trojan-gfw/trojan)

[v2ray](https://github.com/v2ray/)

[smux](https://github.com/xtaci/smux)

[lego](https://github.com/go-acme/lego)

[go-tproxy](https://github.com/LiamHaworth/go-tproxy)

[utls](https://github.com/refraction-networking/utls)

## Stargazers over time

[![Stargazers over time](https://starchart.cc/p4gefau1t/trojan-go.svg)](https://starchart.cc/p4gefau1t/trojan-go)
