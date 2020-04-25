---
title: "SimpleSocks协议"
draft: false
weight: 5
---

SimpleSocks协议是无认证机制的简单代理协议，本质上就是去除了sha224的Trojan协议，设计的目的是减少多路复用时的overhead。只有启用多路复用之后，被复用的连接才会使用这个协议。

SimpleSocks甚至比Socks5更简单，下面是头部结构。

```
+-----+------+----------+----------+-----------+
| CMD | ATYP | DST.ADDR | DST.PORT |  Payload  |
+-----+------+----------+----------+-----------+
|  1  |  1   | Variable |    2     |  Variable |
+-----+------+----------+----------+-----------+
```

各字段定义与Socks5相同，不再赘述。

