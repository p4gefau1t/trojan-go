---
title: "Secondary encryption with Shadowsocks AEAD"
draft: false
weight: 8
---

{{% panel status="caution" title="Compatibility" %}}
Note that Trojan does not support this feature
{{% /panel %}}

The Trojan protocol itself is not encrypted and its security relies on the underlying TLS. in general, TLS security is good and there is no need to encrypt Trojan traffic again. However, there are some scenarios where you may not be able to guarantee the security of a TLS tunnel.

- You use a Websocket, relayed through an untrusted CDN (e.g. a domestic CDN)

- Your connection to the server is subject to a man-in-the-middle attack by GFW against TLS

- Your certificate is invalid and you cannot verify its validity

- You use a pluggable transport layer that cannot guarantee cryptographic security

etc.

Trojan-Go supports encryption of Trojan-Go using Shadowsocks AEAD. The essence of this is to add a layer of Shadowsocks AEAD encryption underneath the Trojan protocol. Both the server and client must be enabled and the password and encryption must be the same, otherwise communication will not be possible.

To turn on AEAD encryption, simply add a ```shadowsocks``` option to

```json
...
"shadowsocks": {
    "enabled": true,
    "method": "AES-128-GCM",
    "password": "1234567890"
}
```

If omitted, AES-128-GCM is used by default. for more information, see the section "Complete Configuration Files".
