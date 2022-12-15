---
title: "Managing users dynamically using the API"
draft: false
weight: 10
---

### Note that Trojan does not support this feature

Trojan-Go provides a set of APIs using gRPC, which supports the following features.

- user information addition, deletion, and checking

- traffic statistics

- speed statistics

- IP connection statistics

Trojan-Go itself has integrated API control, i.e. you can use one Trojan-Go instance to control another Trojan-Go server.

You need to add API settings to the configuration of the server you need to be controlled, e.g.

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

Then start the Trojan-Go server

```shell
. /trojan-go -config . /server.json
```

You can then use another Trojan-Go to connect to that server for administration, with the basic command format

```shell
. /trojan-go -api-addr SERVER_API_ADDRESS -api COMMAND
```

where ```SERVER_API_ADDRESS``` is the API address and port, e.g. 127.0.0.1:10000

```COMMAND``` is the API command, the legal commands are

- list List all users

- get to get information about a user

- set set the information of a user (add/remove/modify)

Here are some examples

1. list all user information

    ```shell
    . /trojan-go -api-addr 127.0.0.1:10000 -api list
    ```

    All user information will be exported as json, the information includes the number of online IPs, live speed, total upload and download traffic, etc. Here is an example of the returned results

    ```json
    [{"user":{"hash":"d63dc919e201d7bc4c825630d2cf25fdc93d4b2f0d46706d29038d01"},"status":{"traffic_total":{"upload_traffic":36393,"download_traffic":186478},"speed_current":{"upload_speed":25210,"download_speed":72384},"speed_limit":{"upload_speed":5242880,"download_speed":5242880},"ip_limit":50}}]
    ```

    All traffic units are bytes.

2. Get a user's information

    You can use -target-password to specify the password, or -target-hash to specify the SHA224 hash of the target user's password. The format is the same as the list command

    ```shell
    . /trojan-go -api-addr 127.0.0.1:10000 -api get -target-password password
    ```

    or

    ```shell
    . /trojan-go -api-addr 127.0.0.1:10000 -api get -target-hash d63dc919e201d7bc4c825630d2cf25fdc93d4b2f0d46706d29038d01
    ```

    The above two commands are equivalent. The following example uses the plaintext password in a uniform way, and the hash specifies a certain user in a similar way.

    The user information will be exported in the form of json in a format similar to the list command. Here is an example of the returned result

    ```json
    {"user":{"hash":"d63dc919e201d7bc4c825630d2cf25fdc93d4b2f0d46706d29038d01"},"status":{"traffic_total":{"upload_traffic":36393,"download_traffic":186478},"speed_current":{"upload_speed":25210,"download_speed":72384},"speed_limit":{"upload_speed":5242880,"download_speed":5242880},"ip_limit":50}}
    ```

3. add a user information

    ```shell
    . /trojan-go -api-addr 127.0.0.1:10000 -api set -add-profile -target-password password
    ```

4. delete a user information

    ```shell
    . /trojan-go -api-addr 127.0.0.1:10000 -api set -delete-profile -target-password password
    ```

5. modify a user information

    ```shell
    . /trojan-go -api-addr 127.0.0.1:10000 -api set -modify-profile -target-password password \
        -ip-limit 3 \
        -upload-speed-limit 5242880 \
        -download-speed-limit 5242880
    ```

    This command limits the upload and download speed to 5MiB/s for users with password, and the number of connected IPs is limited to 3. Note that the unit of 5242880 here is bytes. If you fill in 0 or a negative number, it means no limit is applied.
