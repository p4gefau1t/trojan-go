---
title: "Tunnels and reverse proxies"
draft: false
weight: 5
---

You can use Trojan-Go to set up tunnels. A typical application is to use Trojan-Go to set up a local, unpolluted DNS server, here is an example configuration

```json
{
    "run_type": "forward",
    "local_addr": "127.0.0.1",
    "local_port": 53,
    "remote_addr": "your_awesome_server",
    "remote_port": 443,
    "target_addr": "8.8.8.8",
    "target_port": 53,
    "password": [
        "your_awesome_password"
    ]
}
```

forward is essentially a client, but you need to fill in the ```target_addr``` and ```target_port``` fields to indicate the target of the reverse proxy.

After using this configuration file, the local 53 TCP and UDP ports will be listened to, and all TCP or UDP data sent to the local 53 port will be forwarded to the remote server your_awesome_server via TLS tunnel, and after the remote server gets a response, the data will be returned to the local 53 port via the tunnel. In other words, you can treat 127.0.0.1 as a DNS server, and the results of the local query and the remote server query are the same. You can use this configuration to bypass DNS pollution.

On the same principle, you can build a Google mirror locally

```json
{
    "run_type": "forward",
    "local_addr": "127.0.0.1",
    "local_port": 443,
    "remote_addr": "your_awesome_server",
    "remote_port": 443,
    "target_addr": "www.google.com",
    "target_port": 443,
    "password": [
        "your_awesome_password"
    ]
}
```

Visit ```https://127.0.0.1``` to access the Google homepage, but note that here the browser will raise a certificate error warning because the https certificate provided by the Google server is that of google.com, and the current domain is 127.0.0.1.

Similarly, other proxy protocols can be used for forward transfers. For example, using Trojan-Go to transfer traffic from shadowsocks, the remote host opens the ss server and listens to 127.0.0.1:12345, and the remote server opens the normal Trojan-Go server on port 443. You can specify the configuration like this

```json
{
    "run_type": "forward",
    "local_addr": "0.0.0.0",
    "local_port": 54321,
    "remote_addr": "your_awesome_server",
    "remote_port": 443,
    "target_addr": "www.google.com",
    "target_port": 12345,
    "password": [
        "your_awesome_password"
    ]
}
```

Thereafter, any TCP/UDP connection to the local port 54321 is equivalent to a connection to the remote port 12345. You can use the shadowsocks client to connect to local port 54321 and ss traffic will be transferred to the ss server on the remote port 12345 using trojan's tunnel connection.
