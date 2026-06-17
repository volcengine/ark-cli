# plans get

> **前置条件：** 先阅读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **范围限制：** `plans get` 只回答"我**持有**哪些套餐 + 状态(Effective/Running/...)"，**不回答"用了多少 / 还剩多少 / 几号刷新 / 按模型拆分"**。配额查询走：
> - `arkcli usage plan` —— 完整版，含 used / total / percent / reset_at + 多周期(5h/daily/weekly/monthly)
> - `arkcli usage balance --type plan` —— 精简版（同数据源）
> - `arkcli usage plan-details` —— AgentPlan 按模型 × 套餐内/外的时间序列
>
> 详见 [`../../arkcli-usage/SKILL.md`](../../arkcli-usage/SKILL.md)。

一次性返回当前账号下**实际持有**的套餐：Agent Plan / Coding Plan 的个人版 + 团队版聚合。读操作，无副作用。

## 命令

```bash
arkcli plans get
```

## 参数

零参数。所有 4 路 plan（agent-plan / agent-plan-team / coding-plan / coding-plan-team）一次并发拉取。

## 返回值

```json
{
  "plans": [
    {
      "key": "agent-plan",
      "name": "Agent Plan",
      "scope": "personal",
      "tier": "small",
      "status": "Effective"
    },
    {
      "key": "agent-plan-team",
      "name": "Agent Plan Team",
      "scope": "team",
      "tier": "medium",
      "status": "Running",
      "seat_id": "seat-..."
    }
  ]
}
```

| 字段 | 出现条件 | 含义 |
|------|---------|------|
| `key` | always | 稳定标识：agent-plan / agent-plan-team / coding-plan / coding-plan-team |
| `scope` | always | personal / team |
| `tier` | 持有时 | 个人版：small/medium/large/max（agent）或 lite/pro（coding）；团队版：当前用户席位的 tier |
| `status` | 持有时 | 个人版：Effective / Pending / Expired 等；团队版：Running |
| `seat_id` | 团队版独有 | 当前用户在该 Scene 下的席位 ID（团队版才有） |
| `error` | 调用失败时 | API 透出诊断信息；跟 tier/status 互斥 |

**没持有的套餐 / Reclaimed 历史订阅 / 团队版无席位** 会被过滤掉，不出现在数组里。

## 行为细节

- 内部并发拉 `ListSubscribeTrade`（个人版）+ `GetSeatInfo`（团队版）
- 任一 plan 失败不阻塞其它 plan：失败项保留在数组里，带 `error` 字段
- 团队版仅识别 `BillingStatus=Running` 的有效席位，历史席位（Pending/Expired/Reclaimed）会被过滤

## 常见错误

| 错误 | 原因 | 处理 |
|------|------|------|
| `plans get: ark: API error: AuthFailure` | 未认证 / token 过期 | `arkcli auth login volc-sso` |
| 输出 `{"plans":[]}` | 当前账号没有任何套餐 | 用 `plans buy` 下单或检查账号是否切换 |

## 注意事项

- 这是**唯一一个无需任何参数**的 plans 子命令；用户问 "我有什么套餐 / 我订阅了什么" 直接调用即可
- **想看用量 / 配额（used / total / percent / 几号刷新）**：本命令**不出**这些字段，转 `arkcli usage plan` / `arkcli usage balance --type plan` / `arkcli usage plan-details`（详见顶部范围限制说明）
- 不会列出别人的套餐 — 永远是当前 SSO 身份名下的视图
- 团队版只展示当前用户**自己持有的席位**；要看团队全部席位走 [`plans team seat-list`](arkcli-plans-team-seat-list.md)（基础信息）或 `arkcli usage seats --with-usage`（含席位用量）

## 参考

- [arkcli-plans](../SKILL.md) -- skill 概览
- [`plans model-list`](arkcli-plans-model-list.md) -- 看持有套餐能调哪些模型
- [`plans buy`](arkcli-plans-buy.md) / [`plans renew`](arkcli-plans-renew.md) -- 下单 / 续费
- [arkcli-usage](../../arkcli-usage/SKILL.md) -- 套餐用量 / 配额（本命令不含的字段都在这里）
- [arkcli-shared](../../arkcli-shared/SKILL.md)
