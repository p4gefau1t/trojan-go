---
title: "Trojan协议"
draft: false
weight: 2
---

Trojan-Go遵循原始的Trojan协议，具体格式可以参考[Trojan文档](https://trojan-gfw.github.io/trojan/protocol)，这里不再赘述。

默认情况下，Trojan协议使用TLS作为承载，协议栈如下：

|协议| 
|-|
|真实流量|
|Trojan|
|TLS|
|TCP|
