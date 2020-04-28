---
title: "API"
draft: false
weight: 100
---

Trojan-Go基于grpc实现了API，使用protobuf交换数据，可以通过在配置文件中添加```api```选项激活API模块。下面是一个例子

```
"api": {
    "enabled": true,
    "api_addr": "127.0.0.1",
    "api_port": 10000
}
```

目前API处于开发阶段，使用方法可以参考api下api.proto文件。