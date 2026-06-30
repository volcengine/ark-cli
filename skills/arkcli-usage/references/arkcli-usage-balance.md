# usage balance

> **前置条件:** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **跟 `stats` / `plan` 的区别(务必先看):**
> - `stats` / `plan-details` — **消耗视角**(已经发生的 token / 调用)
> - `plan` — 订阅套餐的 quota **快照**(已用 / 总额 / 重置时间)
> - `balance` — **余额视角**,统一三类"还剩多少 X"的查询入口
>
> 用户问「我账上还有多少 X」时走 `balance`,根据 `--type` 分流到三类数据源。
>
> **`balance --type plan` 跟 `usage plan` 是同一份数据的两个入口**(都调 `QueryPlanUsage`,默认探所有 4 个 SKU,flag 完全一致)。区别只在投影:
> - `balance --type plan` 精简版(去掉 `subscribed` / `updated_at`,补 `"not subscribed"` hint)
> - `usage plan` 完整版(带 `subscribed: true/false`、CodingPlan 的 `updated_at` 等元字段)
>
> 用余额视角(配 free-quota / media-asset 一起看)走 `balance --type plan`;直接看套餐 quota 快照(配 stats / plan-details 一起看)走 `usage plan`。

`arkcli usage balance` 是一个 dispatcher,`--type` 决定数据源。三类目的差异显著,**`--type` 必传**。

## 命令

```bash
# 各模型的免费推理额度 / 资源包余额(走 ListModelChargeItems)
arkcli usage balance --type free-quota
arkcli usage balance --type free-quota --model doubao-seed-1-6
arkcli usage balance --type free-quota --modality LLM --page-size 20

# 媒资库容量(走 /open/GetAssetQuota)
arkcli usage balance --type media-asset
arkcli usage balance --type media-asset --project my-project

# 订阅套餐余额(走 QueryPlanUsage,精简版输出)
arkcli usage balance --type plan
arkcli usage balance --type plan --product agent-plan
arkcli usage balance --type plan --all
```

## 参数白名单

`--type` 是必填,其它 flag **按 type 分组**;跨 type 误传会报 `ErrValidation`(不静默忽略,避免用户以为过滤生效了但实际没生效)。

| `--type` | 允许的额外 flag |
|---|---|
| `free-quota` | `--model` `--modality` `--page-number` `--page-size` |
| `media-asset` | `--project` |
| `plan` | `--product` `--all` `--seat` |

### `--type free-quota`

| 参数 | 类型 | 说明 |
|---|---|---|
| `--model` | string | 单基础模型名;不传 = 列全量已开通模型 |
| `--modality` | string | 后端 `FoundationModelDomain` 枚举:`LLM` / `ComputerVision` / `Audio` / `Embedding` / `Router`(**不接受别名**) |
| `--page-number` | int | 页码,1-based |
| `--page-size` | int | 每页条数,默认 50 |

### `--type media-asset`

| 参数 | 类型 | 说明 |
|---|---|---|
| `--project` | string | 项目名;不传走当前 profile.ProjectName |

### `--type plan`

| 参数 | 类型 | 说明 |
|---|---|---|
| `--product` | string | `agent-plan` / `coding-plan` / `agent-plan-team` / `coding-plan-team` |
| `--all` | bool | 查所有个人版 plan 即使未订阅(跟 `--product` 互斥) |
| `--seat` | string | 仅 team 版,过滤单席位 |

## 返回值

### `--type free-quota`

```json
{
  "page_number": 1,
  "page_size": 50,
  "total_count": 7,
  "items": [
    {
      "model": "doubao-seed-1-6",
      "display_name": "Doubao Seed 1.6",
      "vendor": "Doubao",
      "state": "Available",
      "is_overdue": false,
      "free_usage": { "total": 500000, "consumed": 12345, "remaining": 487655 },
      "resource_packs": [
        {
          "type": "FreeInference",
          "total": 500000, "consumed": 12345, "remaining": 487655,
          "reclaimed": 0, "sync_time": "2026-06-04T12:00:00Z"
        }
      ]
    }
  ]
}
```

**关键口径**:

- `free_usage.remaining = total - consumed`,**已经算好**,不要让 agent 自己算
- 没有任何余额信息(`free_usage` 和 `resource_packs` 都空)的模型行**不会输出**——`pricing models` 才出全表
- `resource_packs[].type` 当前火山只有两个值:`FreeInference`(免费推理) / `DataPermission`(数据权限)。**没有 finetune 资源包**,`Finetune` / `LoraFinetune` 是定价单价(`pricing models` 的 `ChargeItems.Type`),不是余额维度
- 资源包 `remaining` **可能为负**(过期降档时 `consumed > total`),不要 clamp

### `--type media-asset`

```json
{
  "tier": "advanced_monthly",
  "assets":      { "total": 100000, "used": 23456, "remaining": 76544, "percent": 23.456 },
  "asset_groups":{ "total": 100,    "used": 12,    "remaining": 88,    "percent": 12.0 },
  "shared_pool": { "total": 1000000,"used": 0,     "remaining": 1000000,"percent": 0 },
  "write_qpm": 600,
  "capabilities": { "liveness_writable": true, "aigc_readable": true },
  "projects": [
    { "project_name": "default", "used": 23456, "allocated": false, "allocation": 0 }
  ],
  "updated_at_ms": 1717488000123
}
```

**关键口径**:

- `tier` 取值:`free` / `advanced` / `premium` / `advanced_monthly` / `premium_monthly`(月付与年付共用展示)。空值意味着账号没订阅媒资库 → 后端返 free tier 零容量
- **`used > total` 是合法状态**(过期降档时 `tier` 降到 `free` 但 `used` 保留旧值),`percent` 可 > 100,`remaining` 可为负 — **不 clamp**
- `shared_pool` 在没启用共享池的账号下被省略(全 0 不输出)
- `capabilities` 只列**已开启**的能力,空 map 意味着所有能力都关
- `projects[]` 只在**项目级配额分配生效**的企业账号下有数据,个人账号通常空数组

### `--type plan`

```json
{
  "viewer": {
    "auth_method": "sso",
    "user_id": "12345",
    "user_name": "alice",
    "is_root": false,
    "profile": "my-profile",
    "tenant": "volc",
    "region": "cn-beijing"
  },
  "items": [
    {
      "product": "agent-plan",
      "edition": "personal",
      "tier": "small",
      "periods": [
        { "label": "5h",     "used": 10, "total": 100, "percent": 10, "reset_at": "2024-06-04T17:00:00+08:00" },
        { "label": "weekly", "used": 50, "total": 1000, "percent": 5, "reset_at": "2024-06-11T08:00:00+08:00" }
      ]
    },
    {
      "product": "agent-plan-team",
      "edition": "team",
      "tier": "large",
      "seat_id": "seat-001",
      "periods": [
        { "label": "5h",      "used": 50,    "total": 200,    "percent": 25, "reset_at": "2024-06-04T17:00:00+08:00" },
        { "label": "weekly",  "used": 800,   "total": 5000,   "percent": 16, "reset_at": "2024-06-11T08:00:00+08:00" },
        { "label": "monthly", "used": 4000,  "total": 20000,  "percent": 20, "reset_at": "2024-06-29T05:06:40+08:00" }
      ]
    },
    {
      "product": "coding-plan",
      "edition": "personal",
      "error": "not subscribed"
    }
  ]
}
```

**关键口径**:

- **顶层 `viewer`**:身份摘要(子用户 vs root,SSO/AKSK/APIKey)。让 agent 区分"我看的是自己的 vs 主账号的 vs 别人"。详见 [`usage plan` 文档的 viewer 章节](arkcli-usage-plan.md#viewer-字段身份摘要)
- **默认探所有 4 个 SKU**(personal × 2 + team × 2)— 一次拉全你账号下的所有套餐余额。要单查传 `--product=X`
- 比 `usage plan` **精简**:去掉 `subscribed` / `update_timestamp` 元字段;`error` 保留(用于区分"未订阅" / "AccessDenied" / "no seat bound")
- `error: "not subscribed"` 是 balance 层补的 hint,后端原始是 `Subscribed=false` + 空 periods,直接展示对用户不可读
- **Team 套餐自动找 caller seat**:不传 `--seat` 时 service 层自动调 `GetSeatInfo` 找你绑定的席位。订了套餐但没分到席位时返 `error: "no seat bound to caller for ..."`(区分于"未订阅")
- `coding-plan` 后端只返 `percent`,`used` / `total` 字段缺(JSON `omitempty` 不出现)
- `periods[].reset_at` 跟 `usage plan` 同 schema,输出 **RFC3339 北京时间(UTC+08:00)**;周期内无数据时不输出该字段

## Agent 决策

| 用户问法 | 命令 |
|---|---|
| 我哪些模型还有免费额度 | `arkcli usage balance --type free-quota` |
| 模型 X 还能用多少 token | `arkcli usage balance --type free-quota --model X` |
| 媒资库还能上传多少 / 容量满了吗 | `arkcli usage balance --type media-asset` |
| 我的套餐还能用多少 / 几号刷新 | `arkcli usage balance --type plan`(自动探所有订阅) |
| 我特定 product 的套餐余额 | `arkcli usage balance --type plan --product agent-plan-team` |
| Admin 视角看其他席位用量 | `arkcli usage balance --type plan --product agent-plan-team --seat <id>` |

## 常见错误

| 错误 | 原因 | 处理 |
|---|---|---|
| `--type is required` | 没传 `--type` | 三选一:`free-quota` / `media-asset` / `plan` |
| `--product is not valid for --type free-quota` | 跨 type flag | 检查 flag 跟 `--type` 是否匹配(见上表) |
| `tier` 为空 + 全部 0 | 账号没订阅媒资库 | 后端返 free tier 零容量,没问题;要订阅请走火山方舟控制台 |
| `foundation model "X" not found in pricing list` (exit=2) | `--model X` 名字打错或不在 pricing list | 检查拼写;或不传 `--model` 拉全量看支持的名字 |
| `free-quota` 返回 `items: []` | `--modality` 传了别名(`text` / `video`)而不是后端枚举(没传 `--model` 时才会走到这条) | 翻译成 `LLM` / `ComputerVision` 再传 |
| plan 行 `error: "no seat bound"` | 团队版订了但 caller 没分到席位 | 联系 admin 分配席位,或传 `--seat <id>` 显式指定 |

## 跟其它命令的关系

- 想看**单价**(每千 token 多少钱)→ [`arkcli pricing models`](../../arkcli-pricing/references/arkcli-pricing-models.md)
- 想看**已经消耗**(token / 调用次数)→ [`stats`](arkcli-usage-stats.md) / [`plan-details`](arkcli-usage-plan-details.md)
- 想看**订阅套餐 quota 快照(完整版,带 subscribed / 时间戳)**→ [`usage plan`](arkcli-usage-plan.md)

## 参考

- [arkcli-usage](../SKILL.md) — usage skill 概览
- [arkcli-shared](../../arkcli-shared/SKILL.md) — 认证和全局参数
- [arkcli-pricing](../../arkcli-pricing/SKILL.md) — 定价目录视角
