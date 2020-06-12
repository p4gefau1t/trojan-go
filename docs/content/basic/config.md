---
title: "正确配置Trojan-Go"
draft: false
weight: 22
---

下面将介绍如何正确配置Trojan-Go以完全隐藏你的代理节点特征。

在开始之前，你需要

- 一个服务器，且未被GFW封锁

- 一个域名，可以使用免费的域名服务，如.tk等

- Trojan-Go，可以从release页面下载

- 证书密钥对，可以从letsencrpyt等机构免费申请签发

### 服务端配置

我们的目标是，使得你的服务器和正常的HTTPS网站表现相同。

首先你需要一个HTTP服务器，可以使用nginx，apache，caddy等配置一个本地HTTP服务器，也可以使用别人的HTTP服务器。HTTP服务器的作用是，当GFW主动探测时，向它展示一个完全正常的Web页面。

**你需要在```remote_addr```和```remote_port```指定这个HTTP服务器的地址。```remote_addr```可以是IP或者域名。Trojan-Go将会测试这个HTTP服务器是否工作正常，如果不正常，Trojan-Go会拒绝启动。**

下面是一份比较安全的服务器配置server.json，需要你在本地80端口配置一个HTTP服务（必要，你也可以使用其他的网站HTTP服务器，如"remote_addr": "example.com"），在1234端口配置一个HTTPS服务，或是一个展示"400 Bad Request"的静态HTTP网页服务。（可选，可以删除```fallback_port```字段，跳过这个步骤）

```json
{
    "run_type": "server",
    "local_addr": "0.0.0.0",
    "local_port": 443,
    "remote_addr": "127.0.0.1",
    "remote_port": 80,
    "password": [
        "your_awesome_password"
    ],
    "ssl": {
        "cert": "server.crt",
        "key": "server.key",
        "sni": "your-domain-name.com",
        "fallback_port": 1234
    }
}
```

这个配置文件使Trojan-Go在服务器的所有IP地址上(0.0.0.0)监听443端口，分别使用server.crt和server.key作为证书和密钥进行TLS握手。你应该使用尽可能复杂的密码，同时确保客户端和服务端```password```是一致的。注意，**Trojan-Go会检测你的HTTP服务器```http://remote_addr:remote_port```是否正常工作。如果你的HTTP服务器工作不正常，Trojan-Go将拒绝启动。**

当一个客户端试图连接Trojan-Go的监听端口时，会发生下面的事情：

- 如果TLS握手成功，检测到TLS的内容非Trojan协议（有可能是HTTP请求，或者来自GFW的主动探测）。Trojan-Go将TLS连接代理到本地127.0.0.1:80上的HTTP服务。这时在远端看来，Trojan-Go服务就是一个HTTPS网站。

- 如果TLS握手成功，并且被确认是Trojan协议头部，并且其中的密码正确，那么服务器将解析来自客户端的请求并进行代理，否则和上一步的处理方法相同。

- 如果TLS握手失败，说明对方使用的不是TLS协议进行连接。此时Trojan-Go将这个TCP连接代理到本地127.0.0.1:1234上运行的HTTPS服务，返回一个展示400 Bad Reqeust的HTTP页面。```fallback_port```是一个可选选项，如果没有填写，Trojan-Go会直接终止连接。虽然是可选的，但是还是强烈建议填写。

你可以通过使用浏览器访问你的域名```https://your-domain-name.com```来验证。如果工作正常，你的浏览器会显示一个正常的HTTPS保护的Web页面，页面内容与服务器本机80端口上的页面一致。你还可以使用```http://your-domain-name.com:443```验证```fallback_port```工作是否正常。

事实上，你甚至可以将Trojan-Go当作你的HTTPS服务器，用来给你的网站提供HTTPS服务。访客可以正常地通过Trojan-Go浏览你的网站，而和代理流量互不影响。但是注意，不要在```remote_port```和```fallback_port```搭建有高实时性需求的服务，Trojan-Go识别到非Trojan协议流量时会有意增加少许延迟以抵抗GFW基于时间的检测。

配置完成后，可以使用

```shell
./trojan-go -config ./server.json
```

启动服务端。

### 客户端配置

对应的客户端配置client.json

```json
{
    "run_type": "client",
    "local_addr": "127.0.0.1",
    "local_port": 1080,
    "remote_addr": "your_awesome_server",
    "remote_port": 443,
    "password": [
        "your_awesome_password"
    ],
    "ssl": {
        "sni": "your-domain-name.com"
    }
}
```

这个客户端配置使Trojan-Go开启一个监听在本地1080端口的socks5/http代理（自动识别），远端服务器为your_awesome_server:443，your_awesome_server可以是IP或者域名。

如果你在```remote_addr```中填写的是域名，```sni```可以省略。如果你在```remote_addr```填写的是IP地址，```sni```字段应当填写你申请证书的对应域名，或者你自己签发的证书的Common Name，而且必须一致。注意，```sni```字段目前的在TLS协议中是**明文传送**的(目的是使服务器提供相应证书)。GFW已经被证实具有SNI探测和阻断能力，所以不要填写类似```google.com```等已经被封锁的域名，否则很有可能导致你的服务器也被遭到封锁。

配置完成后，可以使用

```shell
./trojan-go -config ./client.json
```

启动客户端。

更多关于配置文件的信息，可以在左侧导航栏中找到相应介绍。
