---
title: "Websocket"
draft: false
weight: 40
---

Since HTTPS is transparent to the CDN when using CDN transit, the CDN can review the content of the Websocket transfer. The Trojan protocol itself is transmitted in clear text, so to ensure security, a layer of Shadowsocks AEAD encryption layer can be added to obfuscate traffic characteristics and ensure security.

**If you are using a CDN provided by an operator in China, please make sure to turn on AEAD encryption**

When AEAD encryption is enabled, traffic carried by Websocket will be encrypted by Shadowsocks AEAD, see Shadowsocks white paper for the specific format of the header.

After enabling Websocket support, the protocol stack is as follows.

| Protocol                 | Remarks                    |
| ------------------------ | -------------------------- |
| Real Traffic             |                            |
| SimpleSocks              | If multiplexing is enabled |
| smux                     | If multiplexing is enabled |
| Trojan                   |                            |
| Shadowsocks              | if encryption is enabled   |
| Websocket                |                            |
| Transport Layer Protocol |                            |
