# plans harness-status

> **前置条件：** 先阅读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **范围限制：** `plans harness-status` 是**只读视图** —— 只回答"本机某个 AI Agent 上，Agent Plan 内置 MCP 装没装、key 就没就绪"，**不安装、不修改、不移除任何配置**。要**注入 / 配置 / 移除** MCP，走 [`../../arkcli-helper/SKILL.md`](../../arkcli-helper/SKILL.md)（`helper mcp` / `helper reset`）。

只读本机 AI Agent（harness）的 MCP 配置文件，报告 Agent Plan 内置的 3 个 MCP server 各自是否已注入、key 是否就绪。纯本地文件读取，**免登录**（不需要 `auth status` 通过）。

## 命令

```bash
arkcli plans harness-status [agent] [--scope project]
```

## 目标 agent 的选择（三种模式）

| 调用 | 行为 |
|---|---|
| `plans harness-status`（在某 agent 里跑） | 自动识别当前宿主（claude-code / opencode / openclaw），只报它 |
| `plans harness-status`（裸终端 / 脚本，识别不出宿主） | **全量兜底**：列出全部支持 MCP 的 agent，不报错 |
| `plans harness-status <agent>` | 只报指定 agent；随处可用，被查 agent 此刻没在跑也能查（读磁盘配置） |

> 宿主识别靠宿主注入的环境变量（`CLAUDECODE` / `OPENCODE` / `OPENCLAW_*`）。脱离宿主就没有这些信号 —— 此时**不报错**，改列全部（只读、零风险）。Trae 不被自动识别（无环境信号），始终在"列全部"兜底里出现，或显式 `harness-status trae`。

## 参数

| 参数 | 必填 | 说明 |
|---|---|---|
| `[agent]` | 否 | `claude-code` / `opencode` / `openclaw` / `trae` 之一。不传 = 按上表自动决定。传非法名或 `hermes`（不支持 MCP）→ `validation` 报错并列出合法值 |
| `--scope` | 否 | **仅 Trae**:`global`(默认)读 `~/.trae/mcp.json`;`project` 读当前目录 `./.trae/mcp.json`。对非 Trae agent 传 `project` 会报错 |

## 返回值

```json
{
  "agents": [
    {
      "agent": "claude-code",
      "mcp_config_path": "/Users/me/.claude.json",
      "mcp_servers": [
        { "name": "mcp-server-askecho-search-infinity", "transport": "stdio", "installed": true,  "key_state": "real" },
        { "name": "dataPro-search",                      "transport": "http",  "installed": true,  "key_state": "real" },
        { "name": "ov-mcp-server",                       "transport": "http",  "installed": false, "key_state": "absent" }
      ]
    }
  ]
}
```

| 字段 | 含义 |
|---|---|
| `agent` | harness 稳定标识：claude-code / opencode / openclaw / trae |
| `mcp_config_path` | 实际读取的配置文件路径（诊断"为什么说没装"时看这里） |
| `mcp_servers[].name` | 内置 MCP server 名（固定 3 个，见下） |
| `mcp_servers[].transport` | `stdio`（子进程）/ `http`（远端） |
| `mcp_servers[].installed` | 该 server 节点是否存在于配置里 |
| `mcp_servers[].key_state` | `real` / `placeholder` / `absent`（见下表） |

### key_state 三态

| key_state | installed | 含义 | 怎么处理 |
|---|---|---|---|
| `real` | true | 已注入且 key 就绪 ✅ | 可用，无需动作 |
| `placeholder` | true | 已注入但 key 还是占位符 / 空 | 带 key 重新注入：`arkcli helper mcp` |
| `absent` | false | 没注入这一个 | 需要的话注入：`arkcli helper mcp` |

> **安全**：本命令**从不输出真实 key**，只输出上面三种状态之一。

## 三个内置 MCP server

| name | 能力 | key 来源 |
|---|---|---|
| `mcp-server-askecho-search-infinity` | 联网搜索（web search） | account 级搜索 key |
| `dataPro-search` | dataPro 专业数据检索 | Agent Plan API Key |
| `ov-mcp-server` | OpenViking 知识库检索 | 绑定库的访问 key（账号下 0 个库时注入会跳过它，故可能 `absent`） |

## 常见错误

| 错误 | 原因 | 处理 |
|---|---|---|
| `{"type":"validation","message":"未知 agent ..."}` exit 2 | `[agent]` 传了不认识的名字 | 用 claude-code / opencode / openclaw / trae |
| `{"type":"validation","message":"agent \"hermes\" 不支持 MCP ..."}` exit 2 | 指定了不支持 MCP 注入的 agent | hermes 不支持 MCP |
| `parse ... 非法 JSON` | agent 配置文件被手改坏了 | 修复该 agent 的 JSON 配置 |

## 注意事项

- 只读、免登录、零副作用 —— 跟 `helper mcp` / `helper reset`（写操作）相反方向
- Trae:默认查全局 `~/.trae/mcp.json`;查项目级用 `harness-status trae --scope project`（读当前目录 `./.trae/mcp.json`）
- `ov-mcp-server` 报 `absent` 不一定是"漏装"：账号下没有 OpenViking 库时，注入本就会跳过它
- 这是 plan 视角下"我的 MCP 落地了没"的快速自检；想看 agent 自身装没装（install）+ 模型 / provider 配没配 → [`../../arkcli-helper/SKILL.md`](../../arkcli-helper/SKILL.md) 的 `helper list`

## 参考

- [arkcli-plans](../SKILL.md) -- skill 概览
- [arkcli-helper](../../arkcli-helper/SKILL.md) -- **注入 / 配置 / 移除** MCP（本命令只读查看，改动走这里）
- [arkcli-shared](../../arkcli-shared/SKILL.md)
