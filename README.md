# Trojan-Go

Trojan proxy written in golang. It is compatiable with the original Trojan protocol and config file. 

It's still currently in heavy development.

## Features

### Compatible

It's fully compatible with the Trojan protocol, so that you can safely replace your client and server program with trojan-go, or even just replace one of them, without changing the config file.

### Easy to use

Trojan-go's configuration file format is compatible with Trojan's, while it's being simplyfied. You can launch your server and client much more easily. Here's an example:

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

run_type supported by Trojan-Go (the same as Trojan):

- Client

- Server

- NAT (transparent proxy, see [here](https://github.com/shadowsocks/shadowsocks-libev/tree/v3.3.1#transparent-proxy))

- Forward

For more infomation, see Trojan's [docs](https://trojan-gfw.github.io/trojan/config) about the configuration file.

### Multiplexing

TLS handshaking may takes much time in a poor network condition.
Trojan-go supports multiplexing([smux](https://github.com/xtaci/smux)), which imporves the performance in the high-concurrency scenario by forcing one single TLS tunnel connection carries mutiple TCP connections.

It's **NOT** compatible with the original trojan protocol, so it's disabled by default. You can enable it by setting up the "mux" field, in the tcp options.

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
        "sni": "your_awesome_domain_name",
    },
    "tcp": {
        "mux": true
    }
}
```

You will only need to set up the client's config file, and the server will automatically detect it.

### Portable

It's written in golang, so it's static compiled by default. You can easily build and deploy it to your server, pc, raspberry pi, and router.

## Build

Just make sure your golang version >= 1.11


```
git clone https://github.com/p4gefau1t/trojan-go.git
cd trojan-go
go build
```

You can cross-compile it by setting up the environment vars, for example
```
GOOS=windows GOARCH=amd64 go build -o trojan-go.exe
```

or

```
GOOS=linux GOARCH=arm go build -o trojan-go
```

## Usage

```
./trojan-go -config your_awesome_config_file.json
```

The format of the configuration file is compatible, see [here](https://trojan-gfw.github.io/trojan/config).


## TODOs

- [x] Server
- [x] Forward
- [x] NAT
- [x] Client
- [x] UDP Tunneling
- [x] Transparent proxy
- [x] Log
- [x] Mux
- [ ] TLS Settings
- [x] TLS redirecting
- [ ] non-TLS redirecting
- [ ] Cert utils
- [ ] Database support
- [x] Traffic stats
- [ ] TCP Settings