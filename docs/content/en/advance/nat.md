---
title: "Transparent Proxy"
draft: false
weight: 11
---

### Note that Trojan does not fully support this feature (UDP)

Trojan-Go supports transparent TCP/UDP proxies based on tproxy.

To enable transparent proxy mode, change ```run_type``` to ```nat``` in a proper client configuration (see the basic configuration section for how to configure it) and modify the local listening port as required.

After that you need to add iptables rules. Assuming that your gateway has two NICs, this configuration below forwards inbound packets from one of the NICs (LAN) to Trojan-Go, which sends them through a tunnel to the remote Trojan-Go server via the other NIC (Internet). You need to replace the following ```$SERVER_IP```, ```$TROJAN_GO_PORT```, ```$INTERFACE``` with your own configuration.

```shell
# New TROJAN_GO chain
iptables -t mangle -N TROJAN_GO

# Bypass Trojan-Go server address
iptables -t mangle -A TROJAN_GO -d $SERVER_IP -j RETURN

# Bypass private addresses
iptables -t mangle -A TROJAN_GO -d 0.0.0.0/8 -j RETURN
iptables -t mangle -A TROJAN_GO -d 10.0.0.0/8 -j RETURN
iptables -t mangle -A TROJAN_GO -d 127.0.0.0/8 -j RETURN
iptables -t mangle -A TROJAN_GO -d 169.254.0.0/16 -j RETURN
iptables -t mangle -A TROJAN_GO -d 172.16.0.0/12 -j RETURN
iptables -t mangle -A TROJAN_GO -d 192.168.0.0/16 -j RETURN
iptables -t mangle -A TROJAN_GO -d 224.0.0.0/4 -j RETURN
iptables -t mangle -A TROJAN_GO -d 240.0.0.0/4 -j RETURN

# Packets that do not hit the rule above, mark them
iptables -t mangle -A TROJAN_GO -j TPROXY -p tcp --on-port $TROJAN_GO_PORT --tproxy-mark 0x01/0x01
iptables -t mangle -A TROJAN_GO -j TPROXY -p udp --on-port $TROJAN_GO_PORT --tproxy-mark 0x01/0x01

# All TCP/UDP packets flowing from $INTERFACE NIC, jump TROJAN_GO chain
iptables -t mangle -A PREROUTING -p tcp -i $INTERFACE -j TROJAN_GO
iptables -t mangle -A PREROUTING -p udp -i $INTERFACE -j TROJAN_GO

# Add routes to re-enter the local loopback with marked packets
ip route add local default dev lo table 100
ip rule add fwmark 1 lookup 100
```

After configuration is complete **start with root privileges** Trojan-Go client.

```shell
sudo trojan-go
```
