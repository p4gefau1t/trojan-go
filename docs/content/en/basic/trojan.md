---
title: "Trojan fundamentals"
draft: false
weight: 21
---

This page will briefly describe the basics of how the Trojan protocol works. If you are not interested in how GFW and Trojan work, you can skip this section. However, for better security of your communication and node concealment, I recommend you read it.

## Why Shadowsocks (with streaming passwords) is vulnerable to blocking

Firewalls in the early days simply intercepted and censored outbound traffic, i.e. **passive detection**. Shadowsocks' encryption protocol was designed so that the transmitted packets themselves had almost no signature and appeared similar to a completely random stream of bits, which did work to bypass GFWs in the early days.

Current GFWs have begun to use **active detection**. Specifically, when GFW finds a suspicious unrecognizable connection (high traffic, random byte streams, high bit ports, etc.), it will **actively connect** to this server port and replay the previously captured traffic (or replay it with some elaborate modifications.) The Shadowsocks server detects an abnormal connection and disconnects it. This abnormal traffic and disconnection is seen as a characteristic of a suspicious Shadowsocks server, and the server is added to GFW's suspicious list. This list does not necessarily take effect immediately, but during certain special sensitive periods, the servers in the suspicious list are blocked temporarily or permanently. Whether or not this suspicious list is blocked may be determined by human factors.

If you want to know more, you can refer to [this article](https://gfw.report/blog/gfw_shadowsocks/).

## How Trojan bypasses GFW

In contrast to Shadowsocks, Trojan does not use custom encryption protocols to hide itself. Instead, it uses the well-characterized TLS protocol (TLS/SSL), which makes the traffic look the same as a normal HTTPS site. TLS is a well-established encryption system, and HTTPS means that it uses TLS to carry HTTP traffic. Using **correctly configured** encrypted TLS tunnels ensures that the transmission is

- Confidentiality (GFW cannot know what is being transmitted)

- Integrity (if the GFW tries to tamper with the transmitted ciphertext, both sides of the communication will find out)

- Non-repudiation (GFWs cannot forge their identities to impersonate the server or client)

- Forward security (GFWs cannot decrypt previously encrypted traffic even if the key is compromised)

For passive detection, Trojan protocol traffic has exactly the same characteristics and behavior as HTTPS traffic. While HTTPS traffic accounts for more than half of the current Internet traffic, and the traffic is ciphertext after a successful TLS handshake, there is almost no feasible way to distinguish Trojan protocol traffic from it.

For active detection, Trojan can correctly identify non-Trojan protocol traffic when the firewall actively connects to the Trojan server for detection. Unlike proxies such as Shadowsocks, Trojan does not disconnect at this point, but proxies this connection to a normal web server. In GFW's view, the server behaves exactly the same as a normal HTTPS website and it is impossible to tell if it is a Trojan proxy node. This is the reason why Trojan recommends using a legitimate domain name with an HTTPS certificate signed by an authoritative CA: this makes your server completely invisible to GFW using active detection to determine if it is a Trojan server.

Therefore, in the current situation, the only way to identify and block Trojan connections is to use indiscriminate blocking (blocking a certain IP segment, a certain type of certificate, a certain type of domain name, or even blocking all outbound HTTPS connections across the country) or to launch a massive man-in-the-middle attack (hijacking all TLS traffic and hijacking the certificate to censor the content). For man-in-the-middle attacks, you can use Websocket's dual TLS countermeasures, which are explained in detail in the advanced configuration.
