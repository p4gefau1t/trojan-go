# Trojan-Go

使用Golang编写的Trojan代理，和Trojan协议以及原版的配置文件格式兼容。

目前仍然处于开发中。

## 特性

### 兼容

与原版Trojan协议以及配置文件兼容，你可以使用Trojan-Go替换客户端和服务端其中任意一个，而无须做额外配置。

### 易用

配置文件格式与原版是兼容的，但做了一些简化。你可以更方便地部署你的服务器和客户端。下面是一个例子：

服务器配置文件

server.json
```
{
	"run_type": "server",
	"local_addr": "0.0.0.0",
	"local_port": 4445,
	"remote_addr": "127.0.0.1",
	"remote_port": 80,
	"password": [
		"your_awesome_password"
	],
	"ssl": {
		"cert": "your_cert.crt",
		"key": "your_key.crt",
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
        "cert": "your_cert.crt",
        "sni": "your_awesome_domain_name",
    }
}
```

Trojan-Go支持的runtype包括（其实和原版是一样的）

- Client

- Server

- NAT (透明代理，[参见这里](https://github.com/shadowsocks/shadowsocks-libev/tree/v3.3.1#transparent-proxy))

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
        "sni": "your_awesome_domain_name",
    },
    "tcp": {
        "mux": true
    }
}
```
你只需要设置客户端的配置文件即可，服务端会自动检测是否启用多路复用并提供支持。

### 移植性

使用Golang编写，而Golang默认进行静态编译，意味着你可以将编译得到的单个可执行文件在目标机器上直接执行而不需要考虑依赖的问题。你可以很方便地编译（或者交叉编译）它，然后在你的服务器，PC，树莓派，甚至路由器上部署。


## 编译构建

确保你的Golang版本 >= 1.11


```
git clone https://github.com/p4gefau1t/trojan-go.git
cd trojan-go
go build
```

Golang支持通过设置环境变量进行交叉编译，例如

```
GOOS=windows GOARCH=amd64 go build -o trojan-go.exe
```
以及

```
GOOS=linux GOARCH=arm go build -o trojan-go
```

## 使用方法

```
./trojan-go -config 你的配置文件.json
```

配置文件格式和Trojan相同, 参见[这里](https://trojan-gfw.github.io/trojan/config)。

