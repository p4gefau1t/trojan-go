---
title: "简介"
draft: false
weight: 10
---

# Trojan-Go

Trojan-Go是使用Golang实现的完整的Trojan代理，和Trojan协议以及原版的配置文件格式兼容。

Trojan-Go的开发将传输安全性和隐蔽性放在首位。在此前提下，尽可能提升传输性能和易用性。

Trojan-Go支持并且兼容原版Trojan的绝大多数功能，包括但不限于：

- TLS隧道传输

- 透明代理 (NAT模式，iptables设置参见[这里](https://github.com/shadowsocks/shadowsocks-libev/tree/v3.3.1#transparent-proxy))

- UDP代理

- 对抗GFW被动/主动检测的机制

- MySQL数据库支持

- 流量统计，用户流量配额限制

- 从数据库中的用户列表进行认证

- TCP性能方面的选项，如TCP Fast Open，端口复用等

同时，Trojan-Go还支持更多高效易用的功能：

- 多路复用，显著提升并发性能

- 自定义路由模块，可实现国内直连/广告屏蔽等功能

- Websocket，用于支持CDN流量中转(基于WebSocket over TLS/SSL)和对抗GFW中间人攻击

- 自动化HTTPS证书申请，使用ACME协议从Let's Encrypt自动申请和更新HTTPS证书

[Telegram交流反馈群](https://t.me/trojan_go_chat)
