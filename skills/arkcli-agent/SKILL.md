---
name: arkcli-agent
version: 1.1.5
description: "arkcli agent：管理 ARK Managed Agents，包括 Agent / Skill / Env / Session / File / Memory Store / Vault / MCP OAuth。控制面优先走 ForTop/OpenTOP，Session 运行时和 Files 走数据面直联。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli agent --help"
---

# arkcli agent

**CRITICAL — 开始前 MUST 先读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)。**

当用户需要创建、查询、调试或联通 ARK Managed Agent 时使用本 skill。核心原则：先用稳定产品命令，不要直接猜 OpenTOP Action；Session 运行时资源、事件、线程、Files API 走数据面直联；MCP OAuth 登录先查后端 provider，再用 `arkcli agent +mcp-login`。

执行前按“先选路径”读取对应 reference；只读相关 reference，不需要一次性加载全部细节。用户请求如果已经明确是创建、复制、挂文件、聊天、MCP 登录等写操作，完成必要确认/消歧后直接执行，不要只给命令建议。

## 创建 Agent 的最小决策表

| 用户输入 | AI Agent 的处理 |
| ---- | ---- |
| 未指定 Agent 主模型 | 先执行一次 `arkcli agent model list --format json`，完整读取候选；不默认加 query、不截断数量、不凭印象拼模型 ID。用户明确要求上下文/模态/能力时，再按具体版本查询 metadata 并筛选 |
| 需要为工具选择绑定模型 | 用 `agent model list --usage tool --format json`；不默认加 `--primary-only`，选中的 `items[].model` 填入 `Tools[].Configs[].Models[].ID`，不修改 Agent 主模型 |
| 未指定 Skill | 先查本账号 custom skill；没有合适候选再查 market/SkillHub skill |
| 给出本地 Skill zip | 先调用 `agent skill create --zip`，拿到返回的 `skill-...` ID 和版本后再创建 Agent |
| 未指定工具 | 使用 CLI 注入的完整默认工具集；显式传 `--tool` 时全量替换默认工具 |
| 未明确要求多模态工具 | 不查询工具模型，不添加多模态工具集、image_generate/video_generate 或其 Models；保留普通默认工具，不能因所选模型支持多模态输入而自行添加 |
| 未指定环境 | 创建 Session 时自动选择当前项目最新环境；没有可用环境才提示创建或传入环境 ID |
| 创建成功 | 必须回读 `agent agent get <agent-id> --format json`，展示服务端最终的 Model、System、Tools、Skills、McpServers 和扩展配置 |
| 用户期待 Agent 回复 | 短请求使用 `+new session ... --message` 或 `events send ... --stream`；大 payload / 长耗时任务使用 `events send --poll`，或 send 立即返回后按 cursor 轮询 events / 使用 `+tail`，不要让所有任务都阻塞等待 |

## 业务目标澄清

模型、Agent、Skill、MCP provider 等目标无法唯一确定时，遵循
[`arkcli-shared`](../arkcli-shared/SKILL.md) 的 0/1/N 结构化选择契约：先做一次最小
只读查询，选项只取本轮完整结构化结果中的真实 ID 和区分字段。多个候选会改变远端结果
时使用宿主结构化选择能力；唯一候选直接继续。选择目标只完成消歧，不等于授权后续创建、
更新或删除。

模型选择是创建 Agent 的强制停点：只能用用户明确给出的硬约束和产品 eligibility 字段
过滤上述查询结果；相关度、返回顺序、推荐或 Agent 自己的“更适合”判断都不能把多个候选
变成唯一候选。过滤后为 N 个（N > 1）时必须把真实 `items[].model` 交给用户选择；
用户选定前不得查询 Agent Skill，也不得执行任何 preview、`agent agent create/update`、
`+new-agent` 或 `+iterate`。这也是当前回合的硬返回点：展示候选后立即结束回合，不再探测
上述写命令的 `--help`，不准备后续参数，也不查询 Skill/MCP 候选。推荐项可以标注理由，
但不能代替用户作出选择。

## 先选路径

| 用户意图 | 首选命令 | 细节 |
| ---- | ---- | ---- |
| 创建 / 更新 / 删除 Managed Agent | `arkcli agent agent ...` | [`references/agent.md`](references/agent.md) |
| 复制已有 Agent 并改名 / 局部覆盖配置 | `arkcli +new-agent --fork <agent-id> [--name <new-name>]` | [`references/agent.md`](references/agent.md#复制-agent) |
| 为创建 Agent 选择可用模型 | `arkcli agent model list` | [`references/agent.md`](references/agent.md#模型选择) |
| 为生图/生视频等支持模型绑定的工具选择模型 | `arkcli agent model list --usage tool` | [`references/agent.md`](references/agent.md#查询工具模型与缓存) |
| 查询模型运行参数的可选值和默认值 | `arkcli agent model config <foundation-model-name>` | [`references/model-config.md`](references/model-config.md) |
| 创建 Agent 时选择 Skill | 默认先查 custom，未命中再查 market；用户明确指定 market 时跳过 custom | [`references/skills.md`](references/skills.md) |
| 查询/使用本账号 custom skill | 先 `agent skill list --source custom --limit 100`，无匹配时沿 `NextPage` 传 `--page` 继续；需要完整候选时再用 `--page-all` / `--skill <skill-id>` | [`references/skills.md`](references/skills.md) |
| 查询 Ark Skill / 跨来源查询 | `agent skill list --source ark` 或 `--source all`；Ark Skill 只读 | [`references/skills.md`](references/skills.md) |
| 上传本地 custom skill zip | `arkcli agent skill create --zip <file>` 或 `agent agent create --skill-zip <file>` | [`references/skills.md`](references/skills.md) |
| 扫描或导入 GitHub Skill | `agent skill github scan <repo>` / `agent skill github import <repo> --path ...|--all` | [`references/skills.md`](references/skills.md#from-github-import) |
| 管理 custom Skill 版本 / 删除 Skill | 先完整列出版本，再按“非 latest → latest → Skill”的依赖顺序删除 | [`references/skills.md`](references/skills.md) |
| 创建运行环境 / 会话 | `arkcli agent env ...` / `arkcli agent session ...` | [`references/session-files.md`](references/session-files.md) |
| 自托管环境 / Work Queue 排障与停止任务 | `agent env create --runtime-type self_hosted` / `agent env work list/get/stats/stop` | [`references/self-hosted.md`](references/self-hosted.md) |
| Environment 初始化脚本 | `arkcli agent env create/update --setup-script @./bootstrap.sh` | [`references/session-files.md`](references/session-files.md#environment-自定义脚本) |
| Session 一次性覆盖 Agent / Environment | `arkcli agent session create --agent-overrides ... --environment-overrides ...` | [`references/session-files.md`](references/session-files.md#session-overrides) |
| 升级已有 Session 的模型 / Agent 版本 / 运行配置 | `arkcli agent session upgrade <session-id>` | [`references/session-upgrade.md`](references/session-upgrade.md) |
| Session 创建时绑定 TOS 目录 | 用户明确提供地址后使用 `arkcli agent session create --tos-path tos://<bucket>/<prefix>/`；未提供时先询问，不猜路径 | [`references/session-files.md`](references/session-files.md#session-tos-资源) |
| 选择/继续 Managed Agent 会话 | `arkcli +new session` | [`references/events-chat.md`](references/events-chat.md) |
| 直接创建新会话并聊天 | `arkcli +new session <agent-id> --environment-id <env-id>` | [`references/events-chat.md`](references/events-chat.md) |
| 给已有会话发消息或实时看回复 | 默认 `events send` 只负责写入；需要 SSE 回复时加 `--stream` 自动跟随 event cursor（`--wait` 保留为兼容别名），实时入口默认请求 `agent.message` / `agent.thinking` Event Deltas；也可使用 `+tail`/`events stream`，`--no-event-deltas` 回退完整事件；`+new session`/`+iterate` 内部自动使用补偿 channel | [`references/events-chat.md`](references/events-chat.md) |
| 主动压缩会话上下文 | `arkcli agent session compact <session-id> [--instructions <text|@file>]` | [`references/events-chat.md`](references/events-chat.md) |
| 看 session 诊断 / 导出诊断包 | `arkcli +debug <session-id>` / `arkcli +export <session-id>` | [`references/debug-export.md`](references/debug-export.md) |
| 上传文件并挂到已有 session | `arkcli agent session resources add <session-id> --path <file>` | [`references/session-files.md`](references/session-files.md) |
| 只上传 / 查询 Files API 文件 | `arkcli agent file upload/list/get/wait/delete` | [`references/session-files.md`](references/session-files.md) |
| 管理 memory store / memories | `arkcli agent memory-store ...` | [`references/interfaces-gaps.md`](references/interfaces-gaps.md) |
| 查询可挂载 MCP / 管理 Vault / Credential / MCP OAuth | `arkcli agent vault oauth-provider list` / `arkcli agent vault ...` / `arkcli agent +mcp-login ...` | [`references/mcp-vault.md`](references/mcp-vault.md) |

## 模型查询：主模型与工具模型

- 默认 `--usage agent` 按 `ep_agent.support` 选择主模型；`--usage tool` 按具体版本的 `epa_tool_model.support` 选择工具模型，两者不取交集。两种查询都不默认加 `--primary-only`，也不默认按 `primary_version` 排除候选；只有用户明确要求“只看主版本”时才加。
- 工具候选使用返回的 `items[].model`，不是内部 `id`（如 epm-*）。只给支持绑定的工具配置 `Models`；字段含义与完整 Tools 替换示例见 [工具模型绑定](references/agent.md#工具绑定模型仅部分工具支持)。
- 两种用途共享 5 分钟全版本 ArkModels 缓存，输出 `cache.source/fetched_at/expires_at`；只有需要刷新时才加 `--refresh-cache`，不要每次查询都强刷。失败冷却为 1 分钟，强刷也不能绕过；错误不等于空候选。
- 列表不截断数量；上下文/模态/能力筛选不再由 list flags 承担。只有用户明确提出这些要求时，才按候选的 `name` + `version` 查询 `ListModelMetaDatas` 后筛选；无额外要求不逐模型补查。流程、字段与缺失值处理见 [按具体版本筛选](references/agent.md#按具体版本筛选)。
- `agent model config` 查询运行参数，不提供工具支持性。用户明确给出的模型 ID 仍交给创建/更新接口判断，不把缓存查询变成写入硬拦截。完整参数、缓存隔离与 query 模式边界见 [查询说明](references/agent.md#查询工具模型与缓存)。

## 认证与 Profile

- 业务命令前先 `arkcli auth status --format json`。未登录、SSO 过期、STS refresh 失败时先处理登录。
- 当前数据面 Files / Session resources / events / threads 需要可用 ARK API Key。CLI 默认使用 profile / identity 解析出的 Key；也可用全局 `--api-key` 做本次调用覆盖。若同时自定义 `--base-url`，必须显式成对提供二者；无 Profile 的 stateless 模式也必须成对提供。
- 线上环境已就位，默认走 `--env prod`，不要再默认跑 stg。
- 非交互 SSO 登录是两段式：先 `arkcli auth login --no-browser` 拿 URL；用户贴回 base64 code 后，再跑 `arkcli auth login --no-browser --code <code>`。

## List 分页

- 支持分页契约的列表可加全局 `--page-all`；未显式传单页大小时，CLI 默认每页取 100 条。默认最多请求 10 页，可用 `--page-limit <N>` 调高，`--page-delay <ms>` 控制页间隔。
- 已支持：Agent/版本、Env、Session、Skill market/custom/ark/all、Memory Store/Memories、Vault/Credentials/OAuth Provider、Files、Session Events/Threads。CLI 会分别按后端契约使用 `Page`、`PageNumber`、`PageToken`、`after`，并合并结果。
- `agent model list`、`memory-store creators`、`session resources list` 没有可用分页契约，不要为它们假设 `--page-all` 能补全结果。命中 `--page-limit` 后应检查返回的 `NextPage`、`has_more` 或 `TotalCount`，判断是否仍有未拉取数据。

## 删除确认

- Managed Agent 的破坏性 `delete` 命令在真实 TTY 且未传 `--yes` 时会显示不可逆警告并询问 `[y/N]`；输入 `y/yes` 才会调用后端，其他输入会取消。
- 非交互环境（AI Agent、CI、管道）不会读取 stdin；未传 `--yes` 时返回 `type=requires_confirmation`，不会调用后端。只有用户已经明确确认删除目标后，调用方才可以补 `--yes` 重试。
- `--dry-run` 不是全域能力。只有命令自身 `--help` 列出该 flag 时才可用；
  当前主要覆盖可由本地 payload 确定的 Env/Session/Memory/Vault/Credential
  写请求，以及 `agent agent create/update/delete`、`agent skill create/update/delete`
  和 `agent file upload/delete`，以及会写本地文件的 `agent skill download`、
  `+export`。`+new-agent`、`+iterate`、MCP login 与所有纯读命令都不注册。
  Client Preview 只生成零网络 `preview.v1` 计划；它不代替真实执行前的开通、
  版本/依赖校验或删除确认。

## 长流程执行规则

- 一个 shell/tool 调用只执行一个 `arkcli` 命令。不要把 `session create`、写入 ID、`events send`、`events list` 用 `&&`、`;`、管道或 heredoc 串成一个长命令。
- `session create`、Agent 创建、文件上传等写操作成功后，立即从结构化输出提取并保存 ID；下一次调用使用已经确认的字面量 ID，不要依赖同一 shell 中的变量赋值继续执行。
- 任一步超过预期时间时，先结束该步并单独执行 `session get`、`events list` 或 `+debug` 诊断；不要让后续命令被前一个阻塞步骤掩盖。
- 大 JSON 或大文本不要在同一个 shell 调用中通过 Python/Node 管道解析；先让 `arkcli --format json` 独立返回，再在下一步解析结果。这样即使请求超时，也能区分是创建、发送还是回查阶段失败。

### 超时与重试

- 只对网络超时、连接中断、429 或 5xx 做有限重试；参数校验、鉴权失败、未开通、权限不足和明确业务错误直接抛出，不要重试。
- `session create` 超时后结果可能未知。先用 `session list/get` 按返回的标题、Agent、环境和创建时间检查是否已经创建，再决定是否重试；没有幂等键时不要盲目重复创建。
- `events send` 超时后也可能已经被服务端接收。先用返回的 event cursor、发送时间或最近的 user event 查询 `events list`；确认没有接收记录后才允许重试，避免重复发送用户消息。
- `events send --stream` 达到 stream 等待上限后会自动切换为 events list polling，默认再等待 120 秒；`--wait` 保留相同行为作为兼容别名。两阶段都超时才返回带 cursor 恢复命令的非 0 错误。此时不要重发原消息，继续 list/poll 同一个 cursor。
- `events list/get` 属于只读请求，可以使用有限次数、指数退避的重试；继续使用同一个 `after` cursor，不要因为重试而从历史开头重新读取。
- 每一步都要保留步骤名、Session ID、event cursor、尝试次数和最后错误；达到重试上限后抛出带上下文的错误，不要把超时伪装成成功。

## 命令速查

| 命令 | 说明 |
| ---- | ---- |
| `arkcli agent agent list/get/create/update/delete/versions` | Agent CRUD + 版本 |
| `arkcli agent model list` | 默认查询主模型，`items[].model` 用于 `--model`；`--usage tool` 查询工具模型，结果用于 `Configs[].Models[].ID`；`--query` 可增强详情/排序 |
| `arkcli +new-agent` | Agent create 增强入口；支持 `--fork/--from` 复制已有 Agent 后创建新 Agent |
| `arkcli +iterate` | 更新 Agent 配置，创建新 Session，并进入 one-shot/REPL 试运行；`--environment-id/--env-id` 可选，省略时自动选择最新环境 |
| `arkcli agent skill search/list/get/create/update/delete/versions/download/set-protection/github` | Market/custom/Ark Skill 查询、GitHub 导入、custom Skill 保护与版本管理；Ark Skill 只读 |
| `arkcli agent env list/get/create/update/delete` | Environment CRUD；当前 `env list` 没有 `--status`，状态筛选规则见 `session-files.md` |
| `arkcli agent session list/get/create/update/delete` | Session CRUD |
| `arkcli agent session upgrade <session-id>` | 升级已有 Session 运行配置，返回受理状态；使用 snake_case 请求并回读验证 |
| `arkcli agent session resources list/add/get` | 数据面 session resources；get 是 CLI 基于 list 的本地筛选 |
| `arkcli agent session events list/send/stream` | 数据面 events；stream 默认请求 Event Deltas 并输出 SSE/NDJSON 行；`--no-event-deltas` 回退完整事件；`user.custom_tool_result` 必须带 `custom_tool_use_id`，`user.tool_result` 仅允许 self_hosted，CLI 会前置校验 |
| `arkcli agent session compact <session-id>` | 主动压缩上下文；正常 thread idle/end_turn + session idle/end_turn 表示命令完成，compacted 事件是可选的实际压缩确认 |
| `arkcli agent session threads list/get` | 数据面 threads |
| `arkcli agent file list/get/upload/wait/delete` | 数据面 Files API |
| `arkcli agent memory-store list/get/create/update/delete` | Memory Store CRUD |
| `arkcli agent memory-store memories list/get/create/batch-create/update/delete` | Memory CRUD |
| `arkcli agent vault list/get/create/update/delete` | Vault CRUD |
| `arkcli agent vault oauth-provider list` | 查询后端已注册 MCP Provider；返回的 MCP server 信息可用于 Agent `--mcp-server` |
| `arkcli agent vault oauth-flow create` | 裸创建 Vault OAuth Flow，适合脚本自带 redirect URL |
| `arkcli agent vault credentials list/get/create/update/delete` | Credential CRUD |
| `arkcli agent +mcp-login` | 托管 MCP OAuth 登录：本地 callback + CreateVaultOAuthFlow + 等待 credential 创建 |
| `arkcli +chat <prompt>` | Responses API 快速对话；不要把它当 Managed Agent session 入口 |
| `arkcli +tail <session-id>` | 人类可读 event stream |
| `arkcli +new session` | Managed Agent session 选择器；可继续已有 session，或选 agent/env 起新 session |
| `arkcli +new session <agent-id> --environment-id <env-id>` | Managed Agent 新 session 直达入口；固定先创建新 session，再 REPL / one-shot |
| `arkcli +debug <session-id>` | 聚合 session、events、resources、threads 做诊断 |
| `arkcli +export <session-id>` | 导出诊断 tar.gz |

复杂字段如 `Tools`、`Skills`、`McpServers`、`Multiagent`、`Metadata`、`Tags` 支持 JSON/YAML 文件、stdin 或结构化 flags。TOP 请求使用 CamelCase，inline 对象兼容常见 lower/snake case alias；`session upgrade` 是例外，使用原生 snake_case，只读取显式文件/flags，不自动读取 stdin。创建成功回显必须展示服务端最终的身份、模型、`System`、Tools、Skills、MCP 和扩展配置，不能只展示摘要；需要核对完整结果时用 `agent agent get <agent-id> --format json` 或 `--format yaml`。

Memory get 的 `--view` 语义：默认按 `basic` 请求，只返回 metadata 和 `ContentSha256`；`--view full` 保留 Content。若服务端仍在 basic 下返回 Content，CLI 会在输出层剥离 Content，但这只能控制输出，不能挽回已经产生的网络传输；服务端仍需正确实现 View 参数以节省带宽。

## 参考

- [`references/agent.md`](references/agent.md)：创建、复制、查找 Agent、默认工具、工具模型绑定与查询缓存。
- [`references/skills.md`](references/skills.md)：Market skill 搜索、custom skill zip 上传、skill ref 组装。
- [`references/session-files.md`](references/session-files.md)：Env / Session / Files API / Session resources。
- [`references/events-chat.md`](references/events-chat.md)：events send/list/stream、`+new session` 选择器、`+tail`。
- [`references/mcp-vault.md`](references/mcp-vault.md)：Vault、Credential、MCP OAuth、Agent 挂 MCP。
- [`references/debug-export.md`](references/debug-export.md)：`+debug`、`+export`、端到端验证模板。
- [`references/interfaces-gaps.md`](references/interfaces-gaps.md)：接口链路和当前 gap。
- [`references/evals.md`](references/evals.md)：Agent/Skill/GitHub/Compact 最小评估用例。
