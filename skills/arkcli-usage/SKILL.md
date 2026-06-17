---
name: arkcli-usage
version: 1.2.0
description: "ARK 用量查询:`usage stats`(Token / 请求数,5-30 分钟延迟)、`usage plan` / `usage balance --type plan`(套餐额度快照)、`usage plan-details`(按模型时间序列,套餐内/外拆分)、`usage balance`(余额:免费额度 / 媒资库 / 套餐)、`usage seats --with-usage`(团队席位用量 by seat)。命中关键词:用量 / 用了多少 / 还剩多少额度 / 套餐用了几成 / 套餐内套餐外 / 每个 seat 消耗。**席位的列表 / 绑定 / 分配 / 轮换 APIKey 属管理范畴,走 arkcli-plans;本 skill 只回答用量。**动词路由:**用 / 消耗 / 多少** → 这里;**列 / 绑 / 分 / 轮换** → arkcli-plans。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli usage stats --help"
---

# arkcli usage

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md),其中包含认证闸门、配置排查与共享安全规则**
**CRITICAL — `usage stats` 在执行之前,务必先用 Read 工具读取 [`references/arkcli-usage-stats.md`](references/arkcli-usage-stats.md),禁止直接盲目调用命令。**
**CRITICAL — `usage plan` 在执行之前,务必先用 Read 工具读取 [`references/arkcli-usage-plan.md`](references/arkcli-usage-plan.md),禁止直接盲目调用命令。**
**CRITICAL — `usage plan-details` 在执行之前,务必先用 Read 工具读取 [`references/arkcli-usage-plan-details.md`](references/arkcli-usage-plan-details.md),禁止直接盲目调用命令。**
**CRITICAL — `usage balance` 在执行之前,务必先用 Read 工具读取 [`references/arkcli-usage-balance.md`](references/arkcli-usage-balance.md),禁止直接盲目调用命令。**
**CRITICAL — `usage seats` 在执行之前,务必先用 Read 工具读取 [`references/arkcli-usage-seats.md`](references/arkcli-usage-seats.md),禁止直接盲目调用命令。**

> **数据时效(务必先看):`usage stats` 走的是上游 BFF 的聚合管道,数据有 5–30 分钟延迟,定位为「日级 / 聚合分析」。** 不要把它用于实时预算监控、限流、实时告警等需要秒级精度的控制场景——这类场景请直接读 `+chat` / `+gen` 等推理命令返回里的 `.usage` 字段(per-request、零延迟,详见 [arkcli-chat](../arkcli-chat/SKILL.md) 与 [arkcli-gen](../arkcli-gen/SKILL.md))。

> **stats / plan / plan-details / balance 区分:**
> - `stats` 出"按 token 计费的 inference 用量"(任何身份,5–30 分钟延迟)
> - `plan` 出"订阅套餐的 quota 快照"(AgentPlan / CodingPlan 个人版 + 团队版的"我用了几成 / 几号刷新")
> - `plan-details` 出"AgentPlan 按模型 / 套餐内外的时间序列调用明细"(AgentPlan 个人版 + 团队版,CodingPlan 没有按模型 endpoint)
> - `balance` 出"还剩多少 X"(余额视角,`--type free-quota / media-asset / plan` 三选一)
> - `seats` 出"团队席位**用量** admin 视角"(每个 seat 用了多少 / 套餐百分比;**只是用量视图**)。**席位的列表 / 绑定 / 分配 / API Key 轮换** 走 [`arkcli-plans`](../arkcli-plans/SKILL.md) 的 `plans team seat-list / seat-assign / rotate-apikey`。
>
> 用户问「我的套餐还剩多少额度」→ `plan` 或 `balance --type plan`(后者输出更精简);「我哪个模型用得最多」→ `plan-details`;「我今天用了多少 token」→ `stats`;「我还剩多少免费额度 / 媒资库容量」→ `balance`;「**每个 seat 用了多少 / 团队席位用量分布**」→ `seats --with-usage`(纯管理类:谁绑了哪个 seat / 列出席位 / 分配席位 / 轮换 key → `arkcli-plans`)。

## 适用场景

- 查看当日或指定日期范围的推理用量(**接受 5–30 分钟延迟**)
- 看订阅套餐(AgentPlan / CodingPlan)的配额消耗 / 重置时间
- 看 AgentPlan 套餐内 vs 套餐外的按模型拆分明细
- 按模型、接入点、API Key 维度分组统计(stats)
- 分析 Token 消耗趋势、出对账报表

## 业务定位

- 本 skill 对应"看存量调用结果和消耗"的场景,是**离线 / 准实时**分析入口
- 它通常发生在两类后置链路之后:
  - Endpoint 已经存在,想按 endpoint 看使用情况
  - 模型已经在被调用,想按 model / API Key 分析消耗
  - 或:订阅了套餐,想看额度还剩多少
- 它不是创建 Endpoint 的入口,也不是试用模型的入口
- **它不是实时预算 / 限流 / 告警的数据源**——实时控制请走 per-request `.usage` 字段(见上方 banner)

## 快速决策

- 用户问"我的套餐还剩多少 / 用了几成 / 几号刷新 / 我有几个套餐":
  - `arkcli usage plan`(默认探所有 4 个 SKU:personal × 2 + team × 2,subscribed 桶才下发 usage 请求)
  - `arkcli usage plan --all`:不探订阅,4 桶全跑(诊断 / Excel)
  - `arkcli usage plan --product=X`:跳过探测,单查
  - **跟 `balance --type plan` 是同一份数据**;`usage plan` 输出完整(带 `subscribed` / `updated_at`),`balance --type plan` 精简(去掉元字段)。一般默认走 `usage plan`,只有用户明确把套餐余额跟"免费额度 / 媒资库容量"一起对账时,才走 `balance --type plan`(三 type 一起跑节省心智)
- 用户问"我哪个模型用得最多 / 套餐内套餐外各占多少 / 团队某个子用户用得多":
  - `arkcli usage plan-details`(默认近 7 天 Day 粒度,AgentPlan personal)
  - `arkcli usage plan-details --product=agent-plan-team`(团队版,自动找 caller seat 或显式 `--seat <id>`)
- 用户问"我的用量 / 我今天用了多少 token / 我的 endpoint 消耗":
  - **火山:** 先 `arkcli usage stats --mine`（默认 endpoint 维度）
    - 返回 `data_count > 0` → 直接从 `totals` 取合计
    - 返回 `data_count=0`（用户没有 Endpoint 或近期未通过 Endpoint 调用）→ 立即重试 `arkcli usage stats --mine --mine-by=apikey`
    - **⛔ 禁止退化为不带 `--mine` 的全量查询** —— 全账号数据（data_count 极大）≠ "我的用量"；必须始终保持 `--mine` 范围
- 用户问"我每个 APIKey 用了多少":
  - **火山:** `arkcli usage stats --mine --mine-by=apikey`,从 `records[].AuthToken` 拆分
- 用户问"我账上还有多少免费 token / 模型免费额度 / 媒资库还能存多少":
  - `arkcli usage balance --type free-quota`(模型免费推理额度)
  - `arkcli usage balance --type media-asset`(素材库容量)
- 用户问"团队席位**用了多少** / 每个 seat 消耗 / 套餐百分比 by seat"(admin 用量视角):
  - `arkcli usage seats --product agent-plan-team --with-usage`(列席位 + 每个 seat 的 AFP 用量)
  - 不带 `--with-usage` 也行,但只有基础元数据,没有用量数字
  - **路由分工**:本命令**只**回答"用量",不负责 seat 管理。"**列席位 / 谁绑了哪个 seat / 给员工分配席位 / 轮换席位 APIKey**" 全部走 [`arkcli-plans`](../arkcli-plans/SKILL.md) 的 `plans team seat-list / seat-assign / rotate-apikey`。Agent 路由这条问句先看动词:**"用了 / 消耗 / 多少"** 留这边,**"绑 / 列 / 分 / 轮换"** 跳过去 plans。
- 数据看起来不对:先查认证和 profile,再怀疑数据本身

## Agent 快速执行顺序

1. 先确认认证状态;不确定时先看 `arkcli auth status`
2. 先做最小查询,再逐步加时间范围和分组条件
3. 结果异常时,先检查 profile / region / API Key 来源是否正确,再怀疑数据本身
4. 如果用户要按 endpoint 看 usage,最好先确认 endpoint-id 已知

## 常见降级

- 鉴权错误:转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
- 查询范围或环境不对:转 [`../arkcli-config/SKILL.md`](../arkcli-config/SKILL.md)
- 用户其实还没有 Endpoint,只是在准备正式接入:转 [`../arkcli-deploy/SKILL.md`](../arkcli-deploy/SKILL.md)
- 用户想看套餐**价格**(不是用量):转 [`../arkcli-pricing/SKILL.md`](../arkcli-pricing/SKILL.md)(`pricing plans` 出 catalog 询价,`usage plan` 出 quota 快照,两者是不同视角)
- 用户想看**结算金额 / 账单 / 花了多少钱**(不是 token 用量):转 [`../arkcli-billing/SKILL.md`](../arkcli-billing/SKILL.md)(数据源是火山计费中心, T+1 出账; usage stats 是 BFF 推理聚合, 近实时)

## 命令一览

| 命令 | 说明 |
|------|------|
| [`usage stats`](references/arkcli-usage-stats.md) | 查询推理用量(Token / 请求数,按日 / 小时聚合,按 token 计费的 inference 视角) |
| [`usage plan`](references/arkcli-usage-plan.md) | 查询订阅套餐的额度快照(AgentPlan / CodingPlan,5h / weekly / monthly / session 周期的"用了多少 / 还剩多少 / 几号刷新")。默认探所有 4 个 SKU |
| [`usage plan-details`](references/arkcli-usage-plan-details.md) | 查询 AgentPlan **个人版 + 团队版** 按模型 / Harness 的时间序列调用明细(套餐内 vs 套餐外拆分)。默认近 7 天,`--product=agent-plan-team` 切团队版 |
| [`usage balance`](references/arkcli-usage-balance.md) | 余额视角统一入口 `--type free-quota / media-asset / plan`(模型免费推理额度 / 素材库容量 / 套餐余额) |
| [`usage seats`](references/arkcli-usage-seats.md) | **Admin 视角** 列举企业版套餐下所有席位(SeatID / 绑定子用户 / billing 状态)。子用户调通常 AccessDenied |

## 自然语言触发词表

| 用户怎么说 | 对应命令 |
|---|---|
| "按模型明细 / 哪个模型用得最多 / 模型维度拆分 / 按模型时序 / 套餐内套餐外按模型" | `arkcli usage plan-details` |
| "coding plan 按模型明细 / coding plan 套餐内套餐外" | `arkcli usage plan-details --product=coding-plan-personal` |
| "配额还剩多少 / 余额 / 还能用多少 / 用了几成 / 套餐额度" | `arkcli usage plan` 或 `arkcli usage balance --type plan` |
| "免费额度还有多少 / 模型免费额度 / 免费 token" | `arkcli usage balance --type free-quota` |
| "AFP 消耗 / 今日 AFP / 本周 AFP" | `arkcli usage plan`（AFP 就是套餐额度单位） |
| "我的用量 / 我今天用了多少 token / 推理用量 / token 消耗" | `arkcli usage stats --mine` |

## 参考

- [arkcli-deploy](../arkcli-deploy/SKILL.md) -- 先创建 Endpoint,再按 endpoint 看 usage
- [arkcli-pricing](../arkcli-pricing/SKILL.md) -- catalog 询价(`pricing plans`),跟 `usage plan` quota 快照互补
- [arkcli-shared](../arkcli-shared/SKILL.md) -- 认证和全局参数
