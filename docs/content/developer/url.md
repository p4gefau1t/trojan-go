---
title: "URL方案（草案）"
draft: false
weight: 200
---

## Changelog

- encrpytion 格式修改为 ss;method:password

## 概述

感谢 @DuckSoft @StudentMain @phlinhng 对 Trojan-Go URL 方案的讨论和贡献。**目前 URL 方案为草案，需要更多的实践和讨论。**

Trojan-Go**客户端**可以接受URL，以定位服务器资源。原则如下:

- 遵守 URL 格式规范

- 保证人类可读，对机器友好

- URL 的作用，是定位 Trojan-Go 节点资源，方便资源分享

需要注意，基于人类可读性的考虑，禁止将 base64 等编码数据嵌入 URL 中。首先， base64 编码不能保证传输安全，其意义在于在 ASCII 信道传输非 ASCII 数据。其次，如果需要保证分享 URL 时的传输安全，请对明文 URL 进行加密，而不是修改 URL 格式。

## 格式

基本格式如下，`$()` 代表此处需要 `encodeURIComponent`。

```text
trojan-go://
    $(trojan-password)
    @
    trojan-host
    :
    port
/?
    sni=$(tls-sni.com)&
    type=$(original|ws|h2|h2+ws)&
        host=$(websocket-host.com)&
        path=$(/websocket/path)&
    encryption=$(ss;aes-256-gcm;ss-password)&
    plugin=$(...)
#$(descriptive-text)
```

例如

```text
trojan-go://password1234@google.com/?sni=microsoft.com&type=ws&host=youtube.com&path=%2Fgo&encryption=ss%3Baes-256-gcm%3Afuckgfw
```

由于 Trojan-Go 兼容 Trojan，所以对于 Trojan 的 URL 方案

```text
trojan://password@remote_host:remote_port
```

可以兼容接受。它等价于

```text
trojan-go://password@remote_host:remote_port
```

需要注意的是，一旦服务器使用了非Trojan兼容的功能，必须使用```trojan-go://```定位服务器。这样设计的目的是使得 Trojan-Go 的 URL 不会被 Trojan 错误接受，避免污染 Trojan 用户的 URL 分享。同时，Trojan-Go 确保可以兼容接受 Trojan 的 URL。

## 详述

注意：所有参数名和常数字符串均区分大小写。

### `trojan-password`

Trojan 的密码。
不可省略，不能为空字符串，不建议含有非 ASCII 可打印字符。
必须使用 `encodeURIComponent` 编码。

### `trojan-host`

节点 IP / 域名。
不可省略，不能为空字符串。
IPv6 地址必须扩方括号。
IDN 域名（如“百度.cn”）必须使用 `xn--xxxxxx` 格式。

### `port`

节点端口。
省略时默认为 `443`。
必须取 `[1,65535]` 中的整数。

### `tls`或`allowInsecure`

没有这个字段。
TLS 默认一直启用，除非有传输插件禁用它。
TLS 认证必须开启。无法使用根CA校验服务器身份的节点，不适合分享。

### `sni`

自定义 TLS 的 SNI。
省略时默认与 `trojan-host` 同值。不得为空字符串。

必须使用 `encodeURIComponent` 编码。

### `type`

传输类型。
省略时默认为 `original`，但不可为空字符串。
目前可选值只有 `original` 和 `ws`，未来可能会有 `h2`、`h2+ws` 等取值。

当取值为 `original` 时，使用原始 Trojan 传输方式，无法方便通过 CDN。
当取值为 `ws` 时，使用 Websocket over TLS 传输。

### `host`

自定义 HTTP `Host` 头。
可以省略，省略时值同 `trojan-host`。
可以为空字符串，但可能带来非预期情形。

警告：若你的端口非标准端口（不是 80 / 443），RFC 标准规定 `Host` 应在主机名后附上端口号，例如 `example.com:44333`。至于是否遵守，请自行斟酌。

必须使用 `encodeURIComponent` 编码。

### `path`

当传输类型 `type` 取 `ws`、`h2`、`h2+ws` 时，此项有效。
不可省略，不可为空。
必须以 `/` 开头。
可以使用 URL 中的 `&` `#` `?` 等字符，但应当是合法的 URL 路径。

必须使用 `encodeURIComponent` 编码。

### `mux`

没有这个字段。
当前服务器默认一直支持 `mux`。
启用 `mux` 与否各有利弊，应由客户端决定自己是否启用。URL的作用，是定位服务器资源，而不是规定用户使用偏好。

### `encryption`

用于保证 Trojan 流量密码学安全的加密层。
可省略，默认为 `none`，即不使用加密。
不可以为空字符串。

必须使用 encodeURIComponent 编码。

使用 Shadowsocks 算法进行流量加密时，其格式为：

```text
ss;method:password
```

其中 ss 是固定内容，method 是加密方法，必须为下列之一：

- `aes-128-gcm`
- `aes-256-gcm`
- `chacha20-ietf-poly1305`

其中的 `password` 是 Shadowsocks 的密码，不得为空字符串。
`password` 中若包含分号，不需要进行转义。
`password` 应为英文可打印 ASCII 字符。

其他加密方案待定。

### `plugin`

额外的插件选项。本字段保留。
可省略，但不可以为空字符串。

### URL Fragment (# 后内容)

节点说明。
不建议省略，不建议为空字符串。

必须使用 `encodeURIComponent` 编码。
