---
title: "Custom Protocol Stack"
draft: false
weight: 8
---

{{% panel status="caution" title="Compatibility" %}}
Note that Trojan does not support this feature
{{% /panel %}}

Trojan-Go allows advanced users to customize the protocol stack. In custom mode, Trojan-Go will relinquish control of the protocol stack and allow users to manipulate the underlying protocol stack combinations. For example:

- Creating one or more layers of TLS encryption on top of one layer of TLS

- Use TLS to transport Websocket traffic, build another layer of TLS on top of the Websocket layer, and then use Shadowsocks AEAD on top of the second layer of TLS for encrypted transport

- Encrypted transport of Trojan protocol using Shadowsocks AEAD on a TCP connection

- Unwrap an inbound Trojan TLS traffic and repackage it with TLS as a new outbound Trojan traffic

And so on.


{{% panel status="warning" title="Caution" %}}
Do not try to use this feature if you don't know anything about networking. Incorrect configuration may cause Trojan-Go to not work properly, or cause performance and security issues.
{{% /panel %}}

Trojan-Go abstracts all protocols into tunnels, each of which may provide a client, which is responsible for sending, a server, which is responsible for receiving, or both. The custom protocol stack is how the custom tunnels are stacked.

{{% panel status="notice" title="Prerequisite" %}}
Before proceeding with the configuration, please read the "Basic Introduction" section of the Developer's Guide to ensure that you understand how Trojan-Go works.
{{% /panel %}}


Here are the tunnels supported by Trojan-Go and their properties:


| Tunnel | Requires lower layer to provide streams | Requires lower layer to provide packages | Provides streams to upper layer | Provides packages to upper layer | Can be used as inbound | Can be used as outbound |
| ----------- | -------------- | -------------- | ------------ | ------------ | ------------ | ------------ |
| transport   | n              | n              | y            | y            | y            | y            |
| dokodemo    | n              | n              | y            | y            | y            | n            |
| tproxy      | n              | n              | y            | y            | y            | n            |
| tls         | y              | n              | y            | n            | y            | y            |
| trojan      | y              | n              | y            | y            | y            | y            |
| mux         | y              | n              | y            | n            | y            | y            |
| simplesocks | y              | n              | y            | y            | y            | y            |
| shadowsocks | y              | n              | y            | n            | y            | y            |
| websocket   | y              | n              | y            | n            | y            | y            |
| freedom     | n              | n              | y            | y            | n            | y            |
| socks       | y              | y              | y            | y            | y            | n            |
| http        | y              | n              | y            | n            | y            | n            |
| router      | y              | y              | y            | y            | n            | y            |
| adapter     | n              | n              | y            | y            | y            | n            |


A custom stack works by defining nodes in a tree/chain and naming them individually (tags) and adding configurations, and then describing the tree/chain using a directed path composed of the tags. For example, for a typical Trojan-Go server, it can be described as follows.

Inbound, there are two paths, and the tls node will automatically identify trojan and websocket traffic and distribute it

- transport->tls->trojan

- transport->tls->websocket->trojan

Outbound, there can only be one path

- router->freedom

For inbound, multiple paths are described starting from the root, forming a **multinomial tree** (which can also degenerate to a chain); graphs that do not satisfy the tree property will result in undefined behavior; for outbound, a **chain** must be described.

Each path must satisfy the condition that.

1. must start with a tunnel that **does not require lower layers to provide streams or packets** (transport/adapter/tproxy/dokodemo, etc.)

2. must terminate with a tunnel that **provides packets and streams** to the upper layer (trojan/simplesocks/freedom, etc.)

3. on outbound single chains, the tunnels must all be available as outbound. On all paths of the inbound, the tunnels must all be available as inbound.

To enable custom stacks, specify ```run_type``` as custom, where all options other than ```inbound``` and ```outbound``` will be ignored.

Here is an example of a protocol node that you can insert or reduce on top of this. The configuration file is configured using YAML for simplicity, you can also configure it using JSON, the effect is equivalent except for the difference in format.

Client client.yaml

```yaml
run-type: custom

inbound:
  node:
    - protocol: adapter
      tag: adapter
      config:
        local-addr: 127.0.0.1
        local-port: 1080
    - protocol: socks
      tag: socks
      config:
        local-addr: 127.0.0.1
        local-port: 1080
  path:
    path: -
      - adapter
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
          - password: 12345678

  path:
    -transport
      - transport
      - tls
      - trojan

```

server server.yaml

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
        remote-addr: 127.0.0.1
        remote-port: 80
        ssl:
          sni: localhost
          key: server.key
          cert: server.crt

    - protocol: trojan
      tag: trojan1
      config:
        remote-addr: 127.0.0.1
        remote-port: 80
        password:
          - 12345678

    - protocol: trojan
      tag: trojan2
      config:
        remote-addr: 127.0.0.1
        remote-port: 80
        password:
          - 87654321

  path:
    -
      - transport
      - tls
      - trojan1
    -transport
      - transport
      - tls
      - websocket
      - trojan2

outbound:
  node:
    - protocol: freedom
      tag: freedom

  path:
    -freedom
      - freedom
```
