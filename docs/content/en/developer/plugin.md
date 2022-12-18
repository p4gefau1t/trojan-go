---
title: "Pluggable transport layer plug-in development"
draft: false
weight: 150
---

Trojan-Go encourages the development of transport layer plugins to enrich protocol types and increase the strategic depth of the fight against GFW.

The role of transport layer plugins is to replace TLS in tansport tunnels for transport encryption and obfuscation.

The plug-in communicates with Trojan-Go based on TCP Socket, there is no coupling with Trojan-Go itself, and you can use any language and design pattern you like for development. We recommend developing with reference to the [SIP003](https://shadowsocks.org/en/spec/Plugin.html) standard. The plugins so developed can be used for both Trojan-Go and Shadowsocks.

Trojan-Go uses only TCP for transmission (plaintext) when the plug-in functionality is enabled. Your plugin only needs to handle inbound TCP requests. You can convert this TCP traffic into any traffic format you like, such as QUIC, HTTP, or even ICMP.

Trojan-Go plugin design principles, slightly different from Shadowsocks.

1. the plugin itself can encrypt, obfuscate and integrity-check the transmitted content, as well as being resistant to replay attacks.

2. the plug-in should forge an existing, common service (noted as X service) and its traffic, and on top of that embed its own encrypted content.

3. The server-side plug-in, upon verifying that the content has been tampered with/replayed, **must hand over the connection to Trojan-Go for processing**. The specific steps are to send the read-in and unread-in content together to Trojan-Go and to establish a two-way connection instead of disconnecting it directly.Trojan-Go will establish a connection to a real X server, allowing the attacker to interact directly with the real X server.

The explanation is as follows.

The first principle, is due to the fact that the Trojan protocol itself is not encrypted. Replacing TLS with a transport layer plug-in will **completely trust the security of the plug-in**.

The second principle, is inherited from the spirit of Trojan. The best place to hide a tree is a forest.

The third principle, to take full advantage of Trojan-Go's anti-active-detection feature. Even if GFW is actively probing your server, your server can behave consistent with the X service and no other features.

To make it easier to understand, let's take an example.

1. Suppose your plugin is masquerading as MySQL traffic. The firewall sniffs through the traffic and finds that your MySQL traffic is unusually large and decides to actively connect to your server for active probing.

2. The firewall connects to your server and sends a probe load to your Trojan-Go server-side plugin, which, after verification, finds that this abnormal connection is not proxy traffic and hands the connection over to Trojan-Go for processing.

3. Trojan-Go finds this connection anomaly and redirects this connection to a real MySQL server. The firewall then starts interacting with a real MySQL server and finds that its behavior is no different from that of a real MySQL server and cannot block the server.

Also, even if your protocol protocols and plug-ins do not satisfy principles 2 and 3, or even principle 1 very well, we encourage development as well. Because GFW only audits and blocks popular protocols, such protocols (earthen cryptography/earth protocols) can also remain very robust as long as they are not publicly published.
