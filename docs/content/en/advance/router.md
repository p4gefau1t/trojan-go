---
title: "Domestic direct connection and ad blocking"
draft: false
weight: 3
---

{{% panel status="caution" title="Compatibility" %}}
Note that Trojan does not support this feature
{{% /panel %}}

Trojan-Go's built-in routing module can help you implement domestic direct connections, i.e., the client can connect directly to domestic websites without going through a proxy.

The routing module can be configured with three policies (```bypass```, ```proxy```, ```block```) on the client side, and only the ```block``` policy can be used on the server side.

Here is an example

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

This configuration file activates the router module, which uses a whitelist mode, and connects directly when matched to an IP/domain name from mainland China or the LAN. If it is the domain name of the advertising operator, it is directly disconnected.

The required databases ```geoip.dat``` and ```geosite.dat``` are included in the release zip and can be used directly. They come from V2Ray's [domain-list-community](https://github.com/v2fly/domain-list-community) and [geoip](https://github.com/v2fly/geoip).

You can specify a certain category of domains using forms like ```geosite:cn```, ```geosite:geolocation-!cn```, ```geosite:category-ads-all```, ```geosite:bilibili```, and all available tags can be found in [ domain-list-community](https://github.com/v2fly/domain-list-community) repository of [data](https://github.com/v2fly/domain-list-community/tree/master/data) directory. ```geosite.dat``` For more detailed usage instructions, refer to [V2Ray/Routing#List of predefined domains](https://www.v2fly.org/config/routing.html#预定义域名列表).

You can specify a certain category of IPs using forms such as ```geoip:cn```, ```geoip:hk```, ```geoip:us```, and ```geoip:private```. ```geoip:private``` is a special entry that encompasses intranet IPs and reserved IPs, and the rest of the categories encompass individual country/region IP address segment. Refer to [Wikipedia](https://zh.wikipedia.org/wiki/%E5%9C%8B%E5%AE%B6%E5%9C%B0%E5%8D%80%E4%BB%A3%E7%A2%BC) for the country/region codes.

You can also configure your own routing rules. For example, to block all example.com domains and their subdomains, as well as 192.168.1.0/24, add the following rule.

```json
"block": [
    "domain:example.com",
    "cidr:192.168.1.0/24"
]
```

The supported formats are

- "domain:", subdomain match

- "full:", full domain match

- "regexp:", regular expression match

- "cidr:", CIDR match

For more details, please refer to the section "The Complete Configuration File".
