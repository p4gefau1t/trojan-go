---
title: "Multiplexing"
draft: false
weight: 30
---

Trojan-Go uses [smux](https://github.com/xtaci/smux) to implement multiplexing. The simplesocks protocol is also implemented for proxy transfers.

When multiplexing is enabled, the client first initiates a TLS connection, using the normal trojan protocol format, but filling in 0x7f (protocol.Mux) in the Command section of the protocol, identifying the connection as a multiplexed connection (similar to the upgrade of http), after which the connection is handed over to the smux client for management. After the server receives the request header, it is handed over to the smux server to parse all traffic for the connection. On each separated smux connection, the proxy destination is marked using the simplesocks protocol (the trojan protocol with the authentication part removed). The top-down protocol stack is as follows.

| protocol             | notes              |
| -------------------- | ------------------ |
| Real Traffic         |
| SimpleSocks          |
| smux                 |
| Trojan               | for Authentication |
| Underlying protocols |                    |
