---
title: "DNS隧道和反向代理"
draft: false
weight: 5
---

你可以使用Trojan-Go建立隧道。一个典型的应用是，使用Trojan-Go在本地建立一个纯净无污染的DNS服务器，下面是一个配置的例子

```
{
    "run_type": "forward",
    "local_addr": "127.0.0.1",
    "local_port": 53,
    "remote_addr": "your_awesome_server",
    "remote_port": 443,
    "target_addr": "8.8.8.8",
    "target_port": 53,
    "password": [
	    "your_awesome_password"
    ]
}
```

forward本质上是一个客户端，不过你需要填入```target_addr```和```target_port```字段，指明反向代理的目标。

使用这份配置文件后，本地53的TCP和UDP端口将被监听，所有的向本地53端口发送的TCP或者UDP数据，都会通过TLS隧道转发给远端服务器your_awesome_server，远端服务器得到回应后，数据会通过隧道返回到本地53端口。 也就是说，你可以将127.0.0.1当作一个DNS服务器，本地查询的结果和远端服务器查询的结果是一致的。你可以使用这个配置避开DNS污染。

同样的原理，你可以在本地搭建一个Google的镜像

```
{
    "run_type": "forward",
    "local_addr": "127.0.0.1",
    "local_port": 53,
    "remote_addr": "your_awesome_server",
    "remote_port": 443,
    "target_addr": "www.google.com",
    "target_port": 443,
    "password": [
	    "your_awesome_password"
    ]
}
```

访问https://127.0.0.1即可访问谷歌主页，但是注意这里由于谷歌服务器提供的https证书是google.com的证书，而当前域名为127.0.0.1，因此浏览器会引发一个证书错误的警告。