# arkcli helper —— 详细参考

## `--dry-run` 边界

`arkcli helper` 整个命令域不注册 `--dry-run`。这些命令负责 TTY
交互、本地 harness 配置、MCP 注入或外部工具安装，不是 API request
参数预览。传入该 flag 应直接报 unknown flag，不能静默忽略，也不能
输出虚假成功。

## 三方渠道订单首用签署

抖店/移动商城等三方渠道购买的个人版 plan，在 `arkcli helper` / `helper configure`
配置前会被数据协议签署闸门拦截——TTY 下走一屏式交互签署（数字键打开协议 · Space 勾选 · Enter 继续 · Esc 退出），
非 TTY 硬拒 `plan_agreement_required`（`--yes` 不放行，逃生 env
`ARKCLI_ALLOW_HEADLESS_PLAN_AGREEMENT=1`）。详见
[`../../arkcli-plans/references/arkcli-plans-agreement-signing.md`](../../arkcli-plans/references/arkcli-plans-agreement-signing.md)。

执行 `configure/reset/mcp/supabase` 前，Agent 必须展示准确 target、
profile、scope 与文件落点并取得确认；只能用 `helper list` 做只读检查。
不要生成 `arkcli --dry-run helper ...` 或
`arkcli helper ... --dry-run`。

## `arkcli helper mcp [target]`

这是**不改 model / provider / base_url**的非交互 Harness 工具入口。不传 `--capability` 时完全保持存量行为：**个人版 `agent-plan` 四台 MCP**(豆包搜索 + DataPro + OpenViking 数据面/控制面)，**团队版 `agent-plan-team` 两台 MCP**(豆包搜索 + DataPro)。传 `--capability` 时只配所选单项，不带入其他已开启工具。

```
arkcli helper mcp [target] [--capability <datapro|web-search|agent-memory|cua>] [--profile <plan-profile>] [--ov-resource <库名>] [--scope global|project] [--keep-native-websearch] [--codex-config-scope profile|global] [--codex-profile <name>] [--dsh-config-scope home|profile] [--dsh-profile <name>]
```

- `target`(可选位置参数):MCP 支持 `claude-code` | `codex` | `opencode` | `openclaw` | `trae` | `deepseek-harness`；CUA 还支持 `hermes` | `pi`。
  - 传了 → 配这个(显式优先于检测)。
  - 不传 → 自动检测当前宿主 agent(host)作 target。
- `--profile`(可选):指定用哪个 Agent Plan profile 的 key/身份。默认自动定位账号下唯一的 Agent Plan profile(`agent-plan` 或 `agent-plan-team` 都可;但**团队版不含 OpenViking**,见下)。
- `--capability`(可选):不传即整组 MCP；传入后仅接受下表稳定能力名。
- `--ov-resource`(可选):仅整组 MCP 或 `--capability agent-memory` 使用，指定 openviking-dataplane 绑定的 OpenViking 库。显式选其它 capability 时传该 flag 会报错。
- `--scope`(可选,**仅 Trae**):`global`(默认)写用户级 `~/.trae/mcp.json`;`project` 写当前目录 `./.trae/mcp.json`。其它 agent 传 `--scope project` 会报错(一律用户级全局)。
- `--keep-native-websearch`(可选):仅整组 MCP 或 `--capability web-search` 使用。豆包搜索配置成功后保留目标 Agent 的原生 WebSearch；不传时采用推荐默认值“关闭原生 WebSearch”。
- `--codex-config-scope`(可选,**仅 Codex**):`profile`(默认)写 `~/.codex/<profile>.config.toml`;`global` 写 `~/.codex/config.toml`。
- `--codex-profile`(可选,**仅 Codex profile 模式**):Codex profile 名,默认 `arkcli`。
- `--dsh-config-scope`(可选,**仅 DeepSeek Harness**):`home`(默认)写 `$DSH_HOME/cordis.patch.yml`;`profile` 写 `$DSH_HOME/profiles/<name>/cordis.patch.yml`。
- `--dsh-profile`(可选,**仅 DSH profile 模式**):与 `--dsh-config-scope profile` 一起使用；Helper 会确保对应 profile 目录存在。

### 单能力矩阵

| `--capability` | profile 范围 | 精确动作 |
|---|---|---|
| `datapro` | 个人版 / 团队版 | 只注入 `dataPro-search`；不查 OpenViking，不安装豆包搜索 Skill |
| `web-search` | 个人版 / 团队版 | 只注入搜索 MCP + 安装 `byted-web-search` + 写 Skill 凭证；默认关闭原生 WebSearch |
| `agent-memory` | **仅个人版** | 只配 OpenViking control-plane，有库时同时配 data-plane；团队版 fail-fast |
| `cua` | **仅个人版 Large/Max** | 只给位置参数指定的 target 安装 `byted-util-ark-cua` Skill；仅本地读 profile 资格，不刷新 Plan Key、不注入 MCP、不写模型、不扫描其他 Agent；安装失败返回非零 |

profile 仍完全复用现有 `--profile`：账号只有一个 Agent Plan profile 时自动选择；个人版与团队版并存时必须显式指定。不增加另一个 agentplan/team 参数。

### target 解析顺序(host ≠ target)

```
1. 给了位置参数 target?        → 用它(忽略检测)
2. 没给 → 检测当前 host:
     CLAUDECODE=1 / AI_AGENT=claude-code_*   → claude-code
     OPENCODE=1                              → opencode
     OPENCLAW_SHELL=… / OPENCLAW_CLI=1       → openclaw
   唯一命中 → 用它
3. 测不出(host 是 cursor/gemini/codex/trae 等不被检测的 agent,或多信号冲突)
     → 报错,要求显式指定;不静默猜
```

> 检测读的是 arkcli 进程从宿主继承的环境变量(skill 用 Bash 调 arkcli 时天然继承)。
> 这些信号均经源码/实测核实;配置类 env(`OPENCODE_CONFIG_DIR` / `OPENCLAW_CONFIG_PATH`)不作信号。
> Codex / Trae(AI IDE)**不做运行态宿主检测**(无经核实的环境信号),只能显式 `arkcli helper mcp codex` / `arkcli helper mcp trae`。

### Agent Plan profile 定位

dataPro / openviking-dataplane / openviking-controlplane 的 key 必须来自 Agent Plan(OpenViking 两台仅个人版注入),所以命令**不看当前 active profile**,而是:

- `--profile P` 给了 → 用 P(若 P 不是 Agent Plan 类型 → 报错)。
- 没给 → 扫所有 profile 找 Agent Plan(`agent-plan` 或 `agent-plan-team` 都行;注意**两者不等价** —— 团队版不注入 OpenViking 两台):0 个 → 引导 `auth login`;1 个 → 直接用;多个(含个人版 + 团队版并存)→ 要求 `--profile` 指定。

豆包搜索(web-search / askecho)、dataPro、openviking-controlplane 三台都用该 profile 的 plan API Key(`RawDefaultAPIKey`,与 `arkcli profile show` 显示的 default key 同一把),不再单独取"搜索专属 key"。豆包搜索的 MCP 子进程使用 `ASK_ECHO_SEARCH_INFINITY_API_KEY`；`byted-web-search` 是另一个进程，Helper 会在安装后通过标准 `skills list --json` 返回的实际路径，将同一把 Key 原子写入 Skill 根目录 `.env` 的 `WEB_SEARCH_API_KEY`，权限为 `0600`。Helper 不修改第三方 Skill 源码、不把 Key 写到全局 shell 环境，也不在输出中展示 Key。Skill 安装或凭证写入失败时，命令报错并保持原生 WebSearch 不变。(**团队版不注入 openviking-controlplane**,实际只前两台用这把 key。)

### OpenViking 库定位(openviking-dataplane;**仅个人版 agent-plan**)

> **团队版 `agent-plan-team` 与 OpenViking 无关**:命令对团队版直接跳过本节整套列库/取 key 流程,两台 OV 都不注入。以下只适用于个人版。

openviking-dataplane 的 `Authorization: Bearer <key>` 绑定到具体 OpenViking 库。命令先 `ListOpenVikingCollections`(走 vikingdb service,Project/ProjectName 传空 = 全账号)拿到 `{库名 ↔ ResourceID}`,再按 `--ov-resource` 选库:

- 0 库 → 跳过 openviking-dataplane(另三台包括 openviking-controlplane 照常注入)+ 引导 create URL。
- 1 库 → 直接用。
- 多库 + 给了 `--ov-resource` → 按库名(或 ResourceID)精确匹配;未命中 → 报错列出可选库名。
- 多库 + 没给 → 报错列出所有库名；skill 只把本次错误返回的真实库名交给宿主结构化选择能力，选定后带 `--ov-resource` 重跑（不写死任何宿主工具名，不凭记忆补库；无结构化选择能力时用精简编号列表）；`arkcli helper` 出 TTY picker。

选定库后 `AccessOpenVikingApiKey(ResourceID, UserID=default, Project=库的 Project)` 取库 key;失败 → 写占位符。

> `openviking-controlplane` 不经过上述流程,只需 Agent Plan API Key,**个人版**始终随 Plan 注入,不受 OV 库数量影响(**团队版不注入**)。

### 输出样例

只给 Codex 安装 CUA：

```
✓ 已为 codex 安装云电脑 skill (plan: agent-plan_cn-beijing)
请重启 Codex 以加载新安装的能力。
```

个人版 `agent-plan` —— 四台全注入:

```
✓ 已为 claude-code 注入 MCP → [mcp-server-askecho-search-infinity dataPro-search openviking-dataplane openviking-controlplane]  (plan: agent-plan_cn-beijing)
✓ claude-code 原生 WebSearch 已关闭（项目级）→ /repo/.claude/settings.json
快捷恢复：arkcli helper reset claude-code（同时移除 ArkCLI 写入的模型/provider/MCP 配置）
提示:Claude Code 需重启后才会加载新注入的 MCP。
```

团队版 `agent-plan-team` —— 只两台(与 OpenViking 无关,自动跳过两台 OV):

```
团队版 Agent Plan 不含 OpenViking,已跳过数据面/控制面两台 MCP。
✓ 已为 claude-code 注入 MCP → [mcp-server-askecho-search-infinity dataPro-search]  (plan: agent-plan-team_cn-beijing_team)
✓ claude-code 原生 WebSearch 已关闭（项目级）→ /repo/.claude/settings.json
快捷恢复：arkcli helper reset claude-code（同时移除 ArkCLI 写入的模型/provider/MCP 配置）
提示:Claude Code 需重启后才会加载新注入的 MCP。
```

### 各 agent 的落点

| target | MCP 写入文件 | 重载方式 |
|--------|-------------|---------|
| claude-code | `~/.claude.json`(`mcpServers.*`)；model 在另一文件 `~/.claude/settings.json`,本命令不碰 | 重启 |
| codex | 默认 profile `~/.codex/arkcli.config.toml`(`mcp_servers.*`);`--codex-config-scope global` → `~/.codex/config.toml` | profile:用 `codex --profile <name>` 启动;global:重启 Codex CLI/TUI/App/IDE |
| opencode | `~/.config/opencode/opencode.json`(`mcp.*`,merge 保留其它键) | 重启 |
| openclaw | `~/.openclaw/openclaw.json`(`mcp.servers.*` + 启用 mcporter skill) | 重启 |
| trae | 默认 `~/.trae/mcp.json`(`mcpServers.*`,与 claude 同构);`--scope project` → `./.trae/mcp.json` | 全局:去「设置 → MCP」确认 MCP 已启用 + 重启;项目级:开「启用项目级 MCP」开关 + 重开项目 |
| deepseek-harness | 默认 `$DSH_HOME/profiles/<name>/cordis.patch.yml`(`--dsh-profile <name>` 必填);`--dsh-config-scope home` → `$DSH_HOME/cordis.patch.yml` | 重启 `dsh --profile <name>`(profile scope);home 层全局生效 |
| zcode | `$ZCODE_HOME/cli/config.json`（默认 `~/.zcode/cli/config.json`，`mcp.servers.*`）；model/provider 在 `$ZCODE_HOME/v2/config.json` | 完全退出 ZCode 主进程后重新打开并新建会话；旧会话不会热加载 provider 配置 |
| workbuddy | 同时写 `~/.workbuddy-ai/mcp.json`、`~/.workbuddy/mcp.json` 与 `~/.codebuddy/mcp.json`(`mcpServers.*`,与 claude 同构);model 清单写各自 `models.json`(OpenAI 兼容,arkcli 项打 `arkcliManaged` marker) | 重启 CodeBuddy IDE |

### 原生 WebSearch 的选择与精确落点

TTY 配置豆包搜索时，在 MCP 与 Skill 都成功后展示“关闭原生 WebSearch（推荐）/保留原生 WebSearch”。非交互 `helper mcp` 与 `helper configure --with-mcp|--with-routing` 默认关闭；传 `--keep-native-websearch` 时不写下表字段。

| target | 关闭方式 | 范围 | reset 行为 |
|--------|----------|------|------------|
| claude-code | 项目根 `.claude/settings.json`：`permissions.deny` 加入 `WebSearch` | 项目级 | 只撤销 ArkCLI 加入的值 |
| codex | 项目根 `.codex/config.toml`：`web_search = "disabled"` | 项目级；与 Codex model/MCP 的 profile/global 范围独立 | 恢复原值或删除 ArkCLI 新增字段 |
| opencode | 项目根 `opencode.json`：`permission.websearch = "deny"` | 项目级 | 恢复原值或删除 ArkCLI 新增字段 |
| openclaw | `~/.openclaw/openclaw.json`：`tools.web.search.enabled = false` | 用户全局 | 恢复原值或删除 ArkCLI 新增字段 |
| trae | 项目根 `.trae/hooks.json`：`PreToolUse` 精确拒绝 `^(WebSearch\|web_search)$` | 项目级；需在 `TRAE Settings > Hooks` 启用项目 | 只移除 ArkCLI 的精确 Hook |
| hermes | `~/.hermes/config.yaml`：`agent.disabled_toolsets` 加入 `search` | 用户全局；当前 Hermes 尚不支持 MCP，暂不会由 helper 触发 | 后续 MCP 支持接入后沿用同一可逆策略 |
| deepseek-harness | patch 层对 `@deepseek-ai/dsh-tool-web` 加 `disabled: true`(`arkcli_managed: tool-web` marker) | profile 级(profile 层 patch) | 只移除 ArkCLI 注入的 marker entry；用户原本关闭不接管 |

边界：只关闭 `WebSearch/web_search`。不得把 `webfetch`、`web_extract`、`x_search`、浏览器工具或 Hermes 的整个 `web` toolset 一起关闭。

触发条件：只有豆包搜索被选择、MCP/Skill 配置成功且用户未选择保留时才写上述配置。普通 `helper configure`、Platform/Coding Plan、选择跳过豆包搜索、MCP 注入失败或传 `--keep-native-websearch` 都保持原路径不变。若关闭原生搜索失败，命令明确输出“豆包搜索已配置，但原生搜索关闭失败”，不得把部分成功包装成全部成功。

### ZCode 模型 limit 与网关边界

`helper configure zcode` 写 `$ZCODE_HOME/v2/config.json`（默认
`~/.zcode/v2/config.json`）。每个模型的 `limit` 来自 ArkModels 元数据，不使用统一
fallback，也不根据 `glm-*` 等名称猜测：

- `limit.context` 原样保留 `context_window`，供 ZCode 计算上下文与自动压缩；
- `limit.output` 来自 `max_completion_tokens`，但写配置时把已知二进制 KiB 近似值
  转成方舟网关的十进制上限，例如 `131072 -> 128000`、
  `262144 -> 256000`、`393216 -> 384000`；已是十进制或不符合已知规则的值不变；
- 任一字段无权威值就只省略该字段；两者都未知时省略整个 `limit`。动态路由别名
  `ark-code-latest` 没有目标元数据时不继承任一静态模型的 limit。

若 ZCode 报 `InvalidParameter` 400 且请求模型的输出上限超过网关限制，先使用新版本
ArkCLI 重新执行原 `arkcli helper configure zcode --profile ... --model ...`，再完全退出
ZCode 主进程、重新打开并新建会话；只切换模型或继续使用旧会话不会热加载 provider
配置。不要把 `limit.context` 一并改成十进制，也不要长期手改生成配置。

## 错误与边界 case

| 现象 | 原因 | 处理 |
|------|------|------|
| `未知 --capability` | 传了服务端 HarnessName、MCP server ID 或 `all` | 只用 `datapro` / `web-search` / `agent-memory` / `cua`；要整组时去掉该 flag |
| `agent-memory 仅支持 Agent Plan 个人版` | `--profile` 指向 `agent-plan-team` | 换个人版 profile；团队版不含 OpenViking |
| `cua 仅支持 Agent Plan 个人版 Large/Max` | profile 是团队版或个人版 Small/Medium | 换合格的个人版 profile，不要改用 `configure --with-cua` 绕过资格闸 |
| `无法确定要配置哪个 agent` | host 不是可检测的 3 个 / 信号冲突 / 无信号(含 Codex / Trae / DSH / ZCode / WorkBuddy 等) | 显式 `arkcli helper mcp <claude-code\|codex\|opencode\|openclaw\|trae\|deepseek-harness\|zcode\|workbuddy>` |
| `<X> 暂不支持 MCP 注入` | target 是 hermes 或未来不支持的 agent | 仅 claude-code/codex/opencode/openclaw/trae/deepseek-harness/zcode/workbuddy 可注入 |
| `--scope project 仅 Trae 支持` | 对非 Trae agent 传了 `--scope project` | 去掉 `--scope`(其它 agent 一律用户级全局) |
| `--codex-config-scope / --codex-profile 仅 Codex harness 支持` | 对非 Codex agent 传了 Codex 专属 flag | 去掉 Codex flag,或 target 改为 `codex` |
| `--dsh-config-scope / --dsh-profile 仅 DeepSeek Harness 支持` | 对非 DSH agent 传了 DSH 专属 flag，或给 `--capability cua` 传了 MCP 配置 flag | 去掉 DSH flag，或对 MCP capability 将 target 改为 `deepseek-harness`；CUA 只安装 Skill，不接受这些 flag |
| `未找到 Agent Plan profile` | 账号无 agent-plan 订阅 / 未登录 | `arkcli auth login` 开通 Agent Plan |
| `检测到多个 Agent Plan profile` | 多个 agent-plan profile | 加 `--profile <名>` 指定 |
| `profile X 不是 Agent Plan` | `--profile` 指了非 Agent Plan(agent-plan / agent-plan-team 之外) | 换成 Agent Plan profile |
| 豆包搜索(web-search)写了占位符 | plan profile 无可用 API Key | `arkcli auth apikey` 选一把,或 `arkcli profile keys refresh` 刷新后重跑 |
| `检测到多个 OpenViking 库` | 账号多个 OpenViking 库且没指定 | 使用宿主结构化选择能力展示本次错误返回的库名；选定后加 `--ov-resource <库名>` 重跑 |
| 跳过了 openviking-dataplane(个人版) | 个人版账号下 0 个 OpenViking 库 | 去 create URL 建库后重跑;或接受跳过(另三台含 openviking-controlplane 已注入)。注:**团队版本就不注入 OV**,不会出现这条 |
| openviking-dataplane 写了占位符 | OpenViking 列库 / 取 key 失败 | 手动填 `Authorization: Bearer <真实 key>`,或重跑 |
| 注入了但 agent 里没生效 | MCP 在 agent 启动时加载 | 重启该 agent |
| TRAE 显示 Hook 文件已写但原生搜索仍可调用 | 项目 Hooks 尚未在 IDE 中启用 | 在 `TRAE Settings > Hooks` 启用当前项目并重开 |
| reset 未恢复原生搜索 | 用户原本就已关闭，或配置后同一字段被外部修改 | ArkCLI 不接管原有禁用；冲突时保留用户当前值并输出警告 |
| DSH 提示检测到旧版配置或 flat credentials | 旧版 ArkCLI 写入格式与当前 DSH provider / credentials 契约不兼容；flat key 可能单独存在，也可能残留在 version 1 文档顶层 | 先运行 `arkcli helper reset deepseek-harness`，再重新执行 `configure`；不要手工删除 `apiKeyEnv` |

DSH model/provider 写入还遵循以下 ownership 规则：`settings.yaml` 使用独立的 `llm-pi-ai.providers.arkcli-<planType>`，凭据使用 `.credentials.yaml` version 1 文档的 `refs.ARKCLI_<PLAN>_API_KEY`，让 DSH credentials service 能按 provider 的 `apiKeyEnv` 解析。两个文件按一个可回滚事务更新。检测到纯 flat credentials、version 1 文档顶层残留的 legacy credential key，或 `llm-deepseek` 同时具有 `apiKeyEnv: DEEPSEEK_API_KEY` 与精确 Ark 数据面 HTTPS URL 的旧 ArkCLI 指纹时，`configure` fail-fast，不做隐式迁移，并明确要求先运行 `arkcli helper reset deepseek-harness`；该 reset 兼容读取旧格式并清理旧 ArkCLI 内容。不要把自定义 `apiKeyEnv` 当成 ArkCLI ownership。`helper reset deepseek-harness` 仅移除 fingerprint 仍完整匹配的 provider、对应凭据以及仍指向该 provider 的 `agent-default-model`。
| ZCode 调用 `glm-5.2` 等模型时报 `InvalidParameter` / token 上限 400 | 旧配置把 ArkModels 的 `131072` 等二进制近似值直接作为 `limit.output` 发给十进制网关 | 升级 ArkCLI 后重跑 `helper configure zcode` 并重启 ZCode；确认 `limit.output=128000`，但保留权威 `limit.context` |

## Platform Endpoint 配置

Platform profile 不使用 Plan 模型列表，而是从本人创建的 Endpoint 中选择。只允许 `Running` 且模型明确为文本输出的 Endpoint；VLM 可用，生图/生视频/生 3D/音频/Embedding/未知模型均拒绝。

```bash
arkcli helper configure codex \
  --profile <platform-profile> \
  --endpoint <ep-id>
```

- Platform 下不能传 `--model`，也不能使用 `--with-mcp`、`--with-supabase` 或 `--with-routing`。
- 没有自己创建的 Endpoint 时，交互向导会打开 `https://ark.volcengine.com/region:cn-beijing/endpoint/create?agentMode=close`，用户创建后选“刷新列表”。
- Endpoint ID 写入 Agent 配置的 `model`；base URL 使用 `https://ark.<region>.volces.com/api/v3`。context window、max completion tokens、输入/输出模态通过 Endpoint 绑定的基础模型名，复用 Plan 当前的 ArkModels `LookupModelMeta → enrichModelMeta` 管道；查询失败时扩展字段保持未知并省略。
- Hermes Agent 使用 `volcengine-platform` provider，把 Endpoint ID 写入 `model.default` 和 provider 模型列表，并写入上述 OpenAI-compatible base URL。Hermes 仅支持 model/provider，不支持 MCP 注入。
- Pi 把 Endpoint ID 作为 model 写入 `~/.pi/agent/models.json` 的 plan provider（`api: openai-completions` + 上述 base URL），并把 `~/.pi/agent/settings.json` 的 `defaultProvider` / `defaultModel` 指过去。Pi 仅支持 model/provider，不支持 MCP 注入。
- Platform 元数据接入不修改 Agent Plan / Coding Plan 的模型列表、默认模型、富化结果或失败退化语义。Chat/Responses 的具体调用方式继续由 Agent 当前 harness 行为决定。

## 与 `arkcli helper` / `configure` 的关系

- `arkcli helper`(交互):选 Platform / Coding Plan 时保持原流程；选择 Agent Plan / Agent Plan Team 时走 profile→model→AI Agent→写 model/provider→**Harness 能力选择页**→逐项真实配置。完成后 Helper 反向读取真实本地状态，动态展示已安装且凭证有效的 Provider，再由用户最终确认 routing.v1；零 Provider 时只能返回配置或不启用退出。只能在 TTY 跑。
- `configure --with-mcp [--keep-native-websearch] [--with-supabase] [--with-routing]`(非交互):配 model/provider **并**注入 MCP(会(重)写 model);加 `--with-supabase` 再连带配 byted-supabase(CLI+Skill+登录态,资格不够/失败只 warn)；`--with-routing` 仅 Volc Agent Plan 家族可用，隐含业务 MCP，并在写 Harness 配置前做版本/能力门禁。豆包搜索成功后默认关闭原生 WebSearch，加 `--keep-native-websearch` 保留；强制路由只管理四类受管 Provider。**这是非交互/agent 一条命令配齐全套 harness 工具的入口**。Codex 默认写 profile `arkcli`,用 `--codex-config-scope global` 才改全局配置。
- `mcp`(非交互):不传 `--capability` 注入整组 MCP；传入则只配所选 Harness 能力。两者都不改 model/provider。**不含 Supabase**。

## Agent Plan Harness 能力矩阵

矩阵数据来自两个互相独立的状态面，禁止混为一谈：

```text
服务端 Agent Plan API                       本机所选 AI Agent
  ├─ Harness 目录 + 平台支持状态              ├─ MCP 配置文件
  ├─ 账号抵扣/超额开关                        ├─ skills CLI 全局安装快照
  └─ 团队版当前席位开关                       ├─ byted-supabase CLI profile
              │                              └─ Evolve CLI/Skill/workspace/credential
              └──────── exact HarnessName join ───────────────┘
                                      │
                                      v
              本次配置 | 套餐抵扣 | 超额后付费 | 配置状态
```

已知服务端名称的精确映射：

| 服务端 `HarnessName` | 展示能力 | 配置适配 |
|---|---|---|
| `datapro` | 专业数据集 | 仅注入 `dataPro-search` MCP |
| `SearchInfinity` | 豆包搜索 | 注入搜索 MCP + 安装 `byted-web-search` Skill |
| `openviking` | Agent 记忆 | 注入 OpenViking control-plane；有库时同时注入 data-plane |
| `aidap-supabase` | AI Native 应用开发底座 | 安装/复用 byted-supabase CLI+Skill，并注入所选 Plan 登录态 |
| `Agent_evolve` | Agent 进化 | 只读检测官方 Evolve 布局；未就绪时展示官方手动安装命令 |

未知 `HarnessName` 不丢弃、不重命名、不按子串猜适配器：保留服务端顺序追加到五项之后，配置状态显示“仅远端能力”。

### 状态与动作

- `本次配置` 是用户操作意图，不是当前状态：`[✓]` 下一步重新配置，`[ ]` 本次不处理，`[-]` 没有本地适配器或当前 Agent/Plan 不支持。已配置能力仍可勾选并强制重跑。
- 默认勾选全部可配置能力；用户可以逐项取消。“进入配置”要求至少一项，也可显式“跳过能力配置”。逐项配置页保留套餐抵扣/超额后付费的显式确认，不因默认勾选而自动开启计费开关。
- 豆包搜索逐项配置先写搜索 MCP，再用持续 Spinner 安装 `byted-web-search` Skill。下载过慢时，真人按 `Ctrl+C` 只取消这次 Skill 下载：保留已经写好的搜索 MCP，跳过 Skill 凭证和原生 WebSearch 选择，并继续处理本次勾选的后续 Harness 能力。不得把该按键解释成退出整个 Helper，也不得显示“豆包搜索配置完成”。
- 主矩阵只回答开关是否开启：启用为 `●`，其余（关闭、受上级限制、平台不支持）均为 `○ 未开启`；进入详情后再区分“受上级限制 / 不支持 / 开启中”。
- 账号超额有效态必须同时满足 `AdminEnabledOverdraft && VendorAllowStatus`。管理员开关已写入、供应方尚未开通时显示“开启中”。
- 团队席位抵扣受账号抵扣总闸限制；席位超额受账号超额有效态限制。普通成员不能修改企业账号总闸。
- 开启账号超额前先校验服务端返回的 `ServiceCode / Action / Version`。坐标不完整时不写账号开关；供应方调用失败时不盲回滚，刷新后显示“开启中”，避免与已成功但响应丢失的供应方请求竞态。
- 每次开关写入都只发送所选字段，不顺带覆盖另一个开关。

### 强制路由最终确认

能力配置结束后才决定是否启用强制路由。最终文案按磁盘验真的 Provider 动态列名；OpenViking control-plane 单独存在不算可检索 Provider，只有 data-plane 真 key 才能显示 Agent 记忆。若集合为空，Helper 不安装 deny-all Gate，只提供“返回配置检索能力 / 不启用强制路由并退出”。用户确认后只安装真实 Provider allowlist Gate；原生 WebSearch 策略已在豆包搜索单项配置中独立决定，不属于路由闭环。

### 本地配置状态

- `✓ 已配置`：该能力要求的所有组件都明确就绪。
- `⚠ 未配置`：至少一个组件明确缺失。
- `? 未知`：探测命令失败或输出无法解析；不能把失败简化成“未配置”。
- `— 不支持`：所选 AI Agent 没有对应交付形态。
- 探测全程只读：不刷新凭证、不调用 Evolve 云端 status、不在渲染时自动安装。

Agent 进化当前官方适配 Claude Code / OpenClaw / Trae；安装命令为：

```bash
curl -fsSL "https://ark-self-evolve.tos-cn-beijing.volces.com/evolve_skill/latest/install.sh" | bash
```

CLI 只展示该命令，不自动执行未固定版本的远程脚本。安装完成后返回矩阵选择“刷新状态”。

## routing.v1 强制检索路由协议

### Route Plan schema

`submit_route_plan` 的参数就是 Route Plan 对象，不再包一层 `plan`：

```json
{
  "version": "routing.v1",
  "turn_id": "<从 enforcement context 原样复制>",
  "tasks": [
    {
      "id": "public-release",
      "intent": "latest_public_web",
      "provider": "doubao_search",
      "query": "ArkCLI 最新公开 release 信息"
    }
  ]
}
```

约束：

- `version` 必须精确为 `routing.v1`；`turn_id` 必须来自本轮 hook 上下文。
- `tasks` 至少一个且 `id` 本轮唯一；非 `none` task 必须有非空 `query`。
- 相同 Provider + 忽略大小写及首尾空格后相同的 query 不得重复。
- `provider` 必须严格匹配 `intent`，不能由模型自由选择。

### 混合请求

用户说：“查今天的 AI 新闻，再看我的知识库里上周的评审结论，并统计本地仓库 Go 文件数。”

```json
{
  "version": "routing.v1",
  "turn_id": "<active-turn-id>",
  "tasks": [
    {"id":"news","intent":"latest_public_web","provider":"doubao_search","query":"今天的 AI 新闻"},
    {"id":"review-notes","intent":"persistent_knowledge","provider":"openviking","query":"上周评审结论"},
    {"id":"local-count","intent":"local_or_compute","provider":"none","query":"统计当前仓库 Go 文件数"}
  ]
}
```

提交成功后分别调用豆包搜索、OpenViking 和本地工具；本地工具不消耗 Provider 授权。意图不明确时提交 `clarification_required -> none` 后追问，不调用搜索工具。`doubao_search` 是 `routing.v1` 中豆包搜索唯一合法的 wire provider 值。

### Provider 调用边界

- 豆包搜索：已加载时优先调用 `byted-web-search` Skill。仅当 Skill 未安装、用户明确要求 MCP，或 Skill 在业务请求发出前因本地命令/配置错误失败时，保留原始错误并只改用一次 `mcp-server-askecho-search-infinity`；鉴权、超时或服务错误表示业务请求已经发出，必须停止，不能换入口重试。
- DataPro：调用 `dataPro-search`。
- OpenViking：检索只调用 `openviking-dataplane`；control-plane 只管理库。
- Supabase：调用 `byted-supabase` Skill/CLI；仅 Helper 验真成功时可用。
- OpenClaw 通过 `mcporter call <server>.<tool>` 包装 MCP 时仍受 gate 校验，不能借 shell 绕过。

总领 Skill / 当前 Agent 根据 Prompt 做意图分类；路由 MCP 只提交计划，gate 只执行受管 Provider 授权。选择豆包搜索、DataPro、OpenViking、Supabase 时必须先提交计划；原生 `WebSearch`、`WebFetch`、`WebExtract` 位于 gate 之外，可直接调用且无需 Route Plan。用户明确指定某个受管 Provider 时不得在失败后静默换成原生工具；任何路径都不得读取、打印或手工传递搜索凭证。

若正确 Provider 不在 `Available providers`，不要提交错误 Provider：说明缺少的能力，建议重新运行 `arkcli helper configure <harness> --with-routing`（必要时加 `--with-supabase`），然后等待用户下一轮决定。

### 授权生命周期

```text
[UserPromptSubmit]
        |
        v
 [new turn_id] ---> [no active authorization]
        |
        v
[submit_route_plan accepted]
        |
        v
[only planned providers allowed]
        |
        v
[next user message] ---> [old authorization revoked]
```

计划无效、Provider 不可用或重放旧 `turn_id` 时保持拒绝状态。修正意图需要进入新的用户回合，禁止原地绕过。

## 移除

```
arkcli helper reset <harness>
```

完整移除 arkcli 注入的 provider/env/model + 内置 MCP server；Volc 还会移除 arkcli 托管的路由 hook/tool/plugin，并恢复 ArkCLI 拥有且未被外部修改的原生 WebSearch 字段(保留用户其它配置)。**该命令不是“只恢复搜索”**。Trae 项目级 MCP 注入用 `arkcli helper reset trae --scope project` 移除；原生搜索 Hook 始终按当前项目根恢复。Codex profile 注入用 `arkcli helper reset codex --codex-profile <name>` 移除;全局注入用 `arkcli helper reset codex --codex-config-scope global` 移除；原生搜索配置始终位于当前项目根 `.codex/config.toml`，与该 flag 独立。
