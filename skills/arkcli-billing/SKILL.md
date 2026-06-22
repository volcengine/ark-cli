---
name: arkcli-billing
version: 1.1.1
description: "查询火山引擎 ARK 拆分账单明细（结算金额、Token 用量计费），支持按账期月、月范围、Endpoint、API Key、产品编码等维度过滤。当用户问账单、花了多少钱、对账、账期、按 EP / API Key 拆账、按产品拆账、月度账单、出账明细时使用。注意 billing 跟 usage stats 不同：stats 出推理量（近实时），billing 出结算金额（T+1 出账，财务口径）。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli billing list --help"
---

# arkcli billing

**执行前 MUST 读取** [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md) (认证 / 全局规则) 和 [`references/arkcli-billing-list.md`](references/arkcli-billing-list.md) (参数详解)。

## 适用场景

- 查询某账期月或月范围内的拆分账单明细(结算金额 / Token 用量)
- 按 Endpoint、API Key、产品编码、账单类型过滤拆账行
- 月度 / 季度 / 年度对账
- 排查"这个 EP / 这把 key 花了多少钱"

## 业务定位 — billing vs usage stats

| 维度 | `usage stats` | `billing list` |
|---|---|---|
| 数据源 | ARK BFF 推理聚合 | 火山计费中心拆账 |
| 时效 | 5–30 分钟延迟 | T+1 出账 |
| 单位 | Token / 请求数 | CNY 金额 |

想看"用了多少 token"→ [`arkcli-usage`](../arkcli-usage/SKILL.md);想看"花了多少钱"→ 本 skill。

## Step 0(MUST):查"我的账单"先按 profile 路由

跟 [`arkcli-usage` 的 Step 0](../arkcli-usage/SKILL.md)(canonical 原则)同一套 —— **先查"自己这档 profile 的套餐账单",再查 endpoint 账单**。查任何"**我**…花了多少"前必走:

1. **探 profile.type**:`arkcli profile show --format json`,读 `type`
2. **定模态**:用户点名模型/模态 → 只查该模态;没点名 → 全模态都覆盖
3. **按 (type × modality) 路由**(`①→②` = 先套餐账单、再 endpoint 账单;单格 = 只查 endpoint 账单):

| profile.type | text | image / video |
|---|---|---|
| `agent-plan` / `agent-plan-team` | ① `arkcli billing list --start <YYYY-MM> --product ark_subscription`<br>② `arkcli billing list --start <YYYY-MM> --mine` | 同 text(全模态都先套餐、再 endpoint) |
| `coding-plan` / `coding-plan-team` | ① `arkcli billing list --start <YYYY-MM> --product ark_subscription`<br>② `arkcli billing list --start <YYYY-MM> --mine` | `arkcli billing list --start <YYYY-MM> --mine`(套餐不覆盖) |
| `platform` | `arkcli billing list --start <YYYY-MM> --mine`(无套餐) | `arkcli billing list --start <YYYY-MM> --mine` |

依据同 usage:Agent Plan 套餐覆盖三模态、Coding Plan 套餐只覆盖文本、Platform 无套餐。套餐账单走订阅类必须显式 `--product ark_subscription`(默认 scope 不含订阅,见下方「Agent 关键纪律」);endpoint 账单 `--mine` 的撞空 fallback 见下。**每条 `billing list` 必带 `--start <YYYY-MM>`(月单位,缺省即报 `--start is required`;查多月范围再加 `--end <YYYY-MM>`)。**

## 快速决策

| 用户问 | 命令 |
|---|---|
| **我**这个月 / 上个月花了多少 (含"我"语义) | **先过 Step 0(本文档顶部)**:plan profile(agent / coding-text)先 `arkcli billing list --start <YYYY-MM> --product ark_subscription` 套餐账单、再 `arkcli billing list --start <YYYY-MM> --mine` endpoint 账单;platform 或非覆盖模态直接 `arkcli billing list --start 2026-05 --mine` — 撞空有 fallback 流程见下 |
| 整个账号花了多少 / 公司账号 / 整体 (无主语自指) | `arkcli billing list --start 2026-05` (账号维度全量) |
| 这个 EP 花了多少 | `arkcli billing list --start 2026-05 --endpoint ep-...` |
| 这把 key 花了多少 | `arkcli billing list --start 2026-05 --apikey ark-...` |
| 账号下每个 EP / key 各花了多少 | `--split-dim endpoint` 或 `--split-dim apikey` |
| 按天 / 按每条结算明细 | `--interval day` 配 `--day YYYY-MM-DD`,或 `--interval detail` |
| 过去 N 个月对账 | `--start ... --end ...` (闭区间, 最多 24 个月) |
| 看 Agent Plan / Coding Plan 订阅类账单 | 显式 `--product ark_subscription` (默认查询不含) |

### `--mine` fallback 流程 (跟 usage stats 对齐)

用户问"**我**...花了多少"时:

- **火山:** 先 `arkcli billing list --start <YYYY-MM> --mine`(默认 `--mine-by=endpoint`)
  - 有数据 → 用 `partial_failures` 检查截断后,告诉用户金额合计
  - 撞空 (`no endpoints owned by current sub-user`) → **立即重试** `--mine --mine-by=apikey`
  - 仍撞空 → 报告用户"该子用户在该账号下零资源,账单可能是主账号 / 其他子用户产生的",问是否切主账号或换 profile
  - **⛔ 禁止退化为不带 `--mine` 的全量查询** — 全账号金额 ≠ "我花的",一旦丢 `--mine` 范围语义就跑偏(对齐 [arkcli-usage 的同条款](../arkcli-usage/SKILL.md))

dim 间 fallback (endpoint→apikey) agent **可以自动重试**,因为同 mine 语义、不改查询范围;但**不能丢 `--mine` 退到全量**(那是改语义)。

## Agent 关键纪律

- **`is_truncated=true` 时停下问用户,不要自动决策** — 撞 cap (preflight 模式下 items 不返,只有 metadata + partial_failures + stderr 警告) 是规模信号不是错误。先把 `total_records` 和 `partial_failures[*].total` 报给用户,列出 4 条出路,**等用户明确选一个再继续**。选哪条要看用户当前任务意图(对账 / 排查 / 导出),agent 不要替用户拍。**永远不要 sum items 当总额**(items 可能为空或部分):
  - 想要全量明细 → `--output FILE` (落盘 stdout 不爆;自动放宽 cap 到 300k 行)
  - 只要总金额 → `--split-dim apikey|endpoint` (服务端聚合到几十行)
  - 缩到一个资源 → `--endpoint <ep-...>` / `--apikey <ark-...>`
  - 强行流进 stdout → `--page-limit=N` (N 见 partial_failures.reason 建议;mind context size)
- **`--limit/--offset` 在 fan-out 时是 per-fan-out 各 N 行** (`--end` 跨月或 `--mine` 多资源时), `partial_failures` 标 `reason="windowed sample"`, agent 据此知 returned ≤ limit × fanout_count
- **金额是 CNY 字符串** — JSON 数字精度有损,加和 / 比较用 decimal 库,不要 `parseFloat`
- **月初对账拉前一个完整账期** — T+1 出账,当月当日数据不全
- **`--mine` ≠ `PayerID`** — `--mine` 按 IAM 子用户过滤资源,`PayerID` 是财务托管 owner 账号 ID,单账户场景手撸 `PayerID` 是 no-op (= 全账号查询)。默认 `--mine-by=endpoint`(infra ownership,对齐 usage stats),要看 cost causation (我的 key 在花钱) 显式 `--mine-by=apikey`
- 默认 scope = ARK 推理 7 个产品(**不含 `ark_subscription`**,要看订阅类显式 `--product ark_subscription`) + profile.project 自动注入(若 profile.project 是具体 id 如 `auto-test`,自动按 project 过滤;`default`/账号全部资源哨兵/空跳过 = 真账号全量;stderr 出软提示)。要强制账号全量传 `--project=` (空值清空默认)

## 常见降级

- 鉴权错误 / SSO 失效 → [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
- 想看推理量 (token / 请求数) → [`../arkcli-usage/SKILL.md`](../arkcli-usage/SKILL.md)
- 想看模型 / 套餐价格目录 → [`../arkcli-pricing/SKILL.md`](../arkcli-pricing/SKILL.md)

## 命令一览

| 命令 | 说明 |
|------|------|
| [`billing list`](references/arkcli-billing-list.md) | 拆分账单明细查询(结算金额 × Token 用量) |
