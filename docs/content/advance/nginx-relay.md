---
title: "一种基于SNI代理的多路径分流中继方案"
draft: false
weight: 6
---

## 前言

Trojan 是一种通过 TLS 封装后进行加密数据传输的工具，利用其 TLS 的特性，我们可以通过 SNI 代理实现在同一主机端口上实现不同路径的分流中继。

## 所需工具及其他准备

- 中继机：nginx 1.11.5 及以上版本
- 落地机：trojan 服务端（无版本要求）

## 配置方法

为了便于说明，这里使用了两台中继主机和两台落地主机。  
四台主机所绑定的域名分别为 (a/b/c/d).example.com。如图所示。  
相互连接一共4条路径。分别为 a-c、a-d、b-c、b-d 。

```text
                        +-----------------+           +--------------------+
                        |                 +---------->+                    |
                        |   VPS RELAY A   |           |   VPS ENDPOINT C   |
                  +---->+                 |   +------>+                    |
                  |     |  a.example.com  |   |       |   c.example.com    |
                  |     |                 +------+    |                    |
  +----------+    |     +-----------------+   |  |    +--------------------+
  |          |    |                           |  |
  |  client  +----+                           |  |
  |          |    |                           |  |
  +----------+    |     +-----------------+   |  |    +--------------------+
                  |     |                 |   |  |    |                    |
                  |     |   VPS RELAY B   |   |  +--->+   VPS ENDPOINT D   |
                  +---->+                 +---+       |                    |
                        |  b.example.com  |           |   d.example.com    |
                        |                 +---------->+                    |
                        +-----------------+           +--------------------+
```

### 配置路径域名和相应的证书

首先我们需要将每条路径分别分配一个域名，并使其解析到分别的入口主机上。  

```text
a-c.example.com CNAME a.example.com  
a-d.example.com CNAME a.example.com  
b-c.example.com CNAME b.example.com  
b-d.example.com CNAME b.example.com
```

然后我们需要在落地主机上部署所有目标路径的证书  
由于解析记录和主机 IP 不符，HTTP 验证无法通过。这里建议使用 DNS 验证方式签发证书。  
具体 DNS 验证插件需要根据您的域名 DNS 解析托管商选择，这里使用了 AWS Route 53。  

```shell
certbot certonly --dns-route53 -d a-c.example.com -d b-c.example.com // 主机 C 上
certbot certonly --dns-route53 -d a-d.example.com -d b-d.example.com // 主机 D 上
```

### 配置 SNI 代理

这里我们使用 nginx 的 ssl_preread 模块实现 SNI 代理。  
请安装 nginx 后按如下方法修过 nginx.conf 文件。  
注意这里不是 HTTP 服务，请不要写在虚拟主机的配置中。

这里给出主机 A 上的对应配置，主机 B 同理。

```nginx
stream {
  map $ssl_preread_server_name $name {
    a-c.example.com   c.example.com;  # 将 a-c 路径流量转发至主机 C
    a-d.example.com   d.example.com;  # 将 a-d 路径流量转发至主机 D

    # 如果此主机上需要配置其他占用 443 端口的服务 （例如 web 服务和 Trojan 服务）
    # 请使那些服务监听在其他本地端口（这里使用了 4000）
    # 所有不匹配上方 SNI 的 TLS 请求都会转发至此端口，如不需要可以删除此行
    default           localhost:4000;
  }

  server {
    listen      443; # 监听 443 端口
    proxy_pass  $name;
    ssl_preread on;
  }
}
```

### 配置落地 Trojan 服务

在之前的配置中我们使用了一个证书签发了所有目标路径的域名，所以这里我们可以使用一个 Trojan 服务端处理所有目标路径的请求。  
Trojan 的配置和通常配置方法无异，这里还是提供一份例子。无关的配置已省略。

```json
{
    "run_type": "server",
    "local_addr": "0.0.0.0",
    "local_port": 443,
    "ssl": {
        "cert": "/path/to/certificate.crt",
        "key": "/path/to/private.key",
    }
    ...
}
```

小提示：如果需要在落地主机上对不同路径分别使用独立的 Trojan 服务端（比如需要分别接入各自的计费服务），可以在落地机上再配置一个 SNI 代理，并分别转发至本地不同的 Trojan 服务端监听端口。由于配置与前面所提到的过程基本相同，这里便不再赘述。

## 总结

通过以上介绍的配置方法，我们可以在单一端口上实现多入口多出口多级中继的 Trojan 流量转发。  
对于多级中继，只需在中间节点上按相同思路配置 SNI 代理即可。
