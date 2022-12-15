---
title: "CDN forwarding and resisting man-in-the-middle attacks using Websocket"
draft: false
weight: 2
---

{{% panel status="caution" title="Compatibility" %}}
Note that Trojan does not support this feature
{{% /panel %}}

Trojan-Go supports using TLS+Websocket to host the Trojan protocol, making it possible to relay traffic using CDNs.

The Trojan protocol itself is not encrypted and relies on the outer layer of TLS for security, but once the traffic passes through the CDN, TLS is transparent to the CDN. Its service provider can review the plaintext content of the TLS. **If you are using an untrusted CDN (any CDN service registered and filed in mainland China should be considered untrusted), please make sure to turn on Shadowsocks AEAD to encrypt Webosocket traffic to avoid being identified and censored.**

Add the websocket option to both the server and client configuration files and set its ```enabled``` field to true, and fill in the ```path``` field and the ```host``` field to enable websocket support. Here is a complete Websocket option:

```json
"websocket": {
    "enabled": true,
    "path": "/your-websocket-path",
    "host": "example.com"
}
```

```host``` is the host name, usually fill in the domain name. Client ```host``` is optional, fill in your domain name. If left blank, it will be filled in with ```remote_addr```.

The ```path``` refers to the URL path where the websocket is located and must start with a slash ("/"). There are no special requirements for the path, just satisfy the basic URL format, but make sure that the ```path``` of the client and server are the same. The ```path``` should be a longer string to avoid direct active detection by GFW.

The client's ```host``` will be included in the Websocket handshake HTTP request sent to the CDN server and must be valid; the server and client ```path``` must be the same, otherwise the Websocket handshake will not work.

Here is an example of a client-side configuration file

```json
{
    "run_type": "client",
    "local_addr": "127.0.0.1",
    "local_port": 1080,
    "remote_addr": "www.your_awesome_domain_name.com",
    "remote_port": 443,
    "password": [
        "your_password"
    ],
    "websocket": {
        "enabled": true,
        "path": "/your-websocket-path",
        "host": "example.com"
    },
    "shadowsocks": {
        "enabled": true,
        "password": "12345678"
    }
}
```
