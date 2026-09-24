---
name: arkcli-profile
version: 1.1.4
description: "arkcli profile 切面管理：列出、查看、新建、切换、删除、重命名 profile；管理 profile 内 API Key 列表；管理五类 profile 的默认资源与持久身份切面。也负责判断 Token 额度包/资源包应继续使用 platform profile 与 `/api/v3`，不能因「套餐」或价格字样误判为 Agent Plan/Coding Plan。临时 API Key/Base URL/Endpoint 调用不写回 profile，按 arkcli-shared 的 execution-context 契约执行。旧 config 子命令已 deprecated。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli profile --help"
---

# arkcli profile

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、配置排查与命令选择顺序**

**CRITICAL — 一旦确定走 `profile create`、`profile delete` 或 `profile project`（重选 project 会重命名/重派生 platform profile），必须先复述对 `config.yaml` 的影响并征得用户确认；其他写操作（`use` / `set-default` / `keys use` / `keys refresh` / `models refresh` / `rename`）执行前也要复述目标 profile 名。**
**CRITICAL — `profile` 是本地身份切面管理，全域不注册 `--dry-run`。`profile show/list/keys list` 会做 best-effort 在线 Key 同步，可能回写本地 Key 库存或默认 Key；只在用户显式要求 Profile 管理时使用，并先说明该影响。其他写操作仍通过明确确认保护。**

## 先区分核对与管理

- 只读核对当前身份、默认资源是否可见或凭证是否兼容，不是 Profile 管理；即使用户提供了 Profile 名称或 type，也先读取 Auth 与 Resources Skill 并交由它们处理。信息缺失不授权进入 `profile show/list/keys list`。
- 用户明确要管理 Profile、准备运行 `show/list/keys list` 时，先在**调用工具之前的用户可见文本**中说明两项具体影响：尝试在线同步远端 API Key；可能回写本地 Key 库存或默认 Key。例如：“该管理查询会尝试在线同步 API Key，可能回写本地 Key 库存或默认 Key。”
- “我会先解释”只是计划，不是披露；工具描述、内部思考和执行后的最终答复也不能替代这条前置说明。执行后不得把这些管理查询称为“纯只读”或据此保证“本地 Key 未变化”。

## 使用原则

- profile 是 0.1.16 引入的 **统一身份切面**，把 `(type × region × project × owner_trn × api_keys)` 五个属性绑成一组
- profile 写操作（create / use / set-default / keys / models / delete / rename）一律走 `arkcli profile <verb>`；旧的 `arkcli config init/list/show/switch/delete` 已 deprecated，不要再引导用户用
- 普通身份准入使用 `arkcli auth status` / `arkcli auth whoami`，默认资源与路由检查使用 `arkcli resources list`；不要为业务准入自动进入 Profile 管理
- 用户明确要查看或管理 Profile 时才运行 `profile show/list/keys list`，执行前先说明它们可能在线同步并回写 Key；脱敏 stdout 不代表本地配置无副作用
- ProfileType 五种：`platform` / `agent-plan` / `agent-plan-team` / `coding-plan` / `coding-plan-team`；其 text/image/video 所需的数据面、凭证与资源不同
- 用户只想临时传 API Key / Base URL / Endpoint 时，不要创建、切换或修改 profile；先读 [`../arkcli-shared/references/execution-context.md`](../arkcli-shared/references/execution-context.md)

## 适用场景

- 用户问"我有哪些 profile / 当前 profile 是哪个 / 切换 profile"
- 用户要创建新 profile（platform / agent-plan / coding-plan）
- 用户问 default 模型 / 资源是什么、想换 default
- 用户的 API Key 列表过期、新 key 还没被拉进来，需要 refresh
- 业务命令 `+chat / +gen / resources list` 报错"profile xxx 缺 ..."，转回这里排查

## 反唤起信号

- 用户问鉴权 / 401 / SSO 失败 → 转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
- 用户问 base URL / region / API Key 优先级被覆盖 → 转 [`../arkcli-config/SKILL.md`](../arkcli-config/SKILL.md)（config 现在专门讲解析归因 + reset）
- 用户问"有哪些 endpoint / 模型可用" → 转 [`../arkcli-resources/SKILL.md`](../arkcli-resources/SKILL.md)
- 用户问"我要找一个模型" / "哪个模型最强" → 转 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md)

## 切换 profile 的硬状态机

1. 执行 `profile use` 前必须先跑一次 `arkcli profile list --format json`；只有本轮已有完整、未截断的同源列表时可复用，不凭记忆认定目标存在。
2. `profile list` 的结构化 stdout 是本地 profile 的唯一事实源；列表为空就报告未配置并停止，不再搜索文件系统、读取状态目录或猜测配置文件。用户给了精确名称时只做精确匹配；名称不存在就如实说明，并只从列表真实结果提供可用 profile。不得把该名称重解释为 project / region，也不得自动 `project` / `create` / `rename`。
3. 用户只给了 type、用途或模糊描述时，只用列表里的 `name / type / region / project` 等权威字段筛选，然后严格按 [`arkcli-shared`](../arkcli-shared/SKILL.md) 的 0 / 1 / N 候选规则消歧；用户选定前不执行写操作。
4. 目标唯一且用户意图明确后才执行 `arkcli profile use <name>`，随后必须跑 `arkcli profile show --format json` 验证实际 active profile。
5. 只有核验结果与目标一致才声称切换成功；报告 `name / type / region / project / base_url` 这些生效切面，不把 `profile use` 的 exit=0 当成最终证据。

## 切换默认资源的硬状态机

1. 用户没给具体资源 ID 时，先确定 modality，再只跑一次 `arkcli resources list --modality <m> --format json` 取得当前或 `--profile` 指定切面的完整候选。
2. 候选只能来自这次真实结果，按 [`arkcli-shared`](../arkcli-shared/SKILL.md) 的 0 / 1 / N 规则处理：多候选优先调用宿主的结构化选择能力，通用 Skill 不写死工具名。
3. 执行前复述目标 profile、modality 和精确 ID；使用 `arkcli profile set-default --modality <m> <id>` 的默认 inline verify。除非用户明确要求并接受跳过校验的风险，不得自行添加 `--skip-verify`。
4. 写入后必须跑 `arkcli profile show --format json` 或 `arkcli resources list --modality <m> --format json` 核验新 default；只有结构化结果与目标 ID 一致才声称切换成功。

## --profile flag 的精确语义（0.1.16 修正）

```
优先级:  --profile flag > ARK_PROFILE env > config.yaml default_profile
         > 第一个 type=platform 的 profile > "default" sentinel
```

**关键修正（codex P0-A）**：在 `resources list / profile keys refresh / profile set-default / profile models refresh` 这些命令上，`--profile` 不再只是改 target 对象名，而是真的切换执行身份 —— 内部用 `Factory.RebuildForProfile(name)` 重建 invoker，所以 `arkcli profile keys refresh --profile B` 会用 B 的 token / UserID 打控制面，而不是 active=A 的身份打完再写到 B。

Agent 行为约定：
- 用户跑 `--profile B` 之类的子命令时，**不要假设它跟 active profile 等价**；告诉用户"将以 B 的身份操作"
- 0.1.16 final clean-slate 模型: 整 arkcli 同一时间只 active 一个 identity. 新 SSO 完成时 `sso.ActivateIdentity` 检测 newKey vs `cfg.DefaultProfile.IdentityKey`: 一致 (alice 重登 alice) → 不动 yaml profile / `--profile B` 用于同 identity 内跨 type 的临时切换; 不一致 (跨 sub / 跨 tenant) → 全清 yaml/.env/identities 三层重建 (BuildFirstProfileSet 火山多 type 派生)

## Agent 快速执行顺序

1. 普通业务只需当前身份/Profile 摘要 → `arkcli auth whoami --format json`；不要调用 Profile 管理命令补齐普通准入
2. 用户明确要查看 active profile 或本地 Profile 清单 → 先说明 Key 同步/回写影响，再用 `arkcli profile show --format json` 或 `arkcli profile list --format json`
3. 用户要切默认 profile → 按上面的硬状态机先 `profile list`，精确定位后 `profile use <name>`，再 `profile show`
4. 用户要新建 profile：先问清楚 type（platform / agent-plan / coding-plan）→ `arkcli profile create --type ... --set-default`
5. 用户问 default 模型是什么 → plan 类用 `arkcli profile models list`，platform 用 `arkcli profile show` 看 `resources` 字段
6. 用户要换 default 资源 → 按上面的硬状态机取真实候选、消歧、`profile set-default`，再只读核验
7. 用户的 default API Key 报错 / key 列表过期 → `arkcli profile keys refresh`，然后 `arkcli profile keys list --format json` 看新清单。refresh 按 profile 类型取对应池：普通池、Agent Plan 个人版专属池或团队席位 Key；空结果会报错并保留本地 Key，不会成功清空。
8. 用户要选别的 key 作 default → `arkcli profile keys use <api-key>`（必须 ∈ `profile.available_api_keys`）
9. 用户要换 active project（不重登）→ `arkcli profile project`（无参拉真实 ListProjects 交互选；先复述「会把 platform profile 重命名/重派生到新 project，个人版 plan profile 保留」并确认）

## Token 额度包 / 资源包的 profile 归属

「19 元 Token 额度包」、「Token 资源包」、「预付费 Token 抵扣包」等产品仍属于标准 platform 按量调用的计费产品，不是 Agent Plan 或 Coding Plan 订阅身份：

- 应沿用 `type=platform` 的 profile，数据面使用 platform `/api/v3`。
- 价格、「额度包 / 套餐」市场名称、Token 数量都不是 ProfileType 判据；只有明确的 Agent Plan / Coding Plan 订阅或团队席位才使用对应 plan profile。
- 用户只问「它属于哪类 profile」时直接解释分类，禁止自动执行 `profile create` / `profile use` / `profile set-default`。
- 用户要实际使用时，先用 `arkcli auth whoami --format json` 核对当前持久 Profile 摘要；需要查看 Profile 详情或修改时，再按本 skill 的同步披露与写操作确认契约执行。

## 命令一览

| 命令 | 说明 | 改动来源 |
|------|------|---------|
| `arkcli profile list` | 列出所有 profile（含 type/region/project 切面） | 替代 `config list` |
| `arkcli profile show [name]` | 显示当前/指定 profile 详细信息 | 替代 `config show` |
| `arkcli profile use [name]` | 切换默认 profile（无参时交互选择） | 替代 `config switch` |
| `arkcli profile create --type ...` | 新建 profile（interactive 或 inline） | 替代 `config init`（type 改为必选） |
| `arkcli profile delete <name>` | 删除 profile（必须 `--yes` 才能跳确认） | 替代 `config delete` |
| `arkcli profile rename <old> --to <new>` | 重命名 profile（校验格式 + 唯一性） | 0.1.16 新增 |
| `arkcli profile project [<name>]` | 重选 active project（无参拉真实 ListProjects 交互选，列表置顶「账号全部资源」=不传 ProjectName/account-wide）；把 platform profile 重派生到新 project，个人版 plan profile 原样保留；不重登 | 0.1.17 新增 |
| `arkcli profile keys list` | 列 default + available API Keys（masked） | 0.1.16 新增 |
| `arkcli profile keys use <key>` | 切 default API Key（key 必须 ∈ available list） | 0.1.16 新增 |
| `arkcli profile keys refresh` | 按 profile 类型重拉普通池 / Agent Plan 专属池 / 团队席位 Key，非空时更新 available list | 0.1.16 新增 |
| `arkcli profile models list` | plan 类 profile 的 PlanTier + Resources defaults | 0.1.16 新增 |
| `arkcli profile models refresh` | 重拉 ListAgentPlanLatestModel，更新 Text.Default | 0.1.16 新增 |
| `arkcli profile set-default --modality <m> <id>` | 设某 modality 的 default 资源 ID | 0.1.16 新增 |

## ProfileType 选型速查

| type | 适用 | 数据面 base URL | 控制面 | 视觉模型 (image/video) |
|------|------|----------------|--------|----------------------|
| `platform` | 火山方舟 console 的标准用法 | `/api/v3` | OpenTOP | ✓ 默认 endpoint |
| `agent-plan` | 火山方舟 Agent Plan 订阅（个人版） | `/api/plan/v3` | OpenTOP + Plan API | ✓ AgentPlanImage/VideoModels 硬编 |
| `coding-plan` | 火山方舟 Coding Plan 订阅（个人版） | `/api/coding/v3` | OpenTOP + CodingPlan API | text: 套餐内文本模型；image/video: 借道 platform 数据面 + 用户 +deploy 的 endpoint id (S10, commit f69be53) |
| `agent-plan-team` | Agent Plan 团队席位 | `/api/plan/v3` | OpenTOP + Plan API | text/image/video 都用套餐模型 + 团队席位 Key |
| `coding-plan-team` | Coding Plan 团队席位 | `/api/coding/v3` | OpenTOP + CodingPlan API | text 用套餐模型 + 团队席位 Key；image/video 用 platform Endpoint + 后付费 API Key |

> Coding Plan 个人版的 text lane 使用后付费 API Key；Coding Plan Team 只有 text
> lane 使用团队席位 Key。完整矩阵见
> [`../arkcli-shared/references/execution-context.md`](../arkcli-shared/references/execution-context.md)。

`plan-tier`：
- agent-plan：`small` / `medium` / `large` / `max`
- coding-plan：`lite` / `pro`

## profile create 决策树

```
用户提到 "Agent Plan / 我买了 plan"?
  yes -> --type agent-plan (--plan-tier 由 Detect 自动识别；后端可见性问题时手动加 --plan-tier=<tier>)
  no  -> 用户提到 "Coding Plan / claude code 整合"?
           yes -> --type coding-plan
           no  -> --type platform (默认场景, region+project 必填)
```

## 常见错误

- `set-default: <id> 不在当前 profile (... ) 可用列表` → 跑 `arkcli resources list --modality <m>` 看可用 ID，或加 `--skip-verify` 强写
- `models refresh: profile %q type=%q (仅 agent-plan 支持)` → 不是 agent-plan profile，先 `profile use <agent-plan-profile>` 或切对 `--profile`
- S10 之后, coding-plan profile 下 `profile set-default --modality image|video <ep>` 不再 fail-fast: verify 会借道 platform 控制面 ListEndpoints 校验 ep-id 是否存在
- `keys refresh: fetch api keys: NotLogin` → 控制面鉴权失败 (如登录态/STS 过期)；fetcher 会用 `.env` 缓存单 key 兜底，profile.available_api_keys 仅含 1 项，恢复后再 refresh
- `keys refresh` 报没有可用 Key → 先按 profile type 判断：`platform` / `coding-plan` 个人版去普通 API Key 管理页；`agent-plan` 个人版去 Agent Plan 使用配置页检查专属 Key；团队版检查 Running 席位。不得把所有类型都引向普通 API Key 页面。
- `arkcli auth apikey` 只操作普通 API Key 池；它返回 `saved=true` 也不表示 Agent Plan 专属 Key 或团队席位 Key 已恢复。
- 切账号后 keys / models refresh 行为奇怪 → 检查 `arkcli auth whoami`，可能 active profile 仍绑旧 identity；P0-D 之后 STS / token 都在 per-identity store，但 active profile.identity_key 是 yaml 字段，跨账号要么 `profile use <new>`，要么走 SSO Gate 2 自动新建

## deprecated 命令自然语言重定向表

| 用户提到 | 实际执行 |
|---|---|
| "config init / 新建配置 / 初始化配置 / 新建 profile" | `arkcli profile create --type ...` |
| "config list / 看看有几个 profile / 列出 profile" | `arkcli profile list` |
| "config show / 看下我的配置 / 显示 profile" | `arkcli profile show` |
| "config switch / 切配置 / 切换 profile / 换 profile" | `arkcli profile use [name]` |
| "config reset / 全清掉 / 恢复出厂设置 / 配置乱了从头来 / 清空所有配置" | **转 `arkcli-config`**：`arkcli config reset`（破坏性操作，必须先确认） |
| "我 agent-plan 下面有哪些可用模型 / 有哪些模型 / 模型列表" | `arkcli profile models list`（plan 类 profile）或 `arkcli resources list` |

## 参考

- [`references/arkcli-profile-create.md`](references/arkcli-profile-create.md)
- [`references/arkcli-profile-keys.md`](references/arkcli-profile-keys.md)
- [`references/arkcli-profile-set-default.md`](references/arkcli-profile-set-default.md)
- [`../arkcli-shared/references/execution-context.md`](../arkcli-shared/references/execution-context.md) — 五类 Profile 的调用矩阵与临时覆盖
