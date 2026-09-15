# Agent

## 创建 Agent SOP

用户说“创建一个 XXX agent / 智能体”时按下面链路执行，不要只拼一个 `agent create`。

1. `arkcli auth status --format json`，确认登录、profile、project、API key。
2. 如果用户没有给精确模型，先完整查询候选：`arkcli agent model list --format json`。用户明确提出上下文/模态/能力要求时，才按[具体版本 metadata](#按具体版本筛选)补查并筛选；无此要求不额外逐模型查询。把返回的 `items[].model` 原样作为 `--model`。随后必须先完成 0/1/N 分支：0 个候选时报告无候选并停下；1 个候选时复述其 ID 后才进入步骤 3；多个候选时展示真实候选并**立即结束当前回合**，用户选定前禁止执行步骤 3 及后续步骤。
3. 模型已经唯一确定后，才把用户意图扩展成 skill 选择上下文。例：数据分析 -> `数据分析 Excel CSV 表格 BI SQL`；代码助手 -> `代码 编程 repo bash`；文档写作 -> `文档 写作 总结 Markdown`。
4. 创建 Agent 默认优先从本账号已有 custom skill 中选择，即使用户没有显式说“使用 custom skill”：按需执行 `arkcli agent skill list --source custom --limit 100 --format json`。由 AI agent 读取这一页全部 `Items`，按名称、描述、能力和版本判断；没有合适候选时，将响应中的 `NextPage` 原样传给 `--page` 继续拉下一页，直到命中或没有下一页。不要只调用 `search --source custom "<query>"` 后选第一条；用户明确要求完整清单或需要离线分析时才使用 `--page-all`。
5. custom skill 分页查完仍没有合适候选，或用户明确要求 market/SkillHub skill 时，再搜索：`arkcli agent skill search "<query>" --limit 10 --format json`。根据名称、描述、能力标签、版本选择 skill。不要臆造 `SkillId`；搜不到时可创建基础 agent，并说明未找到匹配 skill。
6. 组装参数：补领域化 system prompt，按服务端 `ModelMeta` 和用户需求选择 `--thinking`、`--reasoning-effort`、`--service-tier`；不要默认注入旧版 `speed=standard`，也不要把 `speed` 映射成新字段。线上测试资源名使用 `arkcli` 前缀；用户没给名字时生成 `arkcli-<domain>-agent-<YYYYMMDDHHMMSS>`，如 `arkcli-data-agent-20260707153000`。
7. 先执行同一条 `agent agent create ... --dry-run --format json`，读取零网络 `preview.v1`，向用户复述 `DisplayName`、`Model`、`Skills`、默认 `Tools`、`McpServers` 和 `unresolved`。裸 custom Skill 省略 `Version` 表示始终使用服务端最新版本，不需要在线补版本；`--skill-zip` 上传结果仍是实际执行时才能解析的占位符。用户确认后去掉 `--dry-run` 再真实创建。
8. 真实创建时，CLI 会先检查 Managed Agent 能力和模型开通状态：模型未开通会走共享模型开通确认链路；非交互环境不会自动开通，返回 `model_activation_required`。Managed Agent 产品/能力未开通时，TTY 下会提示用户确认并调用前端同款 `OpenChargeItems(ResourceType=DataManagedAgentSum, ResourceNames=[sandbox, web_search])`，非交互环境返回 `managed_agent_activation_required`，不会自动开通。
   - 如果已在对话中拿到用户明确确认，非交互调用可重跑原命令并同时加 `--yes` 和环境变量 `ARKCLI_ALLOW_HEADLESS_ACTIVATION=1`。不要在没有用户确认时设置该环境变量。
9. 真实创建后立刻 `agent agent get <agent-id> --format json` 确认落库。对用户回显时必须展示服务端最终配置，不要只展示“已创建”或单独摘要某个字段。
10. 用户要求端到端验证时，再创建 env/session，发送一条最小消息，拉 events/thread/resources。除非用户明确要求清理，不要删除创建出的资源。

只有模型已经由用户明确给出，或白名单过滤后唯一确定时，skill 与 MCP provider 候选才可以并行查。模型仍有多个候选时不得运行下面的 Skill/MCP 查询，也不得用 `agent agent create --help`、preview 或参数准备来提前推进创建流程：

```bash
arkcli agent model list --format json
arkcli agent model list --query "<capability-query>" --format json
arkcli agent skill search "<capability-query>" --limit 10 --format json
arkcli agent vault oauth-provider list --limit 100 --format json
```

## 模型选择

配置模型的 `speed/thinking/reasoning_effort/service_tier` 时，先按[模型参数 metadata](model-config.md)查询目标模型的可选值与默认值；不要把其他模型的选值当作通用枚举。

`agent agent create --model` 必须传精确可用的模型 ID。不要凭印象写裸模型名或展示名。

默认完整列出可用候选：

```bash
arkcli agent model list --format json
```

`--query` 模式仍以 ArkModels 白名单为主表，不从模型目录反向生成候选。它会调用 `models search` 拿详细信息并增强白名单模型：命中的白名单模型会带 `detail` 字段并排在前面；未命中的白名单模型仍保留，只是没有 `detail`。`detail` 字段包含用于判断适配度的信号，例如 `display_name`、`description`、`context_window`、`input_modalities`、`output_modalities`、`capabilities`、`lifecycle_status`。

需要额外详情和相关度排序时，可选用（不是默认步骤）：

```bash
arkcli agent model list --query "数据分析 Excel CSV SQL agent" --format json
```

选择规则：

- 默认只从 `agent_support=true` 的结果里选；`agent model list` 默认已经过滤非 Agent 模型。
- 主模型和工具模型查询都不默认加 `--primary-only`，也不在本地默认排除 `primary_version=false` 的候选；只有用户明确要求只看主版本时才启用。主版本不等于最新版本，同名多版本时不要跨条目混拼。
- 创建时传用户最终选中条目的 `model` 字段，不要传返回里的 `id`，也不要把列表第一项当成默认选择。
- 先只按用户明确给出的模型族、模态、上下文长度、能力等硬约束，以及所选用途的 `agent_support` / `tool_model_support` 过滤；`primary_version` 仅在用户明确要求主版本时作为过滤条件。查询排序、展示顺序、`router_baseline_support` 或 Agent 对“性能更强 / 更合适”的主观判断都不能把多个候选变成唯一候选。
- 过滤后只有 1 个候选时复述其 `model` 后继续；只要硬约束过滤后仍有多个候选，就必须使用宿主结构化选择能力展示本轮结果中实际存在的 `model`、`name/display_name`、`version`、`context_window`、`capabilities`、`lifecycle_status/status` 等区分字段，并停在模型选择阶段。候选输出就是当前回合的最终输出，之后不再调用任何工具。可以标注推荐项和依据，但用户选定前不得继续查询 Agent Skill/MCP，不得查看创建命令的 `--help`，也不得执行 `agent create/update`、`+new-agent`、`+iterate` 或它们的 `--dry-run`。不要从模型记忆补选项；宿主没有结构化选择能力时退化为精简编号列表。
- 用户明确要求某个模型族时，用 `--name <keyword>` 缩小 Agent 白名单范围：

```bash
arkcli agent model list --name doubao-seed-2-0-pro --format json
```

- 需要排障或确认为什么某模型不可选时，加 `--include-all`，查看 `agent_support=false` 的条目。
- 有明确硬指标时按[具体版本 metadata](#按具体版本筛选)筛选；list 不支持 `--size`、`--modality`、`--input-modality`、`--output-modality`、`--multimodal`、`--min-context-window`、`--capability`、`--strict-filter`。这不改变通用 `models search` 的参数。
- 用户已经给了完整模型 ID 时可以直接用，但如果创建失败提示模型不支持 / 不存在，回到 `agent model list` 重新选择。

### 按具体版本筛选

仅当用户明确提出如“上下文至少 200000 tokens”“支持图片输入”“支持 function calling”等要求时执行：

1. 先 `agent model list`（工具用 `--usage tool`）获取完整资格候选；保留每条的 `name`、`version`、`model`。名称约束可先用 `--name`，不默认只看主版本。
2. 若已有同一 `name` + `version` 的 metadata，直接复用。否则按需查询 `ListModelMetaDatas`。`models get <name> --version <version>` 可查产品详情；需要原始 metadata 时无专用产品命令，使用 [API Explorer](../../arkcli-api-explorer/SKILL.md)，先确认 registry 和本地 preview，再实际只读调用：

```bash
arkcli api model.list_model_meta_datas \
  --params '{"FoundationModelName":"<items[].name>","FoundationModelVersion":"<items[].version>","Keys":["context_window"],"PageNum":1,"PageSize":1000}' \
  --format json
```

占位符替换为候选的真实字段。只查用户要求的 keys，可在同一次请求批量传多个 key：

| 用户约束 | metadata key / 判定 |
| --- | --- |
| 上下文至少 N tokens | `context_window`：将 `Data[0]` 解析为正整数，再与 N 比较 |
| 输入/输出模态 | `input_modalities` / `output_modalities`：检查 `Data` 中的 text/image/video/audio 等实际取值 |
| 思考 / function calling / MCP | `thinking.support` / `functioncall.support` / `mcp.support`：读取实际支持值，不从名称推断 |

3. 从实际响应的 `ModelMetaDatas` 或 `Result.ModelMetaDatas` 读取，核对返回的模型名/版本；顶层分页信息如显示未读完，再增加 `PageNum`，不要依赖 `--page-all`。metadata 行的 `Version` 是编辑版本，不是 `FoundationModelVersion`。
4. 只用对应版本的数据筛选；缺失、空值、无法解析或查询失败表示“未知”，不是 0/不支持，也不能声称满足硬指标。明确区分已满足、不满足、待确认；若用户要求严格满足，只推荐已验证项，并说明未知项未被验证。多个满足项仍交由用户选择。

`--query` 的 `detail` 按名称关联，可能来自同名主版本，不能当作其他版本的硬指标证据。此处 `ListModelMetaDatas` 也不是 `agent model config` 使用的 `ListManagedAgentModelMetaDatas`；后者查询运行参数可选值。工具资格仍仅从 ArkModels 的版本级 `epa_tool_model.support` 判断。不要为无额外约束的列表/创建自动补查 metadata，也不要把这套选型流程变成对用户明确模型 ID 的写入拦截。

### 模型值示例

```bash
arkcli agent agent create \
  --name arkcli-data-analysis-agent-20260706 \
  --model <items[].model from agent model list> \
  --system "You are a data analysis agent. Help users inspect datasets, reason about metrics, write analysis code, and summarize findings clearly." \
  --format json
```

Agent 支持将内部唯一名和展示名分开：`--name` 是稳定资源名，`--display-name` 是用户可见名称。创建和更新模型参数可用独立 flags，也可放进 `--model` 对象；独立 flags 优先：

```bash
arkcli agent agent create \
  --name arkcli-reasoning-agent-20260828 \
  --display-name "推理助手" \
  --model '{id: <items[].model>, thinking: enabled, reasoning_effort: high, service_tier: auto}' \
  --reasoning-effort medium \
  --system @./system.md \
  --format json
```

- `--model` 接受 `id`、`provider`、`speed`、`thinking`、`reasoning_effort`、`service_tier` 及对应 camel/Pascal aliases。
- `speed` 只为旧版兼容保留；不传时保持省略，CLI 不补默认值。
- `agent agent update <id> --display-name ""` 会清空展示名。
- 普通更新不需要传 `--version`，例如 `arkcli agent agent update <id> --description "新描述"`。CLI 在调用 `UpdateAgent` 前自动 `GetAgent`，将当前 `Result.Version` 填入请求；不要让用户手动查询版本再填写。若同时需补齐模型 ID 或合并 Metadata，复用同一次读取。
- `--version <正整数>` 保留给需要锁定已读版本的并发检查；也支持 `--file`/stdin 中的 `Version`，显式 flag 优先。省略或值为 `0` 表示自动读取，负数不是有效版本。自动读取失败或未返回有效版本时不会更新；读取之后发生版本冲突时保留服务端错误，不自动换版本重试覆盖。
- 省略版本时，`--dry-run` 零网络展示 `GetAgent -> UpdateAgent`，`Version` 为 `<current-agent-version>` 且标记 `unresolved`，不是实际版本或版本预留；真实执行时才读取。显式版本且无需模型/Metadata 查询时只展示 `UpdateAgent`。此规则仅适用于 Agent update，不改变 Session 的版本参数语义。
- update 只修改运行时模型字段时不要求再次传模型 ID，例如 `--reasoning-effort high`；真实执行会先读取当前 Agent，并把当前 `Model.Id` 补入 `UpdateAgent` 请求。
- 用户通过 `--model <new-id>`、`--model` 对象或 `--file` 显式提供 `Model.Id` 时，以用户的新 ID 为准，CLI 不会用旧 ID 覆盖；独立 `--thinking`、`--reasoning-effort`、`--service-tier`、`--speed` flag 仍覆盖 model 对象中的同名字段。
- update 的 `--model` 对象不含 ID 时按运行参数 patch 处理，并补当前 `Model.Id`；若当前 Agent 没有返回模型 ID，则在写入前失败。
- `agent agent update ... --dry-run` 始终零网络。缺少 `Model.Id` 时，preview 会展示 `GetAgent -> UpdateAgent` 两步，并用 `<current-agent-model-id>` 与 `unresolved` 表示执行期才能取得的值；不得把 placeholder 当成真实 ID。
- `agent agent list --display-name <keyword>` 按展示名过滤；人类可读表格分别显示 `DISPLAY_NAME` 与稳定 `NAME`。

## 复制 Agent

用户说“复制 / fork / 基于已有 agent 改一个新版”时，优先用 `+new-agent --fork`，不要手动 get 后重新拼完整 create 请求。

```bash
arkcli +new-agent --fork agent-xxx --format json
```

- 用户明确给了源 `agent-id` 但没给新名字时，不要停下来只问名字；CLI 会先 `GetAgent`，默认用源 Agent 的 `Name` 加 `copy-` 前缀，即 `copy-<source-agent-name>`。如果源 name 为空，才 fallback 到 `copy-<agent-id-tail>`。
- 如果用户只说“复制那个数据分析 agent”但没给 ID，先用 `agent agent list --name <keyword>` 查候选；候选唯一时可继续复制，候选多个时使用宿主结构化选择能力展示本轮结果中的 `Id`、`Name`、`Description`、`UpdatedAt/UpdateTime`，拿到准确 `agent-id` 后再继续。
- `+new-agent --fork` 需要在线读取源 Agent，无法提供纯本地 Client Preview，因而不注册 `--dry-run`。先用 `agent agent get <id>` 核对源配置，复述覆盖项并取得确认后执行复制。
- `--fork` / `--from` 会先 `GetAgent`，复制源 Agent 的 `Model`、`System`、`Description`、`Tools`、`Skills`、`McpServers`、`Multiagent`、`Metadata`、`Tags`，再调用 `CreateAgent` 创建新 Agent。
- `--name` 可显式覆盖默认复制名。
- 用户传入的 create flags 会覆盖复制来的配置，例如 `--system`、`--model`、`--description`、`--skill`、`--tool`、`--mcp-server`。
- P0 语义里，`--skill` / `--tool` / `--mcp-server` 是替换对应列表，不是追加；需要追加时先用只读 `agent agent get` 看源配置，再传完整列表。
- 默认创建成功后会再次 `GetAgent` 回显最终配置；只想拿 `CreateAgent` 响应用 `--no-echo`。
- 创建结果中的 `System` 就是实际生效的 system prompt。人类可读回显至少必须展示以下字段：
  - 身份：`Id`、`Name`、`Description`、`Version`、`ProjectName`
  - 模型：`Model.Id`、`Model.Speed` 及服务端返回的其他模型配置
  - Agent 行为：完整 `System`、`Tools`、`Skills`、`McpServers`
  - 扩展配置：`Multiagent`、`Metadata`、`Tags`
  - 服务端返回的时间字段：`CreateTime`/`CreatedAt`、`UpdateTime`/`UpdatedAt`
- 结构化输出使用 `agent agent get <agent-id> --format json` 或 `--format yaml` 保留服务端返回的全部非空字段；调用方不得丢弃、截断或用摘要替换配置字段。人类可读摘要可以压缩时间、ID 等展示格式，但不能隐藏上述配置内容。
- 如果提交的 `request.System` 非空但创建响应或 `GetAgent.Result.System` 为空，优先报告“服务端未回显/未落库”，不要假设 prompt 已生效，并保留请求值与服务端值供排查。
- `+new-agent` 当前不做 LLM 起草和 template；自然语言理解、参数选择、用户确认由调用 arkcli 的 AI agent 完成。
- `+new-agent` 与 `agent agent create` 共用创建链路：真实创建前会做 Managed Agent / 模型开通预检。遇到 `managed_agent_activation_required` 或 `model_activation_required` 时，不要通过自动加 `--yes`、自动购买套餐或自动调用开通接口绕过确认；需要真人在 TTY 确认，或用户显式要求无人值守并设置 `ARKCLI_ALLOW_HEADLESS_ACTIVATION=1`。

## 迭代 Agent 并新建 Session

用户说“改一下这个 agent 然后试试 / 更新 system 后跑一遍 / 调整 tools 后开新 session 验证”时，用 `+iterate`，不要手动串三四条命令。

`+iterate` 会：

1. `GetAgent` 读取当前版本。
2. 如果传了更新 flag，则用当前版本调用 `UpdateAgent`。
3. 调 `CreateSession` 起一个新 session。
4. 有 `--message` 时发送首条消息并流式输出；TTY 且无 `--message` 时进入 `+new session` REPL；非 TTY 无 message 时只输出结构化结果。

```bash
arkcli +iterate agent-xxx \
  --system @./prompts/da-v2.md \
  --message "用新版配置说明你会如何分析 sales.csv" \
  --format json
```

- `--environment-id` / `--env-id` 可选；省略时 CLI 使用当前项目最近创建的 Environment。对环境有明确要求时应显式传入。
- 不要因为用户没有提供环境 ID 就中断或要求补参：真实执行时 CLI 会用 `ListEnvironments` 按 `CreateTime Desc, Limit=1` 自动选择最新环境；只有没有可用环境时，才提示先创建环境或显式传入 `--environment-id`。
- `--diff` 不调用 `ListEnvironments`，会在在线差异预览的 `CreateSession.EnvironmentId` 中显示 `<auto-select-latest-environment>`；这是占位符，不要把它作为真实环境 ID 发送。`+iterate` 不注册 Client Preview `--dry-run`。
- `--resource`、`--vault-id` / `--vault-ids`、`--tags` 会传给新 session。
- `--diff` 会先读取当前 Agent，再只预览 `UpdateAgent`、`CreateSession`、chat send/stream 请求，不改远端；它是命令专用的在线 diff，不是零网络 Client Preview。
- `--no-chat` 只更新 agent 并创建 session，不发送消息、不进 REPL。
- `--tool` / `--skill` / `--mcp-server` 在 iterate 中仍是全量替换语义。
- `+iterate` 单独修改 `--thinking`、`--reasoning-effort`、`--service-tier` 或 `--speed` 时，会复用已读取 Agent 的当前 `Model.Id`；显式新模型 ID 仍优先。
- `--display-name`（包括空字符串清除）属于 Agent 更新，不得被当成“无更新”跳过。
- 使用 `--agent-overrides` 创建 Session 时，Agent 版本写在 `AgentWithOverrides.Version`；最终 CreateSession 请求不得再包含互斥的顶层 `AgentId` / `AgentVersion`。

## 脚本输出

`+new-agent` 输出兼容 `data.agent` 结构，适合脚本提取：

```bash
arkcli +new-agent --fork agent-xxx --format json --transform "data.agent.id"
arkcli +new-agent --fork agent-xxx --format yaml > new-agent.yaml
arkcli +new-agent --fork agent-xxx --no-echo --format json
```

- `--transform "data.agent.id"` 只输出新 Agent ID。
- `--format yaml` 会把结构化结果渲染成 YAML，适合落地为文件检查。
- `--no-echo` 跳过创建后的 `GetAgent` 回显，只输出 `CreateAgent` 响应；默认不加时会再次 `GetAgent` 并回显服务端最终配置。
- 需要确认最终 prompt 时不要使用 `--no-echo`；创建后读取 `data.agent.system`（`+new-agent`）或 `Result.System`（`agent agent create/get`），并原样展示给用户。
- PRD 里的 `arkcli +new-agent "..."` 自然语言起草模式当前未实现；AI agent 应先把自然语言转成结构化 flags，再调用 `+new-agent` 或 `agent agent create`。

## 查找 Agent

- `arkcli agent agent list` 默认只调用一次 TOP `ListAgents`，返回单页 `Items`、`NextPage`、`Total`。
- 需要完整遍历时加全局 `--page-all`；CLI 会沿服务端 `NextPage` 拉取并合并 `Items`，最多拉 `--page-limit` 页，页间隔由 `--page-delay` 控制：

```bash
arkcli agent agent list --page-all --page-limit 10 --format json
```

- 未加 `--page-all` 时，不要默认假设已经拿到账号下全部 Agent。可以读取返回的 `NextPage` 手动继续：

```bash
arkcli agent agent list --page <NextPage> --limit 50 --format json
```

- 用户给明确 Agent ID 时，直接用 `arkcli agent agent get <agent-id> --format json`。
- 用户只给名字或模糊描述时，先用过滤缩小范围：

```bash
arkcli agent agent list --name <keyword> --limit 20 --format json
```

- 如果筛出多个候选，不要臆造选择；使用宿主结构化选择能力展示本轮结果中的 `Id`、`Name`、`Description`、`UpdatedAt/UpdateTime`，宿主不支持时才退化为精简编号列表。用户选定后直接复用该 `Id`，不要为同一批候选重新查询。
- 用户要“复制某个 agent”但没给 ID 时，先通过 `list --name/--ids` 或候选确认拿到准确 ID，再走 `+new-agent --fork <agent-id>`。

## 自定义工具（custom tool）

`agent agent create` / `+new-agent` 支持通过 `--tool` 声明 custom tool；更新时使用同样的结构。它是 `Tools` 数组中 `Type: custom` 的一项，没有单独的 `custom_tools` 字段或 `--custom-tools` 参数，也不是 custom Skill。

| 参数 | 类型 / 要求 | 含义 |
| --- | --- | --- |
| `Type` | string，必填 `custom` | 自定义工具类型 |
| `Name` | string，必填 | 模型调用的工具名；当前后端要求 1–128 个英文字母、数字、下划线或连字符，同一 Agent 内 custom tool 名不能重复 |
| `Description` | string，必填 | 说明用途、适用场景和行为；当前后端上限为 10,000 字符 |
| `InputSchema` | object，custom tool 必填 | 工具入参 schema，直接传对象，不能传 JSON 字符串或 base64 |
| `InputSchema.Type` | string，可选 | 值为 `object`，省略时后端默认 `object`；建议显式写出 |
| `InputSchema.Properties` | object，可选 | 参数名到 JSON Schema 的映射，例如 `city: {type: string, description: 城市名称}`；参数名和内部 schema 键名保留原样，不转为 PascalCase |
| `InputSchema.Required` | string[]，可选 | 必填参数名列表，例如 `[city]`，名称应与 Properties 中的键一致；无必填参数时省略或传 `[]` |

当前后端最多允许一个 Agent 配置 8 个 custom tools。名称、数量等业务校验交给服务端；CLI 不另加一套规则。custom tool 不使用 toolset 的 `Configs`、`DefaultConfig` 或 `McpServerName`。

下面只配置一个 custom tool（不保留默认工具）；模型 ID 替换为已确定的可用 ID：

```bash
arkcli agent agent create \
  --name arkcli-weather-agent \
  --model '<selected-model-id>' \
  --tool '{
    Type: custom,
    Name: get_weather,
    Description: 查询指定城市的天气,
    InputSchema: {
      Type: object,
      Properties: {city: {type: string, description: 城市名称}},
      Required: [city]
    }
  }' \
  --dry-run --format json
```

- `--tool` 接受 JSON/YAML 单个对象、数组或重复参数；工具字段支持 PascalCase / camelCase / snake_case 别名，如 `InputSchema` / `inputSchema` / `input_schema`。文件/stdin 的完整请求建议按 TOP 字段写 `Tools: [...]`，每项使用上表结构。
- 所有传入项合成完整 `Tools` 数组并替换原值，不会追加默认工具。需要保留默认工具时，将 `agent_toolset_20260701` 工具集与 custom tool 一起放入数组；更新追加工具前先 get，再提交完整列表。`--tool '[]'` 清空工具。
- 先检查 dry-run 中 `InputSchema` 为对象、`Properties` 内容未被改写，再真实创建并 get 回读。声明工具不是上传可执行实现，CLI 不会自动执行任意自定义函数；工具调用结果通过 [custom tool result 事件](events-chat.md)回传，使用对应的 `custom_tool_use_id`。

## 工具绑定模型（仅部分工具支持）

创建 Agent 时，只有用户明确要求使用多模态工具（如生图、视频生成），或明确提供这类工具配置，才进入本节查询和配置流程。否则不运行 `agent model list --usage tool`，不添加 `agent_toolset_multimodal_20260825`、`image_generate`、`video_generate` 或其 `Configs[].Models` 参数；不因模型支持图片/视频输入、工具模型有候选或示例包含这些字段而自行添加。未指定工具时继续使用普通默认工具集，不要因此传 `--tool '[]'` 清空它。用户明确给出的配置应保留，不通过此规则静默删除。

`Tools[].Configs[].Models` 不是所有工具的通用参数，也不是 Agent 主模型 `Model`。当前后端仅允许 `Type: agent_toolset_multimodal_20260825` 下的 `image_generate`、`video_generate` 配置它；不要给普通工具、custom tool、MCP 或 `DefaultConfig` 添加 `Models`。CLI 不自动注入模型，也不维护本地工具白名单；是否支持、模型权限和模态匹配由服务端校验。

每个 `Models` 数组元素的字段：

| 字段 | 类型 / 要求 | 含义 |
| --- | --- | --- |
| `ID` | string，必填 | 模型或 Endpoint 的准确 ID，模型身份的唯一依据；注意 TOP 字段为大写 `ID` |
| `Default` | boolean，可选 | 是否为该工具配置的首选模型；显式 `false` 会保留，省略不由 CLI 补值 |
| `Name` | string，可选 | 展示名称，不用于识别模型，不能替代 ID |
| `Version` | string，可选 | 展示版本，不用于识别模型，也不是 Agent 的版本 |

当前服务端每个配置最多允许 10 个模型，ID 不可重复；多个模型时必须且只能有一个 `Default: true`，单个模型可省略 `Default`。不要把 Agent 主模型白名单当成生图/生视频模型清单；先按下面的工具用途查询，再确认目标工具的模态和权限。

### 查询工具模型与缓存

```bash
arkcli agent model list --usage tool --format json
arkcli agent model list --usage tool --name <model-name-keyword> --format json
arkcli agent model list --usage tool --include-all --format json
arkcli agent model list --usage tool --refresh-cache --format json
```

| 参数/输出 | 含义 |
| --- | --- |
| `--usage agent` | 默认用途：Agent 主模型，按版本级 `ep_agent.support` 筛选，保持原有行为。 |
| `--usage tool` | 工具模型，按版本级 `epa_tool_model.support` 筛选，不要求同时支持 `ep_agent.support`。其他 usage 值不支持。 |
| `--include-all` | 跳过所选用途的支持性过滤，用于排障；结果不都是可选模型。 |
| `--primary-only` | 在完整缓存上本地只保留服务端标记为主版本的条目；主模型和工具模型查询都默认不加，只有用户明确要求主版本时使用，以免漏掉支持对应用途的非主版本。 |
| `--name` | 本地按 `name` 或带版本的 `model` 做不区分大小写的子串匹配。 |
| `--query` | 可选的详情增强/排序，仍以所选用途的 ArkModels 候选为主表，未命中搜索的候选仍保留，不截断数量。详情按名称关联，不作为具体版本的硬指标依据。 |
| `items[].tool_model_support` | 对应版本是否声明工具模型支持，与 `agent_support` 分开输出。metadata key 不区分大小写，Data 任一值去空白后为 true/1/yes/y（不区分大小写）表示支持。 |
| `items[].model` | 保留具体版本的模型标识；工具绑定使用此值填 `Tools[].Configs[].Models[].ID`，不是条目内部 `id`（如 epm-*）。 |
| `--refresh-cache` | 同步刷新 ArkModels 缓存，无需配合 query；失败冷却期内仍不重试。query 模式还会刷新原有搜索缓存。 |
| `cache` | CLI 输出 `source`（network/cache）、`fetched_at`、`expires_at`，用于判断本次候选来自请求还是缓存。 |

CLI 首次读取完整 ArkModels 数据，按产品、环境、账号/用户、profile、region、project 隔离落盘，TTL 为 5 分钟；主模型与工具模型、不同名称/主版本过滤共用同一份全版本缓存。有效期内普通查询不重复请求 ArkModels，过期同步刷新；跨进程锁合并并发刷新。缓存不可用时报告错误，不绕过缓存反复请求。请求失败保留旧数据并设置 1 分钟冷却，不用过期数据伪装当前可用结果；冷却结束后可重试。query 的详情搜索使用其原有独立缓存，不包含在上述候选缓存命中保证内。

工具支持性只能读取 ArkModels；不要用 `agent model config --key epa_tool_model.support`（ListManagedAgentModelMetaDatas）或按名称折叠后的主版本搜索数据代替。空候选只表示当前目录没有声明支持的候选，不代表所有工具都不支持 Models。候选不保证对应工具模态、账号权限、开通状态或实际推理成功，也不额外套用前端展示用的体验标签筛选。多个候选交由用户选择，不自行默认绑定；用户已经提供明确 ID 时可以直接交给创建/更新接口，不新增缓存驱动的写入拦截。

以下为工具配置片段，可用于 create/update 的 `--tool`，或完整 JSON/YAML 请求的 `Tools` 数组（占位符需替换为真实模型 ID）：

```yaml
Type: agent_toolset_multimodal_20260825
Configs:
  - Name: image_generate
    Enabled: true
    Models:
      - ID: <selected-image-model-id>
        Default: true
  - Name: video_generate
    Enabled: true
    Models:
      - ID: <selected-video-model-id>
```

`--tool` 仍是完整列表替换：更新前先 get，保留其他需要的工具及配置，再提交完整 `Tools`。create/update 的 Client Preview 只验证本地结构，不表示服务端接受该绑定。get/list/versions 回显和 `+new-agent --fork` 复制会保留工具模型字段；回读核对 `Configs[].Models`，不要只看主模型 `Model.Id`。

## 默认 Agent 工具

创建 Agent 时，如果用户没有显式传 `--tool`，且请求体里没有 `Tools`，CLI 调 CreateAgent 时自动传入默认 `Tools` 数组：

```yaml
Type: agent_toolset_20260701
Name: agent_toolset_20260701
Configs:
  - Name: bash
    Enabled: true
    PermissionPolicy: { Type: always_allow }
  - Name: read
    Enabled: true
    PermissionPolicy: { Type: always_allow }
  - Name: write
    Enabled: true
    PermissionPolicy: { Type: always_allow }
  - Name: edit
    Enabled: true
    PermissionPolicy: { Type: always_allow }
  - Name: glob
    Enabled: true
    PermissionPolicy: { Type: always_allow }
  - Name: grep
    Enabled: true
    PermissionPolicy: { Type: always_allow }
  - Name: web_fetch
    Enabled: true
    PermissionPolicy: { Type: always_allow }
  - Name: web_search
    Enabled: true
    PermissionPolicy: { Type: always_allow }
DefaultConfig:
  Enabled: true
  PermissionPolicy: { Type: always_allow }
```

- 用户明确传 `--tool '[]'` 表示关闭默认工具，必须尊重。
- 用户显式传任意 `--tool` 时，传入值就是完整 `Tools` 数组，CLI 不追加/合并默认工具。
- 需要 `advisor` 且保留默认工具时，必须显式传完整数组，例如 `--tool '[{Type: agent_toolset_20260701, Name: agent_toolset_20260701}, {Type: evolution, Configs: [{Name: advisor, Enabled: true}]}]'`。
- 不要重复手写默认工具，除非用户要调整权限或显式关闭。

### 修改默认工具权限

用户说“把 `<tool-name>` 策略设为总是询问 / 需要确认 / always ask”时，不要只传该工具的一项 config。由于 `--tool` 是完整数组替换，必须传完整默认 toolset，并只覆盖对应工具的 `PermissionPolicy.Type`。

权限策略值：

- 自动放行：`always_allow`
- 总是询问 / 需要用户确认：`always_ask`

例如“创建一个数据分析 agent，write 策略设为总是询问”：

```bash
arkcli agent agent create \
  --name arkcli-data-analysis-agent-20260706 \
  --model <items[].model from agent model list> \
  --system "You are a data analysis agent. Help users inspect datasets, reason about metrics, write analysis code, and summarize findings clearly." \
  --skill '{type: skill_hub, skill_id: skill-xxx, version: "1.0.0"}' \
  --tool '[{
    Type: agent_toolset_20260701,
    Name: agent_toolset_20260701,
    DefaultConfig: {Enabled: true, PermissionPolicy: {Type: always_allow}},
    Configs: [
      {Name: bash, Enabled: true, PermissionPolicy: {Type: always_allow}},
      {Name: read, Enabled: true, PermissionPolicy: {Type: always_allow}},
      {Name: write, Enabled: true, PermissionPolicy: {Type: always_ask}},
      {Name: edit, Enabled: true, PermissionPolicy: {Type: always_allow}},
      {Name: glob, Enabled: true, PermissionPolicy: {Type: always_allow}},
      {Name: grep, Enabled: true, PermissionPolicy: {Type: always_allow}},
      {Name: web_fetch, Enabled: true, PermissionPolicy: {Type: always_allow}},
      {Name: web_search, Enabled: true, PermissionPolicy: {Type: always_allow}}
    ]
  }]' \
  --format json
```

如果用户要求多个工具不同策略，同样在这一个完整 `agent_toolset_20260701` 的 `Configs` 里同时改，不要拆成多个 `agent_toolset`。

## 示例

```bash
# Agent + Skill
arkcli agent skill search "Excel 数据分析" --limit 10 --format json
arkcli agent agent create \
  --name arkcli-data-analysis-agent-20260706 \
  --model <items[].model from agent model list> \
  --system "You are a data analysis agent. Help users inspect datasets, reason about metrics, write analysis code, and summarize findings clearly." \
  --skill '{type: skill_hub, skill_id: skill-xxx, version: "1.0.0"}' \
  --format json

# Fork existing agent
arkcli +new-agent --fork agent-20260707063932-vbfjd --format json
```
