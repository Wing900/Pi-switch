<p align="center">
  <img src="frontend/src/assets/pi-logo.png" width="160" alt="Pi Switch">
</p>

<h1 align="center">Pi Switch</h1>

<p align="center">
  <a href="https://github.com/Wing900/Pi-switch/releases"><img src="https://img.shields.io/github/v/release/Wing900/Pi-switch?style=flat-square" alt="Version"></a>
  <a href="https://github.com/Wing900/Pi-switch/blob/master/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue?style=flat-square" alt="License"></a>
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.22-00ADD8?style=flat-square&logo=go" alt="Go"></a>
  <a href="https://wails.io"><img src="https://img.shields.io/badge/Wails-v2-FF6B6B?style=flat-square" alt="Wails"></a>
  <a href="https://github.com/Wing900/Pi-switch"><img src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey?style=flat-square" alt="Platform"></a>
</p>

## 介绍

Pi Switch 是一个跨平台的 Pi Agent 模型服务商配置工具，基于 Wails 构建。支持管理多个 Provider、获取可用模型、设置默认模型并一键启动 Pi Agent。

## 致谢

- [Pi](https://github.com/earendil-works/pi) - Pi 项目
- [Linux.do](https://linux.do) - 社区支持

## 参考项目

- [pi-model-manager](https://github.com/Qihuanxishini/pi-model-manager) - Provider 配置与自定义请求头实现参考

## 功能特性

- 多 Provider 管理（预设：DeepSeek / OpenAI / Anthropic），支持任意 OpenAI 兼容服务
- Anthropic 原生 API 适配：`anthropic-messages` 模式，通过 `/v1/models` + `x-api-key` 拉取模型
- 模型列表拉取 / 手动导入 / 一键启动 Pi Agent

## 开发

待完善中

## Linux 运行依赖

Linux 版本依赖 GTK3 和 WebKitGTK。发布页同时提供两种 WebKitGTK ABI 构建产物，请按系统选择：

| 系统 | ABI | 发布文件后缀 | 运行依赖 |
| --- | --- | --- | --- |
| Debian 12/13+、Ubuntu 22.04+ | 4.1 | `linux-amd64-webkit2-41` | `sudo apt install libgtk-3-0 libwebkit2gtk-4.1-0` |
| Debian 11、Ubuntu 20.04 | 4.0 | `linux-amd64-webkit2-40` | `sudo apt install libgtk-3-0 libwebkit2gtk-4.0-37` |

例如 Debian 13 请选择 `PiSwitch-<version>-linux-amd64-webkit2-41.tar.gz`。
