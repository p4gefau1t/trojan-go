---
title: "Trojan Protocol"
draft: false
weight: 20
---

Trojan-Go follows the original trojan protocol, the exact format of which can be found in the [Trojan documentation](https://trojan-gfw.github.io/trojan/protocol) and will not be repeated here.

By default, the trojan protocol is carried using TLS, and the protocol stack is as follows.

| Protocol     |
| ------------ |
| Real Traffic |
| Trojan       |
| TLS          |
| TCP          |
