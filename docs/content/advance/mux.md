---
title: "使用多路复用提升并发性能"
draft: false
---

Trojan-Go支持使用多路复用提升并发性能。

Trojan协议基于TLS。在一个TLS安全连接建立之前，连接双方需要进行密钥协商和交换等步骤确保后续通讯的安全性。这个过程即为TLS握手。

目前GFW对于TLS握手有审查和干扰，同时由于出口网络拥塞的原因，普通的线路完成TLS握手通常需要将近一秒甚至更长的时间。这可能会使得浏览网页和观看视频的延迟提高。

Trojan-Go使用多路复用的方式解决这一问题。每个建立的TLS连接将承载多个TCP连接。当新的代理请求到来时，不需要和服务器握手发起一个新的TLS连接，而是尽可能重复使用已有的TLS连接。这样就可以减少TLS握手的带来的延迟。在高并发的情况下，如浏览含有大量图片的网页时，优势尤其突出。

激活mux模块，只需要将```mux```选项中```enabled```字段设为true即可，下面是一个例子

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
    "mux" :{
        "enabled": true
    }
}
```

完整的mux配置如下

```
"mux": {
    "enabled": false,
    "concurrency": 8,
    "idle_timeout": 60
}
```

```concurrency```是每个TLS连接最多可以承载的TCP连接数。这个数值越大，TLS连接被复用的比例就更高，握手导致的延迟越低，但服务器和客户端的计算负担也会越大，这有可能使你的网络吞吐量降低。如果你的线路的TLS握手极端缓慢，你可以将这个数值设置为-1，Trojan-Go将只进行一次TLS握手，只使用唯一的一条TLS连接进行传输。

```idle_timeout```指的是每个TLS连接空闲多长时间后关闭。设置超时时间，**可能**有助于减少不必要的长连接存活确认(Keep Alive)流量传输引发GFW的探测。你可以将这个数值设置为-1，TLS连接将不会因为长时间空闲而被关闭。
