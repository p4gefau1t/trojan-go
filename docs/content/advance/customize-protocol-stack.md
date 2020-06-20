---
title: "自定义协议栈"
draft: false
weight: 8
---

### 注意，Trojan-GFW版本不支持这个特性

Trojan-Go允许高级用户自定义协议栈。在自定义模式下，Trojan-Go将放弃对协议栈的控制，允许用户操作底层协议栈组合。例如

- 在一层TLS上再建立一层或更多层TLS加密

- 使用TLS传输Websocket流量，在Websocket层上再建立一层TLS，在第二层TLS上再使用AEAD进行传输

- 在TCP连接上，使用Shadowsocks的AEAD加密传输Trojan协议

- 将一个入站Trojan的TLS流量解包后重新用TLS包装为新的出站Trojan流量

等等。

**如果你不了解网络相关知识，请不要尝试使用这个功能。不正确的配置可能导致Trojan-Go无法正常工作，或是导致性能和安全性方面的问题。**

Trojan-Go将所有协议抽象为隧道，每个隧道可能提供客户端，负责发送；也可能提供服务端，负责接受；或者两者皆提供。自定义协议栈即自定义隧道的堆叠方式。

### 在继续配置之前，请先阅读开发指南中“基本介绍”一节，确保已经理解Trojan-Go运作方式

下面是Trojan-Go支持的隧道和他们的属性:

|隧道                | 需要下层提供流 | 需要下层提供包 | 向上层提供流 | 向上层提供包 | 可以作为入站 | 可以作为出站 |
|-------------------|--------------|-------------|-------------|------------|------------|------------|
|transport          |      n       |      n      |     y       |     y      |     y      |     y      |
|dokodemo           |      n       |      n      |     y       |     y      |     y      |     n      |
|tproxy             |      n       |      n      |     y       |     y      |     y      |     n      |
|tls                |      y       |      n      |     y       |     n      |     y      |     y      |
|trojan             |      y       |      n      |     y       |     y      |     y      |     y      |
|mux                |      y       |      n      |     y       |     n      |     y      |     y      |
|simplesocks        |      y       |      n      |     y       |     y      |     y      |     y      |
|shadowsocks        |      y       |      n      |     y       |     n      |     y      |     y      |
|websocket          |      y       |      n      |     y       |     n      |     y      |     y      |
|freedom            |      n       |      n      |     y       |     y      |     n      |     y      |
|socks              |      y       |      n      |     y       |     y      |     y      |     n      |
|http               |      y       |      n      |     y       |     n      |     y      |     n      |
|router             |      y       |      y      |     y       |     y      |     n      |     y      |
|adapter            |      y       |      n      |     y       |     n      |     y      |     n      |

自定义协议栈的工作方式是，定义树/链上节点并分别它们起名（tag）并添加配置，然后使用tag组成的有向路径，描述这棵树/链。例如，对于一个典型的Trojan-Go服务器，可以如此描述：

入站，一共两条路径，tls节点将自动识别trojan和websocket流量并进行分发

- transport->tls->trojan

- transport->tls->websocket->trojan

出站，只能有一条路径

- freedom

对于入站，你可以描述一条链，一棵树或者一个有向无环图(DAG)，对于出站，你只能描述一个链。**注意，每条路径，必须以不需要下层提供流或包的隧道开始(transport/tproxy/dokodemo等)，必须以能向上层提供包和流的隧道终止(trojan/simplesocks/freedom)，且必须确认对应隧道是否可作为出站/入站**

要启用自定义协议栈，将```run_type```指定为custom，此时除```inbound```和```outbound```之外的其他选项将被忽略。

下面是一个例子，你可以在此基础上插入或减少协议节点。配置文件为简明起见，使用YAML进行配置，你也可以使用JSON来配置，除格式不同之外，效果是等价的。

客户端 client.yaml

```yaml
run-type: custom

inbound:
  node:
    - protocol: transport
      tag: transport
      config:
        local-addr: 127.0.0.1
        local-port: 1080
    - protocol: socks
      tag: socks
  path:
    -
      - transport
      - socks

outbound:
  node:
    - protocol: transport
      tag: transport
      config:
        remote-addr: you_server
        remote-port: 443

    - protocol: tls
      tag: tls
      config:
        ssl:
          sni: localhost
          key: server.key
          cert: server.crt

    - protocol: trojan
      tag: trojan
      config:
        password:
          - 12345678

  path:
    - 
      - transport
      - tls
      - trojan

```

服务端 server.yaml

```yaml
run-type: custom

inbound:
  node:
    - protocol: websocket
      tag: websocket
      config:
        websocket:
            enabled: true
            hostname: example.com
            path: /ws

    - protocol: transport
      tag: transport
      config:
        local-addr: 0.0.0.0
        local-port: 443
        remote-addr: 127.0.0.1
        remote-port: 80

    - protocol: tls
      tag: tls
      config:
        ssl:
          sni: localhost
          key: server.key
          cert: server.crt

    - protocol: trojan
      tag: trojan1
      config:
        password:
          - 12345678

    - protocol: trojan
      tag: trojan2
      config:
        password:
          - 87654321

  path:
    - 
      - transport
      - tls
      - trojan1
    -
      - transport
      - tls
      - websocket
      - trojan2

outbound:
  node:
    - protocol: freedom
      tag: freedom

  path:
    - 
      - freedom

```
