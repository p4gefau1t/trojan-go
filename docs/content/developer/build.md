---
title: "编译和自定义Trojan-Go"
draft: false
weight: 10
---

编译需要Golang版本号高于0.14.x，请在编译前确认你的编译器版本。推荐使用snap安装和更新go。

编译过程非常简单

```
go mod tidy #更新和下载所需库
go build #编译
```

build.sh中的命令禁用了cgo并去除可执行文件的调试符号以减小体积。

可以通过指定GOOS和GOARCH环境变量，指定交叉编译的目标操作系统和架构，例如

```
GOOS=windows GOARCH=386 go build  #windows x86
GOOS=linux GOARCH=arm64 go build  #linux arm64
```

你可以使用build-all.sh进行批量的大量平台的交叉编译，release版本使用了这个脚本进行构建。

Trojan-Go的大多数模块是可插拔的。你可以在main.go中找到这些模块的导入声明，类似于

```
...
	_ "github.com/p4gefau1t/trojan-go/proxy/client"
	_ "github.com/p4gefau1t/trojan-go/proxy/relay"
	_ "github.com/p4gefau1t/trojan-go/proxy/server"
...
```

如果你不需要其中某些功能，或者需要缩小可执行文件的体积，可以直接将其注释或删除。程序编译后依然可以正常运行，但不再支持这些被移除的功能。