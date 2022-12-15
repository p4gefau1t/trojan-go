---
title: "Using Shadowsocks Plugin/Pluggable Transport Layer"
draft: false
weight: 7
---

### Note that Trojan does not support this feature

Trojan-Go supports a pluggable transport layer. In principle, Trojan-Go can use any software that has TCP tunneling capabilities as a transport layer, such as v2ray, shadowsocks, kcp, etc. Also, Trojan-Go is compatible with Shadowsocks' SIP003 plugin standard, such as GoQuiet, v2ray-plugin, etc. You can also use Tor's transport layer plugins, such as obfs4, meek, etc.

You can use these plugins to replace the TLS transport layer of Trojan-Go.

With pluggable transport layer plugins turned on, the Trojan-Go client will transmit **traffic plaintext** directly to the client local plugins for processing. The client-side plug-in is responsible for encryption and obfuscation, and transmits the traffic to the server-side plug-in. The server-side plugin receives the traffic, decrypts and parses it, and transmits **traffic plaintext** to the server-side local Trojan-Go server.

You can use any plugin to encrypt and obfuscate the traffic, just add the "transport_plugin" option, specify the path to the plugin's executable, and configure it properly.

We recommend that you design your own protocols and develop your own plugins**. Because all existing plugins can not interface with Trojan-Go's features to combat active detection, and some plugins do not have encryption capabilities. If you are interested in developing plugins, feel free to check out the guidelines for plugin design in the "Implementation details and development guidelines" section.

For example, you can use the SIP003-compliant v2ray-plugin, an example of which is shown below:

**This configuration uses websocket to transmit unencrypted trojan protocol in clear text, which is a security risk. This configuration is for demonstration purposes only. **

**Do not use this configuration to penetrate GFW under any circumstances.**

Server-side configuration.

```json
... (omitted)
"transport_plugin": {
    "enabled": true,
    "type": "shadowsocks",
    "command": ". /v2ray-plugin",
    "arg": ["-server", "-host", "www.baidu.com"]
}
```

Client configuration.

```json
... (omitted)
"transport_plugin": {
    "enabled": true,
    "type": "shadowsocks",
    "command": ". /v2ray-plugin",
    "arg": ["-host", "www.baidu.com"]
}
```

Note that the v2ray-plugin plugin needs to specify the ```-server``` parameter to distinguish between client and server. For more detailed instructions on the plugin, refer to the v2ray-plugin documentation.

After starting Trojan-Go, you can see the output of v2ray-plugin startup. The plugin will disguise the traffic as websocket traffic and transmit it.

Non-SIP003 standard plugins may require different configuration, you can specify ```type``` as ```other``` and specify the plugin address, plugin startup parameters, environment variables by yourself.
