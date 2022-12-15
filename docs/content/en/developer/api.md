---
title: "API Development"
draft: false
weight: 100
---

Trojan-Go implements an API based on gRPC, using protobuf to exchange data. The client can get the traffic and speed information; the server can get the traffic, speed, online situation of each user, and dynamically add and delete users and limit the speed. The API module can be activated by adding the ```api`` option to the configuration file. Here is an example, the meaning of each field can be found in the section "Complete configuration file".

```json
...
"api": {
    "enabled": true,
    "api_addr": "0.0.0.0",
    "api_port": 10000,
    "ssl": {
      "enabled": true,
      "cert": "api_cert.crt",
      "key": "api_key.key",
      "verify_client": true,
      "client_cert": [
          "api_client_cert1.crt",
          "api_client_cert2.crt"
      ]
    },
}
```

If you need to implement an API client for interfacing, please refer to the api/service/api.proto file.
