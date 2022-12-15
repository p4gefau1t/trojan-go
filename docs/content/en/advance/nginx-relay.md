---
title: "A multi-path split relaying scheme based on SNI proxy"
draft: false
weight: 6
---

## Preface

Trojan is a tool for encrypted data transmission through TLS encapsulation. By using its TLS feature, we can achieve different paths of relaying on the same host port through SNI proxy.

## Required tools and other preparations

- Relay machine: nginx version 1.11.5 and above
- Landing machine: trojan server (no version required)

## Configuration method

For the sake of illustration, two relay hosts and two landing hosts are used here.
The four hosts are bound to the domain name (a/b/c/d).example.com as shown in the figure.
There are 4 paths to each other. They are a-c, a-d, b-c, and b-d, respectively.

```text
                        +-----------------+           +--------------------+
                        |                 +---------->+                    |
                        |   VPS RELAY A   |           |   VPS ENDPOINT C   |
                  +---->+                 |   +------>+                    |
                  |     |  a.example.com  |   |       |   c.example.com    |
                  |     |                 +------+    |                    |
  +----------+    |     +-----------------+   |  |    +--------------------+
  |          |    |                           |  |
  |  client  +----+                           |  |
  |          |    |                           |  |
  +----------+    |     +-----------------+   |  |    +--------------------+
                  |     |                 |   |  |    |                    |
                  |     |   VPS RELAY B   |   |  +--->+   VPS ENDPOINT D   |
                  +---->+                 +---+       |                    |
                        |  b.example.com  |           |   d.example.com    |
                        |                 +---------->+                    |
                        +-----------------+           +--------------------+
```

### Configure path domain names and corresponding certificates

First we need to assign each path a separate domain name and make it resolve to the respective entry host.

```text
a-c.example.com CNAME a.example.com
a-d.example.com CNAME a.example.com
b-c.example.com CNAME b.example.com
b-d.example.com CNAME b.example.com
```

Then we need to deploy certificates for all target paths on the landing host
HTTP authentication cannot be passed because the resolution record and the host IP do not match. Here it is recommended to use DNS authentication to issue certificates.
The specific DNS validation plugin needs to be chosen according to your domain DNS resolution host, here AWS Route 53 is used.

```shell
certbot certonly --dns-route53 -d a-c.example.com -d b-c.example.com // on host C
certbot certonly --dns-route53 -d a-d.example.com -d b-d.example.com // on host D
```

### Configuring SNI proxy

Here we use the ssl_preread module of nginx to implement the SNI proxy.
Please fix the nginx.conf file as follows after installing nginx.
Note that this is not an HTTP service, so please do not write it in the configuration of the virtual host.

The corresponding configuration for host A is given here, and the same for host B.

```nginx
stream {
  map $ssl_preread_server_name $name {
    a-c.example.com c.example.com; # Forward a-c path traffic to host C
    a-d.example.com d.example.com; # Forward a-d path traffic to host D

    # If you need to configure other services on this host that take up port 443 (such as web services and Trojan services)
    # Please make those services listen on other local ports (4000 is used here)
    # All TLS requests that do not match the SNI above will be forwarded to this port, remove this line if you don't need it
    default localhost:4000;
  default localhost:4000; }

  server {
    listen 443; # listen on port 443
    proxy_pass $name;
    ssl_preread on;
  ssl_preread on; }
}
```

### Configuring the landed Trojan service

In the previous configuration we used a certificate to issue the domain names for all target paths, so here we can use a Trojan server to handle requests for all target paths.
The configuration of Trojan is no different from the usual configuration method, and an example is still provided here. Unrelated configuration has been omitted.

```json
{
    "run_type": "server",
    "local_addr": "0.0.0.0",
    "local_port": 443,
    "ssl": {
        "cert": "/path/to/certificate.crt",
        "key": "/path/to/private.key",
    }
    ...
}
```

Tip: If you need to use separate Trojan servers for different paths on the landed host (for example, if you need to access your own billing service), you can configure an SNI proxy on the landed machine and forward it to a different local Trojan server listening port. Since the configuration is basically the same as the process mentioned above, we will not repeat it here.

## Summary

With the configuration method described above, we can implement multi-entry, multi-exit, multi-stage trunking of Trojan traffic on a single port.
For multi-stage trunking, simply configure the SNI proxy on the intermediate nodes along the same lines.
