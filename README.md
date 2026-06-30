<div align="center">

# Ark CLI

**给你的 AI Agent 赋予火山方舟的全部能力**

[![npm](https://img.shields.io/npm/v/@volcengine/ark-cli?color=brightgreen&label=npm)](https://www.npmjs.com/package/@volcengine/ark-cli)
[![license](https://img.shields.io/badge/license-Apache--2.0-blue)](./LICENSE)
[![node](https://img.shields.io/badge/node-%3E%3D16-brightgreen)](https://nodejs.org)

火山方舟模型、生成、多模态理解、Plan、精调、部署、用量、账单和诊断能力，都可以通过 Ark CLI + Skills 交给 Claude Code、OpenCode、Codex 等本地 Agent 调用。

```shell
npm i -g @volcengine/ark-cli
```

</div>

---

## Ark CLI 能做什么

让你的 AI Agent 具备这些方舟能力，并能在复杂任务中组合调用：

- **模型调用**：文本对话、多模态输入、流式输出、多轮接续
- **图片 / 视频生成**：Seedream 生图、Seedance 生视频，产物自动下载到本地
- **多模态理解**：OCR、视频总结、语音转写、字幕打轴、文档字段抽取
- **Plan 管理**：Agent Plan / Coding Plan 查询、购买、续费、团队席位管理
- **Agent 接入**：安装 Ark Skills，配置 Harness / MCP，让本地 Agent 调用方舟能力
- **模型生产**：精调任务、自定义模型、训练状态跟踪
- **服务上线**：基础模型或自定义模型一键部署为 Endpoint
- **治理闭环**：Usage、Billing、Doctor、监控指标和错误码诊断

## 为什么需要 Ark CLI

Ark CLI 不是让你记住更多命令，而是把火山方舟 MaaS 平台能力变成 Agent 可以调用的工具箱。

```text
消费链路:
选择模型
  -> 开通 Plan 或部署模型
  -> 接入本地 Agent / Harness
  -> 调用、理解、生成
  -> 查看用量、账单、监控和诊断结果

生产链路:
准备数据
  -> 创建精调任务
  -> 得到自定义模型
  -> 部署成 Endpoint
  -> 回到消费链路继续使用
```

## 三步开始

```shell
# 1. 安装
npm i -g @volcengine/ark-cli@latest
arkcli --version

# 2. 登录火山方舟
arkcli auth login volc-sso
arkcli auth status

# 3. 把 Ark Skills 安装进本地 Agent
arkcli +connect
```

无浏览器环境，例如远程开发机、CI 或 Agent 沙箱：

```shell
arkcli auth login --no-browser
arkcli auth login --no-browser --code <授权码>
```

## 一次完整任务可以这样开始

你可以直接告诉 Agent 一个目标：

```text
帮我选择一个适合图文理解的模型，配置到本地 Agent，
然后用它总结一段视频，并告诉我本月用了多少 token。
```

Agent 可以通过 Ark CLI 串起完整流程：

```text
arkcli models search
  -> arkcli plans get / arkcli plans model-list
  -> arkcli +connect / arkcli helper configure
  -> arkcli +understand video-summary
  -> arkcli usage stats / arkcli billing list
```

用户描述目标，Agent 负责选择命令、调用方舟能力、汇总结果。

## 常用任务

### 选择并调用模型

```shell
arkcli +chat "总结一下 RAG 的核心思想"

arkcli +chat --input @photo.jpg "描述这张图片"

arkcli +understand video-summary --input @demo.mp4 "按章节总结这个视频"

arkcli +gen "生成一张未来城市海报"
```

### 开通和管理 Plan

```shell
# 查看当前账号持有哪些套餐
arkcli plans get

# 查看 Plan 支持哪些模型
arkcli plans model-list --plan agent-plan

# 查看团队版席位
arkcli plans team seat-list --plan agent-plan-team

# 查看本机 Agent 上 Plan 相关能力状态
arkcli plans harness-status

# 购买或续费前先查看参数
arkcli plans buy --help
arkcli plans renew --help
```

`plans buy` / `plans renew` 是计费操作，Ark CLI 会先展示协议和确认步骤，不会默认直接扣款。

### 接入方舟服务能力

`+connect` 安装 Ark Skills，让 Agent 知道有哪些方舟任务可以做。`helper` 配置 Harness / MCP / Plan，让 Agent 真正接上模型和工具能力。

```shell
# 查看会检测到哪些本地 Agent
arkcli +connect list

# 安装 Ark Skills
arkcli +connect

# 清理已注入的 Ark Skills
arkcli +connect uninstall

# 查看 Harness / Agent 配置状态
arkcli helper list

# 给当前 Agent 注入 MCP 工具能力
arkcli helper mcp

# 配置指定 Harness，并同步注入 MCP
arkcli helper configure codex --with-mcp
```

`arkcli +connect uninstall` 会清理注入到本地 Agent 的 `ark-` / `arkcli-` 前缀 Skills，适合想恢复干净 Agent 环境、或担心 Skills 过多影响 Agent 路由时使用。它不会卸载 `arkcli` 命令本身。

### 生产模型

```shell
arkcli train finetune --help
arkcli train finetune capability get --model <model> --version <version>
arkcli train finetune create --help
arkcli train finetune watch <job-id>
```

### 上线服务

```shell
arkcli +deploy --name my-endpoint --model doubao-seed-2-0-pro-260215 --dry-run
arkcli +deploy --name my-endpoint --model doubao-seed-2-0-pro-260215
arkcli infer endpoint list
arkcli infer endpoint get <endpoint-id>
```

### 管成本、状态和健康

```shell
arkcli usage plan
arkcli usage stats --start 2026-07-01 --mine
arkcli billing list --start 2026-07
arkcli billing list --start 2026-07 --endpoint ep-xxx

arkcli doctor
arkcli doctor model <model-name>
arkcli doctor infer-endpoint <endpoint-id>
arkcli doctor metrics request.qpm
arkcli doctor error ContentRiskBlocked
```

`doctor` 默认是只读诊断，不会修改资源。需要修复时会通过确认流程执行。

## 命令入口

| 任务 | 命令 |
|---|---|
| 对话和多模态推理 | `arkcli +chat` |
| 图片 / 视频生成 | `arkcli +gen` |
| OCR、视频总结、语音转写、文档抽取 | `arkcli +understand` |
| 模型查询 | `arkcli models` |
| Plan / 套餐管理 | `arkcli plans` |
| Agent 接入和 Harness 配置 | `arkcli +connect`, `arkcli helper` |
| 精调任务 | `arkcli train finetune` |
| 一键部署 Endpoint | `arkcli +deploy` |
| Endpoint 管理 | `arkcli infer endpoint` |
| 用量查询 | `arkcli usage` |
| 账单查询 | `arkcli billing` |
| 诊断和自助排障 | `arkcli doctor` |

每个命令都自带帮助：

```shell
arkcli +chat --help
arkcli +gen --help
arkcli +understand --help
arkcli +deploy --help
arkcli plans --help
arkcli doctor --help
```

## 认证

推荐使用火山 SSO：

```shell
arkcli auth login volc-sso
```

登录后 Ark CLI 会自动完成身份、profile 和 API Key 配置。你通常不需要手动复制密钥。

```shell
arkcli auth status
arkcli auth apikey
arkcli auth logout
```

## 配置

```shell
# 查看当前 profile
arkcli profile show

# 切换 profile
arkcli profile use <profile-name>

# 查看配置
arkcli config show
```

## AI Agent Skills

本仓库的 [`skills/`](https://github.com/volcengine/ark-cli/tree/main/skills) 收录了 Ark CLI 内置的 Agent Skills。执行：

```shell
arkcli +connect
```

Ark CLI 会把这些 Skills 安装到本机 Agent，使 Agent 能通过自然语言调用火山方舟能力，包括对话、多模态理解、图片视频生成、模型查询、Plan 管理、Endpoint 部署、精调、用量、账单和诊断。

## 用户交流群

扫码加入 Ark CLI 飞书用户交流群，获取使用答疑、问题排查、Bug 反馈和使用经验交流支持。你可以在群里咨询安装、登录、模型调用、Skill 使用、版本升级等问题，也欢迎反馈 GitHub Issues 中未覆盖的异常。

请勿在群内发送 API Key、AK/SK、Token 等敏感信息。对于可稳定复现的 Bug，建议同步提交 GitHub Issue，便于跟踪和修复。

<img src="./assets/lark-user-group.png" width="360" alt="Ark CLI 飞书用户交流群二维码" />

## 链接

- 火山方舟：<https://www.volcengine.com/product/ark>
- npm：<https://www.npmjs.com/package/@volcengine/ark-cli>
- GitHub Releases：<https://github.com/volcengine/ark-cli/releases>
- 内置 Skills：<https://github.com/volcengine/ark-cli/tree/main/skills>

## Changelog

Release notes live in [GitHub Releases](https://github.com/volcengine/ark-cli/releases).

## License

[Apache-2.0](./LICENSE)
