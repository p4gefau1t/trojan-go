---
title: "基本介绍"
draft: false
weight: 1
---

Trojan-Go的核心部分有

- protocol 各个协议具体实现

- proxy 代理核心，使用protocol的协议实现，处理入站和出站流量

- conf 配置解析模块

- shadow 主动检测欺骗模块

- stat 用户认证和统计模块

可以在对应文件夹中找到相关源代码。
