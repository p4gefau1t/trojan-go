---
title: "正确配置Trojan-Go"
draft: false
weight: 22
---

下面将介绍如何正确配置Trojan-Go以完全隐藏你的代理节点特征

在开始之前，你需要

- 一个服务器，且未被GFW封锁

- 一个域名，可以使用免费的域名服务，如.tk等

- Trojan-Go，可以从release页面下载

### 配置证书

为了伪装成一个正常的HTTPS站点，也为了保证传输的安全，我们需要一份经过权威证书机构签名的证书。Trojan-Go支持从Let's Encrypt自动申请证书。首先将你的域名正确解析到你的服务器IP。然后准备好一个邮箱地址，合乎邮箱地址规则即可，不需要真实邮箱地址。保证你的服务器443和80端口没有被其他程序（nginx，apache，正在运行的Trojan等）占用。然后执行

```shell
sudo ./trojan-go -autocert request
```

按照屏幕提示填入相关信息。如果操作成功，当前目录下将得到四个文件

- server.key 服务器私钥

- server.crt 经过Let's Encrypt签名的服务器证书

- user.key 用户Email对应的私钥

- domain_info.json 域名和用户Email信息

备份好这些文件，不要将.key文件分享给其他任何人，否则你的身份可能被冒用。

证书的有效期通常是三个月，你可以使用

```shell
sudo ./trojan-go -autocert renew
```

进行证书更新。更新之前请确保同目录下有上述的四个文件。如果你没有指定ACME challenge使用的端口，Trojan-Go将默认使用443和80端口，请确保这两个端口没有被Trojan-Go或者其他程序（nginx, caddy等等）占用。

### 服务端配置

我们的目标是，使得你的服务器和正常的HTTPS网站表现相同。

首先你需要一个HTTP服务器，可以使用nginx，apache，caddy等配置一个本地HTTP服务器，也可以使用别人的HTTP服务器。HTTP服务器的作用是，当GFW主动探测时，向它展示一个完全正常的Web页面。

**你需要在```remote_addr```和```remote_port```指定这个HTTP服务器的地址。```remote_addr```可以是IP或者域名。Trojan-Go将会测试这个HTTP服务器是否工作正常，如果不正常，Trojan-Go会拒绝启动。**

下面是一份比较安全的服务器配置，需要你在本地80端口配置一个HTTP服务（必要，你也可以使用其他的网站HTTP服务器，如"remote_addr": "example.com"），在1234端口配置一个HTTPS服务（可选，可以删除```fallback_port```字段，跳过这个步骤）

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
        "fallback_port": 1234
    }
}
```

这个配置文件使Trojan-Go在服务器的所有网卡上(0.0.0.0)监听443端口，使用server.crt和server.key作为证书和密钥进行TLS握手。你应该使用尽可能复杂的密码，同时确保客户端和服务端```password```是一致的。

如果TLS连接建立后，检测到TLS的内容非法，将TLS连接代理到本地127.0.0.1:80上的HTTP服务，这时远端看起来，Trojan-Go就是一个HTTPS的网站。

如果TLS握手失败了，说明对方使用的不是TLS协议进行主动探测，此时Trojan-Go将连接代理到本地127.0.0.1:1234上运行的HTTPS服务，本地HTTPS服务器也会检测到连接不是TLS连接，返回一个400 Bad Reqeust的HTTP页面。```fallback_port```是一个可选选项，如果没有填写，Trojan-Go会直接终止连接。虽然是可选的，但是还是强烈建议填写。

如果TLS连接建立，并且确认是Trojan协议，而且密码正确，那么服务器将解析来自客户端的请求并进行代理。

你可以通过使用浏览器访问你的域名```https://your_domain_name```来验证。如果工作正常，你的浏览器会显示一个正常的HTTPS保护的Web页面，页面内容与服务器本机80端口上的页面一致。你还可以使用```http://your_domain_name:443```验证```fallback_port```工作是否正常。

事实上，你甚至可以将Trojan-Go当作你的HTTPS服务器，用来给你的网站提供HTTPS服务。访客可以正常地通过Trojan-Go浏览你的网站，而和代理流量互不影响。

### 客户端配置

对应的客户端配置

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
        "fingerprint": "firefox",
        "sni": "your_domain_name"
    }
}
```

这个客户端配置使Trojan-Go开启一个监听在本地1080端口的socks5/http代理（自动识别），远端服务器为your_awesome_server:443，your_awesome_server可以是IP或者域名。

如果你在```remote_addr```中填写的是域名，```sni```可以省略。```sni```字段应当填写你申请证书的对应域名，或者你自己签发证书时证书的Common Name，而且必须一致。注意，```sni```字段目前的在TLS协议中是规定**明文传送**的(目的是使服务器提供相应证书)，所以不要填写类似google.com等已经被封锁的域名，否则很有可能导致你的服务器也被封锁。

```fingerprint```将设置Trojan-Go伪造Firefox浏览器的TLS请求指纹，使得Trojan-Go的流量混杂在正常的HTTPS流量中无法被识别。还可以设置为```ios```，```chrome```等。

更多关于配置文件的信息，可以在左侧导航栏中找到相应介绍。
