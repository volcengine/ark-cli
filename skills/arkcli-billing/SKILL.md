---
name: arkcli-billing
version: 1.1.3
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

## Step 0：确认账单口径与范围

先分清“已结算金额”与“Token / 套餐额度消耗”，后者交给 Usage，不用单价乘估算 Token 冒充账单。用户只查某个 Endpoint、Key 或订阅时，仅查其明确范围；`profile.type=platform` 不能证明没有套餐。

身份上下文已知则复用，确需核实时按共享宿主协议使用 `arkcli auth whoami --format json`；托管宿主的注入协议优先。普通账单查询不调用可能同步/回写 Key 的 `profile show/list/keys list`，不为了补全账单切换身份。

- “我花了多少”先保留 `--mine`，按下面的本人资源流程查；用户要求订阅类账单时显式 `--product ark_subscription`，并解释其账号/订阅范围，不称为个人 Endpoint 用量。
- “整个账号花了多少”才使用账号查询；Project 筛选按本次实际上下文说明，不能为补数擅自清空 Project、换 Profile 或换身份。
- 每条 `billing list` 必带 `--start <YYYY-MM>`，跨月加 `--end <YYYY-MM>`。账期是月份，不是 Usage 的日级参数；T+1 及当期未出账部分必须披露。
- 返回的账单数组是 `items`，不是 Endpoint 列表的 `Items`。解析失败、请求失败、截断与真实空列表分开报告。

## 快速决策

| 用户问 | 命令 |
|---|---|
| **我**这个月 / 上个月花了多少 | `arkcli billing list --start <YYYY-MM> --mine`；明确要订阅账单时按 Step 0 另选产品并披露范围 |
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
  - 仍撞空 → 报告“在本次身份、范围和账期内未查到匹配账单”，核对时间、出账延迟和资源范围；不能据此宣称账号零资源或自动切身份
  - **⛔ 禁止退化为不带 `--mine` 的全量查询** — 全账号金额 ≠ "我花的",一旦丢 `--mine` 范围语义就跑偏(对齐 [arkcli-usage 的同条款](../arkcli-usage/SKILL.md))

dim 间 fallback (endpoint→apikey) agent **可以自动重试**,因为同 mine 语义、不改查询范围;但**不能丢 `--mine` 退到全量**(那是改语义)。

## Agent 关键纪律

- **保留 stderr**：首次查询和用于得出结论的查询不得加 `2>/dev/null`。账号全量 scope 提示、软截断 WARN 与部分失败原因都可能只写 stderr；需要解析 JSON 时只接管 stdout，或优先使用 `--output FILE`。
- **`is_truncated=true` 时停下问用户,不要自动决策** — 撞 cap (preflight 模式下 items 不返,只有 metadata + partial_failures + stderr 警告) 是规模信号不是错误。先把 `total_records` 和 `partial_failures[*].total` 报给用户,列出 4 条出路,**等用户明确选一个再继续**。选哪条要看用户当前任务意图(对账 / 排查 / 导出),agent 不要替用户拍。**永远不要 sum items 当总额**(items 可能为空或部分):
  - 想要全量明细 → `--output FILE` (落盘 stdout 不爆;自动放宽 cap 到 300k 行)
  - 只要总金额 → `--split-dim apikey|endpoint` (服务端聚合到几十行)
  - 缩到一个资源 → `--endpoint <ep-...>` / `--apikey <ark-...>`
  - 强行流进 stdout → `--page-limit=N` (N 见 partial_failures.reason 建议;mind context size)
- **`--limit/--offset` 在 fan-out 时是 per-fan-out 各 N 行** (`--end` 跨月或 `--mine` 多资源时), `partial_failures` 标 `reason="windowed sample"`, agent 据此知 returned ≤ limit × fanout_count
- **金额是 CNY 字符串** — JSON 数字精度有损,加和 / 比较用 decimal 库,不要 `parseFloat`
- **月初对账拉前一个完整账期** — T+1 出账,当月当日数据不全
- **`--mine` ≠ `PayerID`** — `--mine` 按 IAM 子用户过滤资源,`PayerID` 是财务托管 owner 账号 ID,单账户场景手撸 `PayerID` 是 no-op (= 全账号查询)。默认 `--mine-by=endpoint`(infra ownership,对齐 usage stats),要看 cost causation (我的 key 在花钱) 显式 `--mine-by=apikey`
- 默认 scope = ARK 推理 / Agent 9 个支持 API Key 分账的产品(**不含 `ark_subscription`**,要看订阅类显式 `--product ark_subscription`) + profile.project 自动注入(若 profile.project 是具体 id 如 `auto-test`,自动按 project 过滤;`default`/账号全部资源哨兵/空跳过 = 真账号全量;stderr 出软提示)。要强制账号全量传 `--project=` (空值清空默认)

## 常见降级

- 鉴权错误 / SSO 失效 → [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
- 想看推理量 (token / 请求数) → [`../arkcli-usage/SKILL.md`](../arkcli-usage/SKILL.md)
- 想看模型 / 套餐价格目录 → [`../arkcli-pricing/SKILL.md`](../arkcli-pricing/SKILL.md)

## 命令一览

| 命令 | 说明 |
|------|------|
| [`billing list`](references/arkcli-billing-list.md) | 拆分账单明细查询(结算金额 × Token 用量) |
