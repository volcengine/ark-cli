# plans harness-status

> **前置条件：** 先阅读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **范围限制：** `plans harness-status` 是**只读视图** —— 只回答"本机某个 AI Agent 上，Agent Plan 内置 MCP 装没装、key 就没就绪"，**不安装、不修改、不移除任何配置**。要**注入 / 配置 / 移除** MCP，走 [`../../arkcli-helper/SKILL.md`](../../arkcli-helper/SKILL.md)（`helper mcp` / `helper reset`）。

只读本机 AI Agent（harness）的配置，镜像火山方舟控制台「配置 Harness」页：按**能力卡**报告专属 Harness 能力各自装没装、key 就没就绪。纯本地文件读取，**免登录**（不需要 `auth status` 通过）。

输出分两段：

- **`agents[]`** —— per-agent 的 MCP 能力卡：**豆包搜索** / **专业数据集** / **Agent 记忆**。每张卡带 `forms` 交付形态标签 + 卡级 `installed`/`key_state` + 卡内底层 `mcp_servers` 明细。
- **`supabase`** —— 全局的**火山引擎Supabase** 集成态，**与能力卡同形**：`installed` 是「装好没」统一 headline（CLI 与 Skill 都就绪），`cli_installed` / `skill_installed` 是次要细分。它不是 per-agent 的 MCP，而是一次性的 CLI + Skill + 登录态注入，故置顶单列、**与查哪个 agent 无关**。

> **plan 维度提醒：** 命令免登录、不区分 plan 类型，只如实镜像磁盘节点。但「Agent 记忆」卡(OpenViking 两台)仅**个人版 `agent-plan`** 适用；**团队版 `agent-plan-team` 与 OpenViking 无关** —— 团队版机器上该卡 `installed:false` 是预期正确状态，**不是漏装**（详见文末注意事项）。`supabase` 块只报本机有没有装，**不判当前 plan 能不能用**（资格要登录读 profile，与本命令免登录冲突；想确认资格走 `arkcli helper supabase`）。

## 命令

```bash
arkcli plans harness-status [agent] [--scope project] [--codex-config-scope profile|global] [--codex-profile <name>]
```

## 目标 agent 的选择（三种模式）

| 调用 | 行为 |
|---|---|
| `plans harness-status`（在某 agent 里跑） | 自动识别当前宿主（claude-code / opencode / openclaw），只报它 |
| `plans harness-status`（裸终端 / 脚本，识别不出宿主） | **全量兜底**：列出全部支持 MCP 的 agent，不报错 |
| `plans harness-status <agent>` | 只报指定 agent；随处可用，被查 agent 此刻没在跑也能查（读磁盘配置） |

> 宿主识别靠宿主注入的环境变量（`CLAUDECODE` / `OPENCODE` / `OPENCLAW_*`）。脱离宿主就没有这些信号 —— 此时**不报错**，改列全部（只读、零风险）。Codex / Trae 不被自动识别（无环境信号），始终在"列全部"兜底里出现，或显式 `harness-status codex` / `harness-status trae`。

## 参数

| 参数 | 必填 | 说明 |
|---|---|---|
| `[agent]` | 否 | `claude-code` / `codex` / `opencode` / `openclaw` / `trae` 之一。不传 = 按上表自动决定。传非法名或 `hermes`（不支持 MCP）→ `validation` 报错并列出合法值 |
| `--scope` | 否 | **仅 Trae**:`global`(默认)读 `~/.trae/mcp.json`;`project` 读当前目录 `./.trae/mcp.json`。对非 Trae agent 传 `project` 会报错 |
| `--codex-config-scope` | 否 | **仅 Codex**:`profile`(默认)读 `~/.codex/<profile>.config.toml`;`global` 读 `~/.codex/config.toml` |
| `--codex-profile` | 否 | **仅 Codex profile 模式**:Codex profile 名,默认 `arkcli` |

## 返回值

```json
{
  "agents": [
    {
      "agent": "claude-code",
      "mcp_config_path": "/Users/me/.claude.json",
      "capabilities": [
        {
          "key": "web-search", "name": "豆包搜索", "forms": ["skill", "mcp"],
          "installed": true, "key_state": "real",
          "mcp_servers": [
            { "name": "mcp-server-askecho-search-infinity", "transport": "stdio", "installed": true, "key_state": "real" }
          ]
        },
        {
          "key": "datapro", "name": "专业数据集", "forms": ["mcp"],
          "installed": true, "key_state": "real",
          "mcp_servers": [
            { "name": "dataPro-search", "transport": "http", "installed": true, "key_state": "real" }
          ]
        },
        {
          "key": "agent-memory", "name": "Agent 记忆", "forms": ["mcp"],
          "installed": true, "key_state": "real",
          "mcp_servers": [
            { "name": "openviking-dataplane",    "transport": "http",  "installed": false, "key_state": "absent" },
            { "name": "openviking-controlplane", "transport": "stdio", "installed": true,  "key_state": "real" }
          ]
        }
      ]
    }
  ],
  "supabase": {
    "name": "火山引擎Supabase", "forms": ["cli", "skill"],
    "installed": true, "cli_installed": true, "skill_installed": true
  }
}
```

| 字段 | 含义 |
|---|---|
| `agent` | harness 稳定标识：claude-code / codex / opencode / openclaw / trae |
| `mcp_config_path` | 实际读取的配置文件路径（诊断"为什么说没装"时看这里） |
| `capabilities[].key` | 能力稳定标识：`web-search` / `datapro` / `agent-memory` |
| `capabilities[].name` | 能力展示名（对齐控制台卡）：豆包搜索 / 专业数据集 / Agent 记忆 |
| `capabilities[].forms` | 交付形态标签（描述性）：`mcp` / `skill` / `cli`。豆包搜索同时交付 Skill+MCP |
| `capabilities[].installed` | **卡级**：代表性 server 是否注入（Agent 记忆取 `openviking-controlplane`） |
| `capabilities[].key_state` | **卡级**：`real` / `placeholder` / `absent`（见下表） |
| `capabilities[].mcp_servers[]` | 卡内底层 MCP server 明细（诊断用；Agent 记忆含 dataplane+controlplane 两台） |
| `mcp_servers[].name/transport/installed/key_state` | 单个底层 server 的名/传输/是否注入/key 三态 |
| `supabase.name` / `supabase.forms` | 与能力卡同形：展示名「火山引擎Supabase」+ 交付形态 `["cli","skill"]` |
| `supabase.installed` | **统一 headline**：CLI 与 Skill 都就绪才 `true`（= `cli_installed && skill_installed`）。用户/agent 看这一个就知「装好没」，不必自己 AND 两个细分字段 |
| `supabase.cli_installed` | 细分：`byted-supabase-cli` 是否在 PATH |
| `supabase.skill_installed` | 细分：byted-supabase 配套 skill 是否已装（靠 CLI 自带 `skills status`） |

### key_state 三态

| key_state | installed | 含义 | 怎么处理 |
|---|---|---|---|
| `real` | true | 已注入且 key 就绪 ✅ | 可用，无需动作 |
| `placeholder` | true | 已注入但 key 还是占位符 / 空 | 带 key 重新注入：`arkcli helper mcp` |
| `absent` | false | 没注入这一个 | 需要的话注入：`arkcli helper mcp` |

> **安全**：本命令**从不输出真实 key**，只输出上面三种状态之一。

## 能力卡 → 底层 MCP server 对照（个人版 agent-plan / 团队版 agent-plan-team）

| 能力卡 `key`（name） | `forms` | 卡内底层 MCP server | key 来源 |
|---|---|---|---|
| `web-search`（豆包搜索） | skill·mcp | `mcp-server-askecho-search-infinity` | Agent Plan API Key（与 dataPro 同一把） |
| `datapro`（专业数据集） | mcp | `dataPro-search` | Agent Plan API Key |
| `agent-memory`（Agent 记忆，**仅个人版 agent-plan**） | mcp | `openviking-dataplane` + `openviking-controlplane` | dataplane：绑定库访问 key（vikingdb 下发；0 库时跳过 → 可能 `absent`）；controlplane：Agent Plan API Key（有 Agent Plan 即注入，不依赖 OV 库列表）。**团队版两台都从不注入** |

> 卡级 `installed`/`key_state` 取**代表性 server**：豆包搜索取 askecho、专业数据集取 dataPro-search、Agent 记忆取 **controlplane**（最稳的"能力配没配"信号）。逐台细节看 `capabilities[].mcp_servers[]`。

## 全局 Supabase 集成（`supabase` 块）

| 字段 | 含义 | 怎么探测 |
|---|---|---|
| `cli_installed` | `byted-supabase-cli` 在不在 PATH | `exec.LookPath` |
| `skill_installed` | byted-supabase 配套 skill 装没装 | CLI 自带 `byted-supabase-cli skills status`（CLI 未装时直接 `false`，不误报） |

> Supabase 是 `arkcli helper supabase` 配的**全局**能力（CLI + Skill + 火山登录态注入），不是 per-agent 的 MCP，所以不进 `agents[]`、只在顶层 `supabase` 出现一次。本块**只报装没装、不判 plan 资格**（agent-plan 全档均支持，但资格判定要登录读 profile 类型/租户才知道，超出本免登录命令范围）。

## 常见错误

| 错误 | 原因 | 处理 |
|---|---|---|
| `{"type":"validation","message":"未知 agent ..."}` exit 2 | `[agent]` 传了不认识的名字 | 用 claude-code / codex / opencode / openclaw / trae |
| `{"type":"validation","message":"agent \"hermes\" 不支持 MCP ..."}` exit 2 | 指定了不支持 MCP 注入的 agent | hermes 不支持 MCP |
| `parse ... 非法 JSON` | agent 配置文件被手改坏了 | 修复该 agent 的 JSON 配置 |

## 注意事项

- 只读、免登录、零副作用 —— 跟 `helper mcp` / `helper reset`（写操作）相反方向
- Codex:默认查 profile `~/.codex/arkcli.config.toml`;查全局配置用 `harness-status codex --codex-config-scope global`;查其它 profile 用 `--codex-profile <name>`
- Trae:默认查全局 `~/.trae/mcp.json`;查项目级用 `harness-status trae --scope project`（读当前目录 `./.trae/mcp.json`）
- **先分 plan 类型再解读「Agent 记忆」卡**：它仅对**个人版 `agent-plan`** 有意义。**团队版 `agent-plan-team` 与 OpenViking 无关** —— 团队版下该卡 `installed:false`、卡内两台 `absent` 是**预期正确状态**，别据此提示用户去注入 OpenViking（团队版本就不该有这两台）
- 解读「Agent 记忆」卡时**展开卡内 `mcp_servers` 看两台**：（**个人版**）`openviking-dataplane` 报 `absent` 不一定是"漏装"——账号下没有 OpenViking 库时，注入本就会跳过它；`openviking-controlplane` 报 `absent` 才是真漏装（它始终随个人版 Agent Plan 注入、不依赖 OV 库），说明该 agent 从未跑过 `helper mcp`。**卡级 `installed` 取的就是 controlplane**，所以卡级 `installed:false` ≈ controlplane 漏装
- Supabase 块只反映本机 `cli_installed`/`skill_installed`；要**配置/重装**它走 [`../../arkcli-helper/SKILL.md`](../../arkcli-helper/SKILL.md) 的 `helper supabase`
- 这是 plan 视角下"我的能力落地了没"的快速自检；想看 agent 自身装没装（install）+ 模型 / provider 配没配 → [`../../arkcli-helper/SKILL.md`](../../arkcli-helper/SKILL.md) 的 `helper list`

## 参考

- [arkcli-plans](../SKILL.md) -- skill 概览
- [arkcli-helper](../../arkcli-helper/SKILL.md) -- **注入 / 配置 / 移除** MCP（本命令只读查看，改动走这里）
- [arkcli-shared](../../arkcli-shared/SKILL.md)
