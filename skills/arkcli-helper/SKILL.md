---
name: arkcli-helper
version: 1.0.0
description: "arkcli helper:把本机 AI Agent(Claude Code / OpenCode / OpenClaw / Trae)配置到火山方舟 Plan,核心能力是给 Agent 注入 Agent Plan 内置 MCP(联网搜索 mcp-server-askecho-search-infinity + dataPro-search + OpenViking ov-mcp-server)。当用户说『给当前/某个 Agent 配上(我的)MCP / 联网搜索 / dataPro』『把 MCP 装进 Claude Code/OpenCode/OpenClaw/Trae』『设置 Agent Plan 的 MCP』时,用 `arkcli helper mcp`(只注入 MCP、不改 model)。也覆盖 `helper configure`(连带配 model/provider)、`helper list`(查状态)、`helper reset`(移除注入)。反触发:把 arkcli skills 安装进 agent → arkcli-connect;登录/401/鉴权 → arkcli-auth;生图生视频 → arkcli-gen。注意:MCP 注入仅 Agent Plan(含团队版 agent-plan-team)支持,只能配 claude-code/opencode/openclaw/trae(Trae 是 IDE,仅注 MCP、不配 model,支持 --scope project 写项目级 ./.trae/mcp.json)。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli helper mcp --help"
---

# arkcli helper —— 给本机 AI Agent 配置 Plan / 注入 MCP

**前置:** 先用 Read 读 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md) 获取共享安全规则与认证闸门。

把 Agent Plan 内置的三台 MCP server 注入本机 AI Agent 的配置 —— 这正是 `arkcli helper` 交互向导里"注入 MCP"那一步,这里做成**非交互、可被 prompt 触发**。

## 注入的是哪三台 MCP(写死,勿幻觉)

| server | 传输 | key 来源 |
|--------|------|---------|
| `mcp-server-askecho-search-infinity` | stdio(`uvx`) | 联网搜索专属 key,自动从 `ListMCPAPIKeys` 取(**账号级**);取不到则写占位符,提示用户填 |
| `dataPro-search` | http(streamable) | **Agent Plan 的 API Key**(header `X-Agent-Plan-Key`,裸 key) |
| `ov-mcp-server` | http(streamable) | **OpenViking 库的访问 key**(header `Authorization: Bearer <key>`);经 vikingdb 两步取:列库 → 按库取 key。账号多库时要选库(见下) |

## host ≠ target(最关键的概念)

- **host** = 你(这个 AI Agent)此刻跑在哪 —— 命令读环境变量自动检测,无需你判断。
- **target** = 要把 MCP 写进谁的配置 —— 可以是 host 自己,也可以是另一个 agent。
- 二者解耦:人在 OpenCode 里,也能给 Claude Code 配 MCP。

→ 用户在 prompt 里**点名了某个 agent**(如"给 opencode 配"):跑 `arkcli helper mcp opencode`
→ 用户说"**当前 / 这个 Agent**"或没点名:跑 `arkcli helper mcp`(自动检测当前 host)

## 子命令穷举

| 调用 | 说明 |
|------|------|
| `arkcli helper mcp [target] [--ov-resource <库名>] [--scope project]` | **只注入 MCP,不改 model**。不传 target 自动检测当前 agent;账号多个 OpenViking 库时用 `--ov-resource` 指定;`--scope project`(仅 Trae)写项目级 `./.trae/mcp.json` |
| `arkcli helper configure <harness> [--with-mcp]` | 配 model/provider 指向 plan(可选连带注入 MCP) |
| `arkcli helper list` | 查支持的 agent + 安装/配置状态(只读) |
| `arkcli helper reset <harness>` | 移除 arkcli 注入的配置(含 MCP) |
| `arkcli helper` | TTY 交互向导(需终端;非交互场景改用上面的) |

> ⚠️ 想"只加 MCP" → `helper mcp`;想"把 agent 接到 plan、连模型一起" → `helper configure --with-mcp`。别用 `configure` 去只加 MCP(它会一并(重)写 model)。

## 范围边界(管好,别越界)

- **可注入 target 有 4 个**:`claude-code` / `opencode` / `openclaw` / `trae`。
- 本 skill 会被 `arkcli +connect` 装进很多 agent(cursor / gemini-cli / codex …40+),但**只能配上面这 4 个**。host 是其它 agent 时:要么用户点名其一作 target,要么命令会报"请显式指定" —— **绝不静默配错对象**。
- `trae`(字节 AI IDE)是 MCP-only:只注入 MCP、不配 model/provider;**无运行态宿主检测**(不会被自动当成 host),只能显式 `arkcli helper mcp trae`。默认写用户级 `~/.trae/mcp.json`,加 `--scope project` 写项目级 `./.trae/mcp.json`(项目级需在 Trae「设置 → MCP」开启「启用项目级 MCP」开关 + 重开项目)。
- `hermes` 暂不支持 MCP 注入 → 命中就直说"暂不支持"。

## 前提

- **必须有 Agent Plan 订阅**(dataPro / ov 要 Agent Plan 的 key)。命令自动定位账号下的 Agent Plan profile,**与当前 active profile 无关**;**个人版 `agent-plan` 与团队版 `agent-plan-team` 完全等价**,两者都能注入同样的三台 MCP。没有就引导 `arkcli auth login` 开通;账号同时有多个 Agent Plan profile(如个人版 + 团队版)时让用户用 `--profile` 指定。
- 注入后 **agent 需重启**才会加载新 MCP。Trae 还需去「设置 → MCP」面板确认 MCP 已启用(项目级文件额外要开「启用项目级 MCP」开关)后重开项目。

## OpenViking 库的选择(ov-mcp-server 专属)

`ov-mcp-server` 的 key 绑定到某个 OpenViking 库(库名 ↔ ResourceID 1:1)。命令先列账号下的库,按数量分流:

- **0 个库** → 自动跳过 ov-mcp-server(仍注入另两台),并提示去 `https://console.volcengine.com/vikingdb/openviking/region:openviking+cn-beijing/create` 建库后重跑。可直接接受跳过。
- **1 个库** → 直接用,无需选择。
- **多个库** → 命令报错并列出所有库名(形如 `检测到多个 OpenViking 库,请用 --ov-resource <库名> 指定其一:[a, b, c]`)。**此时用 AskUserQuestion 把这些库名作为选项让用户选**,拿到选定库名后带 `--ov-resource <库名>` 重跑同一条命令。
- 取 key 失败(非 0 库)→ ov-mcp-server 写占位符 `Bearer <OPENVIKING_KEY>`,提示用户手动替换。

## 路由判断 / 反触发

- "给 Agent 配 MCP / 联网搜索 / dataPro / web search" → `arkcli helper mcp`
- "把 agent 指向我的 plan(连模型一起)" → `arkcli helper configure`
- 把 arkcli skills **安装**进 agent → 走 [arkcli-connect](../arkcli-connect/SKILL.md),与本 skill 无关
- 401 / 登录 / 鉴权失败 → 走 [arkcli-auth](../arkcli-auth/SKILL.md)
- 生图 / 生视频 → 走 [arkcli-gen](../arkcli-gen/SKILL.md)

详细 flag、输出样例、错误码、边界 case 见 [`references/arkcli-helper.md`](references/arkcli-helper.md)。
