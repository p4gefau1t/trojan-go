---
title: "API"
draft: false
weight: 100
---

Trojan-Go基于gRPC实现了API，使用protobuf交换数据。客户端可获取流量和速度信息；服务端可获取各用户流量，速度，在线情况，并动态增删用户和限制速度。可以通过在配置文件中添加```api```选项激活API模块。下面是一个例子

```json
"api": {
    "enabled": true,
    "api_addr": "127.0.0.1",
    "api_port": 10000
}
```

目前API处于开发阶段，服务和RPC定义可以参考api文件夹下api.proto文件。
