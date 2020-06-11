---
title: "多路复用"
draft: false
weight: 30
---

Trojan-Go使用[smux](https://github.com/xtaci/smux)实现多路复用。同时实现了simplesocks协议用于进行代理传输。

当启用多路复用时，客户端首先发起TLS连接，使用正常trojan协议格式，但协议Command部分填入0x7f(protocol.Mux)，标识此连接为复用连接（类似于http的upgrade），之后连接交由smux客户端管理。服务器收到请求头部后，交由smux服务器解析该连接的所有流量。在每条分离出的smux连接上，使用simplesocks协议（去除认证部分的trojan协议)标明代理目的地。自顶向下的协议栈如下：

|协议            |备注            |
|----------------|---------------|
|真实流量         |
|SimpleSocks     |
|smux            |
|Trojan          |用于鉴权        |
|底层协议         |               |
