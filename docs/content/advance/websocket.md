---
title: "使用Websocket进行CDN转发和抵抗中间人攻击"
draft: false
weight: 2
---

### 注意，Trojan不支持这个特性

Trojan-Go支持使用TLS+Websocket承载Trojan协议，使得利用CDN进行流量中转成为可能。

Trojan协议本身不带加密，安全性依赖外层的TLS。但流量一旦经过CDN，TLS对CDN是透明的。其服务提供者可以对TLS的明文内容进行审查。**如果你使用的是不可信任的CDN（任何在中国大陆注册备案的CDN服务均应被视为不可信任），请务必开启Shadowsocks AEAD对Webosocket流量进行加密，以避免遭到识别和审查。**

服务器和客户端配置文件中同时添加websocket选项，并将其```enabled```字段设置为true，并填写```path```字段和```hostname```字段即可启用Websocket支持。下面是一个完整的Websocket选项:

```json
"websocket": {
    "enabled": true,
    "path": "/imaurlpath",
    "hostname": "www.your_awesome_domain_name.com"
}
```

```hostname```是主机名，一般填写域名。客户端```hostname```是可选的，填写你的域名。如果留空，将会使用```remote_addr```填充，服务端必须填写```hostname```，Trojan-Go可以以此转发Websocket请求，以抵抗针对Websocket的主动检测。

```path```指的是websocket所在的URL路径，必须以斜杠("/")开始。路径并无特别要求，满足URL基本格式即可，但要保证客户端和服务端的```path```一致。```path```应当选择较长的字符串，以避免遭到GFW直接的主动探测。

客户端的```hostname```将发送给CDN服务器，必须有效；服务端和客户端```path```必须一致，否则Websocket握手无法进行。

下面是一个客户端配置文件的例子

```json
{
    "run_type": "client",
    "local_addr": "127.0.0.1",
    "local_port": 1080,
    "remote_addr": "www.your_awesome_domain_name.com",
    "remote_port": 443,
    "password": [
        "your_password"
    ],
    "websocket": {
        "enabled": true,
        "path": "/your-websocket-path",
        "hostname": "www.your_awesome_domain_name.com"
    },
    "shadowsocks": {
        "enabled": true,
        "password": "12345678"
    }
}
```
