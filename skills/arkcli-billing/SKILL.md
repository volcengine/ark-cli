---
name: arkcli-billing
version: 1.0.0
description: "查询火山引擎 ARK 拆分账单明细（结算金额、Token 用量计费），支持按账期月、月范围、Endpoint、API Key、产品编码等维度过滤。当用户问账单、花了多少钱、对账、账期、按 EP / API Key 拆账、按产品拆账、月度账单、出账明细时使用。注意 billing 跟 usage stats 不同：stats 出推理量（近实时），billing 出结算金额（T+1 出账，财务口径）。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli billing list --help"
---

# arkcli billing

**CRITICAL — 开始前 MUST 用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md),其中包含认证闸门、配置排查与共享安全规则。**
**CRITICAL — `billing list` 在执行之前,务必先用 Read 工具读取 [`references/arkcli-billing-list.md`](references/arkcli-billing-list.md),禁止直接盲目调用命令。**

## 适用场景

- 查询某账期月或月范围内的拆分账单明细(结算金额 / Token 用量)
- 按 Endpoint、API Key、产品编码、账单类型、计费模式过滤拆账行
- 月度 / 季度 / 年度对账, 拉前一个完整账期(T+1 数据)
- 排查"这个 EP 花了多少钱 / 这把 key 花了多少钱"

## 业务定位

- **跟 `usage` 故意分开** — 数据源 / 时效 / 鉴权边界都不同:
  | 维度 | `usage stats` | `billing list` |
  |---|---|---|
  | 数据源 | ARK BFF 推理聚合 | 火山计费中心拆账 |
  | 时效 | 5–30 分钟延迟(近实时) | T+1 出账(财务口径) |
  | 单位 | Token / 请求数 | CNY 金额 |
  | 用途 | 用了多少 | 花了多少 |
- 想看"我用了多少 token"→ [`arkcli-usage`](../arkcli-usage/SKILL.md);想看"我花了多少钱"→ 本 skill
- **不暴露 Limit / Offset**: service 层强制全量分页, 翻页对账容易少算 — 这个产品决策不可妥协

## 快速决策

- 用户问"我这个月花了多少 / 上个月账单":
  - `arkcli billing list --start 2026-05`(默认账号维度 ARK 推理全量,Product 自动锁 7 个 arkProducts;**不含 `ark_subscription` / Agent Plan / Coding Plan**,要显式 `--product ark_subscription` 看订阅类)
- 用户问"这个 EP 花了多少":
  - `arkcli billing list --start 2026-05 --endpoint ep-20260415xxx-xxxxx`
- 用户问"这把 key 花了多少"(显式查别的 key):
  - `arkcli billing list --start 2026-05 --apikey ark-xxxxxxxxxxxx`(自动经 arkbff 反查 SID)
- 用户问"账号下每个 EP / API key 各花了多少":
  - `arkcli billing list --start 2026-05 --split-dim endpoint`(账号维度按 EP 切片)
  - `arkcli billing list --start 2026-05 --split-dim apikey`(账号维度按 API key 切片)
- 用户问"我自己花了多少 / 我的费用"(SSO 子用户视角):
  - `arkcli billing list --start 2026-05 --mine`(只看我创建的 apikey 的账单;**不含订阅类**,因为订阅没 IAM 归属)
  - **不要**用 `--params '{"PayerID":[<账号ID>]}'`,那是财务托管 owner-账号过滤,单账户场景是 no-op
- 用户问"按天看 / 看每条结算明细":
  - `--interval day`(同月内单天用 `--day`)/ `--interval detail`
- 用户问"过去 N 个月对账":
  - `--start YYYY-MM --end YYYY-MM`(闭区间, 最多 24 个月; 客户端 fan-out)
- 数据为空 / 出不来:先确认 profile 是火山, 再确认账期是完整月(T+1 出账)

## Agent 快速执行顺序

1. 先确认认证状态;不确定时先看 `arkcli auth status`
2. 月初对账先拉前一个完整账期(T+1 出账, 当月当日数据可能不全)
3. scope 默认 = 账号维度 ARK 推理 7 个产品(不含 `ark_subscription`,stderr 出软提示告知);要缩范围传 `--endpoint` / `--apikey` / `--apikey-sid` 之一,要在账号维度内切片传 `--split-dim apikey` 或 `--split-dim endpoint`(显式 scope 自带 dim,split-dim 必须一致或不传),要看 Agent Plan / Coding Plan 等订阅类账单显式 `--product ark_subscription`
4. 数据量大(单账期 30000+ 行)时加 `--endpoint` / `--apikey` 缩小范围, 或显式 `--page-limit=N` 放宽截断阈值
5. 金额字段是字符串(避免 JSON 数字精度损失), 加和 / 比较请用 decimal 库, 不要 `parseFloat`

## 常见降级

- 鉴权错误 / SSO 失效:转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
- 想看推理量(token / 请求数)而非金额:转 [`../arkcli-usage/SKILL.md`](../arkcli-usage/SKILL.md)
- 想看模型 / 套餐价格(目录):转 [`../arkcli-pricing/SKILL.md`](../arkcli-pricing/SKILL.md)

## 命令一览

| 命令 | 说明 |
|------|------|
| [`billing list`](references/arkcli-billing-list.md) | 拆分账单明细查询(结算金额 × Token 用量), 按账期月或月范围, 单值 scope(EP / API Key) |

## 参考

- [arkcli-usage](../arkcli-usage/SKILL.md) -- 推理用量(token / 请求数), 跟 billing 互补
- [arkcli-pricing](../arkcli-pricing/SKILL.md) -- 模型 / 套餐价格目录(catalog), 跟 billing(实际花费)互补
- [arkcli-shared](../arkcli-shared/SKILL.md) -- 认证和全局参数
