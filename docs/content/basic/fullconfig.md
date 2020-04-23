---
title: "完整的配置文件"
draft: false
weight: 30
---

下面是一个完整的配置文件，其中的必填选项有

- ```run_type```

- ```local_addr```

- ```local_port```

- ```remote_addr```

- ```remote_port```

对于服务器server，```key```和```cert```为必填。

对于客户端client，反向代理隧道forward，以及透明代理nat，```password```必填

其余未填的选项，用下面给出的值进行填充。

```
{
    "run_type": "client/server/nat",
    "local_addr": "127.0.0.1",
    "local_port":  your_port1,
    "remote_addr": "example.com",
    "remote_port": your_port2,
    "log_level": 1,
    "password": [
        "password1",
        "password2"
    ],
    "buffer_size": 512,
    "ssl": {
        "verify": true,
        "verify_hostname": true,
        "cert": "your_crt_file(optional for a client/forward/nat)",
        "key": "your_key_file(optional for a client/forward/nat)",
        "key_password": "",
        "cipher": "",
        "cipher_tls13": "",
        "prefer_server_cipher": false,
        "sni": "",
        "session_ticket": true,
        "reuse_session": true,
        "plain_http_response": "",
        "fallback_port":0
    },
    "tcp": {
        "no_delay": false,
        "reuse_port": false,
        "prefer_ipv4": false,
        "fast_open": false
    },
    "mux": {
        "enabled": false,
        "concurrency": 8,
        "idle_timeout": 60
    },
    "router": {
        "enabled": false,
        "bypass": [],
        "proxy": [],
        "block": [],
        "route_by_ip": false,
        "route_by_ip_on_nonmatch": false，
        "default_policy": "proxy"
    },
    "websocket": {
        "enabled": false,
        "path": "",
        "hostname": "",
        "obfuscation": true,
        "double_tls": true
    },
    "mysql": {
        "enabled": false,
        "server_addr": "",
        "server_port": 0,
        "database": "",
        "username": "",
        "password": "",
        "check_rate": 60
    }
}
```

## 说明

### 一般选项

对于client/nat/forward，```remote_xxxx```应当填写你的trojan服务器地址和端口号，```local_xxxx```对应本地开放的socks5/http代理地址（自动适配）

对于server，```local_xxxx```对应trojan服务器监听地址（强烈建议使用443端口），```remote_xxxx```填写发现非trojan流量时代理到的Web服务地址，通常填写本地80端口。

```log_level```是日志等级，等级越高，输出的信息越少，0输出所有日志，1输出Info及以上日志，2输出Warning及以上日志，3输出Error及以上信息，4输出Fatal及以上信息，5完全不输出日志。

```password```可以填入多个密码。除了使用配置文件配置密码之外，Trojan-Go还支持使用mysql配置密码，参见下文。客户端的密码，只有与服务端配置文件中或者在数据库中的密码记录一致，才能通过服务端的校验，正常使用代理服务。

```buffer_size```为单个连接缓冲区大小，单位KiB，默认512KiB。提升这个数值可以提升网络吞吐量和效率，但是也会增加内存消耗。

### ```ssl```选项

```verify```表示客户端(client/nat/forward)是否校验服务端提供的证书合法性，默认开启。出于安全性考虑，这个选项不应该在实际场景中选择false，否则可能遭受中间人攻击。如果使用自签名或者自签发的证书，开启```verify```会导致校验失败。这种情况下，应当保持```verify```开启，然后在```cert```中填写服务端的证书，即可正常连接。

```verify_hostname```表示客户端(client/nat/forward)是否校验服务端提供的证书的Common Name和本地提供的SNI字段的一致性。

服务端必须填入```cert```和```key```，对应服务器的证书和私钥文件，请注意证书是否有效/过期。如果使用权威CA签发的证书，客户端(client/nat/forward)可以不填写```cert```。如果使用自签名或者自签发的证书，应当在的```cert```处填入服务器证书文件，否则可能导致校验失败。

```sni```指的是证书的Common Name，如果你使用letsencrypt等机构签名的证书，这里填入你的域名。如果这一项未填，将使用```remote_addr```填充。你应当指定一个有效的SNI（和远端证书CN一致），否则客户端可能无法验证远端证书有效性从而无法连接。

```cipher```和```cipher13```指client/server使用的密码学套件。只有在你明确知道自己在做什么的情况下，你才应该去填写cipher/cipher_tls13以修改trojan-go使用的TLS密码学套件。**正常情况下，你应该将其留空或者不填**，trojan-go会根据当前硬件平台以及远端的情况，自动选择最合适的加密算法以提升性能和安全性。如果需要填写，密码学套件名用分号(":")分隔。Golang的TLS库中中弃用了TLS1.2不安全的密码学套件，完全支持TLS1.3。如果你需要较高的安全性，而不担心跨硬件和软件平台的兼容性和性能，你可以强制要求trojan-go只使用TLS1.3密码学套件，设置cipher或者cipher13如下（填写cipher13和cipher是一样的）：

```
cipher13:"TLS_AES_128_GCM_SHA256:TLS_CHACHA20_POLY1305_SHA256:TLS_AES_256_GCM_SHA384"
```

```plain_http_response```指定了当TLS握手失败时，明文发送的原始数据（原始TCP数据）,这个字段填入该文件路径。推荐使用```fallback_port```而不是该字段。

```fallback_port```指TLS握手失败时，Trojan-Go将该连接代理到该端口上。这是Trojan-Go的特性，以便更好地隐蔽Trojan服务器，抵抗GFW的主动检测，使得服务器的443端口在遭遇非TLS协议的探测时，行为与正常服务器完全一致。当服务器接受了一个连接但无法进行TLS握手时，如果```fallback_port```不为空，则流量将会被代理至remote_addr:fallback_port。例如，你可以在本地使用nginx开启一个https服务，当你的服务器443端口被非TLS协议请求时（比如http请求），trojan-go将代理至本地https服务器，nginx将使用http协议明文返回一个400 Bad Request页面。你可以通过使用浏览器访问http://your_domain_name.com:443进行验证。

### ```mux```多路复用选项

多路复用是Trojan-Go的特性。如果服务器和客户端都是Trojan-Go，可以开启mux多路复用以减少高并发情景下的延迟（只需要客户端开启此选项即可，服务端自动适配）。

```enabled```是否开启多路复用

```mux_concurrency```指单个TLS隧道可以承载的最大连接数，默认为8。这个数值越大，多连接并发时TLS由于握手产生的延迟就越低，但网络吞吐量可能会有所降低，填入负数或者0表示所有连接只使用一个TLS隧道承载。

```mux_idle_timeout```指TLS隧道在空闲多久之后关闭，单位为秒。

### ```router```路由选项

路由功能是Trojan-Go的特性。Trojan-Go的路由策略有三种。

- Proxy 代理。将请求通过TLS隧道进行代理，由trojan服务器和目的地址进行连接。

- Bypass 绕过。直接在本地和目的地址进行连接。

- Block 封锁。不代理请求，直接关闭连接。

在```proxy```, ```bypass```, ```block```字段中填入对应列表文件名或者geoip/geosite标签名，Trojan-Go即根据列表中的IP（CIDR）或域名执行相应路由策略。列表文件中每行是一个IP或者域名，Trojan-Go会自动识别。

```enabled```是否开启路由模块。

```route_by_ip``` 开启后，所有域名会被在本地解析为IP后，仅使用IP列表进行匹配。如果开启这个选项，可能导致DNS泄露。

```route_by_ip_on_nonmatch```开启后，如果一个域名不在三个列表中，则会被在本地解析为IP后，仅使用IP列表进行匹配。如果开启这个选项，可能导致DNS泄露。

```default_policy```指的是三个列表匹配均失败后，使用的默认策略，默认为Proxy，即进行代理。

### ```websocket```选项

Websocket传输是Trojan-Go的特性。在**正常的直接连接服务器节点**的情况下，开启这个选项不会提升线路质量（甚至有可能下降），也不会提升你的连接安全性。你只应该在下面两种情况下启用它：

- 你需要利用CDN进行流量中转

- 你到代理节点的直接TLS连接遭到了GFW的中间人攻击

*警告：由于信任CDN证书并使用CDN网络进行传输，HTTPS连接对于CDN是透明的，CDN运营商可以查看Websocket流量传输内容。如果你使用了国内CDN，务必开启double_tls进行双重加密，并使用password进行流量混淆*

```enabled```表示是否启用Websocket承载流量，服务端开启后同时支持一般Trojan协议和基于websocket的Trojan协议，客户端开启后将只使用websocket承载所有Trojan协议流量。

```path```指的是Websocket使用的URL路径，必须以斜杠("/")开头，如"/longlongwebsocketpath"，并且服务器和客户端必须一致。

```hostname```Websocket握手时使用的主机名，如果留空则使用```remote_addr```填充。如果使用了CDN，这个选项一般填入域名。

```double_tls```是否开启双重TLS，默认开启。开启后在TLS+Websocket上将会再承载一次TLS连接。双重TLS的意义在于即使第一层TLS遭到中间人攻击也能保证通信安全。第二层TLS的证书校验被强制打开。客户端和服务端设置必须相同。这个选项对性能有一定影响，如果需要关闭，请自行斟酌安全性和性能的平衡。

```obfuscation```是否启用混淆。用于混淆内层连接以降低遭到国内无良CDN运营商识别的概率。如果需要使用混淆，服务端和客户端必须同时开启这个选项。这个选项对性能有一定影响，如果需要关闭，请自行斟酌安全性和性能的平衡。

### ```mysql```数据库选项

```enabled```表示是否启用mysql数据库进行用户验证。

```check_rate```是Trojan-Go从MySQL获取用户数据，更新缓存的间隔时间，单位是秒。

其他选项可以顾名思义，不再赘述。

users表结构和trojan原版一致，下面是一个创建users表的命令。注意这里的password指的是密码经过SHA224哈希之后的值（字符串），流量download, upload, quota的单位是字节。你可以通过修改数据库users表中的用户记录的方式，添加和删除用户，或者指定用户的流量配额。Trojan-Go会根据所有的用户流量配额，自动更新当前有效的用户列表。如果download+upload>quota，trojan-go服务器将拒绝该用户的连接。

```
CREATE TABLE users (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    username VARCHAR(64) NOT NULL,
    password CHAR(56) NOT NULL,
    quota BIGINT NOT NULL DEFAULT 0,
    download BIGINT UNSIGNED NOT NULL DEFAULT 0,
    upload BIGINT UNSIGNED NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    INDEX (password)
);
```