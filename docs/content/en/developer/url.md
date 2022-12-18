---
title: "URL scheme (draft)"
draft: false
weight: 200
---

## Changelog

- encryption format to ss;method:password

## Overview

Thanks to @DuckSoft @StudentMain @phlinhng for discussing and contributing to the Trojan-Go URL scheme. **The URL scheme is currently a draft and needs more practice and discussion.**

Trojan-Go **client** can accept URLs to locate server resources. The principles are as follows:

- Comply with URL format specifications

- Ensure human readability and machine friendliness

- The purpose of URLs is to locate Trojan-Go node resources and facilitate resource sharing

Note that embedding encoded data such as base64 in URLs is prohibited for human readability reasons. First, base64 encoding does not guarantee secure transmission, but is meant to transmit non-ASCII data over ASCII channels. Second, if you need to secure transmission when sharing URLs, encrypt the plaintext URLs instead of modifying the URL format.

## Format

The basic format is as follows, `$()` means that `encodeURIComponent` is required here.

```text
trojan-go://
    $(trojan-password)
    @
    trojan-host
    :
    port
/?
    sni=$(tls-sni.com)&
    type=$(original|ws|h2|h2+ws)&
        host=$(websocket-host.com)&
        path=$(/websocket/path)&
    encryption=$(ss;aes-256-gcm;ss-password)&
    plugin=$(...)
#$(descriptive-text)
```

For example

```text
trojan-go://password1234@google.com/?sni=microsoft.com&type=ws&host=youtube.com&path=%2Fgo&encryption=ss%3Baes-256-gcm%3Afuckgfw
```

Since Trojan-Go is compatible with Trojan, the URL scheme for Trojan

```text
trojan://password@remote_host:remote_port
```

is compatible and acceptable. It is equivalent to

```text
trojan-go://password@remote_host:remote_port
```

Note that once a server uses a non-Trojan compatible feature, you must use ```trojan-go://``` to locate the server. This is designed so that Trojan-Go URLs are not incorrectly accepted by Trojan, and to avoid contaminating Trojan users with URL sharing. At the same time, Trojan-Go ensures compatibility with accepting Trojan's URLs.

## Details

Note: All parameter names and constant strings are case-sensitive.

### `trojan-password`

Trojan's password.
Cannot be omitted, cannot be an empty string, and is not recommended to contain non-ASCII printable characters.
Must be encoded with `encodeURIComponent`.

### `trojan-host`

Node IP / domain name.
Cannot be omitted and cannot be an empty string.
IPv6 address must be enclosed in square brackets.
IDN domain name (e.g. "Baidu.cn") must be in `xn--xxxxxx` format.

### `port`

Node port.
Default is `443` when omitted.
Must be an integer in `[1,65535]`.

### `tls` or `allowInsecure`

does not have this field.
TLS is always enabled by default, unless a transport plugin disables it.
TLS authentication must be enabled. Nodes that cannot use root CA to verify the identity of the server are not suitable for sharing.

### `sni`

SNI for custom TLS.
Defaults to the same value as `trojan-host` when omitted. Must not be an empty string.

Must be encoded with `encodeURIComponent`.

### `type`

The type of the transfer.
Defaults to `original` when omitted, but may not be an empty string.
Currently the only available values are `original` and `ws`, in the future there may be `h2`, `h2+ws`, etc.

When the value is `original`, the original Trojan transfer method is used and cannot be easily passed through CDN.
When the value is `ws`, Websocket over TLS is used.

### `host`

Custom HTTP `Host` header.
Can be omitted, when omitted the value is the same as `trojan-host`.
Can be an empty string, but may introduce unintended situations.

Warning: If your port is not a standard port (not 80 / 443), the RFC standard states that `Host` should be followed by the port number, e.g. `example.com:44333`. Please use your own discretion as to whether to comply.

The `encodeURIComponent` encoding must be used.

### `path`

Valid when transfer type `type` takes `ws`, `h2`, `h2+ws`.
It may not be omitted and may not be empty.
Must start with `/`.
You can use `&`, `#`, `? ` etc., but it must be a legal URL path.

Must be encoded with `encodeURIComponent`.

### `mux`

does not have this field.
The current server always supports `mux` by default.
There are advantages and disadvantages to enabling `mux` or not, and it is up to the client to decide whether to enable it or not; the purpose of the URL is to locate server resources, not to specify user preferences.

### `encryption`

Encryption layer for securing Trojan traffic cryptographically.
Can be omitted, defaults to `none`, i.e. no encryption is used.
Cannot be an empty string.

Must use encodeURIComponent encoding.

When using the Shadowsocks algorithm for traffic encryption, the format is

```text
ss;method:password
```

where ss is the fixed content and method is the encryption method, which must be one of the following.

- `aes-128-gcm`
- `aes-256-gcm`
- `chacha20-ietf-poly1305`

where `password` is the Shadowsocks password and must not be an empty string.
`password` does not need to be escaped if it contains a semicolon.
`password` should be an English printable ASCII character.

Other encryption schemes to be determined.

### `plugin`

Additional plugin options. This field is reserved.
May be omitted, but may not be an empty string.

### URL Fragment (# followed by content)

Node description.
Not recommended to be omitted, not recommended to be an empty string.

Must be encoded with `encodeURIComponent`.
