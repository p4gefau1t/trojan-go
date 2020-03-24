# Trojan-Go

使用Golang实现的完整Trojan代理，和Trojan协议以及原版的配置文件格式兼容。安全，高效，轻巧，易用。

[English](#English)

## 使用方法

```
./trojan-go -config 你的配置文件.json
```

配置文件格式和Trojan相同, 可以参考Trojan[官方文档](https://trojan-gfw.github.io/trojan/config)。

Trojan-Go支持并且兼容原版Trojan的绝大多数功能，包括

- TLS隧道传输

- 透明代理 (NAT模式)

- UDP代理

- 对抗GFW被动/主动检测的机制

- MySQL数据库支持

- 流量统计，用户流量配额限制

- 从数据库中的用户列表进行认证

- TCP性能方面的选项，如TCP Fast Open，端口复用等等

注意， TLS 1.2密码学套件的名称在golang中有一些不同，并且不安全的TLS 1.2套件已经被弃用，直接使用原版配置文件会引发一个警告，但不影响运行。更多信息参见[Wiki](https://github.com/p4gefau1t/trojan-go/wiki/%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6)。

## 特性

### 移植性

运行Trojan-Go的可执行文件不依赖其他组件。你可以将编译得到的单个可执行文件在目标机器上直接执行而不需要考虑依赖的问题。你可以很方便地编译（或者交叉编译）它，然后在你的服务器，PC，树莓派，甚至路由器上部署。

### 易用

配置文件格式与原版是兼容的，但做了一些简化。未指定的字段会被附上一个初始值。你可以更方便地部署你的服务器和客户端。下面是一个例子，完整的配置文件参见[这里](https://github.com/p4gefau1t/trojan-go/wiki/%E9%85%8D%E7%BD%AE%E6%96%87%E4%BB%B6)。一个完整的配置教程参见[这里](https://github.com/p4gefau1t/trojan-go/wiki/%E5%A6%82%E4%BD%95%E4%BD%BF%E7%94%A8Trojan-Go%E9%9A%90%E8%97%8F%E4%BD%A0%E7%9A%84%E4%BB%A3%E7%90%86%E8%8A%82%E7%82%B9)。

服务器配置文件

server.json
```
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
        "cert": "your_cert.crt",
        "key": "your_key.key",
    }
}

```

客户端配置文件

client.json
```
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
        "sni": "your_awesome_domain_name"
    }
}
```

Trojan-Go支持的runtype包括（其实和原版是一样的）

- Client

- Server

- NAT (透明代理，参见[这里](https://github.com/shadowsocks/shadowsocks-libev/tree/v3.3.1#transparent-proxy))

- Forward

更多关于配置文件的信息，可以参考Trojan的关于配置文件的[文档](https://trojan-gfw.github.io/trojan/config) 。

### 多路复用

在很差的网络条件下，TLS握手可能会花费很多时间。
Trojan-Go支持多路复用([smux](https://github.com/xtaci/smux))。通过使一个TLS隧道连接承载多个TCP连接的方式，减少TLS握手带来的延迟，以期提升高并发情景下的性能。

启用多路复用并不会增加你测速得到的带宽，但是会加速你有大量并发请求时的网络体验，例如浏览含有大量图片的网页等。

注意，这个特性和原版Trojan**不兼容**，所以出于兼容性考虑，这个特性是默认关闭的。但是你可以通过设置tcp选项中的"mux"字段启用它。如下

```
"tcp": {
    "mux": true
}
```

举个例子，上面的客户端的配置文件client.json加上一个tcp选项

client-mux.json
```
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
        "cert": "server.crt",
        "sni": "your_awesome_domain_name"
    },
    "tcp": {
        "mux": true
    }
}
```
你只需要设置客户端的配置文件即可，服务端会自动检测是否启用多路复用并提供支持。

## 构建

确保你的Golang版本 >= 1.11

```
git clone https://github.com/p4gefau1t/trojan-go.git
cd trojan-go
go build
```

Golang支持通过设置环境变量进行交叉编译，例如

```
CGO_ENABLE=0 GOOS=windows GOARCH=amd64 go build -o trojan-go.exe
```
以及

```
CGO_ENABLE=0 GOOS=linux GOARCH=arm go build -o trojan-go
```

---------

<a name="English"></a>

# Trojan-Go

Full-featured Trojan proxy written in golang, compatiable with the original Trojan protocol and config file. It's safe, efficient, lightweight and easy to use.

## Usage

```
./trojan-go -config your_awesome_config_file.json
```

Trojan-Go supports most features of the original trojan, including

- TLS tunneling

- Transparent proxy (NAT mode)

- UDP Relaying

- Mechanism against passive and active detection of GFW

- MySQL Database support

- Traffic statistics, quota limits for each user

- Authentication by users record in database

- TCP performance-related options, like TCP fast open, port reusing, etc.

Note that the name of the TLS 1.2 cipher suite is slightly different in golang, and some of them has been deprecated and disabled. Using the original configuration file directly will cause a warning, but it will not affect the running. See wiki for more information.

The format of the configuration file is compatible, see [here](https://trojan-gfw.github.io/trojan/config).

## Features

### Portable

It's written in Golang, so it will be statically linked by default, which means that you can execute the compiled single executable directly on the target machine without having to consider dependencies. You can easily compile (or cross compile) it and deploy it on your server, PC, Raspberry Pi, or even a router.

### Easy to use

Trojan-go's configuration file format is compatible with Trojan's, while it's being simplyfied. Unspecified fields will be filled in with a default value. You can launch your server and client much easier. Here's an example:

server.json
```
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
        "cert": "your_cert.crt",
        "key": "your_key.key",
    }
}

```

client.json
```
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
        "sni": "your_awesome_domain_name"
    }
}
```

run_type supported by Trojan-Go (the same as Trojan):

- Client

- Server

- NAT (transparent proxy, see [here](https://github.com/shadowsocks/shadowsocks-libev/tree/v3.3.1#transparent-proxy))

- Forward

For more infomation, see Trojan's [docs](https://trojan-gfw.github.io/trojan/config) about the configuration file.

### Multiplexing

TLS handshaking may takes much time in a poor network condition.
Trojan-go supports multiplexing([smux](https://github.com/xtaci/smux)), which imporves the performance in the high-concurrency scenario by forcing one single TLS tunnel connection carries mutiple TCP connections.

Enabling multiplexing does not increase the bandwidth you get from a speed test, but it will speed up the network experience when you have a large number of concurrent requests, such as browsing web pages containing a large number of images, etc.

Note that this feature is not compatible with the original Trojan , so for compatibility reasons, this feature is turned off by default. But you can enable it by setting the "mux" field in the tcp options. as follows

```
"tcp": {
    "mux": true
}
```
for example

client.json
```
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
        "cert": "server.crt",
        "sni": "your_awesome_domain_name"
    },
    "tcp": {
        "mux": true
    }
}
```

You only need to set the client's configuration file, and the server will automatically detect whether to enable multiplexing.

## Build

Just make sure your golang version >= 1.11


```
git clone https://github.com/p4gefau1t/trojan-go.git
cd trojan-go
go build
```

You can cross-compile it by setting up the environment vars, for example
```
CGO_ENABLE=0 GOOS=windows GOARCH=amd64 go build -o trojan-go.exe
```

or

```
CGO_ENABLE=0 GOOS=linux GOARCH=arm go build -o trojan-go
```
