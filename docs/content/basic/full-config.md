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

对于服务器```server```，```key```和```cert```为必填。

对于客户端```client```，反向代理隧道```forward```，以及透明代理```nat```，```password```必填

其余未填的选项，用下面给出的值进行填充。

*Trojan-Go支持对人类更友好的YAML语法，配置文件的基本结构与JSON相同，效果等价。但是为了遵守YAML的命名习惯，你需要把下划线("_")转换为横杠("-")，如```remote_addr```在YAML文件中为```remote-addr```*

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
  "disable_http_check": false,
  "udp_timeout": 10,
  "ssl": {
    "verify": true,
    "verify_hostname": true,
    "cert": *required*,
    "key": *required*,
    "key_password": "",
    "cipher": "",
    "curves": "",
    "prefer_server_cipher": false,
    "sni": "",
    "alpn": [
      "http/1.1"
    ],
    "session_ticket": true,
    "reuse_session": true,
    "plain_http_response": "",
    "fallback_addr": "",
    "fallback_port": 0,
    "fingerprint": "firefox"
  },
  "tcp": {
    "no_delay": true,
    "keep_alive": true,
    "prefer_ipv4": false
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
    "geoip": "$PROGRAM_DIR$/geoip.dat",
    "geosite": "$PROGRAM_DIR$/geosite.dat"
  },
  "websocket": {
    "enabled": false,
    "path": "",
    "host": ""
  },
  "shadowsocks": {
    "enabled": false,
    "method": "AES-128-GCM",
    "password": ""
  },
  "transport_plugin": {
    "enabled": false,
    "type": "",
    "command": "",
    "plugin_option": "",
    "arg": [],
    "env": []
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
  "api": {
    "enabled": false,
    "api_addr": "",
    "api_port": 0,
    "ssl": {
      "enabled": false,
      "key": "",
      "cert": "",
      "verify_client": false,
      "client_cert": []
    }
  }
}
```

## 说明

### 一般选项

对于client/nat/forward，```remote_xxxx```应当填写你的trojan服务器地址和端口号，```local_xxxx```对应本地开放的socks5/http代理地址（自动适配）

对于server，```local_xxxx```对应trojan服务器监听地址（强烈建议使用443端口），```remote_xxxx```填写识别到非trojan流量时代理到的HTTP服务地址，通常填写本地80端口。

```log_level```指定日志等级。等级越高，输出的信息越少。合法的值有

- 0 输出Debug以上日志（所有日志）

- 1 输出Info及以上日志

- 2 输出Warning及以上日志

- 3 输出Error及以上日志

- 4 输出Fatal及以上日志

- 5 完全不输出日志

```log_file```指定日志输出文件路径。如果未指定则使用标准输出。

```password```可以填入多个密码。除了使用配置文件配置密码之外，trojan-go还支持使用mysql配置密码，参见下文。客户端的密码，只有与服务端配置文件中或者在数据库中的密码记录一致，才能通过服务端的校验，正常使用代理服务。

```buffer_size```为单个连接缓冲区大小，单位KiB，默认32KiB。适当提升这个数值可以提升网络吞吐量和效率，但是也会增加内存消耗。对于路由器等嵌入式系统，建议根据实际情况，适当减小该数值。

```disable_http_check```是否禁用HTTP伪装服务器可用性检查。

```udp_timeout``` UDP会话超时时间。

### ```ssl```选项

```verify```表示客户端(client/nat/forward)是否校验服务端提供的证书合法性，默认开启。出于安全性考虑，这个选项不应该在实际场景中选择false，否则可能遭受中间人攻击。如果使用自签名或者自签发的证书，开启```verify```会导致校验失败。这种情况下，应当保持```verify```开启，然后在```cert```中填写服务端的证书，即可正常连接。

```verify_hostname```表示服务端是否校验客户端提供的SNI与服务端设置的一致性。如果服务端SNI字段留空，认证将被强制关闭。

服务端必须填入```cert```和```key```，对应服务器的证书和私钥文件，请注意证书是否有效/过期。如果使用权威CA签发的证书，客户端(client/nat/forward)可以不填写```cert```。如果使用自签名或者自签发的证书，应当在的```cert```处填入服务器证书文件，否则可能导致校验失败。

```sni```指的是TLS客户端请求中的服务器名字段，一般和证书的Common Name相同。如果你使用let'sencrypt等机构签发的证书，这里填入你的域名。对于客户端，如果这一项未填，将使用```remote_addr```填充。你应当指定一个有效的SNI（和远端证书CN一致），否则客户端可能无法验证远端证书有效性从而无法连接；对于服务端，若此项不填，则使用证书中Common Name作为SNI校验依据，支持通配符如*.example.com。

```fingerprint```用于指定客户端TLS Client Hello指纹伪造类型，以抵抗GFW对于TLS Client Hello指纹的特征识别和阻断。trojan-go使用[utls](https://github.com/refraction-networking/utls)进行指纹伪造，默认伪造Firefox的指纹。合法的值有

- ""，不使用指纹伪造

- "firefox"，伪造Firefox指纹（默认）

- "chrome"，伪造Chrome指纹

- "ios"，伪造iOS指纹

一旦指纹的值被设置，客户端的```cipher```，```curves```，```alpn```，```session_ticket```等有可能影响指纹的字段将使用该指纹的特定设置覆写。
```alpn```为TLS的应用层协议协商指定协议。在TLS Client/Server Hello中传输，协商应用层使用的协议，仅用作指纹伪造，并无实际作用。**如果使用了CDN，错误的alpn字段可能导致与CDN协商得到错误的应用层协议**。

```prefer_server_cipher```客户端是否偏好选择服务端在协商中提供的密码学套件。

```cipher```TLS使用的密码学套件。```cipher13``字段与此字段合并。只有在你明确知道自己在做什么的情况下，才应该去填写此项以修改trojan-go使用的TLS密码学套件。**正常情况下，你应该将其留空或者不填**，trojan-go会根据当前硬件平台以及远端的情况，自动选择最合适的加密算法以提升性能和安全性。如果需要填写，密码学套件名用分号(":")分隔，按优先顺序排列。Go的TLS库中弃用了TLS1.2中部分不安全的密码学套件，并完全支持TLS1.3。默认情况下，trojan-go将优先使用更安全的TLS1.3。

```curves```指定TLS在ECDHE中偏好使用的椭圆曲线。只有你明确知道自己在做什么的情况下，才应该填写此项。曲线名称用分号(":")分隔，按优先顺序排列。

```plain_http_response```指服务端TLS握手失败时，明文发送的原始数据（原始TCP数据）。这个字段填入该文件路径。推荐使用```fallback_port```而不是该字段。

```fallback_addr```和```fallback_port```指服务端TLS握手失败时，trojan-go将该连接重定向到该地址。这是trojan-go的特性，以便更好地隐蔽服务器，抵抗GFW的主动检测，使得服务器的443端口在遭遇非TLS协议的探测时，行为与正常服务器完全一致。当服务器接受了一个连接但无法进行TLS握手时，如果```fallback_port```不为空，则流量将会被代理至fallback_addr:fallback_port。如果```fallback_addr```为空，则用```remote_addr```填充。例如，你可以在本地使用nginx开启一个https服务，当你的服务器443端口被非TLS协议请求时（比如http请求），trojan-go将代理至本地https服务器，nginx将使用http协议明文返回一个400 Bad Request页面。你可以通过使用浏览器访问```http://your-domain-name.com:443```进行验证。

```key_log```TLS密钥日志的文件路径。如果填写则开启密钥日志。**记录密钥将破坏TLS的安全性，此项不应该用于除调试以外的其他任何用途。**

### ```mux```多路复用选项

多路复用是trojan-go的特性。如果服务器和客户端都是trojan-go，可以开启mux多路复用以减少高并发情景下的延迟（只需要客户端开启此选项即可，服务端自动适配）。

注意，多路复用的意义在于降低握手延迟，而不是提升链路速度。相反，它会增加客户端和服务端的CPU和内存消耗，从而可能造成速度下降。

```enabled```是否开启多路复用。

```concurrency```指单个TLS隧道可以承载的最大连接数，默认为8。这个数值越大，多连接并发时TLS由于握手产生的延迟就越低，但网络吞吐量可能会有所降低，填入负数或者0表示所有连接只使用一个TLS隧道承载。

```idle_timeout```空闲超时时间。指TLS隧道在空闲多长时间之后关闭，单位为秒。如果数值为负值或0，则一旦TLS隧道空闲，则立即关闭。

### ```router```路由选项

路由功能是trojan-go的特性。trojan-go的路由策略有三种。

- Proxy 代理。将请求通过TLS隧道进行代理，由trojan服务器和目的地址进行连接。

- Bypass 绕过。直接在本地和目的地址进行连接。

- Block 封锁。不代理请求，直接关闭连接。

在```proxy```, ```bypass```, ```block```字段中填入对应列表geoip/geosite或路由规则，trojan-go即根据列表中的IP（CIDR）或域名执行相应路由策略。客户端(client)可以配置三种策略，服务端(server)只可配置block策略。

```enabled```是否开启路由模块。

```default_policy```指的是三个列表匹配均失败后，使用的默认策略，默认为"proxy"，即进行代理。合法的值有

- "proxy"

- "bypass"

- "block"

含义同上。

```domain_strategy```域名解析策略，默认"as_is"。合法的值有：

- "as_is"，只在域名列表中进行匹配。

- "ip_if_non_match"，在域名列表中进行匹配，如果不匹配，解析为IP后在IP列表中匹配。该策略可能导致DNS泄漏或遭到污染。

- "ip_on_demand"，域名均解析为IP，在IP列表中匹配。该策略可能导致DNS泄漏或遭到污染。

```geoip```和```geosite```字段指geoip和geosite数据库文件路径，默认使用程序所在目录的geoip.dat和geosite.dat。也可以通过指定环境变量TROJAN_GO_LOCATION_ASSET指定工作目录。

### ```websocket```选项

Websocket传输是trojan-go的特性。在**正常的直接连接代理节点**的情况下，开启这个选项不会改善你的链路速度（甚至有可能下降），也不会提升你的连接安全性。你只应该在需要利用CDN进行中转，或利用nginx等服务器根据路径分发的情况下，使用websocket。

```enabled```表示是否启用Websocket承载流量，服务端开启后同时支持一般Trojan协议和基于websocket的Trojan协议，客户端开启后将只使用websocket承载所有Trojan协议流量。

```path```指的是Websocket使用的URL路径，必须以斜杠("/")开头，如"/longlongwebsocketpath"，并且服务器和客户端必须一致。

```host```Websocket握手时，HTTP请求中使用的主机名。客户端如果留空则使用```remote_addr```填充。如果使用了CDN，这个选项一般填入域名。不正确的```host```可能导致CDN无法转发请求。

### ``shadowsocks`` AEAD加密选项

此选项用于替代弃用的混淆加密和双重TLS。如果此选项被设置启用，Trojan协议层下将插入一层Shadowsocks AEAD加密层。也即（已经加密的）TLS隧道内，所有的Trojan协议将再使用AEAD方法进行加密。注意，此选项和Websocket是否开启无关。无论Websocket是否开启，所有Trojan流量都会被再进行一次加密。

注意，开启这个选项将有可能降低传输性能，你只应该在不信任承载Trojan协议的传输信道的情况下，启用这个选项。例如：

- 你使用了Websocket，经过不可信的CDN进行中转（如国内CDN）

- 你与服务器的连接遭到了GFW针对TLS的中间人攻击

- 你的证书失效，无法验证证书有效性

- 你使用了无法保证密码学安全的可插拔传输层

等等。

由于使用的是AEAD，trojan-go可以正确判断请求是否有效，是否遭到主动探测，并作出相应的响应。

```enabled```是否启用Shadowsocks AEAD加密Trojan协议层。

```method```加密方式。合法的值有：

- "CHACHA20-IETF-POLY1305"

- "AES-128-GCM" (默认)

- "AES-256-GCM"

```password```用于生成主密钥的密码。如果启用AEAD加密，必须确保客户端和服务端一致。

### ```transport_plugin```传输层插件选项

```enabled```是否启用传输层插件替代TLS传输。一旦启用传输层插件支持，trojan-go将会把**未经TLS加密的trojan协议流量明文传输给插件**，以允许用户对流量进行自定义的混淆和加密。

```type```插件类型。目前支持的类型有

- "shadowsocks"，支持符合[SIP003](https://shadowsocks.org/en/spec/Plugin.html)标准的shadowsocks混淆插件。trojan-go将在启动时按照SIP003标准替换环境变量并修改自身配置(```remote_addr/remote_port/local_addr/local_port```)，使插件与远端直接通讯，而trojan-go仅监听/连接插件。

- "plaintext"，使用明文传输。选择此项，trojan-go不会修改任何地址配置(```remote_addr/remote_port/local_addr/local_port```)，也不会启动```command```中插件，仅移除最底层的TLS传输层并使用TCP明文传输。此选项目的为支持nginx等接管TLS并进行分流，以及高级用户进行调试测试。**请勿直接使用明文传输模式穿透防火墙。**

- "other"，其他插件。选择此项，trojan-go不会修改任何地址配置(```remote_addr/remote_port/local_addr/local_port```)，但会启动```command```中插件并传入参数和环境变量。

```command```传输层插件可执行文件的路径。trojan-go将在启动时一并执行它。

```arg```传输层插件启动参数。这是一个列表，例如```["-config", "test.json"]```。

```env```传输层插件环境变量。这是一个列表，例如```["VAR1=foo", "VAR2=bar"]```。

```option```传输层插件配置（SIP003)。例如```"obfs=http;obfs-host=www.baidu.com"```。

### ```tcp```选项

```no_delay```TCP封包是否直接发出而不等待缓冲区填满。

```keep_alive```是否启用TCP心跳存活检测。

```prefer_ipv4```是否优先使用IPv4地址。

### ```mysql```数据库选项

trojan-go兼容trojan的基于mysql的用户管理方式，但更推荐的方式是使用API。

```enabled```表示是否启用mysql数据库进行用户验证。

```check_rate```是trojan-go从MySQL获取用户数据并更新缓存的间隔时间，单位为秒。

其他选项可以顾名思义，不再赘述。

users表结构和trojan版本定义一致，下面是一个创建users表的例子。注意这里的password指的是密码经过SHA224散列之后的值（字符串），流量download, upload, quota的单位是字节。你可以通过修改数据库users表中的用户记录的方式，添加和删除用户，或者指定用户的流量配额。trojan-go会根据所有的用户流量配额，自动更新当前有效的用户列表。如果download+upload>quota，trojan-go服务器将拒绝该用户的连接。

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

```ssl``` TLS相关设置。

- ```enabled```是否使用TLS传输gRPC流量。

- ```key```，```cert```服务器私钥和证书。

- ```verify_client```是否认证客户端证书。

- ```client_cert```如果开启客户端认证，此处填入认证的客户端证书列表。

警告：**不要将未开启TLS双向认证的API服务直接暴露在互联网上，否则可能导致各类安全问题。**
