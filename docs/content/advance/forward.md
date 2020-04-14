---
title: "流量中继"
draft: false
---

Trojan-Go支持进行流量转发。一个典型的使用场景是，你所使用的ISP提供的网络服务，出境的线路质量并不理想。这时你可以使用国内的一些线路更好的服务器，作为中继，将你的流量转发给trojan服务器。

Forward中继的配置很简单，下面是一个Forward的配置

```
{
    "run_type": "forward",
    "local_addr": "0.0.0.0",
    "local_port": 1234,
    "remote_addr": "your_trojan_server",
    "remote_port": 443,
}

```

Forwad启动后，客户端连接该主机的1234端口，和直接连接Trojan服务器443端口是等效的。为了保证安全性和稳定性，Forward只做简单的流量转发，本地不需要任何的证书文件和密钥文件。

你可以使用多个Forward链接起来作为多重跳板，如果你忍受这种做法带来的延迟升高和吞吐量下降的话。