# arkcli helper —— 详细参考

## `arkcli helper mcp [target]`

只把内置 MCP server 注入目标 agent 的配置文件,**不动 model / provider / base_url**。**个人版 `agent-plan` 四台**(豆包搜索 web-search + dataPro-search + openviking-dataplane + openviking-controlplane);**团队版 `agent-plan-team` 与 OpenViking 无关,只两台**(豆包搜索 + dataPro)。

```
arkcli helper mcp [target] [--profile <plan-profile>] [--ov-resource <库名>] [--scope global|project]
```

- `target`(可选位置参数):`claude-code` | `opencode` | `openclaw` | `trae`。
  - 传了 → 配这个(显式优先于检测)。
  - 不传 → 自动检测当前宿主 agent(host)作 target。
- `--profile`(可选):指定用哪个 Agent Plan profile 的 key/身份。默认自动定位账号下唯一的 Agent Plan profile(`agent-plan` 或 `agent-plan-team` 都可;但**团队版不含 OpenViking**,见下)。
- `--ov-resource`(可选):指定 openviking-dataplane 绑定的 OpenViking 库(按库名,也接受 ResourceID)。账号有多个库且未指定时命令会报错并列出可选库名。
- `--scope`(可选,**仅 Trae**):`global`(默认)写用户级 `~/.trae/mcp.json`;`project` 写当前目录 `./.trae/mcp.json`。其它 agent 传 `--scope project` 会报错(一律用户级全局)。

### target 解析顺序(host ≠ target)

```
1. 给了位置参数 target?        → 用它(忽略检测)
2. 没给 → 检测当前 host:
     CLAUDECODE=1 / AI_AGENT=claude-code_*   → claude-code
     OPENCODE=1                              → opencode
     OPENCLAW_SHELL=… / OPENCLAW_CLI=1       → openclaw
   唯一命中 → 用它
3. 测不出(host 是 cursor/gemini/trae 等不被检测的 agent,或多信号冲突)
     → 报错,要求显式指定;不静默猜
```

> 检测读的是 arkcli 进程从宿主继承的环境变量(skill 用 Bash 调 arkcli 时天然继承)。
> 这些信号均经源码/实测核实;配置类 env(`OPENCODE_CONFIG_DIR` / `OPENCLAW_CONFIG_PATH`)不作信号。
> Trae(AI IDE)**不做运行态宿主检测**(无经核实的环境信号),只能显式 `arkcli helper mcp trae`。

### Agent Plan profile 定位

dataPro / openviking-dataplane / openviking-controlplane 的 key 必须来自 Agent Plan(OpenViking 两台仅个人版注入),所以命令**不看当前 active profile**,而是:

- `--profile P` 给了 → 用 P(若 P 不是 Agent Plan 类型 → 报错)。
- 没给 → 扫所有 profile 找 Agent Plan(`agent-plan` 或 `agent-plan-team` 都行;注意**两者不等价** —— 团队版不注入 OpenViking 两台):0 个 → 引导 `auth login`;1 个 → 直接用;多个(含个人版 + 团队版并存)→ 要求 `--profile` 指定。

豆包搜索(web-search / askecho)、dataPro、openviking-controlplane 三台都用该 profile 的 plan API Key(`RawDefaultAPIKey`,与 `arkcli profile show` 显示的 default key 同一把),不再单独取"搜索专属 key";plan key 取不到才写占位符。(**团队版不注入 openviking-controlplane**,实际只前两台用这把 key。)

### OpenViking 库定位(openviking-dataplane;**仅个人版 agent-plan**)

> **团队版 `agent-plan-team` 与 OpenViking 无关**:命令对团队版直接跳过本节整套列库/取 key 流程,两台 OV 都不注入。以下只适用于个人版。

openviking-dataplane 的 `Authorization: Bearer <key>` 绑定到具体 OpenViking 库。命令先 `ListOpenVikingCollections`(走 vikingdb service,Project/ProjectName 传空 = 全账号)拿到 `{库名 ↔ ResourceID}`,再按 `--ov-resource` 选库:

- 0 库 → 跳过 openviking-dataplane(另三台包括 openviking-controlplane 照常注入)+ 引导 create URL。
- 1 库 → 直接用。
- 多库 + 给了 `--ov-resource` → 按库名(或 ResourceID)精确匹配;未命中 → 报错列出可选库名。
- 多库 + 没给 → 报错列出所有库名(skill 据此 AskUserQuestion 选库后带 `--ov-resource` 重跑;`arkcli helper` 出 TTY picker)。

选定库后 `AccessOpenVikingApiKey(ResourceID, UserID=default, Project=库的 Project)` 取库 key;失败 → 写占位符。

> `openviking-controlplane` 不经过上述流程,只需 Agent Plan API Key,**个人版**始终随 Plan 注入,不受 OV 库数量影响(**团队版不注入**)。

### 输出样例

个人版 `agent-plan` —— 四台全注入:

```
✓ 已为 claude-code 注入 MCP → [mcp-server-askecho-search-infinity dataPro-search openviking-dataplane openviking-controlplane]  (plan: agent-plan_cn-beijing)
提示:Claude Code 需重启后才会加载新注入的 MCP。
```

团队版 `agent-plan-team` —— 只两台(与 OpenViking 无关,自动跳过两台 OV):

```
团队版 Agent Plan 不含 OpenViking,已跳过数据面/控制面两台 MCP。
✓ 已为 claude-code 注入 MCP → [mcp-server-askecho-search-infinity dataPro-search]  (plan: agent-plan-team_cn-beijing_team)
提示:Claude Code 需重启后才会加载新注入的 MCP。
```

### 各 agent 的落点

| target | MCP 写入文件 | 重载方式 |
|--------|-------------|---------|
| claude-code | `~/.claude.json`(`mcpServers.*`)；model 在另一文件 `~/.claude/settings.json`,本命令不碰 | 重启 |
| opencode | `~/.config/opencode/opencode.json`(`mcp.*`,merge 保留其它键) | 重启 |
| openclaw | `~/.openclaw/openclaw.json`(`mcp.servers.*` + 启用 mcporter skill) | 重启 |
| trae | 默认 `~/.trae/mcp.json`(`mcpServers.*`,与 claude 同构);`--scope project` → `./.trae/mcp.json` | 全局:去「设置 → MCP」确认 MCP 已启用 + 重启;项目级:开「启用项目级 MCP」开关 + 重开项目 |

## 错误与边界 case

| 现象 | 原因 | 处理 |
|------|------|------|
| `无法确定要配置哪个 agent` | host 不是可检测的 3 个 / 信号冲突 / 无信号(含 Trae 等 IDE) | 显式 `arkcli helper mcp <claude-code\|opencode\|openclaw\|trae>` |
| `<X> 暂不支持 MCP 注入` | target 是 hermes 或未来不支持的 agent | 仅 claude-code/opencode/openclaw/trae 可注入 |
| `--scope project 仅 Trae 支持` | 对非 Trae agent 传了 `--scope project` | 去掉 `--scope`(其它 agent 一律用户级全局) |
| `未找到 Agent Plan profile` | 账号无 agent-plan 订阅 / 未登录 | `arkcli auth login` 开通 Agent Plan |
| `检测到多个 Agent Plan profile` | 多个 agent-plan profile | 加 `--profile <名>` 指定 |
| `profile X 不是 Agent Plan` | `--profile` 指了非 Agent Plan(agent-plan / agent-plan-team 之外) | 换成 Agent Plan profile |
| 豆包搜索(web-search)写了占位符 | plan profile 无可用 API Key | `arkcli auth apikey` 选一把,或 `arkcli profile keys refresh` 刷新后重跑 |
| `检测到多个 OpenViking 库` | 账号多个 OpenViking 库且没指定 | 加 `--ov-resource <库名>`(skill 用 AskUserQuestion 让用户选库) |
| 跳过了 openviking-dataplane(个人版) | 个人版账号下 0 个 OpenViking 库 | 去 create URL 建库后重跑;或接受跳过(另三台含 openviking-controlplane 已注入)。注:**团队版本就不注入 OV**,不会出现这条 |
| openviking-dataplane 写了占位符 | OpenViking 列库 / 取 key 失败 | 手动填 `Authorization: Bearer <真实 key>`,或重跑 |
| 注入了但 agent 里没生效 | MCP 在 agent 启动时加载 | 重启该 agent |

## 与 `arkcli helper` / `configure` 的关系

- `arkcli helper`(交互):一条龙选 plan→model→harness→安装→**注入 MCP**→末尾问 byted-supabase。只能在 TTY 跑。
- `configure --with-mcp [--with-supabase]`(非交互):配 model/provider **并**注入 MCP(会(重)写 model);加 `--with-supabase` 再连带配 byted-supabase(CLI+Skill+登录态,资格不够/失败只 warn)。**这是非交互/agent 一条命令配齐全套 harness 工具的入口**。
- `mcp`(非交互):**只**注入 MCP,零副作用 —— 适合"已经配好 agent、只想加豆包搜索"或 prompt 触发。**不含 Supabase**。

## 移除

```
arkcli helper reset <harness>
```

移除 arkcli 注入的 provider/env/model + 内置 MCP server(保留用户其它配置)。Trae 项目级注入用 `arkcli helper reset trae --scope project` 移除。
