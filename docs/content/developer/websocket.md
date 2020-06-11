---
title: "Websocket"
draft: false
weight: 40
---

由于使用CDN中转时，HTTPS对CDN透明，CDN可以审查Websocket传输内容。而Trojan协议本身是明文传输，因此为保证安全性，可添加一层Shadowsocks AEAD加密层以混淆流量特征并保证安全性。

**如果你使用的是中国境内运营商提供的CDN，请务必开启AEAD加密**

开启AEAD加密后，Websocket承载的流量将被Shadowsocks AEAD加密，头部具体格式参见Shadowsocks白皮书。

开启Websocket支持后，协议栈如下：

|协议              |备注       |
|-----------------|----------|
|真实流量           |         |
|SimpleSocks      |如果开启多路复用|
|smux             |如果开启多路复用|
|Trojan           |          |
|Shadowsocks      |如果开启加密|
|Websocket        |          |
|传输层协议         |          |
