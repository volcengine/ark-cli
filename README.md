# Ark CLI · `@volcengine/ark-cli`

火山方舟（Volcengine Ark）官方命令行工具 —— 在终端里直接**对话推理、生图生视频、多模态理解、部署模型**，并能把这些能力作为 **Skill** 一键注入到你的 AI Agent（Claude Code / OpenCode 等）。

[![npm](https://img.shields.io/npm/v/@volcengine/ark-cli?color=brightgreen)](https://www.npmjs.com/package/@volcengine/ark-cli)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue)](./LICENSE)

---

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

## 环境要求

- **Node.js ≥ 16**（自带 npm / npx）
- 支持 macOS、Linux、Windows（arm64 / amd64）

## 第 1 步 · 安装

```shell
npm config set registry https://registry.npmjs.org/
npm i -g @volcengine/ark-cli@latest

# 查看装上的版本
arkcli --version
```

## 第 2 步 · 登录（火山引擎 SSO）

运行后会输出一个授权链接，在浏览器完成 SSO 授权即可：

```shell
arkcli auth login volc-sso
```

> **无浏览器的环境**（CI / 远程开发机 / Agent 沙箱）：加 `--no-browser`，在任意有浏览器的设备完成授权后，把页面显示的 base64 授权码用 `--code <码>` 粘回完成登录。

## 第 3 步 · 验证

```shell
arkcli auth status
```

## 第 4 步 · 把能力装进你的 AI Agent

```shell
arkcli +connect
```

`+connect` 会自动检测本机的 AI Agent（Claude Code / OpenCode 等）并安装 Ark Skills。装好后，你在 Agent 里用**自然语言**就能驱动上面所有能力（“帮我生成一张图”“把这段录音转写成文字”“把某个模型部署成 Endpoint”）。

## 快速体验

```shell
# 一句话对话
arkcli +chat "用一句话解释什么是 RAG"

# 生成一张图（产物自动下载并打开）
arkcli +gen "赛博朋克风格的城市夜景，霓虹灯，雨夜"

# 看看当前可用的模型
arkcli models list
```

## 链接

- 火山方舟产品与文档：<https://www.volcengine.com/product/ark>
- npm 包：<https://www.npmjs.com/package/@volcengine/ark-cli>

## License

[Apache-2.0](./LICENSE)
