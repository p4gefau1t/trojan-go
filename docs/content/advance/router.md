---
title: "国内直连和广告屏蔽"
draft: false
weight: 3
---

### 注意，Trojan不支持这个特性

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
        "sni": "your-domain-name.com"
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

所需要的数据库```geoip.dat```和```geosite.dat```已经包含在release的压缩包中，直接使用即可。它们来自V2Ray的[domain-list-community](https://github.com/v2fly/domain-list-community)和[geoip](https://github.com/v2fly/geoip)。

你可以使用如```geosite:cn```、```geosite:geolocation-!cn```、```geosite:category-ads-all```、```geosite:bilibili```的形式来指定某一类域名，所有可用的tag可以在[domain-list-community](https://github.com/v2fly/domain-list-community)仓库的[```data```](https://github.com/v2fly/domain-list-community/tree/master/data)目录中找到。```geosite.dat``` 更详细使用说明，参考[V2Ray/Routing路由#预定义域名列表](https://www.v2fly.org/config/routing.html#预定义域名列表)。

你可以使用如```geoip:cn```、```geoip:hk```、```geoip:us```、```geoip:private```的形式来指定某一类IP。`geoip:private`为特殊项，囊括了内网IP和保留IP，其余类别囊括了各个国家/地区的IP地址段。各国家/地区的代号参考[维基百科](https://zh.wikipedia.org/wiki/%E5%9C%8B%E5%AE%B6%E5%9C%B0%E5%8D%80%E4%BB%A3%E7%A2%BC)。

你也可以配置自己的路由规则。例如，想要屏蔽所有example.com域名以及其子域名，以及192.168.1.0/24，添加下面的规则。

```json
"block": [
    "domain:example.com",
    "cidr:192.168.1.0/24"
]
```

支持的格式有

- "domain:"，子域名匹配

- "full:"，完全域名匹配

- "regexp:"，正则表达式匹配

- "cidr:"，CIDR匹配

更详细的说明参考"完整的配置文件"一节。
