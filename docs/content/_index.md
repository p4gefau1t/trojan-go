---
title: "简介"
draft: false
weight: 10
---

# Trojan-Go

这里是Trojan-Go的文档，你可以在左侧的导航栏中找到一些使用技巧，以及完整的配置文件说明。

Trojan-Go是使用Golang实现的完整的Trojan代理，和Trojan协议以及原版的配置文件格式兼容。

Trojan-Go的的首要目标是保障传输安全性和隐蔽性。在此前提下，尽可能提升传输性能和易用性。

Trojan-Go支持并且兼容原版Trojan的绝大多数功能，包括但不限于：

- TLS/SSL隧道传输

- 透明代理 (NAT模式，iptables设置参见[这里](https://github.com/shadowsocks/shadowsocks-libev/tree/v3.3.1#transparent-proxy))

- UDP代理

- 对抗GFW被动/主动检测的机制

- MySQL数据库支持

- 流量统计，用户流量配额限制

- 从数据库中的用户列表进行认证

- TCP性能方面的选项，如TCP Fast Open，端口复用等

同时，Trojan-Go还有更多高效易用的功能特性：

- 简易模式，快速部署使用

- Socks5/HTTP代理自动适配

- 多平台和多操作系统支持，无特殊依赖

- 多路复用，显著提升并发性能

- 自定义路由模块，可实现国内直连/广告屏蔽等功能

- Websocket，用于支持CDN流量中转(基于WebSocket over TLS/SSL)和对抗GFW中间人攻击

- 自动化HTTPS证书申请，使用ACME协议从Let's Encrypt自动申请和更新HTTPS证书

如果你遇到配置方面的问题，或是遇到了软件Bug，或是有更好的想法，欢迎加入[Telegram交流反馈群](https://t.me/trojan_go_chat)



----

> Across the Great Wall, we can reach every corner in the world.
>
> (越过长城，走向世界。)