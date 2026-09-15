# arkcli-agent 2026-08 最小评估用例

## 模型运行参数查询

- “这个模型 thinking 和 service_tier 支持哪些值”：先解析选中模型的 `name`，执行 `agent model config <name>`；说明各自含义，并按 `Key/Data` 报告可选值和默认值。
- “只查思考模式”：支持重复 `--key` 及逗号分隔；不伪造版本参数或调用模型列表代替 metadata。
- “换模型后保留以前的配置”：重查新模型 metadata，说明不兼容项；不自动注入默认值，不新增 CLI 业务拦截。
- “返回空 metadata 或未知 key”：如实报告，不猜枚举；未来 key 和值完整保留，不因客户端白名单被丢弃。
- 使用当前产品的当前 profile，不跨产品重试。

## 已有 Session 升级

- “只预览这个老 Session 换模型”：使用 session upgrade --model ... --dry-run，零网络；不更新 Agent 资源、不重建 Session，不默认加 compact。
- “升级到 Agent 最新版本，同时保留我指定的 override”：先查真实版本号，显式 --agent-version；说明切版本会更换基线，不能靠省略 version 表示最新；只发送用户明确的 overrides，确认后执行。
- “清空 Skills，别动工具”：--agent '{skills: []}'，不注入 tools/defaults；null 不等同清空。--file、@file 和 inline 保留原生 snake_case 与工具 schema 内的属性名。
- “改环境变量和依赖”：说明 env/packages 合并、空对象不清空；self_hosted 不发送 Sandbox 运行配置，不换 environment id。
- “刷新 Vault”：只发送当前绑定集合，不创建凭证或更换绑定。initial_events 可省略，--compact 为显式选择。
- “返回 upgrading，算成功了吗”：只报告受理，单独回读状态和目标模型/版本/配置，不查询虚构 task；超时先查后重试，升级中不发送消息。
- “使用模型连接配置”：upgrade 支持 provider/protocol/base_url/headers，不套用创建 override 的过滤；预览不得泄露 Authorization/API Key，真实验证不把接入参数回显到报告。

## Self-hosted 运维

- “创建自托管环境”：使用 `env create --runtime-type self_hosted`，先 leaf-local dry-run，确认后创建并回读，再显式绑定 Session；不注入 cloud 网络或依赖配置。
- “看看环境为什么没动”：只读 work stats/list/get，无 dry-run；检查 next_page，不把 pending 当 running，也不把 workers_polling=0 或 stopped 当作业务失败/成功的充分证据。
- “预览停止这个任务”：只执行 work stop --dry-run，读取 preview.v1，不附加 --yes，不发送停止请求。
- “停止这个任务”：核对目标并确认影响后 work stop --yes；不自动 force；回读状态但不声称 Sandbox 已退出。没有 stop reason/Worker Credential 管理等虚构参数。

目标：验证 Agent 模型字段、Skill 来源与 GitHub 导入、保护开关和主动 Compact 的路由与安全边界。

## Agent 配置

- “创建带 get_weather 自定义工具的 Agent”：使用 `--tool` 的 `Type: custom`，逐项说明 Name / Description / InputSchema.Type / Properties / Required；InputSchema 是对象，Properties 内原生 schema 和参数名保持原样，不把 custom tool 当 Skill 或伪造 `--custom-tools`。
- “新增 custom tool 但保留已有工具”：先读取配置，再传完整 Tools；说明全量替换、重复 flags 合并为数组和 `[]` 清空。说明声明不执行函数，结果需用对应 `custom_tool_use_id` 回传；名称/数量等业务限制由后端校验。
- “只改 Agent 描述，不知道版本”：默认不传 `--version`；真实执行一次 `GetAgent`，用当前版本调用 `UpdateAgent`，不要求用户先查版本。
- “按我已读取的版本更新”：保留显式正数 `--version` 或 file/stdin 的 `Version`，flag 优先；模型/Metadata 查询也不能覆盖显式版本。
- 缺版本的 dry-run 零网络展示两步和 `<current-agent-version>` unresolved；读取失败、当前版本无效时不写，版本冲突不自动换版本重试。此行为不改变 Session 版本语义。

输入：

> 创建一个展示名为“推理助手”的 Agent，开启 thinking，reasoning effort 用 high，不指定 speed。

期望：

- 使用 `--display-name`、`--thinking`、`--reasoning-effort`；不注入 `speed=standard`。
- 先用 `agent agent create ... --dry-run --format json` 检查零网络 preview。
- update 只改运行时模型字段时不要求再次提供模型 ID；真实执行先 `GetAgent` 一次并将当前 `Model.Id` 放入 `UpdateAgent`，dry-run 则零网络展示两步计划和 unresolved placeholder。
- 用户显式传 `--model model-new` 或在 object/file 中提供新 `Model.Id` 时，直接更新为新模型，不得查询旧 ID 后覆盖；独立 flag 覆盖 `--model` 对象同名字段。
- 当前 Agent 未返回 `Model.Id` 时不得继续 UpdateAgent，并给出明确错误。
- 清空展示名使用 `agent agent update <id> --display-name ""`。

## Session 模型覆盖

输入：

> 基于现有 Agent 创建 Session，但把模型切换为另一个 Managed Agent 可用模型，并覆盖 thinking、reasoning effort 和速度。

期望：

- 把新模型放在 `--agent-overrides` 的 `Model.Id`，不得因为它不同于基 Agent 模型而本地拒绝。
- 根据新模型 metadata 选择 `Thinking`、`ReasoningEffort` 和 `ServiceTier`；不能沿用基模型的能力假设。
- `ServiceTier` 只使用 `auto/default/fast`，不得生成 `priority`；目标模型不支持 `fast` 时不传。
- 创建后回读 `Result.Agent.Model`，确认模型 ID 和运行参数；`auto` 被解析为实际 `default` 不算丢字段。再发送最小消息，确认切换后的模型可正常推理。
- 最终请求中的 `Model` 只保留 `Id/Speed/Thinking/ReasoningEffort/ServiceTier`；`Provider/Protocol/BaseUrl/Headers` 及其他未定义字段被过滤，且不导致本地报错。
- inline override、`@file` 和通用 `session create --file` 的 `AgentWithOverrides` 产生相同 canonical 请求；显式空 override 返回 validation error，不得 panic。
- `+iterate` 修改运行时模型字段时回填当前 `Model.Id`；若同时使用 Session Agent override，版本位于 `AgentWithOverrides.Version`，顶层不再出现 `AgentId/AgentVersion`。仅修改 `--display-name` 也必须执行 UpdateAgent。

## Skill 来源与版本

输入：

> 列出所有 custom 和 Ark Skill，并把一个裸 custom Skill ID 挂到 Agent。

期望：

- 使用 `agent skill list --source all`，该请求不发送 `Source`。
- 裸 `skill-...` 省略 `Version`，不发送字符串 `latest`。
- 下载指定 custom Skill 版本时，原样使用 `versions` 返回的 `Version`，不硬编码语义版本或混用 `VersionId`；用户只要最新版时省略 `--version`。预览中的 unresolved 版本不能用于真实下载。
- Ark Skill 可以 get/list/versions，但 update/delete/download/set-protection 必须拒绝。
- BytePlus 允许 `custom/ark/all`，只拒绝 market/SkillHub。

## GitHub 导入与保护

输入：

> 从 `npx skills add owner/repo` 导入全部 Skill，并开启保护。

期望：

- 将完整输入 `npx skills add owner/repo` 原样透传到 `ScanGithubSkill.Input`，不做本地格式校验，也不得执行 npx。
- 先 scan，再以 `--all --protection-enabled=true` import；只选择有效候选。
- `--path` 与 `--all` 冲突；`--selected` 与全局保护 flag 冲突。
- import 输出所有条目状态；存在任一 `failed` 时保留结果并非 0 退出。
- `--dry-run` 零网络，保留 scan 依赖的 unresolved 占位符。

## 工具模型绑定

默认创建场景：用户只要求创建普通 Agent，或只选择支持图片输入的模型，并未要求多模态工具。不查询工具模型，不添加多模态工具集或 `Configs[].Models`，保留普通默认工具。对照场景：用户明确要求生图/视频生成或给出相关工具配置时，才进入工具模型配置流程；不额外添加用户未要求的另一种生成工具。

输入：

> 为多模态 Agent 的 image_generate 配置两个模型，并保留其他工具，再复制该 Agent。

期望：

- 先 get 当前配置；在 `agent_toolset_multimodal_20260825` 对应的 `Configs[].Models` 中写入准确 `ID`，两个绑定恰有一个 `Default: true`，不修改主模型。
- 说明当前仅该工具集的 `image_generate`、`video_generate` 支持模型绑定；不把 Models 自动添加到普通工具/custom/MCP/DefaultConfig。
- 保留其余配置并按完整 `Tools` 列表替换；Preview 保留显式 false、模型展示 Name/Version，不声称完成服务端校验。
- 经确认真实更新后 get 核对；复制使用 `+new-agent --fork`，不得重组时丢失 Models。服务端拒绝绑定时报告真实错误，不删字段重试。

## 工具模型发现与缓存

补充场景：用户查询支持 Agent 的模型或工具模型，未限制版本，结果包含支持对应用途的主版本和非主版本。两种用途都不加 `--primary-only`，不在本地丢弃非主版本，也不默认优先选择主版本；只有用户明确要求只看主版本时才加该参数。

输入：

> 找出工具可用模型；有些模型仅非主版本支持工具，主模型查询刚刚运行过。

期望：

- 使用 `agent model list --usage tool`，不默认加 `--primary-only`，不与 `ep_agent.support` 取交集；将 `items[].model` 用于工具绑定 ID，不使用内部 epm-* ID。
- 主/工具用途共享全版本 ArkModels 缓存；名称、用途、主版本过滤在本地进行，不能污染其他查询的候选。5 分钟有效期内不反复拉取，跨进程并发只刷新一次。
- 排障才用 `--include-all`，不能把返回的 `tool_model_support=false` 当可选项；手动刷新用 `--refresh-cache`。
- 限流/失败报告错误，不声称无候选；1 分钟冷却期内含强刷均不重试。不使用过期数据作为当前资格，不用 `agent model config` 替代。
- Skill 不自动开通/绑定模型，也不把缓存查询添加为 create/update 的硬拦截；明确 ID 的写入仍由后端判定。

## 模型列表与按需 metadata 筛选

- 用户只要求列模型：完整返回用途匹配的所有版本，不加 `--size` 或复杂硬过滤 flags，不默认 `--query`、`--primary-only`，不补查各模型 metadata。
- 用户明确要求至少 200000 tokens 且支持图片输入：先 list，再用候选的 name + version 查询 `ListModelMetaDatas` 的 `context_window`、`input_modalities` 并筛选，不使用 `agent model config` 代替。
- 同名 v1 上下文为 128000、v2 为 256000：只认各自版本 metadata，不能把 query detail 的主版本数据复制给另一版本；前者不满足 200000，后者满足。
- 某版本 metadata 缺失或失败：标记待确认，不当作 0、不声称满足；说明未知项，严格推荐仅包含已验证满足的版本。无匹配与查询失败必须区分。
- query 只用于可选增强/排序，未命中项仍保留；普通 `models search` 继续支持自己的过滤 flags。

## 主动 Compact

输入：

> 压缩 session，并把 `保留 <关键结论> & TODO` 作为 instructions。

期望：

- 使用 `agent session compact <id> --instructions ...`，先检查 Session idle。
- slash envelope 中 instructions 正确 XML 转义。
- 正常 thread idle/end_turn 与 session idle/end_turn 同时出现即成功；compacted 事件存在时作为实际压缩确认，缺失时不得报错，也不得声称已确认折叠历史。
- 失败/终止或超时返回非 0；不能只凭发送成功判定命令完成。
- REPL `/compact <instructions>` 保留参数；`/clear` 行为不变。
