---
title: "使用Websocket进行CDN转发和抵抗中间人攻击"
draft: false
weight: 2
---

### 注意，Trojan-GFW版本不支持这个特性

Trojan-Go支持使用TLS+Websocket承载Trojan协议，使得利用CDN进行流量中转成为可能。这个特性的设计考虑了将来GFW部署大规模HTTPS中间人攻击的情景。开启Websocket后，因为Trojan-Go支持使用多重TLS，即使遭受GFW的HTTPS中间人攻击，**在正确的配置下**，连接的安全性依然能得到保证。

服务器和客户端配置文件中同时添加websocket选项，并将其```enabled```字段设置为true，并填写```path```字段和```hostname```字段即可启用Websocket支持。下面是一个完整的Websocket选项

```json
"websocket": {
    "enabled": true,
    "path": "/imaurlpath",
    "hostname": "www.your_awesome_domain_name.com",
    "obfuscation_password": "another_password",
    "double_tls": true,
    "ssl": {
      "verify": true,
      "verify_hostname": true,
      "cert": "",
      "key": "",
      "key_password": "",
      "prefer_server_cipher": false,
      "sni": "",
      "session_ticket": true,
      "reuse_session": true,
      "plain_http_response": "",
    }
}
```

```hostname```是主机名，一般填写域名。客户端```hostname```是可选的，填写你的域名。如果留空，将会使用```remote_addr```填充，服务端必须填写```hostname```，Trojan-Go可以以此转发Websocket请求，以抵抗针对Websocket的主动检测。

```path```指的是websocket所在的URL路径，必须以斜杠("/")开始。路径并无特别要求，满足URL基本格式即可，但要保证客户端和服务端的```path```一致。```path```应当选择较长的字符串，以避免遭到GFW直接的主动探测。

```double_tls```表示是否开启双重TLS，如果省略，默认设置为true。因为Trojan-Go与CDN进行了TLS握手，对于CDN而言，TLS流量内容是明文。为了保证安全性，Trojan-Go默认将在Websocket连接上再建立一次TLS连接（双重TLS）。此时传输实际上经过了两次TLS握手。

```ssl```第二层TLS的配置。如果未填写，使用全局的```ssl```选项填充

```obfuscation_password```为Websocket流量混淆密码。如果你使用了国内的CDN，建议设置```obfuscation_password```字段进行流量混淆。Trojan-Go将对Websocket承载的流量再进行一次加密(AES-128-CTR)。注意这个字段的主要目的仅仅是**混淆**上层流量的特征(TLS/Trojan)，防止被国内的CDN识别和封锁，**它无法确保传输数据安全性**。安全性应该由第二层TLS隧道保证。

服务器开启Websocket支持后可以同时支持Websocket和一般Trojan流量，未配置Websocket选项的客户端依然可以正常使用。

由于Trojan-GFW版本并不支持Websocket，因此，虽然开启了Websocket支持的服务端仍然可以兼容原版Trojan客户端，但是如果要使用Websocket承载流量进行CDN中转等，请确保双方都使用Trojan-Go。

如果你想提高传输的性能和吞吐量，可以将```double_tls```设为false或者将```obfuscation_password```设为空，此时websocket将会直接承载Trojan协议。但是出于安全性考虑，还是建议至少开启混淆和双重TLS中的至少一项。

**如果你使用的是国内的CDN，务必保证两者均开启。最坏情况下也应当保持混淆和双重TLS之一是打开的。**

CDN转发的场景，和GFW在2020年3月29日进行的对包括github pages等站点进行的HTTPS流量劫持和中间人攻击,是类似的。它们的共同点是，第一层TLS承载的流量明文均可以被第三者窃听（CDN或GFW）。**如果你使用了Websocket模式**，你可以将客户端全局```ssl```选项中```verify```字段填写为false。并指定```websocket```选项的```ssl```选项并且打开证书校验。在这样的设置下，即使第一层TLS传输的明文遭到审查或攻击，由于第二层TLS的保护，传输的内容依旧安全。

Trojan-Go同样具有针对Websocket的主动探测的欺骗能力。当一个合法的Webosocket握手完成，但密码不匹配或内容不合法时，将会尝试与```http://remote_addr:remote_port/path```的HTTP服务器进行Websocket握手，并将连入的Websocket连接代理给它。如果连接失败，Websocket会被直接关闭。

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
        "path": "/imaurlpath",
        "hostname": "www.your_awesome_domain_name.com",
        "obfuscation_password": "another_password",
        "double_tls": true
    }
}

```
