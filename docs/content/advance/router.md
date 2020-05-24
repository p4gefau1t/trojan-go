---
title: "国内直连和广告屏蔽"
draft: false
weight: 3
---

### 注意，Trojan-GFW版本不支持这个特性

Trojan-Go内建的路由模块可以帮助你实现国内直连，即客户端对于国内网站不经过代理，直接连接。

路由模块在客户端可以配置三种策略(```bypass```, ```proxy```, ```block```)，在服务端只可使用```block```策略。

下面是一个例子

```json
{
    "run_type": "client",
    "local_addr": "127.0.0.1",
    "local_port": 1080,
    "remote_addr": "your_server",
    "remote_port": 443,
    "password": [
        "your_password"
    ],
    "ssl": {
        "sni": "your_domain_name"
    },
    "mux" :{
        "enabled": true
    },
    "router":{
        "enabled": true,
        "bypass": [
            "geoip:cn",
            "geoip:private",
            "geosite:cn",
            "geosite:geolocation-cn"
        ],
        "block": [
            "geosite:category-ads"
        ],
        "proxy": [
            "geosite:geolocation-!cn"
        ]
    }
}
```

这个配置文件激活了router模块，使用的是白名单的模式，当匹配到中国大陆或者局域网的IP/域名时，直接连接。如果是广告运营商的域名，则直接断开连接。

所需要的数据库geoip.dat和geosite.dat已经包含在release的压缩包中，直接使用即可。它们来自v2ray的[domain-list-community](https://github.com/v2ray/domain-list-community)和[geoip](https://github.com/v2ray/geoip)。你可以使用如```geosite:cn```，```geosite:bilibili```的形式来指定某一类域名，使用如```geoip:cn```，```geoip:private```的形式来指定某一类IP。所有可用的tag可以在[domain-list-community](https://github.com/v2ray/domain-list-community)仓库中找到。

你也可以配置自己的列表文件，列表文件每一行是一个域名或者IP子网（CIDR）。例如，想要屏蔽所有example.com域名以及其子域名，以及192.168.1.0/24，只需要编写一个txt文件

test_list.txt

```text
example.com
192.168.1.0/24
```

然后在修改上面的block选项，添加一行指定该文件名即可

```json
"block": [
    "geosite:category-ads",
    "test_list.txt"
]
```

更详细的说明参考"完整的配置文件"一节。
