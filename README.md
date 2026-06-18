<div align="center">

# Ark CLI

**火山方舟（Volcengine Ark）官方命令行工具**

在终端里一句话**对话 · 生图 · 生视频 · 多模态理解 · 部署模型** —— 还能把这些能力一键装进你的 AI Agent。

[![npm](https://img.shields.io/npm/v/@volcengine/ark-cli?color=brightgreen&label=npm)](https://www.npmjs.com/package/@volcengine/ark-cli)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue)](./LICENSE)
[![node](https://img.shields.io/badge/node-%E2%89%A516-brightgreen)](https://nodejs.org)

```shell
npm i -g @volcengine/ark-cli
```

</div>

---

## ✨ 一分钟感受

```shell
# 对话推理
arkcli +chat "用一句话解释什么是 RAG"

# 文生图 —— 产物自动下载并打开
arkcli +gen "赛博朋克风格的城市夜景，霓虹灯，雨夜"

# 看看现在能用哪些模型
arkcli models list
```

生视频、图生图、语音转写、文档抽取、视频总结、模型部署…… 都是一条命令的事。

## 为什么用它

- 🚀 **一条命令，不写代码**：生图 / 对话 / 转写 / 部署，终端直接出结果，不用翻 SDK、不用搭环境。
- 🎨 **图片 + 视频 + 多模态全覆盖**：Seedream 生图、Seedance 生视频，OCR / 视觉定位 / 视频总结 / 语音转写一站式。
- 🤖 **装进你的 AI Agent**：一句 `arkcli +connect`，把上面所有能力作为 Skill 注入 Claude Code / OpenCode，让 Agent 用自然语言直接驱动火山方舟。
- 🔌 **一套工具多种场景**：platform / agent-plan / coding-plan 多 profile，一致体验、随手切换。

## 能做什么

| 能力 | 命令 | 说明 |
|------|------|------|
| 💬 对话推理 | `arkcli +chat` | 多模态对话（文本 / 图片 / 视频 / 音频），流式输出、多轮接续 |
| 🎨 生图生视频 | `arkcli +gen` | Seedream 文生图 / 图生图、Seedance 文生视频 / 图生视频，产物自动下载 |
| 🔍 多模态理解 | `arkcli +understand` | OCR、视觉定位（框选）、文档抽取、视频总结、语音转写 / 翻译、字幕打轴 |
| 🚀 部署模型 | `arkcli +deploy` · `arkcli infer endpoint` | 一键把模型部署成在线推理接入点（Endpoint） |
| 📋 模型 / 用量 / 账单 | `arkcli models` · `arkcli usage` · `arkcli billing` | 查公共基础模型、推理用量、结算账单 |
| 🤖 注入 AI Agent | `arkcli +connect` | 把上面这些能力作为 Skill 装进本机的 AI Agent |

> 每个命令都自带 `--help`，例如 `arkcli +gen --help`。

## 快速开始

**环境要求**：Node.js ≥ 16（自带 npm / npx）·支持 macOS / Linux / Windows（arm64 / amd64）

```shell
# 1. 安装
npm config set registry https://registry.npmjs.org/
npm i -g @volcengine/ark-cli@latest
arkcli --version

# 2. 登录（火山引擎 SSO，运行后在浏览器完成授权）
arkcli auth login volc-sso
arkcli auth status

# 3. 上手
arkcli +chat "你好，介绍下你自己"
```

> **无浏览器环境**（CI / 远程开发机 / Agent 沙箱）：`arkcli auth login --no-browser`，在任意设备完成授权后用 `--code <码>` 粘回。

## 把能力装进你的 AI Agent

```shell
arkcli +connect
```

`+connect` 会自动检测本机的 AI Agent（Claude Code / OpenCode 等）并安装 Ark Skills。装好后，你在 Agent 里用**自然语言**就能驱动上面所有能力——“帮我生成一张图”“把这段录音转写成文字”“把某个模型部署成 Endpoint”。

## 浏览内置 Skills

本仓库的 [`skills/`](https://github.com/volcengine/ark-cli/tree/main/skills) 收录了 arkcli 全部内置 AI Agent Skills（对话 / 生成 / 理解 / 部署 / 模型 / 用量 / 账单 / 套餐 …），每个都是一份可读的 `SKILL.md`。`arkcli +connect` 装进你 Agent 的就是它们。

## 链接

- 🌋 火山方舟控制台 & 文档：<https://www.volcengine.com/product/ark>
- 📦 npm：<https://www.npmjs.com/package/@volcengine/ark-cli>

## License

[Apache-2.0](./LICENSE)
