# usage plan

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **跟 `stats` / `billing` 的区别（务必先看）：**
> - `usage stats` — 按 token 计费的 **inference 用量**（按时间序列,有 5–30 分钟延迟）
> - `usage billing` — 按月**结算金额**(火山计费中心拆账,T+1 出账)
> - `usage plan` — **订阅类套餐(AgentPlan / CodingPlan)的 quota 快照**(后端"我的套餐"实时数据,跟 maas-fe 控制台 1:1)
>
> 用户问「我用了多少 token / 多少次请求」走 stats;问「花了多少钱」走 billing;问「我的套餐还剩多少额度 / 用了百分之几 / 几号刷新」走 **plan**。

查询订阅套餐的额度用量快照。Plan 不是按时间序列查,而是**按时间窗口(5h / weekly / monthly / session)给当前生效订阅的"已用 / 总额 / 重置时间"**。

## 命令

```bash
# 默认:并发 ListSubscribeTrade × 4 探所有 SKU (personal × 2 + team × 2),
# 探到的桶才下发 usage 请求 — 一次拉全你账号下所有生效订阅的余额
arkcli usage plan

# 显式指定 product (跳过订阅探测,直接单桶查)
arkcli usage plan --product agent-plan
arkcli usage plan --product coding-plan
arkcli usage plan --product agent-plan-team
arkcli usage plan --product coding-plan-team

# 强制查所有 4 个桶 (personal + team), 不管订阅与否 (诊断 / Excel 视角)
arkcli usage plan --all

# Team 版查指定席位 (绕过 GetSeatInfo 的自动 caller seat 解析)
arkcli usage plan --product agent-plan-team --seat seat-001
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--product` | 否 | string | 显式 product id: `agent-plan` / `coding-plan` / `agent-plan-team` / `coding-plan-team`。不传则默认走自动探订阅 |
| `--all` | 否 | bool | 强制查所有 4 个 product,没订阅的桶 endpoint 自己返 subscribed=false 行。**跟 `--product` 互斥** |
| `--seat` | 否 | string | 仅 team product 有意义。不传时 service 层自动调 `GetSeatInfo` 找 caller 绑定的 seat;传了则跳过该 RPC,直接用提供的 SeatID(管理员视角查别人的 seat) |

## 默认行为(无 `--product` / `--all`)

不传任何 product 时,**默认行为统一**(profile.Type 不再参与派生 — 历史版本曾按 profile 单桶查,V2 起改为统一探所有订阅,因为同一身份常常同时订了多个 product):

1. 4 桶并发探订阅状态:
   - **Personal × 2** 走 `ListSubscribeTrade`(`ResourceNames=[RealAgentPlanPersonal]` / `[""]`)
   - **Team × 2** 走 `GetSeatInfo(Scene="agent_plan_enterprise" / "")`(企业版不在 ListSubscribeTrade 暴露,实测返 null;改通过 caller seat 绑定状态探)
2. 探到的桶下发对应 usage RPC:
   - `agent-plan` (personal) → `GetAFPUsage`
   - `coding-plan` (personal) → `GetCodingPlanUsage`
   - `agent-plan-team` → 已经拿到 SeatID → `GetSeatAFPUsage(SeatIDs=[id])`
   - `coding-plan-team` → 已经拿到 SeatID → `GetSeatInfoUsage(SeatID=id, Scene="")`
3. 一桶失败不挡其它桶 (per-bucket error isolation)

> **多 1 次 RPC 换零 flag 体验**:`ListSubscribeTrade × 2 + GetSeatInfo × 2` 并发约 ~50ms,远低于循环逐 product 查的成本。要跳过这个探测就传具体 `--product`。
>
> **Team 桶静默规则**:caller 没绑 seat 时,team 桶在默认探测里**不出现**(避免 admin 没买企业版还看到 "no seat bound" 误导)。要看完整 4 桶(包括 "no seat bound" error 行)用 `--all` 强制。

## 返回值

JSON 形态(顶层 `viewer` + `items` 数组):

```json
{
  "viewer": {
    "auth_method": "sso",
    "is_root": false,
    "user_id": "12345",
    "user_name": "alice",
    "account_id": "210000",
    "profile": "my-profile",
    "tenant": "volc",
    "region": "cn-beijing",
    "project_name": "default"
  },
  "items": [
    {
      "product": "agent-plan",
      "edition": "personal",
      "tier": "medium",
      "subscribed": true,
      "periods": [
        { "label": "5h",      "used": 250,    "total": 1000,   "percent": 25,   "reset_at": 1717018000000 },
        { "label": "weekly",  "used": 12500,  "total": 50000,  "percent": 25,   "reset_at": 1717604800000 },
        { "label": "monthly", "used": 50000,  "total": 200000, "percent": 25,   "reset_at": 1719608800000 }
      ]
    },
    {
      "product": "agent-plan-team",
      "edition": "team",
      "tier": "large",
      "seat_id": "seat-001",
      "subscribed": true,
      "periods": [...]
    }
  ]
}
```

### `viewer` 字段(身份摘要)

```
viewer.auth_method  = sso | aksk | apikey | none
viewer.is_root      = SSO 主账号(JWT trn 带 :root)
viewer.user_id      = SSO 子用户 ID(AK/SK / APIKey 通常空)
viewer.user_name    = SSO 子用户名
viewer.account_id   = JWT sub claim
viewer.profile      = 当前激活 profile 名
viewer.tenant       = volc
viewer.region       = active region
viewer.project_name = active project
```

**用途**:子用户跟主账号视图易混淆 — `viewer` 字段告诉用户/agent"这条数据是给谁看的"。具体匹配规则:
- 子用户 SSO 登录 → `is_root=false`,`user_id` / `user_name` 是子用户身份,看到的是这个子用户绑定的 seat
- 主账号 SSO 登录 → `is_root=true`,看到的是主账号视图(整个 account 下的订阅)
- AK/SK 登录 → `auth_method=aksk`,`user_id` 通常空,后端按 AK/SK 解析的 IAM 身份找数据
- APIKey 登录 → `auth_method=apikey`,数据范围跟 API Key 绑定的身份对齐

### 字段说明

| 字段 | 说明 |
|---|---|
| `product` | `agent-plan` / `coding-plan` / `agent-plan-team` / `coding-plan-team`(跟 `profile.Type` 一致) |
| `edition` | `personal` / `team` |
| `tier` | 实际订阅档位(`small/medium/large/max` 或 `lite/pro`),由后端返;AgentPlan 才有,CodingPlan 通常缺 |
| `subscribed` | 当前身份是否有有效订阅。AgentPlan: `Result != nil`; CodingPlan: `QuotaUsage` 数组非空(跟 maas-fe `if (!data?.QuotaUsage)` 一致) |
| `periods[].label` | AgentPlan: `5h` / `weekly` / `monthly`(跟 maas-fe `useGetAgentPlanAFPUsage` UI 渲染一致,后端返的 `daily` 不展示)。CodingPlan: `session` / `weekly` / `monthly`,按 canonical order 稳定排序 |
| `periods[].used` / `total` | AgentPlan 给的是绝对值,CodingPlan 后端只返 `Percent`,`used`/`total` 字段缺(JSON 通过 `omitempty` 不出现) |
| `periods[].percent` | 已用百分比(0-100). AgentPlan 由 `used/total*100` 算; CodingPlan 直接透传后端 `Percent` |
| `periods[].reset_at` | 下次刷新时间,**统一 epoch ms**。AgentPlan 后端原本就是 ms; CodingPlan 后端返秒,service 层 ×1000 归一。`-1` 表示该周期内无数据(sentinel,不参与单位换算) |
| `updated_at` | 仅 CodingPlan,后端 `UpdateTimestamp` 透传(epoch ms) |
| `error` | 该桶失败原因(per-bucket error isolation,一桶失败不挡其它桶)。team product NotImpl 也走这个字段 |

## 错误处理

- **AccessDenied / 网络错误**: 写到对应 `items[].error`,其它桶不受影响,整体不阻塞
- **未订阅**: AgentPlan `Result=nil` → `subscribed: false, periods: []`; CodingPlan `QuotaUsage` 空 → 同样 `subscribed: false`
- **Team 订了但 caller 没绑 seat**: `subscribed: true` + `error: "no seat bound to caller for ..."`(区分于"未订阅"),传 `--seat <id>` 显式指定可绕过
- **typo product**: 不在 `agent-plan` / `coding-plan` / `*-team` 集合内 → ErrValidation 立即返,不发请求
- **`--all` + `--product` 同传**: ErrValidation 互斥
- **自动探全失败**: 上抛 `auto-discover subscriptions: <err>`(如 `MissingParameter.ResourceNames` 之类的后端拒绝),提示用户改用显式 `--product`

## AI Agent 决策路径

- 用户问「我的套餐用了多少 / 还剩多少」→ `arkcli usage plan`(默认探所有订阅,一把拉全)
- 用户问「我有几个套餐」→ `arkcli usage plan --all`(展开所有 4 个桶,subscribed=true 的就是有的)
- 用户问「team 套餐 admin 视角看其他 seat 用量」→ `arkcli usage plan --product agent-plan-team --seat <id>`
- 用户问「为啥 plan 出了空 items」→ 大概率当前身份名下所有 product 都没订阅,转 [arkcli-pricing](../../arkcli-pricing/SKILL.md) 查 catalog 价格
- 用户问「按模型看我哪个模型用得多」→ 这条命令只给 quota 快照,**没有按模型的拆分**,转 [`usage plan-details`](arkcli-usage-plan-details.md)(仅 AgentPlan)

## 跟 maas-fe 调用面对齐

- AgentPlan personal: `/open/GetAFPUsage`(OpenAPI,AK/SK + SSO 都支持)。maas-fe 控制台用的是 `ArkService.GetAgentPlanAFPUsage`(`/api/...` 路由,SSO-only),response shape 一致;arkcli 用公网 OpenTOP 入口,不带 AgentPlan 前缀
- CodingPlan personal: `/open/GetCodingPlanUsage`(OpenAPI),跟 maas-fe `useGetCodingPlanUsage` 共享同一 endpoint
- AgentPlan team: `/open/GetSeatAFPUsage`(**公网 OpenAPI**,AK/SK + SSO 都可调)。maas-fe 控制台用的是 `ArkService.GetAgentPlanSeatAFPUsage`(`/api/...` 路由,SSO-only),wire schema 一致;arkcli 用公网 OpenTOP 不带 AgentPlan 前缀的入口
- CodingPlan team: `/open/GetSeatInfoUsage`(OpenAPI),跟 maas-fe CodingPlanEnterprisePane `useUserState` 同源
- 自动探的 `ListSubscribeTrade` payload(personal × 2)跟 maas-fe `useGetAgentPlanInfo` / `useGetCodingPlanInfo` 一字不差;**企业版**不在 `ListSubscribeTrade` 暴露,改通过 `GetSeatInfo(Scene)` 探 caller seat 绑定状态(跟 maas-fe `AgentPlanEnterprisePane/useUserState` 同模式)
- Team SeatID 解析: `/open/GetSeatInfo` 不传 SeatID,后端按 caller 身份 + Scene 返绑定的 seat。**Scene 必须用 `agent_plan_enterprise`**(AgentPlan 企业版)或**空字符串**(CodingPlan 企业版,后端默认即 enterprise)— 早期错用 `agent_plan` / `coding_plan` 会返空 SeatID 不报错
