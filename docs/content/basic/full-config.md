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

```json
{
  "run_type": *required*,
  "local_addr": *required*,
  "local_port": *required*,
  "remote_addr": *required*,
  "remote_port": *required*,
  "log_level": 1,
  "log_file": "",
  "password": [],
  "buffer_size": 32,
  "dns": [],
  "ssl": {
    "verify": true,
    "verify_hostname": true,
    "cert": *required*,
    "key": *required*,
    "key_password": "",
    "cipher": "",
    "cipher_tls13": "",
    "curves": "",
    "prefer_server_cipher": false,
    "sni": "",
    "alpn": [
      "h2",
      "http/1.1"
    ],
    "session_ticket": true,
    "reuse_session": true,
    "plain_http_response": "",
    "fallback_port": 0,
    "fingerprint": "firefox",
    "serve_plain_text": false
  },
  "tcp": {
    "no_delay": true,
    "keep_alive": true,
    "reuse_port": false,
    "prefer_ipv4": false,
    "fast_open": false,
    "fast_open_qlen": 20
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
    "default_policy": "proxy",
    "domain_strategy": "as_is",
    "geoip": "./geoip.dat",
    "geosite": "./geoip.dat"
  },
  "websocket": {
    "enabled": false,
    "path": "",
    "hostname": "",
    "obfuscation_password": "",
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
  },
  "forward_proxy": {
    "enabled": false,
    "proxy_addr": "",
    "proxy_port": 0,
    "username": "",
    "password": ""
  },
  "mysql": {
    "enabled": false,
    "server_addr": "localhost",
    "server_port": 3306,
    "database": "",
    "username": "",
    "password": "",
    "check_rate": 60
  },
  "redis": {
    "enabled": false,
    "server_addr": "localhost",
    "server_port": 6379,
    "password": ""
  },
  "api": {
    "enabled": false,
    "api_addr": "",
    "api_port": 0
  }
}
```

## 说明

### 一般选项

对于client/nat/forward，```remote_xxxx```应当填写你的trojan服务器地址和端口号，```local_xxxx```对应本地开放的socks5/http代理地址（自动适配）

对于server，```local_xxxx```对应trojan服务器监听地址（强烈建议使用443端口），```remote_xxxx```填写识别到非trojan流量时代理到的HTTP服务地址，通常填写本地80端口。

```log_level```指定日志等级。等级越高，输出的信息越少，0输出Debug以上日志（所有日志），1输出Info及以上日志，2输出Warning及以上日志，3输出Error及以上信息，4输出Fatal及以上信息，5完全不输出日志。

```log_file```指定日志输出文件路径。如果未指定则使用标准输出。

```password```可以填入多个密码。除了使用配置文件配置密码之外，trojan-go还支持使用mysql配置密码，参见下文。客户端的密码，只有与服务端配置文件中或者在数据库中的密码记录一致，才能通过服务端的校验，正常使用代理服务。

```dns```指定trojan-go使用的DNS服务器列表，如果不指定则使用主机默认DNS。如果指定了服务器，按照列表顺序依次查询，支持UDP/TCP/DOT类型的DNS，查询结果会被缓存五分钟。使用URL格式描述服务器，例如

- "udp://1.1.1.1"，基于UDP的DNS服务器，默认53端口

- "udp://1.1.1.1:53"，与上一项等价

- "1.1.1.1"，与上一项等价

- "tcp://1.1.1.1"，基于TCP的DNS服务器，默认53端口

- "dot://1.1.1.1"，基于DOT(DNS Over TLS)的DNS服务器，默认853端口

使用DOT可以防止DNS请求泄露，但由于TLS的握手耗费更多时间，查询速度也会有一定的下降，请自行斟酌性能和安全性的平衡。

```buffer_size```为单个连接缓冲区大小，单位KiB，默认32KiB。提升这个数值可以提升网络吞吐量和效率，但是也会增加内存消耗。对于路由器等嵌入式系统，建议根据实际情况，适当减小该数值。

### ```ssl```选项

```verify```表示客户端(client/nat/forward)是否校验服务端提供的证书合法性，默认开启。出于安全性考虑，这个选项不应该在实际场景中选择false，否则可能遭受中间人攻击。如果使用自签名或者自签发的证书，开启```verify```会导致校验失败。这种情况下，应当保持```verify```开启，然后在```cert```中填写服务端的证书，即可正常连接。

```verify_hostname```表示客户端(client/nat/forward)是否校验服务端提供的证书的Common Name和本地提供的SNI字段的一致性。

服务端必须填入```cert```和```key```，对应服务器的证书和私钥文件，请注意证书是否有效/过期。如果使用权威CA签发的证书，客户端(client/nat/forward)可以不填写```cert```。如果使用自签名或者自签发的证书，应当在的```cert```处填入服务器证书文件，否则可能导致校验失败。

```sni```指的是证书的Common Name，如果你使用letsencrypt等机构签名的证书，这里填入你的域名。如果这一项未填，将使用```remote_addr```填充。你应当指定一个有效的SNI（和远端证书CN一致），否则客户端可能无法验证远端证书有效性从而无法连接。

```alpn```为TLS的应用层协议协商指定协议。在TLS Client/Server Hello中传输，协商应用层使用的协议，仅用作指纹伪造，并无实际作用。**如果使用了CDN，错误的alpn字段可能导致与CDN协商错误的应用层协议**。

```prefer_server_cipher```客户端是否偏好选择服务端在协商中提供的密码学套件。

```cipher```和```cipher13```指TLS使用的密码学套件。只有在你明确知道自己在做什么的情况下，才应该去填写此项以修改trojan-go使用的TLS密码学套件。**正常情况下，你应该将其留空或者不填**，trojan-go会根据当前硬件平台以及远端的情况，自动选择最合适的加密算法以提升性能和安全性。如果需要填写，密码学套件名用分号(":")分隔。Go的TLS库中弃用了TLS1.2中不安全的密码学套件，并完全支持TLS1.3。默认情况下，trojan-go将优先使用更安全的TLS1.3。

```curves```指定TLS在ECDHE中偏好使用的椭圆曲线。只有你明确知道自己在做什么的情况下，才应该填写此项。曲线名称用分号(":")分隔。

```fingerprint```用于指定TLS Client Hello指纹伪造类型，以抵抗GFW对于TLS Client Hello指纹的特征识别和阻断。trojan-go使用[utls](https://github.com/refraction-networking/utls)进行指纹伪造，默认伪造Firefox的指纹。合法的值有

- ""，不使用指纹伪造

- "auto"，自动尝试并选择

- "firefox"，伪造Firefox指纹（默认）

- "chrome"，伪造Chrome指纹

- "ios"，伪造iOS指纹

一旦指纹的值被设置，```cipher```，```curves```，```alpn```，```session_ticket```等有可能影响指纹的字段将使用该指纹的特定设置覆写。

```plain_http_response```指定了当TLS握手失败时，明文发送的原始数据（原始TCP数据）。这个字段填入该文件路径。推荐使用```fallback_port```而不是该字段。

```fallback_port```指TLS握手失败时，trojan-go将该连接代理到该端口上。这是trojan-go的特性，以便更好地隐蔽Trojan服务器，抵抗GFW的主动检测，使得服务器的443端口在遭遇非TLS协议的探测时，行为与正常服务器完全一致。当服务器接受了一个连接但无法进行TLS握手时，如果```fallback_port```不为空，则流量将会被代理至remote_addr:fallback_port。例如，你可以在本地使用nginx开启一个https服务，当你的服务器443端口被非TLS协议请求时（比如http请求），trojan-go将代理至本地https服务器，nginx将使用http协议明文返回一个400 Bad Request页面。你可以通过使用浏览器访问```http://your_domain_name.com:443```进行验证。

```serve_plain_text```服务端直接是否直接接受TCP连接并处理trojan协议明文。开启此选项后，```ssl```的其他选项将失效，trojan-go将直接处理连入的TCP连接而不使用TLS。此选项的意义在于支持nginx等Web服务器的分流。如果开启，请不要将trojan-go服务对外暴露。

### ```mux```多路复用选项

多路复用是trojan-go的特性。如果服务器和客户端都是trojan-go，可以开启mux多路复用以减少高并发情景下的延迟（只需要客户端开启此选项即可，服务端自动适配）。

注意，开启多路复用不会提升你的链路速度。相反，它会增加客户端和服务端的CPU和内存消耗，从而可能造成速度下降。多路复用的意义在于降低延迟。

```enabled```是否开启多路复用

```concurrency```指单个TLS隧道可以承载的最大连接数，默认为8。这个数值越大，多连接并发时TLS由于握手产生的延迟就越低，但网络吞吐量可能会有所降低，填入负数或者0表示所有连接只使用一个TLS隧道承载。

```idle_timeout```指TLS隧道在空闲多久之后关闭，单位为秒。如果数值为负或0，则一旦TLS隧道空闲，则立即关闭

### ```router```路由选项

路由功能是trojan-go的特性。trojan-go的路由策略有三种。

- Proxy 代理。将请求通过TLS隧道进行代理，由trojan服务器和目的地址进行连接。

- Bypass 绕过。直接在本地和目的地址进行连接。

- Block 封锁。不代理请求，直接关闭连接。

在```proxy```, ```bypass```, ```block```字段中填入对应列表文件名或者geoip/geosite标签名，trojan-go即根据列表中的IP（CIDR）或域名执行相应路由策略。列表文件中每行是一个IP或者域名，trojan-go会自动识别。

```enabled```是否开启路由模块。

```default_policy```指的是三个列表匹配均失败后，使用的默认策略，默认为"bypass"，即进行代理。合法的值有

- "proxy"

- "bypass"

- "block"

含义同上。

```domain_strategy```域名解析策略，默认"as_is"。合法的值有：

- "as_is"，只在域名列表中进行匹配。

- "ip_if_nonmatch"，在域名列表中进行匹配，如果不匹配，解析为IP后在IP列表中匹配。该策略可能导致DNS泄漏或遭到污染。

- "ip_on_demand"，域名均解析为IP，在IP列表中匹配。该策略可能导致DNS泄漏或遭到污染。

```geoip```和```geosite```字段指geoip和geosite数据库文件路径，默认使用当前目录的geoip.dat和geosite.dat。

### ```websocket```选项

Websocket传输是trojan-go的特性。在**正常的直接连接代理节点**的情况下，开启这个选项不会改善你的链路速度（甚至有可能下降），也不会提升你的连接安全性。你只应该在下面两种情况下启用它：

- 你需要利用CDN进行流量中转

- 你到代理节点的直接TLS连接遭到了GFW的中间人攻击

警告：**由于信任CDN证书并使用CDN网络进行传输，HTTPS连接对于CDN是透明的，CDN运营商可以查看Websocket流量传输内容。如果你使用了国内的CDN，应当假定CDN不可信任，请务必开启double_tls进行双重加密，并使用obfuscation_password进行流量混淆**

```enabled```表示是否启用Websocket承载流量，服务端开启后同时支持一般Trojan协议和基于websocket的Trojan协议，客户端开启后将只使用websocket承载所有Trojan协议流量。

```path```指的是Websocket使用的URL路径，必须以斜杠("/")开头，如"/longlongwebsocketpath"，并且服务器和客户端必须一致。

```hostname```Websocket握手时使用的主机名，客户端如果留空则使用```remote_addr```填充。如果使用了CDN，这个选项一般填入域名。

```double_tls```是否开启双重TLS，默认开启。开启后在TLS+Websocket上将会再承载一次TLS连接。双重TLS的意义在于即使第一层TLS遭到中间人攻击也能保证通信安全。第二层TLS的证书校验被强制打开。客户端和服务端设置必须相同。这个选项对性能有一定影响，请自行斟酌安全性和性能的平衡。

```ssl```如果```double_tls```启用，这个选项用于配置第二层TLS，如果没有填写则使用全局的```ssl```填充。各字段定义与全局```ssl```相同。

```obfuscation_password```指定混淆密码。用于混淆内层连接以降低遭到国内无良CDN运营商识别的概率。如果需要使用混淆，服务端和客户端必须同时设置相同密码。这个选项对性能有一定影响，请自行斟酌安全性和性能的平衡。

### ```tcp```选项

```no_delay```是否禁用纳格算法(Nagle’s algorithm)，即TCP封包是否直接发出而不等待缓冲区填满。

```keep_alive```是否启用TCP心跳存活检测。

```reuse_port```是否启用端口复用。由于trojan-gfw版本对多线程支持不佳，因而服务器使用此选项开启多个进程监听同一端口以提升并发性能。trojan-go本身的并发性能足够优秀，并无必要开启此选项。该选项仅为兼容而保留。

```prefer_ipv4```是否优先使用IPv4地址。

```fast_open```是否启用TCP Fast Open。开启此选项需要操作系统支持。考虑到TFO开启后的TCP封包特征明显，容易被GFW阻断，且可能存在安全性问题，trojan-go仅仅出于兼容目的在服务端实现TFO支持。

```fast_open_qlen```TCP Fast Open的qlen值，即允许的同时发起的未经三次握手的TFO连接数量。

### ```mysql```数据库选项

trojan-go兼容trojan-gfw的基于mysql的用户管理方式，但更推荐的方式是使用API。

```enabled```表示是否启用mysql数据库进行用户验证。

```check_rate```是trojan-go从MySQL获取用户数据，更新缓存的间隔时间，单位是秒。

其他选项可以顾名思义，不再赘述。

users表结构和trojan-gfw定义一致，下面是一个创建users表的例子。注意这里的password指的是密码经过SHA224散列之后的值（字符串），流量download, upload, quota的单位是字节。你可以通过修改数据库users表中的用户记录的方式，添加和删除用户，或者指定用户的流量配额。trojan-go会根据所有的用户流量配额，自动更新当前有效的用户列表。如果download+upload>quota，trojan-go服务器将拒绝该用户的连接。

```mysql
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

### ```forward_proxy```前置代理选项

前置代理选项允许使用其他代理承载trojan-go的流量

```enabled```是否启用前置代理(socks5)。

```proxy_addr```前置代理的主机地址。

```proxy_port```前置代理的端口号。

```username``` ```password```代理的用户和密码，如果留空则不使用认证。

### ```api```选项

trojan-go基于gRPC提供了API，以支持服务端和客户端的管理和统计。可以实现客户端的流量和速度统计，服务端各用户的流量和速度统计，用户的动态增删和限速等。

```enabled```是否启用API功能。

```api_addr```gRPC监听的地址。

```api_port```gRPC监听的端口。

警告：**不要将API直接暴露在互联网上，否则可能导致各类安全问题**
