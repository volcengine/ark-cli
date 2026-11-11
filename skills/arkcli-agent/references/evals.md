# arkcli-agent 2026-08 最小评估用例

## Todo 配置决策

- 用户只要求给已有 Agent 添加 Todo：先 get，合并到多模态 toolset；新添该 toolset 时 DefaultConfig.Enabled=false，只启用 todo_write，不查询工具模型、不加 Models、不丢失基础工具。
- 已有多模态 toolset 启用了生图并设置 always_ask：添加/禁用 Todo 不改变已有 DefaultConfig、图片模型及权限，不追加重复 toolset。
- Todo 被错误放入基础 toolset：指出归属错误，不把拒绝归因为工具名应改成 TodoWrite；写入名始终为 todo_write。
- 普通 Agent 遭白名单拒绝：报告前提未满足并停止，不切账号/产品或 Raw API 重试。仅配置回读成功但没有工具事件时，不声称实际调用通过。

## 消息执行失败退出码

- `+new session --message` 已创建 Session，但模型重试耗尽并退出 5：报告执行失败，保留 Session ID/cursor，不因创建成功而报告任务成功，不自动重发。
- 用户提供 MA 数据面地址和临时 Key 挂载文件：使用 `session resources add`，随后用 `resources list/get` 核对；不误用 BYOK 模型地址，也不要求修改全局 Profile。

- 假 BYOK 凭证：`events send --stream/--wait/--poll` 收到模型失败与 `retries_exhausted`，应识别退出码 5 和 `agent_execution_failed`；说明消息已受理、执行失败，不自动重发。
- SSE 超时转轮询：先收到 exhausted/terminal 或无 retry_status 的 `session.error`，补偿阶段才收到 idle，仍报告执行失败；retrying 后正常 end_turn 则不误报失败。
- 无历史上下文恢复：用户仅提供 Session ID 和发送 event ID；首轮列表只有旧 idle，后续分页才出现发送 ID、回复、thread idle、usage、session idle。先定位精确边界再处理，不排序、不重复输出、不在 thread idle 提前完成，不重发 send。
- after 被忽略、边界未落库或分页截断：不能宣称任务完成；有界补查仍无边界时报告未确认。没有发送 ID 且多条同文 user event 时先消歧。
- 工具报错后正常 `end_turn`：不因单次 `is_error=true` 把整轮判失败；`requires_action` 则报告待确认，不宣称已完成。
- 不带等待参数：退出码 0 只代表受理，不宣称模型已成功回复。

## 模型运行参数查询

- “这个模型 thinking 和 service_tier 支持哪些值”：先解析选中模型的 `name`，执行 `agent model config <name>`；说明各自含义，并按 `Key/Data` 报告可选值和默认值。
- “只查思考模式”：支持重复 `--key` 及逗号分隔；不伪造版本参数或调用模型列表代替 metadata。
- “换模型后保留以前的配置”：重查新模型 metadata，说明不兼容项；不自动注入默认值，不新增 CLI 业务拦截。
- “返回空 metadata 或未知 key”：如实报告，不猜枚举；未来 key 和值完整保留，不因客户端白名单被丢弃。
- 使用当前产品内已确认的 Platform profile，不跨产品重试。

## Profile 类型边界

- 默认是 Agent Plan / Coding Plan（含团队版），用户要操作 MA：识别套餐与 MA 的区别，对本次命令指定已确认的 Platform profile；不自动改全局默认，不猜 profile 名或跨账号。
- 用户只提供 Key，当前为套餐 Profile 或无 Profile：数据面直接使用显式 Key 和当前产品/Region/环境的 MA 标准地址，不要求额外提供地址，不继承套餐路由，不修改保存的 Profile。
- 用户提供 stg Key 和地址，当前为套餐 Profile：数据面保留显式地址，不因套餐类型拒绝；控制面仍需要有效登录。仅提供地址而没有显式 Key 时拒绝；错误 Key 如实报告服务端认证失败，不切换凭证重试。
- 套餐 profile 下 Preview 成功：只报告离线计划，不声称真实执行资格通过。真实执行遇到 `managed_agent_profile_required` 时停止，不重复发送。

## Session 创建项目一致性

- account-wide 查询返回 Agent 属于项目 A、最新 Environment 属于项目 B：不能直接创建；查询 A 的环境并显式绑定，先确定新 Session 的项目上下文。
- 环境或其他项目级关联资源与 Session 项目不同，或归属未知：补查/确认，不因同账号可见、override 或 preview 成功而继续；不擅自改默认 profile。
- Agent、Environment 和其他项目级关联资源均已确认属于目标项目：使用该项目 Platform profile 创建，回读 Session 项目和绑定。覆盖 `session create`、`+new session`、`+iterate`。

## 已有 Session 升级

- “只预览这个老 Session 换模型”：使用 session upgrade --model ... --dry-run，零网络；不更新 Agent 资源、不重建 Session，不默认加 compact。
- “升级到 Agent 最新版本，同时保留我指定的 override”：先查真实版本号，显式 --agent-version；说明切版本会更换基线，不能靠省略 version 表示最新；只发送用户明确的 overrides，确认后执行。
- “清空 Skills，别动工具”：--agent '{skills: []}'，不注入 tools/defaults；null 不等同清空。--file、@file 和 inline 保留原生 snake_case 与工具 schema 内的属性名。
- “改环境变量和依赖”：说明 env/packages 合并、空对象不清空；self_hosted 不发送 Sandbox 运行配置，不换 environment id。
- “刷新 Vault”：发送当前绑定集合；“更换/解绑 Vault”：先确认目标全集及差异，再发送目标数组或 `[]`，不自动创建凭证；回读验证，不能仅凭受理成功声称已生效。initial_events 可省略，--compact 为显式选择。
- “返回 upgrading，算成功了吗”：只报告受理，单独回读状态和目标模型/版本/配置，不查询虚构 task；超时先查后重试，升级中不发送消息。
- BYOK：Agent create/update、Session override/upgrade 均保留显式模型接入字段；请求凭证只在 events send、compact、+new session 使用 `--model-header-env`，不写入静态 Headers。Preview 只读取引用。

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
- BYOK：Agent create/update、Session override/upgrade 均保留显式模型接入字段；请求凭证只在 events send、compact、+new session 使用 `--model-header-env`，不写入静态 Headers。Preview 只读取引用。
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
# Toolset / Dream 回归补充

- 火山新建未传工具：普通 20260916 的八个工具，不附带多模态或提示词增强；显式 20260701 和 [] 保持原意。
- 用户只改旧 Agent 名称、复制、Session override/upgrade：不升级 Tools；用户明确要求升级才在完整数组里只改普通工具集 Type/Name，保留权限、Models、custom/MCP。
- Dream：先选模型与输入；create 仅一次，输出 ID 后 get/wait；不把 pending 当成功完成，不把 null Usage 当零费用。
- wait failed/canceled/未知状态非零退出，超时不取消；cancel 需确认且保留输出。分页读取 NextPage，page-limit 不是单页大小。
- create/cancel Preview 不发请求；Dream 仅火山，BytePlus 不推荐该命令。真实端到端未执行时不得以本地测试替代线上验收。

## Dream 无历史上下文用例

- 用户给出候选包版本并报告 404：记录实际 CLI 版本、加载的 Skill 来源和失败命令；不能用本机旧 Skill 推断用户也加载了旧版，不能将 unknown command 当作 HTTP 404。
- 默认套餐 Profile，用户提供可用 API Key、未指定地址：允许显式 Key，使用 MA 标准地址；不要求切换全局默认 Profile，不复用套餐或 BYOK 上游地址。
- 用户要求线上执行，但进程遗留 stg 或自定义 base URL：先核对环境与地址来源；不带着旧覆盖继续创建，也不擅自覆盖用户明确配置。
- list 成功但 create 返回 404：区分控制面与数据面，保留错误码和 RequestId；不直接认定服务未发布，不循环试写或猜 raw API。
- 已选 agent model list 返回的模型：Dream 直接使用模型 ID，不额外创建 Endpoint，不把 BYOK Model 对象传入 Dream。
- create 超时无 ID，或 wait completed 但只读到输出文件列表：前者先核对已有任务、不重复创建；后者只报告完成与文件元数据，未读内容不得声称内容验证通过。
## Vision 类型与配置保真

- Vision 查询返回 `model_access.status=requires_whitelist`：说明是当前账号的模型/版本可见性问题，展示该项申请链接与模型 ID，不建议 agent activate/models activate 替代开白，不自动提交申请或替换模型。
- Vision 查询返回 `unknown` 或旧版本缺少 model_access：不把未知当作未开白或已开白，区分查询失败与正常空结果。visible 也不声称已完成模型计费开通或推理验证。

- “创建一个会生图的 Agent”：不自动选 ultra_vision_agent_preview，先区分普通工具能力和明确 Vision 类型。
- 用户明确要求 Ultra、但未提供主模型：读 vision-agent.md，执行 `agent model list --usage vision`；按返回的完整 ID 和 default_model 选择，不用普通 ArkModels 候选、不猜模型、不遍历源码。stg 查询不得回退 prod；失败与空目录分别说明。
- 已确认 Vision 模型和生图模型：构造完整两组工具、high 参数、当前必需工具；只预览时零 API 调用，真实执行仍需本轮确认。
- 复制 Vision / 旧 multimodal 类型：BaseAgent 和 Models 均保留；显式基座覆盖优先，普通 Agent 未请求多模态仍保留原默认工具。
- 用户仅要求预览复制：`+new-agent --help` 没有 `--dry-run`，只读取源 Agent 并说明拟复制配置，不真实创建、不伪造 preview。
- 创建返回生图工具默认模型未配置：查询真实工具候选并核对输出模态，确认后补齐 image_generate.Models；不更换主模型、不重放原请求、不执行 MA 开通。成功后回读完整绑定。
- get/list/versions：JSON/YAML 回显 BaseAgent；不能只凭 table 摘要判断 Models 丢失。
- FunctionNotAllowed：区分 Vision 开白与 MA 开通，不切账号、不降级类型、不重复创建。
- 不将火山 Vision 模板套用到其他产品。未执行线上生图不得报告生图端到端通过。

## 独立开通 MA

- 用户只要求开通 MA：选择 `agent activate`，不创建 Agent 或 Session，不查询或开通模型。
- 非交互且尚未授权：不自行设置 `ARKCLI_ALLOW_HEADLESS_ACTIVATION`；解释账号级计费能力开通并等待确认，不能用 `--yes` 绕过。
- 已有明确授权：使用独立开通命令；根据 `already_activated` 区分本次开通与原已开通，不声称创建了资源。
- 要求只预览：说明此命令不支持 `--dry-run`，不发起开通；未知状态不当作已开通。
