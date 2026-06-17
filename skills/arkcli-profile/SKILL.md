---
name: arkcli-profile
version: 1.0.0
description: "arkcli profile 切面管理：列出、查看、新建、切换、删除、重命名 profile；管理 profile 内 API Key 列表；管理 plan 类 profile 的 default 模型（text/image/video）；设置某 modality 的默认资源 ID。0.1.16 把原 `arkcli config init/list/show/switch/delete` 全部迁过来，是 profile 类操作的唯一入口；旧 config 子命令已 deprecated。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli profile --help"
---

# arkcli profile

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、配置排查与命令选择顺序**

**CRITICAL — 一旦确定走 `profile create`、`profile delete` 或 `profile project`（重选 project 会重命名/重派生 platform profile），必须先复述对 `config.yaml` 的影响并征得用户确认；其他写操作（`use` / `set-default` / `keys use` / `keys refresh` / `models refresh` / `rename`）执行前也要复述目标 profile 名。**

## 使用原则

- profile 是 0.1.16 引入的 **统一身份切面**，把 `(type × region × project × owner_trn × api_keys)` 五个属性绑成一组
- profile 写操作（create / use / set-default / keys / models / delete / rename）一律走 `arkcli profile <verb>`；旧的 `arkcli config init/list/show/switch/delete` 已 deprecated，不要再引导用户用
- 只读排障优先 `arkcli profile show` 或 `arkcli profile list`，不要上来就改
- ProfileType 三种：`platform` / `agent-plan` / `coding-plan`；选错 type 后续 `+chat / +gen` 默认模型派发会错

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

1. 不确定当前 active profile → `arkcli profile show --format json`
2. 不确定有哪些 profile → `arkcli profile list --format json`
3. 用户要切默认 profile → `arkcli profile use <name>`（无 `<name>` 时会弹 promptui）
4. 用户要新建 profile：先问清楚 type（platform / agent-plan / coding-plan）→ `arkcli profile create --type ... --set-default`
5. 用户问 default 模型是什么 → plan 类用 `arkcli profile models list`，platform 用 `arkcli profile show` 看 `resources` 字段
6. 用户要换 default 资源 → `arkcli profile set-default --modality <m> <id>`（默认 inline verify <id> ∈ 可用列表，加 `--skip-verify` 强写）
7. 用户的 default API Key 报错 / key 列表过期 → `arkcli profile keys refresh`，然后 `arkcli profile keys list --format json` 看新清单
8. 用户要选别的 key 作 default → `arkcli profile keys use <api-key>`（必须 ∈ `profile.available_api_keys`）
9. 用户要换 active project（不重登）→ `arkcli profile project`（无参拉真实 ListProjects 交互选；先复述「会把 platform profile 重命名/重派生到新 project，个人版 plan profile 保留」并确认）

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
| `arkcli profile keys refresh` | 重拉控制面 ListApiKeys，更新 available list | 0.1.16 新增 |
| `arkcli profile models list` | plan 类 profile 的 PlanTier + Resources defaults | 0.1.16 新增 |
| `arkcli profile models refresh` | 重拉 ListAgentPlanLatestModel，更新 Text.Default | 0.1.16 新增 |
| `arkcli profile set-default --modality <m> <id>` | 设某 modality 的 default 资源 ID | 0.1.16 新增 |

## ProfileType 选型速查

| type | 适用 | 数据面 base URL | 控制面 | 视觉模型 (image/video) |
|------|------|----------------|--------|----------------------|
| `platform` | 火山方舟 console 的标准用法 | `/api/v3` | OpenTOP | ✓ 默认 endpoint |
| `agent-plan` | 火山方舟 Agent Plan 订阅（个人版） | `/api/plan/v3` | OpenTOP + Plan API | ✓ AgentPlanImage/VideoModels 硬编 |
| `coding-plan` | 火山方舟 Coding Plan 订阅（个人版） | `/api/coding/v3` | OpenTOP + CodingPlan API | text: 套餐内文本模型；image/video: 借道 platform 数据面 + 用户 +deploy 的 endpoint id (S10, commit f69be53) |

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
