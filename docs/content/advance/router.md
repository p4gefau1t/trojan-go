---
title: "国内直连和广告屏蔽"
draft: false
weight: 3
---

Trojan-Go内建的路由模块可以帮助你实现国内直连，即国内网站不经过代理，直接连接。

下面是一个例子

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
    "ssl": {
        "sni": "your_domain_name"
    },
    "mux" :{
        "enabled": true
    },
    "router":{
        "enabled": true,
        "bypass": [
            "geosite:cn",
            "geoip:cn",
            "geoip:private"
        ],
        "block": [
            "geosite:category-ads"
        ]
    }
}
```

所需要的geoip.dat和geosite.dat已经包含在release的压缩包中，直接使用即可。它们来自v2ray的[domain-list-community](https://github.com/v2ray/domain-list-community)和[geoip](https://github.com/v2ray/geoip)。

这个配置文件激活了router模块，使用的是白名单的模式，当匹配到中国大陆的ip或域名时，将使用直接连接，否则使用trojan代理进行连接。

你也可以配置自己的列表文件，列表文件每一行是一个域名或者IP子网（CIDR）。例如，你想要屏蔽所有example.com域名以及其子域名，以及192.168.1.0/24，只需要编写一个txt文件

test_list.txt
```
example.com
192.168.1.0/24
```

然后在block字段中填入该文件名

同时geosite中也含有广告提供商的域名，可以通过"geosite:category-ads"指定屏蔽它们。下面这个例子使用了一个列表文件，和geosite的category-ads标签，对相关连接进行屏蔽

```
"router":{
    "enabled": true,
    "bypass": [
        "geosite:cn",
        "geoip:cn",
        "geoip:private"
    ],
    "block": [
        "test_list.txt"，
        "geosite:category-ads"
    ]
}
```

下面介绍完整的路由功能

路由策略有三种

- Proxy 代理。将请求通过TLS隧道进行代理，由trojan服务器和目的地址进行连接。

- Bypass 绕过。直接在本地和目的地址进行连接。

- Block 封锁。不代理请求，直接关闭连接。

```
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
    ]
    "proxy": [
        "proxy_list.txt"
    ]
}
```

其中```bypass```,```block```, ```proxy```字段中填入相应的列表文件或者geo数据库tag。列表文件每行是一个域名或者IP地址段(CIDR)。geo数据库geoip和geosite为IP数据库和域名数据库。一旦匹配，则执行相应策略。