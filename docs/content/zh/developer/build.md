---
title: "编译和自定义Trojan-Go"
draft: false
weight: 10
---

编译需要Go版本号高于1.14.x，请在编译前确认你的编译器版本。推荐使用snap安装和更新go。

编译方式非常简单，可以使用Makefile预设步骤进行编译：

```shell
make
make install #安装systemd服务等，可选
```

或者直接使用Go进行编译：

```shell
go build -tags "full" #编译完整版本
```

可以通过指定GOOS和GOARCH环境变量，指定交叉编译的目标操作系统和架构，例如

```shell
GOOS=windows GOARCH=386 go build -tags "full" #windows x86
GOOS=linux GOARCH=arm64 go build -tags "full" #linux arm64
```

你可以使用release.sh进行批量的多个平台的交叉编译，release版本使用了这个脚本进行构建。

Trojan-Go的大多数模块是可插拔的。在build文件夹下可以找到各个模块的导入声明。如果你不需要其中某些功能，或者需要缩小可执行文件的体积，可以使用构建标签(tags)进行模块的自定义，例如

```shell
go build -tags "full" #编译所有模块
go build -tags "client" -trimpath -ldflags="-s -w -buildid=" #只有客户端功能，且去除符号表缩小体积
go build -tags "server mysql" #只有服务端和mysql支持
```

使用full标签等价于

```shell
go build -tags "api client server forward nat other"
```
