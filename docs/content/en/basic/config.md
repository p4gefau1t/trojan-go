---
title: "Configuring Trojan-Go correctly"
draft: false
weight: 22
---

The following section will describe how to properly configure Trojan-Go to completely hide your proxy node features.

Before you start, you need

- A server that is not blocked by GFW

- A domain name, you can use a free domain name service such as .tk

- Trojan-Go, which can be downloaded from the release page

- Certificates and keys, which can be issued for free from agencies such as letsencrypt

### Server-side configuration

Our goal is to make your server behave the same as a normal HTTPS site.

First you need an HTTP server, you can configure a local HTTP server using nginx, apache, caddy etc. or you can use someone else's HTTP server. the role of the HTTP server is to show a fully functional web page to GFW when it actively detects it.

**You need to specify the address of this HTTP server in ```remote_addr``` and ```remote_port```. The ```remote_addr``` can be either an IP or a domain name. Trojan-Go will test if this HTTP server is working properly, and if not, Trojan-Go will refuse to start it.**

The following is a more secure server configuration server.json, which requires you to configure an HTTP service on local port 80 (if necessary, you can also use another HTTP server for your website, such as "remote_addr": "example.com") and an HTTPS service on port 1234, or a static HTTP service that displays "400 Bad Request" static HTTP web service. (Optionally, you can delete the ```fallback_port``` field and skip this step)

```json
{
    "run_type": "server",
    "local_addr": "0.0.0.0",
    "local_port": 443,
    "remote_addr": "127.0.0.1",
    "remote_port": 80,
    "password": [
        "your_awesome_password"
    ],
    "ssl": {
        "cert": "server.crt",
        "key": "server.key",
        "fallback_port": 1234
    }
}
```

This configuration file causes Trojan-Go to listen on port 443 on all IP addresses of the server (0.0.0.0), using server.crt and server.key as the certificate and key for the TLS handshake, respectively. You should use as complex a password as possible, while making sure that the client and server ```password``` are the same. Note that **Trojan-Go will check if your HTTP server ```http://remote_addr:remote_port``` is working properly. If your HTTP server is not working properly, Trojan-Go will refuse to start it.**

When a client tries to connect to Trojan-Go's listening port, the following happens.

- If the TLS handshake is successful and the contents of the TLS are detected as non-Trojan protocol (possibly an HTTP request, or an active probe from GFW). Trojan-Go proxies the TLS connection to the HTTP service on local 127.0.0.1:80. At this point it appears to the remote end that the Trojan-Go service is an HTTPS site.

- If the TLS handshake is successful and is confirmed to be a Trojan protocol header with the correct password in it, then the server will parse the request from the client and proxy it, otherwise it is handled in the same way as in the previous step.

- If the TLS handshake fails, the other party is using a protocol other than TLS to connect. At this point Trojan-Go proxies the TCP connection to an HTTPS service (or HTTP service) running on local 127.0.0.1:1234, returning an HTTP page displaying 400 Bad Reqeust. The ```fallback_port``` is an optional option, if not filled in, Trojan-Go will simply terminate the connection. Although it is optional, it is highly recommended to fill it in.

You can verify this by visiting your domain ```https://your-domain-name.com``` using your browser. If it works properly, your browser will display a normal HTTPS-protected web page with the same content as the page on the server's native port 80. You can also use ```http://your-domain-name.com:443``` to verify that ```fallback_port``` is working properly.

In fact, you can even use Trojan-Go as your HTTPS server to serve HTTPS to your website. Visitors can browse your site through Trojan-Go normally, without affecting each other and the proxy traffic. Note, however, that you should not build services with high real-time requirements at ```remote_port``` and ```fallback_port```. Trojan-Go will intentionally add a small delay when it identifies non-Trojan protocol traffic to resist GFW's time-based detection.

Once configured, you can use the

```shell
./trojan-go -config ./server.json
```

Start the server.

### Client configuration

The corresponding client configuration client.json

```json
{
    "run_type": "client",
    "local_addr": "127.0.0.1",
    "local_port": 1080,
    "remote_addr": "your_awesome_server",
    "remote_port": 443,
    "password": [
        "your_awesome_password"
    ],
    "ssl": {
        "sni": "your-domain-name.com"
    }
}
```

This client configuration enables Trojan-Go to open a socks5/http proxy listening on local port 1080 (auto-aware) with a remote server of your_awesome_server:443, your_awesome_server can be an IP or domain name.

If you fill in ```remote_addr``` with a domain name, ```sni``` can be omitted. If you fill in ```remote_addr``` with the IP address, the ```sni``` field should be filled in with the corresponding domain name of the certificate you applied for, or the Common Name of the certificate you issued yourself, and it must be consistent. Note that the ```sni``` field is currently in the TLS protocol **explicit transmission** (the purpose is to make the server provide the corresponding certificate). GFW has been proven to have SNI detection and blocking capabilities, so do not fill in similar ```google.com``` and other domains that have been blocked, otherwise it is likely to cause your server to be blocked as well.

After the configuration is done, you can use

```shell
./trojan-go -config ./client.json
```

Start the client.

More information about configuration files can be found in the left navigation bar.
