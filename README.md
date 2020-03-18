# Trojan-Go

Trojan proxy written in golang. It is compatiable with the original Trojan protocol and config file. 

It's still currently in heavy development.


## Build

Just make sure your golang version >= 1.11


```
git clone https://github.com/p4gefau1t/trojan-go.git
cd trojan-go
go build
```

You can cross-compile it by setting up the environment vars, for example
```
GOOS=windows GOARCH=386 go build -o trojan-go.exe
```

or

```
GOOS=linux GOARCH=amd64 go build -o trojan-go
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
- [ ] Mux
- [ ] TLS Settings
- [x] TLS redirecting
- [ ] non-TLS redirecting
- [ ] Cert utils
- [ ] Database support
- [x] Traffic stats