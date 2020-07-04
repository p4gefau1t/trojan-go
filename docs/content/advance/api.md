---
title: "使用API动态管理用户"
draft: false
weight: 10
---

### 注意，Trojan不支持这个特性

Trojan-Go使用gRPC提供了一组API，API支持以下功能：

- 用户信息增删改查

- 流量统计

- 速度统计

- IP连接数统计

Trojan-Go本身集成了API控制功能，也即可以使用一个Trojan-Go实例控制另一个Trojan-Go服务器。

你需要在你需要被控制的服务端配置添加API设置，例如：

```json
{
    ...
    "api": {
        "enabled": true,
        "api_addr": "127.0.0.1",
        "api_port": 10000,
    }
}
```

然后启动Trojan-Go服务器

```shell
./trojan-go -config ./server.json
```

然后可以使用另一个Trojan-Go连接该服务器进行管理，基本命令格式为

```shell
./trojan-go -api-addr SERVER_API_ADDRESS -api COMMAND
```

其中```SERVER_API_ADDRESS```为API地址和端口，如127.0.0.1:10000

```COMMAND```为API命令，合法的命令有

- list 列出所有用户

- get 获取某个用户信息

- set 设置某个用户信息（添加/删除/修改）

下面是一些例子

1. 列出所有用户信息

    ```shell
    ./trojan-go -api-addr 127.0.0.1:10000 -api list
    ```

    所有的用户信息将以json的形式导出，信息包括在线IP数量，实时速度，总上传和下载流量等。下面是一个返回的结果的例子

    ```json
    [{"user":{"hash":"d63dc919e201d7bc4c825630d2cf25fdc93d4b2f0d46706d29038d01"},"status":{"traffic_total":{"upload_traffic":36393,"download_traffic":186478},"speed_current":{"upload_speed":25210,"download_speed":72384},"speed_limit":{"upload_speed":5242880,"download_speed":5242880},"ip_limit":50}}]
    ```

    流量单位均为字节。

2. 获取一个用户信息

    可以使用 -target-password 指定密码，也可以使用 -target-hash 指定目标用户密码的SHA224散列值。格式和list命令相同

    ```shell
    ./trojan-go -api-addr 127.0.0.1:10000 -api get -target-password password
    ```

    或者

    ```shell
    ./trojan-go -api-addr 127.0.0.1:10000 -api get -target-hash d63dc919e201d7bc4c825630d2cf25fdc93d4b2f0d46706d29038d01
    ```

    以上两条命令等价，下面的例子统一使用明文密码的方式，散列值指定某个用户的方式以此类推。

    该用户信息将以json的形式导出，格式与list命令类似。下面是一个返回的结果的例子

    ```json
    {"user":{"hash":"d63dc919e201d7bc4c825630d2cf25fdc93d4b2f0d46706d29038d01"},"status":{"traffic_total":{"upload_traffic":36393,"download_traffic":186478},"speed_current":{"upload_speed":25210,"download_speed":72384},"speed_limit":{"upload_speed":5242880,"download_speed":5242880},"ip_limit":50}}
    ```

3. 添加一个用户信息

    ```shell
    ./trojan-go -api-addr 127.0.0.1:10000 -api set -add-profile -target-password password
    ```

4. 删除一个用户信息

    ```shell
    ./trojan-go -api-addr 127.0.0.1:10000 -api set -delete-profile -target-password password
    ```

5. 修改一个用户信息

    ```shell
    ./trojan-go -api-addr 127.0.0.1:10000 -api set -modify-profile -target-password password \
        -ip-limit 3 \
        -upload-speed-limit 5242880 \
        -download-speed-limit 5242880
    ```

    这个命令将密码为password的用户上传和下载速度限制为5MiB/s，同时连接的IP数量限制为3个，注意这里5242880的单位是字节。如果填写0或者负数，则表示不进行限制。
