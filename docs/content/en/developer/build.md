---
title: "Compiling and customizing Trojan-Go"
draft: false
weight: 10
---

Compilation requires Go version number higher than 1.14.x. Please check your compiler version before compiling. It is recommended to use snap to install and update go.

The compilation is very simple and can be done using the Makefile preset steps

```shell
make
make install #Install systemd services etc., optional
```

Or you can compile directly using Go:

```shell
go build -tags "full" #compile full version
```

You can specify the target OS and architecture for cross-compilation by specifying the GOOS and GOARCH environment variables, for example

```shell
GOOS=windows GOARCH=386 go build -tags "full" #windows x86
GOOS=linux GOARCH=arm64 go build -tags "full" #linux arm64
```

You can use release.sh for batch cross-compilation of multiple platforms, the release version uses this script for building.

Most modules of Trojan-Go are pluggable. Import declarations for individual modules can be found in the build folder. If you don't need some of these features, or need to reduce the size of the executable, you can customize the module using build tags, for example

```shell
go build -tags "full" #compile all modules
go build -tags "client" -trimpath -ldflags="-s -w -buildid=" #only client functionality, and remove symbol table to reduce size
go build -tags "server mysql" #Only server-side and mysql support
```

Using the full tag is equivalent to

```shell
go build -tags "api client server forward nat other"
```
