---
title: "API开发"
draft: false
weight: 100
---

Trojan-Go基于gRPC实现了API，使用protobuf交换数据。客户端可获取流量和速度信息；服务端可获取各用户流量，速度，在线情况，并动态增删用户和限制速度。可以通过在配置文件中添加```api```选项激活API模块。下面是一个例子

```json
"api": {
    "enabled": true,
    "api_addr": "0.0.0.0",
    "api_port": 10000,
    "api_tls": true,
    "ssl": {
      "cert": "api_cert.crt",
      "key": "api_key.crt",
      "key_password": "",
      "client_cert": [
          "api_client_cert1.crt",
          "api_client_cert2.crt"
      ]
    },
}
```

如果需要实现API客户端进行对接，请参考api/service/api.proto文件。
