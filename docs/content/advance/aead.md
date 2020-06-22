---
title: "使用Shadowsocks AEAD进行二次加密"
draft: false
weight: 8
---

### 注意，Trojan-GFW版本不支持这个特性

Trojan协议本身无加密，其安全性依赖于下层的TLS。在一般情况下，TLS安全性很好，并不需要再次加密Trojan流量。但是，某些场景下，你可能无法保证TLS隧道的安全性：

- 你使用了Websocket，经过不可信的CDN进行中转（如国内CDN）

- 你与服务器的连接遭到了GFW针对TLS的中间人攻击

- 你的证书失效，无法验证证书有效性

- 你使用了无法保证密码学安全的可插拔传输层

等等。

Trojan-Go支持使用Shadowsocks AEAD对Trojan-Go进行加密。其本质是在Trojan协议下方加上一层Shadowsocks AEAD加密。服务端和客户端必须同时开启，且密码和加密方式必须一致，否则无法进行通讯。

要开启AEAD加密，只需添加一个```shadowsocks```选项：

```json
...
"shadowsocks": {
    "enabled": true,
    "method": "AES-128-GCM",
    "password": "1234567890"
}
```

```method```如果省略，则默认使用AES-128-GCM。更多信息，参见“完整的配置文件”一节。
