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

## 到期时间 / 还剩多少天

这是订阅有效期问题，不是 Usage 的额度刷新时间。标准 `plans get` JSON **没有到期字段**，不能从 `status`、`tier`、购买月份或 `reset_at` 推算。

1. 用户问到期时间时，可用 `arkcli --debug plans get --format json` 取证；但 debug 的团队响应可能含席位 Key，**必须先把 stdout/stderr 分别捕获到权限受限的临时文件，只向工具返回/对话输出目标时间和必要标识**，不能直接打印原始 debug。`--debug` 的上游响应在 stderr，不是给 stdout 增加字段，不要用 `2>&1` 混成 JSON。当前环境无法安全捕获并脱敏时不打开 debug，转 API Explorer 核对已注册的订阅读接口及字段投影，或说明缺少安全取证条件。
2. 个人版从对应 `ListSubscribeTrade` 响应的 `Result.InfoList` 定位订阅，读取 `EndTime`；团队版从对应 `GetSeatInfo` 的 `Result` 读取 `SeatID` 与 `ExpiredTime`。区分 Agent/Coding、个人/团队、订阅/席位，不能把四路响应里的首个时间当作目标套餐到期时间。已有团队管理清单时可复用目标席位的 `expired_time`，不为查个人到期时间扩大到全团队。
3. 上述字段是 Unix **秒**，可能以数字或数字字符串返回。转成北京时间 `Asia/Shanghai`（UTC+8）后交付完整日期和时间，并标明时区；不要把 UTC 原值直接标成北京时间，也不要误按毫秒转换。若上游已给带时区的 ISO 时间，按其原始时区解析后转换，不重复加 8 小时。
4. “还剩几天”按本次查询时间与到期时间计算，注明查询时间及取整口径；已经到期明确说已到期。缺失、0、解析失败、目标记录不唯一或该路 API 失败时，只能说到期时间未确认，不能报 1970 年、永久有效或用别的套餐时间补齐。

最终只给套餐/席位标识、到期时间、时区及必要状态；不回显 debug 原文、API Key 或其他身份信息。诊断只读失败时保留脱敏错误，遵循当前宿主的鉴权恢复规则，不自动购买、续费或切换身份。

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
