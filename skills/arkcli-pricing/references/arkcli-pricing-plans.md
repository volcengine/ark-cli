# pricing plans

> **前置条件:** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

列出 AgentPlan / CodingPlan 套餐订阅价格(个人版 + 企业版,共 12 档)。`Price` 是后端按当前账号计算的**最终订阅单价(含折扣)**;`OriginalPrice` 是公示原价。

`pricing plans` 与 `pricing models` 是**两类完全独立的命令**:
- `pricing models` — 按 token 计费(模型按量调用)
- `pricing plans` — 包月订阅(整体套餐)

## 命令

```bash
# 列全 12 档套餐(personal 8 + enterprise 4 = 12)
arkcli pricing plans

# 单档查询
arkcli pricing plans --plan agent-plan-personal-small
arkcli pricing plans --plan coding-plan-personal-pro

# 按产品过滤
arkcli pricing plans --product agent-plan        # 只 AgentPlan,共 8 档
arkcli pricing plans --product coding-plan       # 只 CodingPlan,共 4 档

# 按版本过滤
arkcli pricing plans --edition personal          # 只个人版,共 6 档
arkcli pricing plans --edition enterprise        # 只企业版,共 6 档

# 组合
arkcli pricing plans --product agent-plan --edition personal   # AgentPlan 个人版 4 档
```

## Plan ID 全枚举

| Plan ID | 产品 | 版本 | 档位 |
|---|---|---|---|
| `agent-plan-personal-small` | AgentPlan | personal | small |
| `agent-plan-personal-medium` | AgentPlan | personal | medium |
| `agent-plan-personal-large` | AgentPlan | personal | large |
| `agent-plan-personal-max` | AgentPlan | personal | max |
| `agent-plan-enterprise-small` | AgentPlan | enterprise | small |
| `agent-plan-enterprise-medium` | AgentPlan | enterprise | medium |
| `agent-plan-enterprise-large` | AgentPlan | enterprise | large |
| `agent-plan-enterprise-max` | AgentPlan | enterprise | max |
| `coding-plan-personal-lite` | CodingPlan | personal | lite |
| `coding-plan-personal-pro` | CodingPlan | personal | pro |
| `coding-plan-enterprise-lite` | CodingPlan | enterprise | lite |
| `coding-plan-enterprise-pro` | CodingPlan | enterprise | pro |

`--plan` 接受的就是上面这 12 个 ID(三段式 kebab-case)。错拼时命令会返回结构化错误 + 列出全部合法 ID。

## 输出字段

```json
{
  "items": [
    {
      "plan_id": "agent-plan-personal-small",
      "product": "AgentPlan",
      "edition": "personal",
      "tier": "small",
      "price": 40,
      "original_price": 40,
      "period": "monthly",
      "currency": "CNY"
    },
    {
      "plan_id": "agent-plan-enterprise-small",
      "product": "AgentPlan",
      "edition": "enterprise",
      "tier": "small",
      "price": 0,
      "original_price": 0,
      "period": "monthly",
      "currency": "CNY",
      "error": "AccessDenied: caller lacks permission"
    }
  ]
}
```

| 字段 | 含义 |
|---|---|
| `plan_id` | 三段式 ID,可用作 `--plan` 入参 |
| `product` | `AgentPlan` 或 `CodingPlan` |
| `edition` | `personal` 或 `enterprise` |
| `tier` | 档位标识,personal/enterprise 共用同套(small/medium/large/max 或 lite/pro) |
| `price` | 当前账号最终订阅单价(后端结算后) |
| `original_price` | 公示原价 |
| `period` | 订阅周期,目前固定 `monthly` |
| `currency` | 币种,目前固定 `CNY` |
| `error` | 该行查询失败的原因。**只有该行**填,其它行不影响 |

## 错误隔离机制

后端把"个人版批量"和"企业版批量"分成两个 Action,所以**个人版和企业版的成功 / 失败彼此独立**:

- 你账号没有企业版权限 → 后端返回 `AccessDenied`
- 现象:6 行 enterprise 的 `error` 字段填错误信息,`price = 0`;6 行 personal 正常返价
- 命令本身**不报错** —— `arkcli pricing plans` 退出码仍是 0,JSON 输出完整 12 行

这跟火山方舟控制台的 UI 行为一致:UI 也是分 4 个 tab(AgentPlan / Agent 企业版 / Coding Plan / Coding 企业版),每个 tab 独立调询价、独立失败,不会因一个挂了影响另一个。

**Agent 在响应处理时**:
1. 先看用户问的是哪一档 → 直接读那一行的 `price`
2. 如果该行 `error` 非空 → 告诉用户"账号没有 X 版权限,联系管理员"
3. 不要因为某行有 `error` 就否定其他行的价格

## 参数细节

### `--plan <plan-id>`
精确匹配 12 个 canonical ID 之一。命中 → 只查那一档(只调 1 个 API 即可)。错拼 → 错误信息 + 列出全部合法 ID。

### `--product agent-plan | coding-plan`
也接受 PascalCase(`AgentPlan` / `CodingPlan`)。其他值 → 返回空 items 数组(不报错,因为已知值集合是封闭的,但不阻塞 agent 链路)。

### `--edition personal | enterprise`
精确匹配。

### 组合规则
- 不传任何 flag → 全 12 档
- `--plan X` 优先级最高(传 plan 时其它 flag 被忽略,因为 plan 本身已经唯一指定产品+版本+档位)
- `--product` + `--edition` 可叠加(逻辑 AND)

## 与其他命令的对照

- 用户问"DeepSeek 多少钱"、"图生图多少钱"、"做精调多少钱" → [`pricing models`](arkcli-pricing-models.md)
- 用户问"Agent Plan 多少钱"、"Coding Plan small 多少钱"、"个人版 / 企业版价格" → 本命令
- 用户问"我**已经**用了多少 / 消耗多少" → [`../arkcli-usage/SKILL.md`](../../arkcli-usage/SKILL.md)

## 调用涉及的后端 Action(供 raw API 排错时参考)

| Action | 用于 |
|---|---|
| `EstimateSubscribeNewOrderPrice` | AgentPlan 个人版批量(4 档一次返回) |
| `EstimateSubscribePrice` | CodingPlan 个人版单条(每档单独调一次) |
| `EstimateAgentPlanEnterpriseNewOrderPrice` | AgentPlan + CodingPlan 企业版(按 Scene 区分,Agent=`agent_plan_enterprise`,Coding 不传) |

如果 `pricing plans` 失败到完全无响应,可以用 raw API 自检:

```bash
arkcli api trade.estimate_subscribe_new_order_price --params '{
  "Items": [
    {"BizInfo":"small","ResourceType":"AgentPlan","ResourceName":"RealAgentPlanPersonal","Period":"monthly","Times":1}
  ]
}'
```
