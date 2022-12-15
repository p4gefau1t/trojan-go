---
title: "Full profile"
draft: false
weight: 30
---

The following is a complete configuration file with the required options

- ```run_type```

- ```local_addr```

- ```local_port```

- ```remote_addr```

- ```remote_port```

For server ```server```，```key``` and ```cert```are required.

For the client ```client```, the reverse proxy tunnel ```forward```, and the transparent proxy ```nat```, ```password``` is required

The rest of the unfilled options are filled with the values given below.

*Trojan-Go supports the more human-friendly YAML syntax, and the basic structure of the configuration file is the same as JSON, with equivalent effects. However, to comply with YAML naming conventions, you need to convert underscores ("_") to horizontal bars ("-"), e.g. ```remote_addr``` is ```remote-addr``` in YAML files*

```json
{
  "run_type": *required*,
  "local_addr": *required*,
  "local_port": *required*,
  "remote_addr": *required*,
  "remote_port": *required*,
  "log_level": 1,
  "log_file": "",
  "password": [],
  "disable_http_check": false,
  "udp_timeout": 60,
  "ssl": {
    "verify": true,
    "verify_hostname": true,
    "cert": *required*,
    "key": *required*,
    "key_password": "",
    "cipher": "",
    "curves": "",
    "prefer_server_cipher": false,
    "sni": "",
    "alpn": [
      "http/1.1"
    ],
    "session_ticket": true,
    "reuse_session": true,
    "plain_http_response": "",
    "fallback_addr": "",
    "fallback_port": 0,
    "fingerprint": ""
  },
  "tcp": {
    "no_delay": true,
    "keep_alive": true,
    "prefer_ipv4": false
  },
  "mux": {
    "enabled": false,
    "concurrency": 8,
    "idle_timeout": 60
  },
  "router": {
    "enabled": false,
    "bypass": [],
    "proxy": [],
    "block": [],
    "default_policy": "proxy",
    "domain_strategy": "as_is",
    "geoip": "$PROGRAM_DIR$/geoip.dat",
    "geosite": "$PROGRAM_DIR$/geosite.dat"
  },
  "websocket": {
    "enabled": false,
    "path": "",
    "host": ""
  },
  "shadowsocks": {
    "enabled": false,
    "method": "AES-128-GCM",
    "password": ""
  },
  "transport_plugin": {
    "enabled": false,
    "type": "",
    "command": "",
    "option": "",
    "arg": [],
    "env": []
  },
  "forward_proxy": {
    "enabled": false,
    "proxy_addr": "",
    "proxy_port": 0,
    "username": "",
    "password": ""
  },
  "mysql": {
    "enabled": false,
    "server_addr": "localhost",
    "server_port": 3306,
    "database": "",
    "username": "",
    "password": "",
    "check_rate": 60
  },
  "api": {
    "enabled": false,
    "api_addr": "",
    "api_port": 0,
    "ssl": {
      "enabled": false,
      "key": "",
      "cert": "",
      "verify_client": false,
      "client_cert": []
    }
  }
}
```

## Description

### General options

For client/nat/forward, ```remote_xxxx``` should be filled in with your trojan server address and port number, ```local_xxxx``` corresponds to the local open socks5/http proxy address (auto-adapted)

For server, ```local_xxxx``` corresponds to the trojan server listening address (port 443 is highly recommended), ```remote_xxxx``` fills in the address of the HTTP service that proxies to when non-trojan traffic is recognized, usually filling in local port 80.

```log_level``` specifies the log level. The higher the level, the less information will be output. Legitimate values are

- 0 Outputs logs above Debug (all logs)

- 1 Output Info and above logs

- 2 Output Warning and above logs

- 3 Output Error and above logs

- 4 Output Fatal and above logs

- 5 Do not output logs at all

```log_file``` specifies the log output file path. If not specified, standard output is used.

```password``` can be filled with multiple passwords. In addition to configuring passwords using the configuration file, trojan-go also supports configuring passwords using mysql, see below. The client's password can only pass the server's checksum and use the proxy service properly if it matches the password record in the server's configuration file or in the database.

```disable_http_check``` Whether to disable HTTP masquerade server availability check.

```udp_timeout``` UDP session timeout time.

### ```ssl``` options

```verify``` indicates whether the client (client/nat/forward) verifies the legitimacy of the certificate provided by the server, and is enabled by default. For security reasons, this option should not be selected false in real scenarios, otherwise it may suffer from man-in-the-middle attacks. If you use self-signed or self-issued certificates, turning on ```verify``` will cause the verification to fail. In this case, you should keep ```verify``` on, and then fill in the certificate of the server side in ```cert``` to connect normally.

```verify_hostname``` indicates whether the server side verifies the consistency between the SNI provided by the client and the server side settings. If the server-side SNI field is left blank, the authentication will be forced off.

The server side must fill in ```cert``` and ```key```, corresponding to the server's certificate and private key file, please pay attention to whether the certificate is valid/expired. If you use the certificate issued by authoritative CA, the client (client/nat/forward) can not fill in ```cert```. If you use self-signed or self-issued certificate, you should fill in the server certificate file at the ```cert```, otherwise it may cause the verification failure.

```sni``` refers to the server name field in the TLS client request, which is generally the same as the Common Name of the certificate. If you use a certificate issued by an organization such as let'sencrypt, fill in your domain name here. For clients, if this is not filled in, it will be populated using ```remote_addr```. You should specify a valid SNI (consistent with the remote certificate CN), otherwise the client may not be able to verify the validity of the remote certificate and thus cannot connect; for the server, if this item is not filled in, the Common Name in the certificate will be used as the basis for SNI verification, supporting wildcards such as *.example.com.

```fingerprint``` is used to specify the type of client-side TLS Client Hello fingerprint forgery to resist GFW's feature identification and blocking for TLS Client Hello fingerprint. trojan-go uses [utls](https://github.com/refraction-networking/utls) to perform fingerprint forgery, and Firefox fingerprints are forged by default. Legitimate values are

- "", no fingerprint forgery is used (default)

- "firefox", forging Firefox fingerprints

- "chrome", to forge Chrome fingerprints

- "ios", to forge iOS fingerprints

Once the value of the fingerprint is set, client-side fields such as ```cipher```, ```curves```, ```alpn```, ```session_ticket``` that have the potential to affect the fingerprint will be overridden using the specific settings of that fingerprint.
```alpn``` specifies the protocol for application layer protocol negotiation for TLS. Transferring in TLS Client/Server Hello and negotiating the protocol used by the application layer is only used for fingerprint forgery and has no practical effect. **If a CDN is used, the wrong alpn field may result in negotiating with the CDN to get the wrong application layer protocol**.

```prefer_server_cipher``` Does the client prefer to select the cryptographic suite provided by the server in the negotiation.

```cipher``` The cryptography suite used by TLS. The ```cipher13``` field is merged with this field. You should only go ahead and fill this in to modify the TLS cipher suite used by trojan-go if you know exactly what you are doing. **Normally, you should leave this blank or leave it blank**. trojan-go will automatically choose the most appropriate cryptographic algorithm to improve performance and security based on the current hardware platform and the remote end. If required, cryptography suite names are separated by semicolons (":") in order of preference. go's TLS library deprecates some of the insecure cryptography suites in TLS1.2 and fully supports TLS1.3. by default, trojan-go will give preference to the more secure TLS1.3.

```curves``` specifies the elliptic curve that TLS prefers to use in ECDHE. This should only be filled in if you know exactly what you are doing. The curve names are separated by semicolons (":") and are listed in order of preference.

```plain_http_response``` refers to the raw data sent in plaintext (raw TCP data) when the server-side TLS handshake fails. This field is filled with the path to that file. It is recommended to use ```fallback_port``` instead of this field.

The ```fallback_addr``` and ```fallback_port``` refer to the address to which trojan-go redirects the connection if the server-side TLS handshake fails. This is a feature of trojan-go to better conceal the server against active GFW detection, making the server's port 443 behave exactly the same as a normal server when encountering probes from non-TLS protocols. When a server accepts a connection but cannot perform a TLS handshake, if ```fallback_port``` is not empty, traffic will be proxied to fallback_addr:fallback_port. if ```fallback_addr``` is empty, it will be populated with ```remote_addr``` . For example, you can enable an https service locally using nginx, and when your server port 443 is requested by a non-TLS protocol (such as an http request), trojan-go will proxy to the local https server and nginx will return a 400 Bad Request page using the http protocol in plaintext. You can verify this by visiting ```http://your-domain-name.com:443``` using your browser.

The path to the ```key_log``` file for the TLS key log. If filled in then key logging is turned on. **Logging keys will break TLS security, and this item should not be used for anything other than debugging.**

### ```mux``` multiplexing options

Multiplexing is a feature of trojan-go. If both the server and client are trojan-go, you can turn on mux multiplexing to reduce latency in high concurrency scenarios (you only need to turn on this option on the client side, the server side adapts automatically).

Note that the point of multiplexing is to reduce handshake latency, not to improve link speed. Instead, it increases CPU and memory consumption on both the client and server side, which may cause speed degradation.

```enabled``` Enables or disables multiplexing.

```concurrency``` refers to the maximum number of connections a single TLS tunnel can carry, the default is 8. The higher the value, the lower the latency of TLS due to handshaking when multiple connections are concurrent, but the network throughput may be reduced, filling in a negative number or 0 means that all connections are carried using only one TLS tunnel.

```idle_timeout``` Idle timeout. Indicates how long the TLS tunnel will be idle before it is closed, in seconds. If the value is negative or 0, then the TLS tunnel is closed as soon as it becomes idle.

### ```router``` routing options

Routing is a feature of trojan-go. trojan-go has three types of routing policies.

- Proxy Proxy. Proxy the request through a TLS tunnel, with the trojan server and the destination address connected.

- Bypass bypassing. Connects directly locally to the destination address.

- Block blocking. Does not proxy the request and closes the connection directly.

Fill in the ```proxy```, ```bypass```, ```block``` fields with the corresponding list geoip/geosite or routing rules, trojan-go then executes the corresponding routing policy according to the IP (CIDR) or domain in the list. The client (client) can configure three policies, the server (server) can only configure block policies.

```enabled``` Enables or disables the routing module.

```default_policy``` refers to the default policy to be used when all three lists fail to match. Legitimate values are

- "proxy"

- "bypass"

- "block"

Same meaning as above.

```domain_strategy``` Domain resolution strategy, default "as_is". Legitimate values are.

- "as_is" to match only within the domain rules in each list.

- "ip_if_non_match", which first matches within the rules of the domain name in each list; if it does not match, it resolves to IP and then matches within the rules of the IP address in each list. This policy may lead to DNS leaks or contamination.

- "ip_on_demand", resolves to IP first and matches within the IP address rules in each list; if it does not match, it matches within the domain name rules in each list. This policy may lead to DNS leaks or contamination.

The ```geoip``` and ```geosite``` fields refer to the geoip and geosite database file paths, which by default use geoip.dat and geosite.dat in the directory where the program is located. you can also specify the working directory by specifying the environment variable TROJAN_GO_LOCATION_ASSET.

### ```websocket``` option

Websocket transfers are a feature of trojan-go. In the case of **normal direct connections to proxy nodes**, turning on this option will not improve your link speed (or possibly even drop it), nor will it improve your connection security. You should only use websocket in cases where you need to use a CDN for relaying, or use a server such as nginx to distribute based on paths.

```enabled``` indicates whether to enable Websocket to carry traffic, the server side supports both general Trojan protocol and websocket-based Trojan protocol when enabled, the client side will only use websocket to carry all Trojan protocol traffic when enabled.

```path``` refers to the URL path used by websocket, which must start with a slash ("/"), such as "/longlongwebsocketpath", and the server and client must be the same.

```host``` is the host name used in HTTP requests during the Websocket handshake. The client uses ```remote_addr``` to fill in if left blank. If a CDN is used, this option is usually filled with the domain name. An incorrect ```host``` may cause the CDN to fail to forward the request.

### ```shadowsocks``` AEAD encryption option

This option is used to replace the deprecated obfuscated encryption and double TLS. if this option is set to enabled, a Shadowsocks AEAD encryption layer will be inserted under the Trojan protocol layer. That is, within the (already encrypted) TLS tunnel, all Trojan protocols will again be encrypted using the AEAD method. Note that this option is independent of whether Websocket is enabled or not. All Trojan traffic will be encrypted again, regardless of whether Websocket is on or not.

Note that turning this option on will potentially degrade transport performance, and you should only enable this option if you do not trust the transport channel carrying the Trojan protocol. Example.

- You are using a Websocket, relayed through an untrusted CDN (e.g. a domestic CDN)

- Your connection to the server is subject to a man-in-the-middle attack by GFW against TLS

- Your certificate is invalid, and you cannot verify the validity of the certificate

- You use a pluggable transport layer that cannot guarantee cryptographic security

etc.

Thanks to the use of AEAD, trojan-go can correctly determine if the request is valid and is being actively probed, and respond accordingly.

```enabled``` Enables or disables the Shadowsocks AEAD encryption Trojan protocol layer.

```method``` encryption method. Legitimate values are.

- "CHACHA20-IETF-POLY1305"

- "AES-128-GCM" (default)

- "AES-256-GCM"

```password``` is used to generate the password for the master key. If AEAD encryption is enabled, you must ensure that the client and server are identical.

### ```transport_plugin``` Transport layer plugin options

```Enabled`` Whether to enable transport layer plug-in instead of TLS transport. Once transport layer plugin support is enabled, trojan-go will transmit **unencrypted TLS-encrypted trojan protocol traffic in plaintext to the plugin** to allow user-defined obfuscation and encryption of the traffic.

```type``` plug-in types. The currently supported types are

- "shadowsocks", which supports the shadowsocks obfuscation plugin conforming to the [SIP003](https://shadowsocks.org/en/spec/Plugin.html) standard. trojan-go will replace environment variables and modify its own configuration at startup according to the SIP003 standard (```remote_addr/remote_port/local_addr/local_port```) so that the plugin communicates directly with the remote end, while trojan-go only listens/connects to the plugin.

- "plaintext", to use plaintext transfer. By selecting this option, trojan-go does not modify any address configuration (```remote_addr/remote_port/local_addr/local_port```) and does not start the plugin in ```command```, only the lowest TLS transport layer is removed and TCP plaintext transport is used. The purpose of this option is to support nginx and others to take over TLS and perform triage, and advanced users to perform debugging tests. **Please do not use plaintext transport mode directly to penetrate the firewall. **

- "other", other plugins. By selecting this, trojan-go will not modify any address configuration (```remote_addr/remote_port/local_addr/local_port```), but will start the plugins in ```command``` and pass in parameters and environment variables.

The path to the ```command``` transport layer plugin executable. trojan-go will execute it along with it at startup.

The ```arg``` transport layer plugin startup parameters. This is a list, e.g. ```["-config", "test.json"]```.

```env``` transport layer plugin environment variables. This is a list, e.g. ```["VAR1=foo", "VAR2=bar"]```.

```option``` Transport layer plug-in configuration (SIP003). For example ```"obfs=http;obfs-host=www.baidu.com"```.

### ```tcp``` option

```no_delay``` Whether TCP packets are sent directly without waiting for the buffer to fill up.

```keep_alive``` Whether or not to enable TCP heartbeat survival detection.

```prefer_ipv4``` Whether to give preference to IPv4 addresses.

### ```mysql``` database options

trojan-go is compatible with trojan's mysql-based approach to user management, but the more recommended approach is to use the API.

```enabled``` indicates whether to enable the mysql database for user authentication.

```check_rate``` is the interval in seconds for trojan-go to fetch user data from MySQL and update the cache.

The other options can be named as such and will not be repeated.

The users table structure is the same as the trojan version definition, here is an example of creating a users table. Note that password here refers to the value (string) of the password after SHA224 hashing, and the units of traffic download, upload, quota are bytes. You can add and remove users or specify their traffic quota by modifying their records in the users table of the database. trojan-go will automatically update the list of currently active users based on all their traffic quotas. If download+upload>quota, the trojan-go server will refuse the connection for that user.

```mysql
CREATE TABLE users (
    id INT UNSIGNED NOT NULL AUTO_INCREMENT,
    username VARCHAR(64) NOT NULL,
    password CHAR(56) NOT NULL,
    quota BIGINT NOT NULL DEFAULT 0,
    download BIGINT UNSIGNED NOT NULL DEFAULT 0,
    upload BIGINT UNSIGNED NOT NULL DEFAULT 0,
    PRIMARY KEY (id),
    INDEX (password)
);
```

### ```forward_proxy``` preproxy option

The forward proxy option allows to use other proxies to carry trojan-go's traffic

```enabled``` Enables or disables the forward proxy (socks5).

```proxy_addr``` The host address of the preproxy.

```proxy_port``` The port number of the preproxy.

```username``` ```password``` The user and password of the proxy, if left blank no authentication will be used.

### ```api``` options

trojan-go provides an API based on gRPC to support server-side and client-side management and statistics. You can implement client-side traffic and speed statistics, server-side traffic and speed statistics for each user, dynamic addition and deletion of users and speed limits, etc.

```enabled``` Enables or disables the API function.

```api_addr``` The address of the gRPC listener.

```api_port``` The port gRPC is listening on.

```ssl``` TLS related settings.

- ```enabled``` Whether to use TLS for gRPC traffic.

- ```key```, ```cert``` server private key and certificate.

- ```verify_client``` Whether to certify client certificates.

- ```client_cert``` If client authentication is enabled, fill in the list of certified client certificates here.


{{% panel status="warning" title="Warning" %}}
Do not expose API services without TLS bi-directional authentication directly to the Internet, otherwise it may lead to various security problems.
{{% /panel %}}
