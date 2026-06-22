# usage plan-details

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **跟 `usage plan` / `usage stats` 的关系（务必先看）：**
> - `usage plan` — AgentPlan / CodingPlan **额度快照**(5h / weekly / monthly 用了多少 / 还剩多少 / 几号刷新)
> - `usage plan-details` — **AgentPlan 个人版**按模型 / Harness 的**时间序列调用明细**(每个 ObjectName 在每个时间桶的 token / 张数,以及套餐内 vs 套餐外计费)
> - `usage stats` — 按 token 计费的 inference 用量(任意身份,跟订阅无关)
>
> 用户问「我哪个模型用得最多 / 套餐内用了多少 / 套餐外又用了多少」走 plan-details;问"还剩多少额度"走 plan。

查询 AgentPlan 套餐的 Model / Harness 调用明细时间序列。**支持个人版 (默认) + 团队版**(`--product=agent-plan-team`)。CodingPlan 后端没暴露按模型的明细 API,传 `coding-plan` 类 product 直接 ErrValidation。

## 命令

```bash
# 个人版,默认近 7 天 Day 粒度
arkcli usage plan-details

# 指定时间范围
arkcli usage plan-details --start 2026-05-01 --end 2026-05-31

# Hour 粒度 (单次范围 ≤ 31 天约束不变)
arkcli usage plan-details --start 2026-06-01 --end 2026-06-04 --interval Hour

# 过滤到特定模型 (--model 可重复)
arkcli usage plan-details --model doubao-1-5-pro

# === 团队版 ===
# 自动找 caller 绑定的 seat,跟 usage plan team 一样
arkcli usage plan-details --product agent-plan-team

# 显式指定一个或多个 seat (admin 视角看其他 seat,可重复)
arkcli usage plan-details --product agent-plan-team --seat seat-001
arkcli usage plan-details --product agent-plan-team --seat seat-001 --seat seat-002

# 全队聚合 (admin 全景视图,自动 ListSeatInfos 列全队后批量查)
arkcli usage plan-details --product agent-plan-team --all-seats

# Admin server-side 按子用户过滤 (后端 Filter,跟 --all-seats 配套用最方便)
arkcli usage plan-details --product agent-plan-team --all-seats --seat-user-name alice
arkcli usage plan-details --product agent-plan-team --all-seats --seat-user-id 12345
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--start` | **是** | string | 开始日期(`YYYY-MM-DD`)。后端约束: 单次范围 ≤ 31 天 / 历史 ≤ 6 个月 |
| `--end` | 否 | string | 结束日期(`YYYY-MM-DD`,闭区间)。默认今天 |
| `--interval` | 否 | string | 查询粒度: `Day`(默认) / `Hour`。大小写不敏感 |
| `--model` | 否 | []string | 对客模型 / Harness 名称,可重复 |
| `--product` | 否 | string | `agent-plan`(默认,个人版,OpenAPI)\| `agent-plan-team`(团队版,**OpenAPI**,AK/SK + SSO 都行) |
| `--seat` | 否 | []string | **仅团队版**;SeatID 可重复(单次最多 1000)。不传时默认 `GetSeatInfo` 找 caller 单 seat |
| `--all-seats` | 否 | bool | **仅团队版,admin 视角**;先 `ListSeatInfos` 列全队 SeatIDs 再批量查。`--seat` 显式给值时优先,`--all-seats` 被覆盖 |
| `--seat-user-id` | 否 | []string | **仅团队版,admin 视角**;server-side 按子用户 ID 过滤(可重复) |
| `--seat-user-name` | 否 | []string | **仅团队版,admin 视角**;server-side 按子用户名过滤(可重复) |

### 后端约束(client-side 校验)

- **单次范围 ≤ 31 天**(`end - start`):超出立即 ErrValidation 不发请求
- **历史窗口 ≤ 6 个月**(`end` 离 now 不能超过 ~180 天):超出立即 ErrValidation
- **end ≥ start**:反向时间范围立即 ErrValidation
- 时间格式必须 `YYYY-MM-DD`,其它格式立即 ErrValidation

## 返回值

JSON 形态:

```json
{
  "details": [
    {
      "time": 1717000000000,
      "object_name": "doubao-1-5-pro",
      "usage": 1024,
      "unit": "Tokens",
      "billing_type": "WithinPlan"
    },
    {
      "time": 1717000000000,
      "object_name": "doubao-1-5-pro",
      "usage": 256,
      "unit": "Tokens",
      "billing_type": "OutsideOfPlan",
      "seat_id": "seat-001"
    }
  ]
}
```

### 字段说明

| 字段 | 说明 |
|---|---|
| `time` | 时间桶起点,**epoch ms**(后端 IDL 标 ms,跟 `usage plan` 的 CodingPlan ResetTimestamp 不同,这里不需要换算) |
| `object_name` | 对客展示的模型 / Harness 名称(如 `doubao-1-5-pro`) |
| `usage` | 该 (time, object_name, billing_type) 三元组的用量。整数,单位看 `unit` |
| `unit` | `Tokens` / `Images` / 其它(后端按模型类型决定) |
| `billing_type` | `WithinPlan`(套餐内,免费配额消耗)/ `OutsideOfPlan`(套餐外,走基础按 token 计费) |
| `seat_id` | **仅团队版**填充(`SeatUsageDetails[]` 展平时附带);个人版不出此字段 |

> **聚合提示**: 同一时间点可能 `WithinPlan` + `OutsideOfPlan` 两行并列。对接时:
> - 想看「这个模型今天总共调用了多少 token」→ 同 object_name 行 `usage` 累加
> - 想看「套餐内 vs 套餐外的用量比」→ 按 `billing_type` 分组合计

## V1 限制

- **仅 AgentPlan 个人版**(`profile.Type=agent-plan` 或当前身份名下有 AgentPlan personal 订阅时数据非空)。AgentPlan 团队版(席位维度明细)对应后端 endpoint `GetSeatUsageDetails`,V1 未实装
- **不依赖 `ListAgentPlanModelMappingMeta`**: 控制台用它做"模型下拉框 + 'all models' 时塞全量 ObjectName 列表",CLI V1 简化掉:不传 `--model` 就让后端返聚合行,要拆模型自己传 `--model`
- **没有 PlanType filter**: 后端支持按套餐档位过滤(small/medium/large/max),CLI V1 没暴露;通常按身份当前生效订阅自动返,用户没有跨档位查询的必要
- **CodingPlan 不支持**: CodingPlan 后端没有按模型的调用明细 API(其 `GetCodingPlanUsage` 只返周期 `Percent`),这条命令对 CodingPlan profile 也只能查 AgentPlan,**故不开 `--product` flag**

## AI Agent 决策路径

- 用户问「我哪个模型用得最多 / 套餐内套餐外各占多少」→ `arkcli usage plan-details`(默认近 7 天聚合)
- 用户问「上周哪天 token 消耗最大」→ `--start 2026-05-26 --end 2026-06-01`,output 按 `time` 排序
- 用户问「按小时看 5/15 当天的调用」→ `--start 2026-05-15 --end 2026-05-15 --interval Hour`
- 用户问「`doubao-1-5-pro` 的趋势」→ `--model doubao-1-5-pro`(单模型时间序列,套餐内+套餐外两行/桶)
- 用户问「我的套餐用了百分之几 / 几号刷新」→ 这条命令是**逐次累加视角**,看 quota 快照走 [`usage plan`](arkcli-usage-plan.md)
- 用户问「按 token 计费的 inference 用量,不是套餐」→ 转 [`usage stats`](arkcli-usage-stats.md)

## 错误降级

- **未订阅 AgentPlan**: 后端返 Result=nil,CLI 输出 `{"details":[]}`(空数组而非 null)。要确认是不是订阅问题,先跑 `arkcli usage plan --product=agent-plan` 看 `subscribed` 字段
- **AccessDenied**: 通常是 sub-user 没 AgentPlan 权限,转 [arkcli-auth](../../arkcli-auth/SKILL.md) 排查身份
- **范围错误**: 看 ErrValidation 提示(31 天限 / 6 月历史窗口 / YYYY-MM-DD 格式)直接修 flag

## 调用面对齐

- 个人版调 OpenAPI **`/open/GetUsageDetails`**(AK/SK + SSO 都支持);arkcli 用公网 OpenTOP 不带 AgentPlan 前缀的入口
- 团队版调 OpenAPI **`/open/GetSeatUsageDetails`** + **`/open/GetSeatInfo`**(找 caller seat),都是公网入口,AK/SK + SSO 都行
- 默认窗口 `今天 - 6 天 ~ 今天`
- 后端 31 天 / 6 月窗口约束 client-side 前置校验,不发请求,把错误暴露在 CLI 层
