---
title: "Basic introduction"
draft: false
weight: 1
---

The core parts of Trojan-Go are

- tunnel Specific implementation of each protocol

- proxy proxy core

- config configuration registration and parsing module

- redirector Active spoof detection module

- statistics user authentication and statistics module

The source code can be found in the corresponding folder.

## tunnel.Tunnel tunnel

Trojan-Go abstracts all protocols (including routing functions, etc.) into tunnels (tunnel.Tunnel interface), each of which can be opened with a server (tunnel.Server interface) and a client (tunnel.Client). Each server can, from its underlying tunnel, strip and accept streams (tunnel.Conn) and packets (tunnel.PacketConn). Clients can create streams and packets to the underlying tunnel.

Each tunnel does not care what the tunnels below it are, but each tunnel knows exactly what information is relevant for the other tunnels above this it.

All tunnels require either stream or packet transport support from the lower layer, or both. All tunnels must provide stream transport support to the upper tunnels, but not necessarily packet transport.

Tunnels may be server-only, client-only, or both. Tunnels that have both may be used as transport tunnels between Trojan-Go clients and servers.

Note that a distinction is made between a server/client for Trojan-Go, and a server/client for a tunnel. Here is an example of a diagram that is easy to understand.

```text

Inbound                              GFW                              Outbound
-------> Tun.A server->Tun.B client -----> Tun.B server->Tun.C client ------->
            (Trojan-Go client)                  (Trojan-Go server)

````

The bottom tunnel is the transport layer, i.e. the tunnel that does not get or create streams and packets from other tunnels, acting as tunnel A or C in the above diagram.

- transport, the pluggable transport layer

- socks, socks5 proxy, only the tunnel server

- tproxy, transparent proxy, tunnel server only

- dokodemo, reverse proxy, tunnel server only

- freedom, free outbound, tunnel client only

These tunnels create streams and packets directly from TCP/UDP sockets and do not accept any tunnels added for their underlying layers.

The other tunnels, in principle, can be combined and stacked in any way and in any number, as long as the lower layer can satisfy the packet and stream transport needs of the upper layer. These tunnels act as Tunnel B in the above diagram, and they have

- trojan

- websocket

- mux

- simplesocks

- tls

- router, routing function, tunneling client only

None of them care about their lower layer tunneling implementation. However, they can be distributed to the upper layer tunnels based on the incoming flows and packets.

For example, in this diagram, which is a typical Trojan-Go client and server, the individual tunnels are stacked from bottom to top in the following order

- Tunnel A: transport->socks

- Tunnel B: transport->tls->trojan

- Tunnel C: freedom

The actual tunnel stacking situation is a bit more complicated than this. A typical inbound tunnel is in the form of a multinomial tree, not a chain. See below for a detailed explanation.

## proxy.Proxy core

The role of the proxy core is to listen to the stack formed by combining and stacking the above tunnels, and to forward all the streams and packets extracted from the inbound stack (end nodes of multiple tunnel servers, see below), along with the corresponding meta information, to the outbound stack (a tunnel Client).

Note that there can be multiple inbound protocol stacks, e.g., the client can extract streams and packets from both Socks5 and HTTP protocol stacks, the server can extract streams and packets from both the Websocket-hosted Trojan protocol and the TLS-hosted Trojan protocol, etc. However, there can be only one outbound protocol stack, e.g. only the TLS-bearing Trojan protocol is used for outbound.

To describe how the inbound protocol stacks (tunnel servers) are combined and stacked, a multinomial tree is used to describe all the protocol stacks. You can see the process of building the tree in the proxy folder, for each component.

The outbound protocol stack, on the other hand, is simpler and can be described using a simple list.

So in effect, for a typical client/server with Websocket and Mux turned on, the tunnel stacking model in the above figure is

Client

- inbound (tree)
  - transport (root)
    - adapter Able to recognize HTTP and Socks traffic and distribute to upper layer protocols
      - http (endpoint node)
      - socks (end node)

- outbound (chain)
  - transport (root)
  - tls
  - websocket
  - trojan
  - mux
  - simplesocks

server-side

- inbound (tree)
  - transport (root)
    - tls can identify HTTP and non-HTTP traffic and distribute
      - websocket
        - trojan (endpoint node)
          - mux
            - simplesocks (endpoint node)
      - trojan can identify and distribute mux and normal trojan traffic (end node)
        - mux
          - simplesocks (end node)

- Outbound (chain)
  - freedom

Note that the proxy core only extracts flows and packets from the end nodes of the tree formed by the tunnel and forwards them to a unique outbound station. The purpose of the multiple terminal nodes design is to make Trojan-Go compatible with both Websocket and Trojan protocol inbound connections, inbound connections with/without Mux enabled, and HTTP/Socks5 auto-identification. Each node in the tree with multiple sons has the ability to precisely identify and distribute flows and packets to different son nodes. This is consistent with our assumption that each protocol understands its upper layer bearer protocols.
