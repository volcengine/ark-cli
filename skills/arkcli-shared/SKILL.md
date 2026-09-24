---
name: arkcli-shared
version: 2.2.3
description: "arkcli 共享执行协议：首次配置入口、业务命令执行前的认证闸门、命令路由与选择顺序、输出/安全/二次确认规则。深度细节（身份解析、AK-SK 边界、API Key 恢复、实名闸门、profile 默认与漂移、临时数据面执行上下文、版本检查与显式升级、全局 flags、故障分流）按需在 references/ 加载。当用户第一次使用 arkcli、遇到未登录/鉴权失败、询问版本是否最新或要求升级 arkcli、需要判断该走产品命令还是 raw api、或任何 arkcli-* skill 需要公共上下文时触发。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli --help"
---

# arkcli 共享规则

本 skill 是 `arkcli` 的统一执行协议**入口**。所有 `arkcli-*` skill 在执行前都应先读取本文件。

> **分层约定**：本文正文只放"几乎每个任务都命中"的规则。稀有路径 / 查表类细节下沉到 reference，命中对应场景时再读：
>
> | 场景 | 读哪个 |
> |---|---|
> | "我的 / my xxx" 自指资源过滤 | [`../arkcli-auth/references/identity-resolution.md`](../arkcli-auth/references/identity-resolution.md) |
> | AK/SK 态能调什么 / 数据面 API Key 报错恢复 | [`../arkcli-auth/references/auth-modes.md`](../arkcli-auth/references/auth-modes.md) |
> | 开通 / 部署 / 精调前的实名检查 | [`../arkcli-auth/references/realname-gate.md`](../arkcli-auth/references/realname-gate.md) |
> | `+chat`/`+gen`/`+deploy` 的默认资源与漂移 nudge | [`references/profile-defaults.md`](references/profile-defaults.md) |
> | 临时传入 Profile / API Key / Base URL / Endpoint / 模型名 | [`references/execution-context.md`](references/execution-context.md) |
> | 版本是否最新 / 刷新版本信息 / 显式升级 arkcli | [`references/update.md`](references/update.md) |
> | 全局 flags 速查 | [`references/global-flags.md`](references/global-flags.md) |
> | 报错不知归类 / 故障分流 | [`references/troubleshooting.md`](references/troubleshooting.md) |

## 配置与首次使用

- 首次使用或怀疑配置归因（`--profile` / `ARK_PROFILE` / 全局 flags / `.env` 谁覆盖谁）不对时，先看 [`../arkcli-config/SKILL.md`](../arkcli-config/SKILL.md)
- profile 类写操作（create / use / set-default / keys / models / delete / rename）走 [`../arkcli-profile/SKILL.md`](../arkcli-profile/SKILL.md)
- 看可用资源（endpoint / plan 模型 ID）走 [`../arkcli-resources/SKILL.md`](../arkcli-resources/SKILL.md)
- 首次远端调用前，先看 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
- `arkcli` 是产品 CLI，不是 OpenAPI Action 浏览器；不要从 Action 名反推命令设计
- 安装: `npm i @volcengine/ark-cli -g`（公开版）
- 版本检查与显式升级按需读取 [`references/update.md`](references/update.md)；`update.mode` 的读取、写入与 automatic gate 仍走 [`../arkcli-config/SKILL.md`](../arkcli-config/SKILL.md)

## 调用归因协议

当 AI Agent 通过任意 `arkcli-*` 业务 skill 执行 `arkcli` 命令时，必须使用下面的单命令环境变量前缀。它既标记调用来源和 owning skill，也冻结本次 Skill workflow 的 CLI 版本：不做隐式 registry 更新检查、不打印更新提示、不在任一子命令结束后调度 automatic apply。人类直接在终端调用 `arkcli` 时不要手动补这些变量。

```bash
ARKCLI_NO_UPDATE_NOTIFIER=1 \
ARKCLI_CALLER_TYPE=ai_agent \
ARKCLI_CALLER_NAME=<agent-id> \
ARKCLI_SKILL_NAME=<current-arkcli-skill> \
arkcli <command> ...
```

约定：

- `ARKCLI_CALLER_NAME` 使用稳定 Agent ID，例如 `codex` / `claude-code` / `opencode` / `openclaw` / `trae` / `cursor`；无法可靠判断时用 `unknown_agent`
- `ARKCLI_SKILL_NAME` 填当前业务 skill 名，例如 `arkcli-gen` / `arkcli-chat` / `arkcli-models` / `arkcli-deploy`
- 一般业务 workflow 不要把 `arkcli-shared` 填进 `ARKCLI_SKILL_NAME`，归因必须落到实际 owning skill；仅当本 skill 直接处理版本检查或显式升级时，它就是 owning skill，使用 `ARKCLI_SKILL_NAME=arkcli-shared`
- `ARKCLI_NO_UPDATE_NOTIFIER=1` 的兼容变量名虽然只写了 notifier，但契约覆盖全部**隐式**更新活动：缓存读取/刷新、提示和 automatic 调度；它不阻止用户明确要求的 `arkcli update` / `arkcli update --check`
- Skill 内每一条 `arkcli` 子命令都必须带完整前缀，不能只给第一条加；这样多命令 workflow 从开始到结束都使用同一已安装版本
- 只给当前命令加前缀，不要 `export` 到整个 shell 会话，避免串到后续无关命令

## 统一 CLI 与 Profile

当前 `arkcli` 的产品身份在编译时固定；Profile 只选择该产品内的身份与消耗切面，不能切换产品：

```bash
arkcli profile create --type platform --set-default          # 新建火山 profile（旧 config init/switch 已 deprecated）
arkcli profile use <name>                                    # 切换默认 profile
```


切换 profile 会联动切换登录身份、API Key、控制面路由等上下文。详细命令树看 [`../arkcli-profile/SKILL.md`](../arkcli-profile/SKILL.md)。不能从 Profile 的名称或 `tenant` 字段推断、切换当前编译产品。

只查当前身份/Profile 时，普通 CLI 使用 `arkcli auth status` / `arkcli auth whoami`；默认模型与路由使用 `arkcli resources list --modality <text|image|video>` 并按 Resources Skill 验证。`profile show/list/keys list` 可能同步远端 Key 并回写本地库存或默认 Key，不能作为常规 Chat/Gen 准入或“不改配置/Key”请求的无副作用查询。显式 Profile 管理任务保留这些命令，但先说明同步影响；不改用 deprecated `config show/list` 绕过限制。此边界同样约束业务 Skill/reference 的旧建议。

## 命令路由与执行顺序

优先按**用户目标**判断，而不是按命令名思考：

| 用户目标 | 路径 | 关键点 |
|---|---|---|
| 试用模型 / 快速验证效果 | `auth` → `resources` / `models`（可选）→ [`+chat`](../arkcli-chat/SKILL.md) / [`+gen`](../arkcli-gen/SKILL.md) | Plan lane 用模型名；Platform lane 用 Endpoint |
| 专项多模态理解（转写/抽取/字幕/框目标…） | `auth` → [`+understand`](../arkcli-understand/SKILL.md) | 有明确产出形态时走 understand，不是 chat |
| 语音模型发现 / 选型（TTS / ASR / 播客 / 音色 / 实时语音交互） | `auth` → [`models search`](../arkcli-models/SKILL.md) | **仅支持广场检索**；不支持 `+chat` / `+gen` / `+deploy` / `+code-example` / `usage` / `pricing` / `onboard` |
| 正式接入 / 稳定调用 | `auth` → `models` → [`infer endpoint list`](../arkcli-infer-endpoint/SKILL.md) → 没有就 [`+deploy`](../arkcli-deploy/SKILL.md) | 核心资源是 Endpoint，不是 +chat/+gen |
| 排查存量调用 / 看消耗 | `auth` → [`usage`](../arkcli-usage/SKILL.md) | — |
| 本地 AI Agent 集成 | [`+connect`](../arkcli-connect/SKILL.md) | — |
| 明确查官方文档 / 读取文档 URL / 业务 Skill 缺少产品知识 | [`docs`](../arkcli-docs/SKILL.md) | 搜索后读取正文；不替代资源操作、鉴权、诊断和 CLI help |

**易混动词路由**（避免选错 skill）：

- **列 / 绑 / 分 / 轮换席位 APIKey** → [`arkcli-plans`](../arkcli-plans/SKILL.md)；**用 / 消耗 / 还剩多少额度** → [`arkcli-usage`](../arkcli-usage/SKILL.md)
- profile **写操作**（create/use/set-default/keys）→ [`arkcli-profile`](../arkcli-profile/SKILL.md)；配置**排障 / 老 yaml** → [`arkcli-config`](../arkcli-config/SKILL.md)
- 开放式带图**对话** → `+chat`；**生成**图/视频 → `+gen`；有产出形态的**理解** → `+understand`
- **语音合成 / TTS / 配音 / 朗读**、或用户点名 `doubao-seed-tts-*` 等广场语音模型 → 只允许 `models search` 做发现与选型说明；不要转 `+chat` / `+gen` / `+deploy` / `+code-example` / `usage` / `pricing`。不要主动补充控制台 / OpenAPI / SDK 等非 arkcli 接入路径或链接，除非用户另问"官方文档在哪里"。如果用户只是要"把一个音频文件转文字"且未要求使用广场语音模型，可另走 `+understand asr`；不要把 `+understand` 解释成支持广场 ASR 模型。

**命令选择顺序**（始终按此）：

1. 先用产品命令 `arkcli <domain> <verb>` 或 `arkcli +<workflow>`
2. 有对应 reference 文档先读 reference 再执行
3. 产品命令确实不覆盖，最后才走 `arkcli api`（**不要**把 Raw API Explorer 当默认入口）
4. 读操作优先直接执行；写 / 删除 / 切换默认配置前必须确认用户意图

## 认证闸门

除 `arkcli auth ...`、`arkcli profile list/show`、`arkcli +connect list`、`arkcli update ...` 外，默认认为业务命令**需要先过认证检查**。不要跳过认证检查就连续重试一串业务命令。

1. 先运行 `arkcli auth status`；已登录就继续目标命令
2. 未登录 / 凭证失效：保持当前编译产品，按其认证 Skill 选择登录命令，不靠 Profile 推断或切换产品
   - **火山**：直接 Bash 执行 `arkcli auth login volc-sso`
   - 执行前一句话告知用户"检测到未登录，正在启动 SSO 登录，请在浏览器完成授权"；Bash 调用设 `timeout=600000`（10 分钟）；成功后立即回到原始任务，不要停在 auth 结果
   - SSO 同时覆盖控制面 BFF 和数据面，所以默认走 SSO；**AK/SK 登录通道 0.1.16 暂关**。CI / agent / 沙箱（非 TTY）走 `arkcli auth login --no-browser` **两段式**：Phase 1 跑它拿 `authorize_pending` JSON 里的 `authorize_url` 转发给用户，待其浏览器授权后回粘 base64 授权码，再跑 `arkcli auth login --no-browser --code <授权码>` 完成（细节见 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)）。**不要**在非 TTY 直接指望它阻塞等粘贴（旧版会 `EOF` 崩）
   - 启动 SSO 失败（无浏览器 / `open` 失败 / 端口占用 / 超时）→ 不原地重试，把 stderr 原样贴回用户、请其手动在终端登录后回来
   - 登录流程细节、whoami、apikey 选择见 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)

**追加闸门 / 错误恢复**（命中才读，避免常驻）：

- 命中**开通 / 部署 / 精调 / 激活模型**（`+deploy` / `infer endpoint create` / `train finetune create` / `models activate`）→ **第一步**先做实名检查，见 [`../arkcli-auth/references/realname-gate.md`](../arkcli-auth/references/realname-gate.md)
- 当前是 **AK/SK 态**、要判断能调什么，或**数据面 API Key** 报 401/403/AccessDenied 要恢复 → [`../arkcli-auth/references/auth-modes.md`](../arkcli-auth/references/auth-modes.md)
- 用户说"**我的 / my** xxx"要按身份过滤资源 → [`../arkcli-auth/references/identity-resolution.md`](../arkcli-auth/references/identity-resolution.md)

## 输出规则

- 全局 `--format` 支持 `json`、`yaml`、`table`、`csv`、`jsonl`、`pretty`；脚本场景优先用 `json`/`yaml`，需要抽字段时配合 `--transform`
- `stdout` 只放结构化结果；解释 / 调试 / 错误都走 `stderr`
- 消费结构化输出时，禁止先用 `head` / `tail` / `sed -n` / `head -c` 按行或字节截断再解析。这样既可能破坏 JSON，也会漏掉数组后部的目标项、default 标记或 warning 对应的数据。输出较大时，优先使用命令自己的 `--output <file>`；没有该 flag 时只把完整 stdout 重定向到文件，再用 `jq` / `yq` 提取所需字段。
- 首次调用和决定结论的调用不得用 `2>/dev/null` 吞掉 stderr；scope 提示、软截断告警和可恢复错误可能只出现在 stderr。

## 安全规则

- 禁止输出完整 AK/SK、token、secret
- 写入、删除、切换配置前需要确认用户意图
- 只有已注册 Action 才能通过 `arkcli api` 调用
- 涉及创建 Endpoint、修改 profile、清理凭证等操作时，先看叶子命令 `--help`：支持 Client Preview 才能用 `--dry-run`；不支持时改用只读命令核对并取得明确确认，绝不生成不存在的 flag

## 意图澄清与结构化选择

意图澄清只解决“用户要操作哪个真实目标”，不代替写操作授权。用户目标已经唯一，或可从
本轮输入与当前权威结果安全确定时，不要额外提问。确有歧义时遵守以下硬规则：

1. 先做一次最小、有界的只读查询；候选只能来自本轮完整结构化结果，不能从模型记忆、示例或截断输出补全。
2. 只允许用用户明确给出的硬约束和产品权威 eligibility 字段过滤候选；相关度排序、展示顺序、推荐语和 Agent 自己的优劣判断都不是硬约束，不得用主观“最合适”把 N 个候选自行收敛成 1 个。
3. 按 0 / 1 / N 收敛：0 个时补充一个最关键条件；1 个时复述精确 ID 后继续；N 个且选择会改变远端结果时必须询问用户，并在用户选定前停止，不得进入下游 preview、create、update 或 delete。
4. N 个候选优先使用当前宿主提供的结构化选择能力，选项直接携带区分目标所需的真实 ID 与关键字段；可以标注推荐及依据，但推荐不能代替用户选择。通用 Skill 不写死任何宿主工具名，也不重复添加宿主自动提供的自由输入项。
5. 宿主没有结构化选择能力时，退化为精简编号列表并要求用户回复精确 ID。用户选定后只沿原 workflow 继续，不重新查询同一批候选。

## 二次确认错误处理（human-in-the-loop）

高危操作（删除资源、变更凭证、产生费用等）需要二次确认。CLI 自动检测环境：交互式终端显示 Y/N 提示；非交互式（Agent 调用）返回 `ExitValidation` 错误、`type="requires_confirmation"`。

当 arkcli 返回 `ExitValidation` 且 `type="requires_confirmation"` 时：

1. **不要直接报错给用户** —— 这是正常的二次确认流程
2. **调用宿主提供的用户确认能力** —— 提示内容可从 CLI 错误的 `hint` 字段提取，或通用提示"即将执行高危操作，确认继续吗？"
3. 用户确认 → 给原命令加 `--yes` flag 重试
4. 用户取消 → 返回"操作已取消"

当前会触发二次确认的命令：`plans personal rotate-apikey`、`plans team rotate-apikey`、`models activate <model-name>`、`profile delete <profile-name>`、`profile project [<project-name>]`、`config delete <profile-name>`。后续新增高危命令遵循同一约定。

## Agent 禁止行为

- 不要把 `arkcli api` 当默认入口
- 不要在未检查认证状态前连续重试业务命令
- 不要把中间步骤当最终结果；登录、查模型、切 profile 完成后应回到用户原始任务
- 不要在业务 skill 里重复共享规则；共享规则统一以本 skill（及其 references）为准
- 不要一概声称“试用不需要 Endpoint”：Plan lane 使用模型名，Platform lane 必须使用 Endpoint。按 [`references/execution-context.md`](references/execution-context.md) 判断；正式接入仍走 `+deploy`
- 不要把广场可搜到的语音模型误写成 arkcli 已支持调用、部署、示例、用量或费用查询；语音模型在 arkcli 当前只承认 `models search` 发现能力。
- 不要把语音模型边界回答扩展成"去控制台开通 / 用 OpenAPI / 用 SDK 接入"的替代方案；当前 skill 只负责说明 arkcli 支持边界。
- 不要给业务命令增加临时 Project/Region override；根命令已删除 `--project-name` 与 `--region`。持久上下文通过 `arkcli profile create/use` 管理，单次切换使用 `--profile`。

## 参考

- [arkcli-auth](../arkcli-auth/SKILL.md) — 认证状态检查、登录与退出登录；其 `references/` 放身份解析 / AK-SK 边界 / API Key 恢复 / 实名闸门
- [arkcli-config](../arkcli-config/SKILL.md) — profile、base-url、region 配置排障
- [arkcli-api-explorer](../arkcli-api-explorer/SKILL.md) — 产品命令未覆盖时的 raw API 兜底入口
- [references/profile-defaults.md](references/profile-defaults.md) — profile 默认资源、漂移检测、跨模态
- [references/execution-context.md](references/execution-context.md) — 五类 Profile × 三种模态、临时 Key/Base URL/Endpoint 组合与 Client Preview 边界
- [references/global-flags.md](references/global-flags.md) — 常用全局 flags 速查
- [references/troubleshooting.md](references/troubleshooting.md) — 故障分流与能力边界
