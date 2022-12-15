---
title: "Enabling multiplexing to improve network concurrency performance"
draft: false
weight: 1
---

### Note that Trojan does not support this feature

Trojan-Go supports the use of multiplexing to improve network concurrency performance.

The Trojan protocol is based on TLS, and before a secure TLS connection can be established, both sides of the connection need to negotiate and exchange keys to ensure the security of the subsequent communication. This process is known as the TLS handshake.

The GFW currently censors and interferes with the TLS handshake, and due to egress network congestion, it usually takes nearly a second or more for an ordinary line to complete the TLS handshake. This can lead to increased latency for web browsing and video viewing.

Trojan-Go uses multiplexing to solve this problem. Each established TLS connection will host multiple TCP connections. When a new proxy request arrives, instead of handshaking with the server to initiate a new TLS connection, existing TLS connections are reused whenever possible. This reduces the latency caused by frequent TLS handshakes and TCP handshakes.

Enabling multiplexing will not increase your link speed (or even decrease it), and may increase the computational burden on the server and client. Roughly speaking, multiplexing can be interpreted as sacrificing network throughput and CPU power for lower latency. It can improve the usage experience in high concurrency scenarios, such as when browsing web pages containing a large number of images, or when sending a large number of UDP requests.

To activate the ```mux``` module, just set the ```enabled``` field in the ```mux``` option to true, here is a client-side example

```json
...
"mux" :{
    "enabled": true
}
```

Just configure the client side, the server side can be adapted automatically without configuring the ```mux``` option.

The complete mux configuration is as follows

```json
"mux": {
    "enabled": false,
    "concurrency": 8,
    "idle_timeout": 60
}
```

```concurrency``` is the maximum number of TCP connections that each TLS connection can carry. The larger this value is, the more each TLS connection is reused and the lower the latency due to handshaking. However, the greater the computational burden on the server and client will also be, which could potentially make your network throughput lower. If your line's TLS handshake is extremely slow, you can set this value to -1 and Trojan-Go will perform only one TLS handshake, using only one TLS connection for transmission.

The ```idle_timeout``` refers to how long each TLS connection is idle before it is closed. Setting the timeout time **may** help reduce unnecessary long connection live confirmation (Keep Alive) traffic transmissions triggering GFW probes. You can set this value to -1 and TLS connections will be closed immediately when idle.
