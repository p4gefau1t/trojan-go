---
title: "Websocket"
draft: false
weight: 4
---

开启Websocket的客户端可以使用```password```字段开启混淆/启用双重TLS以确保连接安全性。

由于使用CDN中转时，HTTPS对CDN透明，CDN可以审查Websocket传输内容，而Trojan协议头部特征过于明显（SHA224，"\r\n"等），为了保证不被审查和传输安全，默认情况下还会进行一次AES加密(混淆层)和TLS连接(双重TLS)。

混淆层使用AES-CTR-128加密方式，使用```password```的md5作为主密钥对承载流量进行加密。

协议栈如下：

|协议| 
|-|
|真实流量|
|Trojan|
|TLS(如果开启双重TLS)|
|混淆层(如果开启混淆)|
|Websocket|
|TLS|
|TCP|