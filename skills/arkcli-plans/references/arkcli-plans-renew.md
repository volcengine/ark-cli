# plans renew

> **前置条件：** 先阅读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **⚠️ 写操作 + 计费：** 跟 `plans buy` 同样 `IsAutoPay=true`，加 `--yes` 后自动扣款。失败语义（payment_failed envelope）也跟 buy 一致。

> **🔒 协议闸门 (强制流程,跟 `plans buy` 完全同套路)：** 不传 `--yes` 时,本命令**不续费**,只返协议清单。Agent **必须**先把 `agreements` 数组逐条展示给用户,等用户**明确同意**全部协议后才能加 `--yes` 重跑 `next_step`。详见 [`plans buy` 协议闸门说明](arkcli-plans-buy.md)。续费走的是跟新购完全一样的协议清单 (frontend BuyWindow 同时承载 buy + upgrade)。

续费已有套餐。个人版自动从 `ListSubscribeTrade` 反查 InstanceID；团队版按 `--seat-ids` 拼 `Items` 一次走 `RenewAgentPlanEnterpriseTrade`。

## 选择续费对象

- 复用本轮持有列表；只有一个与用户目标相符的套餐时直接复述并沿用，不再要求用户选同一个对象。多个套餐/席位仍可能匹配时，让用户选目标，不默认第一条。
- 选项应带套餐家族、个人/团队、当前档位、真实席位 ID（团队版）与已确认的到期时间；到期时间按 [get 的有效期规则](arkcli-plans-get.md#到期时间--还剩多少天) 转北京时间，未查到就标“到期时间未知”，不伪造。
- 只补缺少的续费月数或目标席位，每轮一个决策。1 / 3 / 6 / 12 个月可作快捷选项，但 CLI 合法范围是 1–12，不能拒绝用户明确给出的 2、4 等合法月份。
- 保持原档位；不询问新购档位，不传 `--type` / `--quantity`。参数完整也不能跳过下述本轮协议与价格披露。

## 三种调用形态

跟 `plans buy` 一致:

| 形态 | 行为 |
|---|---|
| 不传 `--yes` 不传 `--estimate` | 协议闸门,返协议清单 + **价格** + `next_step`,内部走 EstimatePrice (IsRenew=true / Scene=RENEW),不下单 |
| `--estimate` | 在线询价，不带协议清单 (脚本"只询价"出口) |
| `--yes` | 真续费,自动扣款 |

> 闸门返回值会**反查 tier**填回:个人版从 `ListSubscribeTrade.InfoList[0].BizInfo` 来,团队版从 `ListSeatInfos` 用 SeatIDs[0] 反查。`echo.tier` 字段供 agent 跟用户确认"还是续这个档位吗"。

## 命令

```bash
# 第 1 步: 引导 — 拿到协议清单
arkcli plans renew --plan agent-plan
arkcli plans renew --plan agent-plan-team --duration 3 --seat-ids seat-001,seat-002

# 第 2 步: 询价 (可选)
arkcli plans renew --plan agent-plan --duration 6 --estimate

# 第 3 步: 用户同意协议后, 真续费
arkcli plans renew --plan agent-plan --yes
arkcli plans renew --plan coding-plan --duration 6 --yes
arkcli plans renew --plan agent-plan-team --duration 3 --seat-ids seat-001,seat-002 --yes
arkcli plans renew --plan coding-plan-team --seat-ids seat-aaa --yes
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--plan` | 是 | string | `agent-plan` / `coding-plan` / `agent-plan-team` / `coding-plan-team` |
| `--duration` | 否 | int | 续费时长（月），1-12，默认 1 |
| `--seat-ids` | 团队版 必填 | string | 逗号分隔，要续费的席位 ID（个人版**不允许**传） |
| `--yes` | 否 | bool | 跳过协议闸门，真续费 (`IsAutoPay=true`) |

> **关键差异**：跟 `plans buy` 不同，`renew` 不需要 `--type` —— 续费保持原档位。也不需要 `--quantity` —— 续费数量由 `--seat-ids` 个数决定。

## 返回值 (协议闸门, 默认形态)

不传 `--yes` 不传 `--estimate` 时:

```json
{
  "status": "agreement_required",
  "plan": "agent-plan",
  "tier": "small",
  "duration": 1,
  "echo": {"plan": "agent-plan", "tier": "small", "duration": 1},
  "agreements": [
    {"title": "火山引擎数据授权协议", "url": "..."},
    ...
  ],
  "total_amount_cny": 99.5,
  "original_amount_cny": 100.0,
  "confirm_text": "续费将立即扣款 (IsAutoPay=true)。请把上述协议链接展示给最终用户...",
  "next_step": "arkcli plans renew --plan agent-plan --duration 1 --yes"
}
```

`tier` / `echo.tier`:从已有订阅 / 席位反查的当前档位 — agent 拿这个跟用户确认续费档位。

团队版 `next_step` 会带上 `--seat-ids seat-001,seat-002`,`echo.seat_ids` 也会列出 — 直接抄给用户即可。

## 返回值 (真续费)

加 `--yes` 后:

```json
{
  "status": "success",
  "plan": "agent-plan-team",
  "tier": "",
  "duration": 3,
  "quantity": 2,
  "order_number": "<本次真实订单号>",
  "seat_ids": ["<已确认席位1>", "<已确认席位2>"]
}
```

这是结构示例，不是成功回执。个人版可含 `instance_id` 与反查得到的 `tier`；团队版列出本次 `seat_ids`，`tier` 可能为空，不自行补值。订单相关可选字段以实际返回为准；不能把 `agreement_required`、询价结果或存在订单号当成支付成功。

## 失败语义

跟 [`plans buy`](arkcli-plans-buy.md#失败语义) 完全一致：
- transport error → 普通 error；超时/断连后的订单与扣款状态可能未知，先核对，不自动重放续费
- `payment_failed` envelope → 订单已落库但 auto-pay 没完成；hint 带 `Order ID: ORD-...`，**用户去 console 补单**，CLI 不要重试
- `ProductStockNotEnough` → 库存不足；不重试、不重问已确认参数，按 buy 的控制台兜底规则处理，不改成新购

这些失败均保留本轮已经确认的 `--plan` / `--duration` / `--seat-ids`，不重新启动参数 question。用户明确改变目标时才重新收集变化项并重新披露价格和协议。

## 常见错误

| 错误 | 原因 | 处理 |
|------|------|------|
| `--plan must be one of ...` | 拼错 plan key | 严格用 4 选 1 |
| `--seat-ids is required for team plans` | 团队版漏传 | 列出要续费的席位 ID |
| `--seat-ids only applies to team plans` | 个人版多传了 | 删掉 `--seat-ids`，个人版不需要 |
| `--seat-ids must contain at least one non-empty seat ID` | 给了空串 / 全是逗号 | 检查输入 |
| `--duration must be 1..12` | 越界 | 1-12 之间 |
| `payment_failed` envelope | 余额不足 | console 手动补单 |

## 注意事项

- 团队版的 `--seat-ids` 不去重 —— 重复 ID 由服务端按最后一条覆盖
- 个人版续费**不需要预先查 InstanceID** —— service 层会自动 `ListSubscribeTrade` 反查；用户只关心 `--plan`
- 续费时长不会叠加到现有过期时间之外的限制；`--duration` 1-12 是单次调用上限

## 参考

- [arkcli-plans](../SKILL.md) -- skill 概览
- [`plans buy`](arkcli-plans-buy.md) -- 同套 payment_failed 语义
- [`plans get`](arkcli-plans-get.md) -- 看续费后过期时间
- [`plans team seat-list`](arkcli-plans-team-seat-list.md) -- 列出团队版可续费的席位
- [arkcli-shared](../../arkcli-shared/SKILL.md)
