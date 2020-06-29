---
title: "透明代理"
draft: false
weight: 11
---

### 注意，Trojan不完全支持这个特性（UDP）

Trojan-Go支持基于tproxy的透明TCP/UDP代理。

要开启透明代理模式，将一份正确的客户端配置（配置方式参见基本配置部分）其中的```run_type```修改为```nat```，并按照需求修改本地监听端口即可。

之后需要添加iptables规则。这里假定你的网关具有两个网卡，下面这份配置将其中一个网卡（局域网）的入站包转交给Trojan-Go，由Trojan-Go通过隧道，通过另一个网卡（互联网）发送到远端Trojan-Go服务器。你需要将下面的```$SERVER_IP```，```$TROJAN_GO_PORT```，```$INTERFACE```替换为自己的配置。

```shell
# 新建TROJAN_GO链
iptables -t mangle -N TROJAN_GO

# 绕过Trojan-Go服务器地址
iptables -t mangle -A TROJAN_GO -d $SERVER_IP -j RETURN

# 绕过私有地址
iptables -t mangle -A TROJAN_GO -d 0.0.0.0/8 -j RETURN
iptables -t mangle -A TROJAN_GO -d 10.0.0.0/8 -j RETURN
iptables -t mangle -A TROJAN_GO -d 127.0.0.0/8 -j RETURN
iptables -t mangle -A TROJAN_GO -d 169.254.0.0/16 -j RETURN
iptables -t mangle -A TROJAN_GO -d 172.16.0.0/12 -j RETURN
iptables -t mangle -A TROJAN_GO -d 192.168.0.0/16 -j RETURN
iptables -t mangle -A TROJAN_GO -d 224.0.0.0/4 -j RETURN
iptables -t mangle -A TROJAN_GO -d 240.0.0.0/4 -j RETURN

# 未命中上文的规则的包，打上标记
iptables -t mangle -A TROJAN_GO -j TPROXY -p tcp --on-port $TROJAN_GO_PORT --tproxy-mark 0x01/0x01
iptables -t mangle -A TROJAN_GO -j TPROXY -p udp --on-port $TROJAN_GO_PORT --tproxy-mark 0x01/0x01

# 从$INTERFACE网卡流入的所有TCP/UDP包，跳转TROJAN_GO链
iptables -t mangle -A PREROUTING -p tcp -i $INTERFACE -j TROJAN_GO
iptables -t mangle -A PREROUTING -p udp -i $INTERFACE -j TROJAN_GO

# 添加路由，打上标记的包重新进入本地回环
ip route add local default dev lo table 100
ip rule add fwmark 1 lookup 100
```

配置完成后以root权限启动Trojan-Go客户端：

```shell
sudo trojan-go
```
