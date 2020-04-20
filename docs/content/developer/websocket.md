---
title: "Websocket"
draft: false
weight: 4
---

由于使用CDN中转时，HTTPS对CDN透明，CDN可以审查Websocket传输内容，而Trojan协议头部特征过于明显（SHA224，"\r\n"等），为了保证不被审查和传输安全，默认情况下还会进行一次AES加密(混淆层)和TLS连接(双重TLS)。

开启Websocket的客户端可以使用```password```字段开启混淆，以及使用```double_tls```启用双重TLS以确保连接安全性。

混淆层使用AES-CTR-128加密方式，使用```password```的md5作为主密钥对承载流量进行加密，这层加密作用仅仅是混淆流量特征，而不是保护数据安全。它不保证数据完整性和身份认证，因此可能遭受中间人攻击和重放攻击。如果CDN不可信，或者遭受了基于HTTPS劫持的中间人攻击，应启用双重TLS保证数据传输安全。

如果使用了双重TLS，握手造成的延迟可能略有增加，但是只要开启```session_reuse```，```session_ticket```复用TLS连接，以及开启```mux```启用TLS多路复用，只会刚刚开启Trojan-Go时察觉明显的延迟。

协议栈如下：

|协议| 
|-|
|真实流量|
|simplesocks(如果使用mux)|
|smux(如果使用mux)|
|trojan|
|TLS(如果开启双重TLS)|
|混淆层(如果开启混淆)|
|Websocket|
|TLS|
|TCP|