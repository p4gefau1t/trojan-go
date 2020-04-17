---
title: "使用Websocket进行CDN转发"
draft: false
weight: 2
---

Trojan-Go支持使用TLS+Websocket承载Trojan协议，使得利用CDN进行流量中转成为可能。这个特性的同时考虑了将来GFW部署大规模HTTPS中间人攻击的情景。开启Websocket后，因为Trojan-Go支持使用多重TLS，即使遭受GFW的HTTPS中间人攻击，**在正确的配置下**，连接的安全性依然能得到保证。

服务器和客户端配置文件中同时添加websocket选项，并将其```enabled```字段设置为true，并填写```path```字段即可启用Websocket支持。下面是一个完整的Websocket选项

```
"websocket": {
    "enabled": true,
    "path": "/imaurlpath",
    "hostname": "www.your_awesome_domain_name.com",
    "password": "another_password",
    "double_tls": true
}
```

客户端```hostname```是可选的，填写你的域名。如果留空，将会使用```remote_addr```填充，服务端可以省略```hostname```。

```path```指的是websocket所在的URL路径，必须以斜杠("/")开始。路径并无特别要求，满足URL基本格式即可，但要保证客户端和服务端的```path```一致。```path```应当选择较长的字符串，以避免遭到GFW主动探测。

服务器开启Websocket支持后可以同时支持Websocket和一般Trojan流量，未配置Websocket选项的客户端依然可以正常使用。

由于原版Trojan并不支持Websocket，因此，虽然开启了Websocket支持的服务端可以兼容原版Trojan客户端，但是如果要使用Websocket承载流量进行CDN中转等，请确保双方都使用Trojan-Go。

因为Trojan-Go与CDN进行了TLS握手，对于CDN而言，TLS流量内容是明文。为了保证安全性，Trojan-Go默认将在Websocket连接上再建立一次TLS连接（双重TLS）。此时传输实际上经过了两次TLS握手，并且这个TLS隧道的证书校验被**强制开启**。

如果你使用了国内的CDN，建议设置```password```字段进行流量混淆，Trojan-Go将使用该密码对Websocket承载的流量再进行一次加密(AES-128-CTR)。注意这个字段的作用仅仅是**混淆**TLS的特征，防止被国内的CDN识别和封锁Trojan流量。无论是否使用二次加密，传输的安全性都可以由第二层TLS隧道保证。注意确保服务端和客户端混淆密码一致。

如果你想提高传输的性能和吞吐量，可以将```double_tls```设为false或者将```password```设为空，此时websocket将会直接承载Trojan协议。但是出于安全性考虑，还是建议至少开启混淆和双重TLS中至少一项。

CDN转发的场景和在GFW在2020年3月29日进行的HTTPS流量劫持和中间人攻击类似。它们的共同点是，第一层TLS承载的流量明文均可以被第三者窃听。**如果你使用了websocket模式**，你可以将客户端的```verify```字段填写为false，并指定```cert```字段。在这样的设置下，即使第一层TLS传输的明文遭到审查，由于第二层TLS的保护（证书校验强制开启），传输的内容依旧安全。

下面是一个客户端配置文件的例子

```
{
    "run_type": "client",
    "local_addr": "127.0.0.1",
    "local_port": 1080,
    "remote_addr": "your_server",
    "remote_port": 443,
    "password": [
        "your_password"
    ],
    "websocket": {
        "enabled": true,
        "path": "/imaurlpath",
        "hostname": "www.your_awesome_domain_name.com"
        "password": "another_password",
        "double_tls": true
    }
}

```
