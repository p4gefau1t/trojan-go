---
title: "SimpleSocks Protocol"
draft: false
weight: 50
---

SimpleSocks protocol is a simple proxy protocol with no authentication mechanism, essentially Trojan protocol with sha224 removed. The purpose of using this protocol is to reduce the overhead when multiplexing.

Only when multiplexing is enabled, the connections being multiplexed will use this protocol. That is, SimpleSocks is always carried by SMux.

SimpleSocks is even simpler than Socks5, here is the header structure.

```text
+-----+------+----------+----------+-----------+
| CMD | ATYP | DST.ADDR | DST.PORT |  Payload  |
+-----+------+----------+----------+-----------+
|  1  |  1   | Variable |    2     |  Variable |
+-----+------+----------+----------+-----------+
```

The definitions of the fields are the same as for the Trojan protocol and will not be repeated.
