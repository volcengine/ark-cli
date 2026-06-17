# usage seats

> **前置条件:** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **跟其它命令的区别:**
> - `usage plan` / `balance --type plan` — **caller 自己**的订阅余额(子用户视角看自己的 seat)
> - `usage seats` — **admin 视角** 列举团队下**所有 seat**(管理员看团队席位)
>
> 子用户调 `usage seats` 通常 `AccessDenied`(后端权限校验)。需要 admin / root SSO 身份。

## 命令

```bash
# 列出 AgentPlan 团队版下所有 seat (默认自动翻页直到末页)
arkcli usage seats --product agent-plan-team

# 列出 CodingPlan 团队版的所有 seat
arkcli usage seats --product coding-plan-team

# 按 BizInfo 档位过滤
arkcli usage seats --product agent-plan-team --biz-info large

# 按 BillingStatus 过滤(回收态准备清理)
arkcli usage seats --product agent-plan-team --billing-status Reclaimed

# 查特定子用户绑定的 seat
arkcli usage seats --product agent-plan-team --user-id 12345
arkcli usage seats --product agent-plan-team --user-name alice

# 多个用户(client-side fan-out 并行查 N 次再合并)
arkcli usage seats --product agent-plan-team --user-name alice --user-name bob --user-name charlie

# 跨字段组合 = AND (alice 名下的 + 状态是 Running 的)
arkcli usage seats --product agent-plan-team --user-name alice --billing-status Running

# 单页查询(关闭 auto-paginate)
arkcli usage seats --product agent-plan-team --page-number 1 --page-size 50

# === --with-usage:同时拉每个 seat 的用量 ===
arkcli usage seats --product agent-plan-team --with-usage   # AFP 用量(used/total/percent)
arkcli usage seats --product coding-plan-team --with-usage  # 仅 percent(CodingPlan 后端只暴露百分比)
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|---|---|---|---|
| `--product` | 是 | string | 团队版 plan id:`agent-plan-team` / `coding-plan-team`。**personal 不支持**(没 seat 概念) |
| `--biz-info` | 否 | string | 档位过滤:`small/medium/large/max`(AgentPlan)\|`lite/pro`(CodingPlan) |
| `--billing-status` | 否 | string slice | 计费状态(可重复):`Running` / `Expired` / `Reclaimed` / `Inactive` |
| `--seat-status` | 否 | string | 席位状态:`Active` / `Inactive` |
| `--user-id` | 否 | string slice | 按子用户 ID 过滤(可重复;>1 触发 client-side fan-out 并行查) |
| `--user-name` | 否 | string slice | 按子用户名字过滤(可重复;>1 触发 client-side fan-out 并行查) |
| `--seat-id` | 否 | string slice | 指定具体 SeatID(可重复;wire 原生 array,单次调用) |
| `--project` | 否 | string | 项目过滤;不传走 active profile project |
| `--page-size` | 否 | int | 单页大小,默认 100 |
| `--page-number` | 否 | int | 1-based 页码;**显式传则关闭 auto-paginate**,只拉单页 |
| `--sort-by` / `--sort-order` | 否 | string | 排序字段 / 升降序 |
| `--with-usage` | 否 | bool | **Team product only**:join 每个 seat 的用量。AgentPlan team 走 `ListSeatAFPUsage` 出 `afp_usage{plan_type, periods[5h/weekly/monthly with used/total/percent]}`;CodingPlan team 走 `ListSeatInfoUsages` 出 `coding_usage{periods[session/weekly/monthly with percent only]}` |

## 默认行为:auto-paginate

不传 `--page-number` 时 service 层自动用 `NextToken` 翻页直到末页(最多 30 页 × 100 = 3000 seat 上限,触发上限会报错引导手动分页)。

## 多值过滤语义(必读)

flag 之间是 **AND** 关系(每个 flag 是 narrowing filter,跟 SQL `WHERE` / kubectl `-l` 同),**同一个 flag 多值是 OR** 关系。

| 例子 | 语义 | wire 调用 |
|---|---|---|
| `--seat-id sa --seat-id sb` | seat ∈ {sa, sb} | 1 次,wire `Filter.SeatIDs:["sa","sb"]` 原生 array |
| `--user-name a --user-name b` | user = a OR user = b | **2 次并行**(client-side fan-out;wire `Filter.UserName` 单值不接 array) |
| `--seat-id sa --user-name a` | seat ∈ {sa} **AND** user = a | 1 次 |
| `--seat-id sa --user-name a --user-name b` | seat ∈ {sa} **AND** (user ∈ {a,b}) | 2 次并行,每次都带同一个 SeatIDs filter |
| `--user-name alice --billing-status Reclaimed` | user = alice **AND** billing = Reclaimed | 1 次 |

**等价 SQL**:`WHERE seat_id IN (...) AND user_name IN (...) AND billing_status IN (...)`。

### 不会拿到的"并集" — 是刻意的

```bash
# 不会返回 "seat sa 加上 alice 的所有 seat 的并集"
arkcli usage seats --seat-id sa --user-name alice
# 实际:只返回 sa(且必须属于 alice;不属于则 0 行)
```

如果真的想要"seat sa **或** alice 的 seat" 这种并集语义,跑两条命令客户端合并:

```bash
{ arkcli usage seats --product ... --seat-id sa
  arkcli usage seats --product ... --user-name alice; } | jq -s '...'
```

故意不在 CLI 里隐式 OR — filter 的安全契约是"只能缩小结果",意外扩大容易导致误统计 / 误清理。

### Fan-out 是什么

只跟 wire 打交道时才需要懂这个细节:wire 上 `Filter.UserID / UserName` 是单值字段(传 array 会被 InvalidParameter 拒)。多用户场景 service 层起 N 个 goroutine 并行各发一次单值 RPC,客户端按 `SeatID` 去重合并。后果:

- **延迟** = `max(N 个并发的最慢一个)`,不是 sum(并行)
- **后端负担** = N 倍(用户传越多名字,RPC 越多)— 几十个用户没问题,几百个建议改 `--all` + 客户端再筛
- **一致性** = 多次调用之间有秒级时间窗,理论上有 race(seat 此刻被解绑),admin 盘点场景可忽略
- 任一 user 查询失败整体报错(标明哪个 user 失败),不返回部分结果

## 返回值

```json
{
  "items": [
    {
      "seat_id": "seat-001",
      "account_id": "210000",
      "project_name": "default",
      "biz_info": "large",
      "billing_status": "Running",
      "seat_status": "Active",
      "identity_type": "IAMUser",
      "user_id": "12345",
      "user_name": "alice",
      "instance_id": "inst-xxx",
      "order_time": 1700000000000,
      "expired_time": 1717000000000,
      "bind_count": 1,
      "create_time": 1700000000000,
      "update_time": 1716000000000
    }
  ],
  "total": 30,
  "biz_summaries": [
    { "biz_info": "large", "total_count": 30, "active_count": 25 }
  ]
}
```

### 字段说明

| 字段 | 含义 |
|---|---|
| `seat_id` | 唯一 seat 标识 |
| `biz_info` | 档位 |
| `billing_status` | 计费状态(`Running` 才计费;`Reclaimed` 已回收) |
| `seat_status` | 席位本身状态(跟 billing 解耦,例如 admin 主动停用) |
| `user_id` / `user_name` | **绑定的子用户**(后端原 schema 是 `IdentityId` / `IdentityDetail`,arkcli 改名跟 maas-fe SeatManageTable 列名对齐) |
| `bind_count` | 本月绑定换绑次数 |
| `order_time` / `expired_time` | 订单 / 过期 epoch ms |
| `total` | 后端口径下满足 filter 的总数 |
| `biz_summaries[]` | 按档位的 summary(`large` 档买了 30 个,激活 25 个) |

## Admin 工作流场景

**场景 1**:看团队整体订阅情况
```bash
arkcli usage seats --product agent-plan-team
# 看 biz_summaries 知道每档买了几个、激活了几个
```

**场景 2**:找到回收态(过期)的 seat 准备处理
```bash
arkcli usage seats --product agent-plan-team --billing-status Reclaimed
```

**场景 3**:查特定子用户绑了什么
```bash
arkcli usage seats --product agent-plan-team --user-name alice
```

**场景 4**:查特定 SeatID 的当前状态
```bash
arkcli usage seats --product agent-plan-team --seat-id seat-001 --seat-id seat-002
```

## 常见错误

| 错误 | 原因 | 处理 |
|---|---|---|
| `--product is required and must be a team plan` | 没传 / 传了 personal | 传 `agent-plan-team` 或 `coding-plan-team` |
| `AccessDenied` | 调用者不是 admin | 让 admin 跑;或子用户改用 `usage plan` 看自己的那个 seat |
| `exceeded auto-paginate limit (30 pages)` | 团队规模 > 3000 seat | 用 `--page-number` 手动分页 |

## 跟其它命令的关系

- 子用户视角看**自己**绑的 seat → [`usage plan`](arkcli-usage-plan.md) / [`usage balance --type plan`](arkcli-usage-balance.md)
- Admin 想看团队**用量明细**(不只 seat 列表) → [`usage plan-details --product agent-plan-team`](arkcli-usage-plan-details.md)
- 查模型免费额度(账户级) → [`usage balance --type free-quota`](arkcli-usage-balance.md)

## 跟 maas-fe 调用面对齐

- wire path:`/open/ListSeatInfos`(OpenAPI)。`--with-usage` 增量调:
  - AgentPlan team → `/open/ListSeatAFPUsage`(admin 列全 + 翻页)
  - CodingPlan team → `/open/ListSeatInfoUsages`(SeatIDs 必传,从 ListSeatInfos 收齐后批量喂,单次 ≤ 1000)
- 入参 schema 跟 maas-fe `useAdminListSeatInfos` 一致
- 翻页用 `NextToken`(cursor)而不是 PageNum,跟 maas-fe `useSeatManageTableModel` 同模式
- `IdentityId` / `IdentityDetail` 字段在前端 `SeatManageTable` 也叫 `user_id` / `user_name`,arkcli 跟前端列名对齐而非 wire 字段名
- `ListSeatAFPUsage` 用的是 page-based 分页(`PageNumber/PageSize`,后端约束 PageSize 10-100),跟 `ListSeatInfos` 的 cursor 翻页不同;service 层独立翻页拉全后再 join
