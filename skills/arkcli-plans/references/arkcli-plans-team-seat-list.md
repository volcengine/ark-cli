# plans team seat-list

> **前置条件：** 先阅读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

列出 Agent Plan / Coding Plan **企业版** 席位，支持 7 种过滤维度 + 分页。读操作。

## 命令

```bash
# 列 Agent Plan 团队版全部席位（默认页大小）
arkcli plans team seat-list --plan agent-plan-team

# Coding Plan 团队版
arkcli plans team seat-list --plan coding-plan-team

# 按席位档位过滤
arkcli plans team seat-list --plan agent-plan-team --type medium

# 精准查指定席位（最多 1000 个 SeatID）
arkcli plans team seat-list --plan agent-plan-team --seat-ids seat-001,seat-002

# 按绑定子用户名过滤
arkcli plans team seat-list --plan agent-plan-team --user-name yinfan.ivan

# 仅看空闲席位（Idle=1）
arkcli plans team seat-list --plan agent-plan-team --seat-status 1

# 仅看 Running 计费状态
arkcli plans team seat-list --plan agent-plan-team --billing-status 2

# 翻页
arkcli plans team seat-list --plan agent-plan-team --page-number 2 --page-size 50
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--plan` | 是 | string | `agent-plan-team` / `coding-plan-team`（个人版 plan 会被拒绝） |
| `--type` | 否 | string | tier 过滤：agent-plan-team 取 `small/medium/large/max`；coding-plan-team 取 `lite/pro` |
| `--seat-ids` | 否 | string | 逗号分隔精准过滤；**单次上限 1000**（客户端先拦截） |
| `--user-name` | 否 | string | 按绑定的子用户名过滤 |
| `--seat-status` | 否 | int | 1=Idle（空闲未绑定）/ 2=Active（已绑定）。其它值会被拒 |
| `--billing-status` | 否 | int | 1=Pending（待付费） / 2=Running（生效中）/ 3=Expired（已过期）/ 4=Reclaimed（已回收） |
| `--project-name` | 否 | string | 项目名，默认 `default` |
| `--page-number` | 否 | int | 页码（≥1） |
| `--page-size` | 否 | int | 页大小（≥1） |

`--plan` 与 `--type` 兼容性跟 `plans buy` 一致。

## 返回值

```json
{
  "plan": "agent-plan-team",
  "scene": "agent_plan_enterprise",
  "total": 42,
  "seats": [
    {
      "seat_id": "seat-20260608110804-gjrwr",
      "tier": "medium",
      "seat_status": "Active",
      "billing_status": "Running",
      "user_id": "83144215",
      "user_name": "yinfan.ivan",
      "project_name": "default",
      "instance_id": "ins-...",
      "order_time": 1700000000000,
      "expired_time": 1800000000000,
      "create_time": 1690000000000,
      "update_time": 1700000000000
    }
  ]
}
```

| 字段 | 说明 |
|------|------|
| `scene` | agent-plan-team 是 `agent_plan_enterprise`；coding-plan-team **空字符串**（服务端默认 coding_plan） |
| `total` | 服务端匹配过滤后的总数（不一定等于当前页大小） |
| `seat_status` | "Idle" / "Active" / "Unknown"（已经把数字枚举翻译成可读字符串） |
| `billing_status` | "Pending" / "Running" / "Expired" / "Reclaimed" / "Unknown" |
| `user_id` / `user_name` | 绑定的子用户标识（IAM `IdentityId` / `IdentityDetail`） |
| `*_time` | epoch ms |

`coding-plan-team` 时输出里 `scene` 是空字符串 —— 这是预期行为，服务端按空 Scene 默认到 coding_plan。

## 注意事项

- 如果用户问 "我有多少个席位" 不区分 plan：先 `arkcli plans get` 看持有哪种团队版，再分别 `seat-list` 一次
- **本命令只列基础信息**（SeatID / 绑定身份 / billing 状态等）。**想看每个席位的用量**（5h/weekly/monthly percent + reset_at）走 `arkcli usage seats --product <agent-plan-team|coding-plan-team> --with-usage` —— 数据源同 `ListSeatInfos`，多 join 一层 `ListSeatAFPUsage` / `ListSeatInfoUsages`
- 数字枚举 (`--seat-status` / `--billing-status`) 已经在 reference 里说明数字含义，**不要让用户去查文档**；提示时直接给意义
- `--seat-ids` 不去重，重复 ID 由服务端按最后一条覆盖
- 默认 `BillingStatus` 不过滤（返回所有状态）；frontend 上 BindSeatDrawer 默认只看 `[Running, Expired]`，CLI 更"raw"，要按需自己加 `--billing-status`

## 参考

- [arkcli-plans](../SKILL.md) -- skill 概览
- [`plans team seat-assign`](arkcli-plans-team-seat-assign.md) -- 给空闲席位绑定子用户
- [`plans team rotate-apikey`](arkcli-plans-team-rotate-apikey.md) -- 轮换席位 APIKey
- [`plans get`](arkcli-plans-get.md) -- 先看持有哪种团队版套餐
- [arkcli-usage](../../arkcli-usage/SKILL.md) -- 席位**用量**视图（`usage seats --with-usage`），跟本命令的"基础列表"互补
- [arkcli-shared](../../arkcli-shared/SKILL.md)
