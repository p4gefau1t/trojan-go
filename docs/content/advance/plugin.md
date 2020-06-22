---
title: "使用Shadowsocks插件/可插拔传输层"
draft: false
weight: 7
---

### 注意，Trojan-GFW版本不支持这个特性

Trojan-Go支持可插拔的传输层。原则上，Trojan-Go可以使用任何有TCP隧道功能的软件作为传输层，如v2ray、shadowsocks、kcp等。同时，Trojan-Go也兼容Shadowsocks的SIP003插件标准，如GoQuiet，v2ray-plugin等。也可以使用Tor的传输层插件，如obfs4，meek等。

你可以使用这些插件，替换Trojan-Go的TLS传输层。

开启可插拔传输层插件后，Trojan-Go客户端将会把**流量明文**直接传输给客户端本地的插件处理。由客户端插件负责进行加密和混淆，并将流量传输给服务端的插件。服务端的插件接收到流量，进行解密和解析，将**流量明文**传输给服务端本地的Trojan-Go服务端。

你可以使用任何插件对流量进行加密和混淆，只需添加"transport_plugin"选项，并指定插件的可执行文件的路径，并做好配置即可。

我们更建议**自行设计协议并开发相应插件**。因为目前现有的所有插件无法对接Trojan-Go的对抗主动探测的特性，而且部分插件并无加密能力。如果你对开发插件有兴趣，欢迎在"实现细节和开发指南"一节中查看插件设计的指南。

例如，可以使用符合SIP003标准的v2ray-plugin，下面是一个例子:

**这个配置中使用了websocket明文传输未经加密的trojan协议，存在安全隐患。这个配置仅作为演示使用。**

**不要在任何情况下使用这个配置穿透GFW。**

服务端配置：

```json
...（省略）
"transport_plugin": {
    "enabled": true,
    "type": "shadowsocks",
    "command": "./v2ray-plugin",
    "arg": ["-server", "-host", "www.baidu.com"]
}
```

客户端配置：

```json
...（省略）
"transport_plugin": {
    "enabled": true,
    "type": "shadowsocks",
    "command": "./v2ray-plugin",
    "arg": ["-host", "www.baidu.com"]
}
```

注意，v2ray-plugin插件需要指定```-server```参数来区分客户端和服务端。更多关于该插件详细的说明，参考v2ray-plugin的文档。

启动Trojan-Go后，你可以看到v2ray-plugin启动的输出。插件将把流量伪装为Websocket流量并传输。

非SIP003标准的插件可能需要不同的配置，你可以指定```type```为"other"，并自行指定插件地址，插件启动参数、环境变量。
