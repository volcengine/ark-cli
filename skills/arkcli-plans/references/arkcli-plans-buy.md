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
| 不传 `--yes` 不传 `--estimate` | **协议闸门**:返协议清单 + **价格** + `next_step`,内部走 EstimatePrice,**不下单** | 默认引导步骤;agent 走这个一次拿"协议 + 价格" |
| `--estimate` (无论是否带 `--yes`) | **在线询价**:调 EstimatePrice,返价格,**不下单** | 跟闸门重叠,留作脚本"只询价不看协议"的快捷出口 |
| `--yes` 不带 `--estimate` | **真下单**:`IsAutoPay=true` 自动扣款 | 用户已确认协议 + 价格,真下单 |

> **闸门里已经包含价格** — 不需要先 `--estimate` 再看协议。一次调用同时拿到 `agreements` + `total_amount_cny` + `original_amount_cny`；展示后等待明确同意，再走 `next_step`。

## 收参与协议披露模板

先读本 reference，再构造第一条命令。家族与个人/团队共同映射为 `--plan`，档位映射为 `--type`，团队席位数映射为 `--quantity`；不能发明 `--product`、`--edition` 或 `--tier`。

一次只问一个尚未确定的决策；已提供的值不重复问。按以下顺序展开，不能在家族未定时混列两家的档位：

1. 家族：Agent Plan（智能体任务）／Coding Plan（AI 编程工具）。
2. 版本与档位：已选家族的个人版合法档位，或一个团队版入口。个人版一次选择确定版本和档位；选团队版后，再单问同家族档位，然后单问席位数，不同时把后续层选项堆进当前问题。
3. 时长：提供 1／3／6／12 个月快捷选项；用户明确给出其他 1–12 的整数仍合法，不能强改为快捷值。
4. 收齐参数后进入协议闸门。续费使用 [renew](arkcli-plans-renew.md) 的短决策链，不重新询问档位或购买数量。

每个档位选项的说明必须包含当前价格、典型场景及关键限制，按下文权益说明填充；团队数量是席位数，不是月份。没有可靠当前权益数据就明确标注未知，不捏造价格或资格。

协议披露使用以下结构，**所有占位符都取本次返回值**，不要复用历史询价或把协议缩写为“相关条款”：

```text
协议确认
购买/续费：<plan> / <tier>，<duration> 个月
数量/目标席位：<quantity 或 seat_ids；个人版注明不适用>
本次实付：¥<total_amount_cny>（有原价时同时列原价）
请阅读以下全部协议：
- [<agreements[0].title>](<agreements[0].url>)
- ... 按返回顺序逐条列出，不截断
明确同意全部协议后，继续执行将按上述实付价格自动扣款。
继续：已阅读并同意全部协议；停止：取消购买/续费。
```

必须实际等待用户回应。拒绝、取消、空答案、未回答、工具拒绝或模糊回应均不放行；不得代答或自行加 `--yes`。宿主支持结构化交互时在其正文中完整展示这些信息；具体 question、授权 Pane / Web Widget 的分工以共享宿主协议为准。工具执行 permission 不等于同意业务协议，协议同意也不等于已取得工具执行授权。价格、套餐或席位变化后须重新披露并确认。

## 命令

```bash
# 第 1 步: 引导 (不传 --yes) — 拿到协议清单
arkcli plans buy --plan agent-plan --type small --duration 1

# 第 2 步 (可选): 询价
arkcli plans buy --plan agent-plan --type small --duration 1 --estimate

# 第 3 步: 用户阅读并同意全部协议后, 真下单
arkcli plans buy --plan agent-plan --type small --duration 1 --yes

# 团队版各档位示例
arkcli plans buy --plan agent-plan-team --type medium --duration 2 --quantity 5 --yes
arkcli plans buy --plan coding-plan-team --type pro --quantity 10 --yes
```

## 参数

### 档位选项的业务说明

不能只让用户在裸 `small/pro` 等名称中盲选。每个档位选项应说明**价格、典型场景、关键限制**；已明确的参数不再问，Agent 与 Coding 的档位不能混在同一组。

价格先复用本轮 `pricing plans` 的对应产品/版本/档位记录，没有时按 [pricing plans](../../arkcli-pricing/references/arkcli-pricing-plans.md) 查询；带 `error` 的记录不是零元档。团队价不能套个人价，多月和多席位最终以本次协议闸门的 `total_amount_cny` 为准。

以下保留 Vaka 1.0.32 业务说明的场景和权益对照，**数值及视频资格是历史核对基线，不是当前权益承诺**：

| 家族 / 个人档位 | 典型场景 | 原业务说明的权益基线（使用前核对） |
| --- | --- | --- |
| Coding / `lite` | 中等强度开发、一般开发任务 | AI 编程工具场景；价格以当前查询为准 |
| Coding / `pro` | 复杂项目、高强度开发 | AI 编程工具场景；价格以当前查询为准 |
| Agent / `small` | 轻量体验 | ¥40/月、20,000 AFP、不支持视频生成 |
| Agent / `medium` | 1–2 个项目并行的参考档位 | ¥200/月、100,000 AFP、支持视频生成 |
| Agent / `large` | 2 个以上项目并行的参考档位 | ¥500/月、250,000 AFP |
| Agent / `max` | 2 个以上项目并行、较高额度需求 | ¥1,000/月、500,000 AFP |

- 场景是选择参考，不是并发数或任务成功保证。团队版沿用同家族档位集合，但计价与权益按当前团队商品核对，不直接套个人 AFP。
- 当前价格用查询返回；AFP、视频资格和使用限制用当前明确返回或官方权益条款核对，`pricing plans` 只有价格时不能声称已核实全部权益。未取得当前证据的项在选项说明中标“待核实”，不把历史值写成当前可用额度。
- 当前 AFP 快照与上述历史基线不同，只能报告两者数值、查询时间和范围；没有订单周期或官方折算规则的证据，不得推断是按剩余天数折算、赠送、升降档或权益缩水。
- Coding 的适用场景不能简化成“技术上完全没有 API”：客户端接入方式、套餐允许用途和实际 API 支持分别核对，不能用一把 Key 能调通推导任意用途都合规。
- 宿主使用结构化选项时将这些信息放进 `option.description`；普通终端用等价简短文字。参数已完整时跳过收参，但不跳过协议和价格披露。

### 参数表

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

不传 `--yes` 不传 `--estimate` 时:

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

**失败类型要分清；已确认的参数不得重新收集：**

1. **transport error**：返回普通 error，但超时、断连或未解析出订单字段不能证明订单没落库/没扣钱。先核对订单状态；不得仅因失败退出就自动重新下单。
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

3. **`ProductStockNotEnough`（商品库存不足）**：不是用户参数填错。不自动重试，不重新问 plan / tier / 时长 / 席位数，不擅自换档或拆单；说明库存不足，保留已返回的 Order ID / SeatID，引导到对应 Agent Plan / Coding Plan 控制台页面处理。没有返回订单号只表示本次响应未提供订单号，不能据此宣称订单未落库、没有扣款或无需补付/取消；是否已有订单须核对控制台记录。GUI 只是后续处理入口，不表示购买已成功或页面一定有库存；Vaka 的卡片格式见主 Skill 宿主说明。

transport error、库存不足、支付失败都保留本轮已确认参数。只有用户主动改变购买目标时才重新收集变更项；不得为了“恢复流程”再次创建收费订单。

## 常见错误

| 错误 | 原因 | 处理 |
|------|------|------|
| `--plan must be one of ...` | 拼错 plan key | 严格用 4 选 1 |
| `--type %q not allowed for --plan %q` | tier 跟 plan 不兼容 | 见上方表格 |
| `--quantity is required for team plans` | 团队版漏传 | 加 `--quantity N` |
| `--duration must be 1..12` | 越界 | 1-12 之间 |
| `payment_failed` envelope | 余额不足 / auto-pay 失败 | 去 console 补单，不要 CLI 重试 |
| `ProductStockNotEnough` | 商品库存不足 | 不重试、不重问参数；转对应套餐控制台处理，保留已有订单/席位信息 |

## 注意事项

- **真实扣款**：CLI 没有 "下单不支付" 模式；要询价用 `--estimate`，不得为了验证流程擅自换轻档位真实试买
- 团队版下单后席位会以 `BillingStatus=Running` 状态出现，但**没绑定子用户**；后续走 [`plans team seat-assign`](arkcli-plans-team-seat-assign.md) 分配
- 重试要谨慎：`payment_failed` 的订单**已经在服务端**，重跑会再开一单

## 参考

- [arkcli-plans](../SKILL.md) -- skill 概览
- [`plans renew`](arkcli-plans-renew.md) -- 续费已有套餐
- [`plans get`](arkcli-plans-get.md) -- 下单后查持有状态
- [`plans team seat-assign`](arkcli-plans-team-seat-assign.md) -- 团队版下单后绑定子用户
- [arkcli-shared](../../arkcli-shared/SKILL.md)
