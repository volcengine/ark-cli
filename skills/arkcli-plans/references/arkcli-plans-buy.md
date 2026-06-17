# plans buy

> **前置条件：** 先阅读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **⚠️ 写操作 + 计费：** `plans buy` 强制 `IsAutoPay=true`，加 `--yes` 调用成功即真实扣款。**不要替用户做选择**：`--plan`、`--type`、`--duration`、团队版 `--quantity` 必须由用户明确给出。

> **🔒 协议闸门 (强制流程，agent 必须遵守)：** 不传 `--yes` 时,本命令**不下单**，只返协议清单。Agent **必须**:
> 1. 拿到 `agreements` 数组后,把每条协议的 `title` 和 `url` **逐条**展示给最终用户
> 2. 等待用户**明确表达**"已阅读并同意全部协议"(不能用模糊措辞默认同意)
> 3. 用户同意后才能加 `--yes` 重跑 `next_step` 字段里的命令
>
> **禁止**在用户没看过协议的情况下直接补 `--yes`。这是合规要求,跟 frontend 购买面板的 "我已阅读并同意《...》" checkbox 等价。

## 三种调用形态

| 形态 | 行为 | 何时用 |
|---|---|---|
| 不传 `--yes` 不传 `--dry-run` | **协议闸门**:返协议清单 + **价格** + `next_step`,内部走 EstimatePrice,**不下单** | 默认引导步骤;agent 走这个一次拿"协议 + 价格" |
| `--dry-run` (无论是否带 `--yes`) | **询价**:调 EstimatePrice,返价格,**不下单** | 跟闸门重叠,留作脚本"只询价不看协议"的快捷出口 |
| `--yes` 不带 `--dry-run` | **真下单**:`IsAutoPay=true` 自动扣款 | 用户已确认协议 + 价格,真下单 |

> **闸门里已经包含价格** — 不需要先 `--dry-run` 再看协议。一次调用同时拿到 `agreements` + `total_amount_cny` + `original_amount_cny`,展示给用户后直接走 `next_step`。

## 命令

```bash
# 第 1 步: 引导 (不传 --yes) — 拿到协议清单
arkcli plans buy --plan agent-plan --type small --duration 1

# 第 2 步 (可选): 询价
arkcli plans buy --plan agent-plan --type small --duration 1 --dry-run

# 第 3 步: 用户阅读并同意全部协议后, 真下单
arkcli plans buy --plan agent-plan --type small --duration 1 --yes

# 团队版各档位示例
arkcli plans buy --plan agent-plan-team --type medium --duration 2 --quantity 5 --yes
arkcli plans buy --plan coding-plan-team --type pro --quantity 10 --yes
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--plan` | 是 | string | `agent-plan` / `coding-plan` / `agent-plan-team` / `coding-plan-team` |
| `--type` | 是 | string | tier 档位：agent-plan* 取 `small/medium/large/max`；coding-plan* 取 `lite/pro` |
| `--duration` | 否 | int | 订阅时长（月），1-12，默认 1 |
| `--quantity` | 团队版 必填 | int | 席位数量（≥1）；个人版传了会被忽略 |
| `--yes` | 否 | bool | 跳过协议闸门，真下单 (`IsAutoPay=true`) |

`--plan` 与 `--type` 兼容性：

| Plan 系 | 合法 tier |
|---|---|
| agent-plan / agent-plan-team | small, medium, large, max |
| coding-plan / coding-plan-team | lite, pro |

## 返回值 (协议闸门, 默认形态)

不传 `--yes` 不传 `--dry-run` 时:

```json
{
  "status": "agreement_required",
  "plan": "agent-plan",
  "tier": "small",
  "duration": 1,
  "echo": {"plan": "agent-plan", "tier": "small", "duration": 1},
  "agreements": [
    {"title": "火山引擎数据授权协议", "url": "https://www.volcengine.com/docs/82379/1928265"},
    {"title": "方舟平台专用条款", "url": "https://www.volcengine.com/docs/82379/1104498"},
    {"title": "免责声明", "url": "https://www.volcengine.com/docs/82379/1108564"},
    {"title": "豆包模型服务协议", "url": "https://www.volcengine.com/docs/82379/1142195"},
    {"title": "开源模型许可证", "url": "https://www.volcengine.com/docs/82379/1454060"},
    {"title": "语音模型服务协议", "url": "https://www.volcengine.com/docs/6561/1866421?lang=zh"},
    {"title": "Harness权益说明和产品专用条款", "url": "https://www.volcengine.com/docs/82379/2516291?lang=zh"}
  ],
  "total_amount_cny": 99.5,
  "original_amount_cny": 100.0,
  "confirm_text": "下单将立即扣款 (IsAutoPay=true)。请把上述协议链接展示给最终用户...",
  "next_step": "arkcli plans buy --plan agent-plan --type small --duration 1 --yes"
}
```

**闸门同时返回价格**:`total_amount_cny`(实付)+ `original_amount_cny`(原价,有折扣时低于原价)。Agent 应该把价格 + 协议一起展示给用户。

**`agreements` 数组按 plan 类型不同**(agent-plan 系含 Harness 协议;团队版多一条对应套餐专用条款,不带《数据授权协议》)。**直接用返回值里的清单**,不要替换或省略。

## 返回值 (真下单)

加 `--yes` 后:

```json
{
  "status": "success",
  "plan": "agent-plan-team",
  "tier": "medium",
  "duration": 2,
  "quantity": 5,
  "order_number": "ORD-...",
  "instance_id": "ins-...",
  "seat_ids": ["seat-...", "seat-..."]
}
```

`seat_ids` 仅团队版填充。

## 失败语义

**两类失败要分清：**

1. **transport error**（订单**没**落库）：返回普通 error，exit code = 1。**没扣钱**。
2. **payment_failed**（订单**已**落库但支付未完成）：典型如余额不足。返回 `output.ErrWithHint` envelope：
   ```json
   {
     "ok": false,
     "error": {
       "type": "payment_failed",
       "message": "INSUFFICIENT_BALANCE_ERROR",
       "hint": "Order ID: ORD-...; please complete the payment manually in console"
     }
   }
   ```
   exit code = 5（`ExitAPI`）。**用户必须去 console 手动补单或取消订单**，不要在 CLI 重试 `plans buy`（会重复下单）。

   团队版 dangling：第一步 `CreateSeatInfo` 成功 + 第二步 trade 失败 → hint 里会带 `(Created seats: seat-001,seat-002; bound to the unpaid order above.)`，**席位资源已生成**，不会自动清理。

## 常见错误

| 错误 | 原因 | 处理 |
|------|------|------|
| `--plan must be one of ...` | 拼错 plan key | 严格用 4 选 1 |
| `--type %q not allowed for --plan %q` | tier 跟 plan 不兼容 | 见上方表格 |
| `--quantity is required for team plans` | 团队版漏传 | 加 `--quantity N` |
| `--duration must be 1..12` | 越界 | 1-12 之间 |
| `payment_failed` envelope | 余额不足 / auto-pay 失败 | 去 console 补单，不要 CLI 重试 |

## 注意事项

- **真实扣款**：CLI 没有 "下单不支付" 模式；要预演用 `--dry-run`（若支持）或换轻档位 + 1 月先试
- 团队版下单后席位会以 `BillingStatus=Running` 状态出现，但**没绑定子用户**；后续走 [`plans team seat-assign`](arkcli-plans-team-seat-assign.md) 分配
- 重试要谨慎：`payment_failed` 的订单**已经在服务端**，重跑会再开一单

## 参考

- [arkcli-plans](../SKILL.md) -- skill 概览
- [`plans renew`](arkcli-plans-renew.md) -- 续费已有套餐
- [`plans get`](arkcli-plans-get.md) -- 下单后查持有状态
- [`plans team seat-assign`](arkcli-plans-team-seat-assign.md) -- 团队版下单后绑定子用户
- [arkcli-shared](../../arkcli-shared/SKILL.md)
