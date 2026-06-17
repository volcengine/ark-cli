# plans team seat-assign

> **前置条件：** 先阅读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

> **写操作：** 修改企业版席位与子用户的绑定关系。一次调用 `GrantSeats`，可批量；服务端 per-item 状态（成功 / 失败可同时存在）。

把 Agent Plan / Coding Plan 企业版席位绑定到 IAM 子用户。同一 Action 同时支持两种企业版（agent-plan-team / coding-plan-team），服务端按 SeatID 反推套餐线，**不需要 Scene**。

## 不知道员工的精确 UserID 怎么办

如果你只知道员工用户名（或前缀），先用 `arkcli iam userid` 反查：

```bash
# 单个前缀
arkcli iam userid --username ivan

# 多个员工一次性查
arkcli iam userid --username ivan,bob,carol
```

输出形态：
```json
{
  "queries": [
    {
      "query": "ivan",
      "matches": [
        {"user_id": "12345", "user_name": "ivan"},
        {"user_id": "67890", "user_name": "ivanka"}
      ]
    }
  ]
}
```

- 严格**前缀匹配**：`iv` 不会命中 `kevin`，只命中 `ivan` / `ivanka`
- 一个前缀命中多个时全部返回，让你挑
- 0 命中也保留该 query 记录，方便脚本判定
- 不传 `--username` → 返回当前身份的 UserID（即 "我是谁"）

把拿到的 `user_id` 喂给 `seat-assign --bind`，跳过下面 IAM 反查也省一次 round-trip。

## 命令

```bash
# 单条绑定（自动调 IAM ListUsers 反查 UserName）
arkcli plans team seat-assign --plan agent-plan-team \
    --bind seat-001=83144215

# 批量绑定（重复 --bind）
arkcli plans team seat-assign --plan agent-plan-team \
    --bind seat-001=83144215 \
    --bind seat-002=83143634

# 显式指定 UserName，跳过 IAM 反查（escape hatch — 没 IAM 权限或要省调用时用）
arkcli plans team seat-assign --plan coding-plan-team \
    --bind seat-001=83144215:ivan \
    --bind seat-002=83143634:bob

# 自定义 project
arkcli plans team seat-assign --plan agent-plan-team \
    --bind seat-001=83144215 \
    --project-name my-project
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--plan` | 是 | string | `agent-plan-team` / `coding-plan-team` |
| `--bind` | 是 | stringArray | 重复多次实现批量。格式 `seat-id=user-id` 或 `seat-id=user-id:user-name` |
| `--project-name` | 否 | string | 项目名，默认 `default` |

> **`--bind` 而不是 `--seat-id=...`：** 用户原始 spec 写的 `--<seat-id>=<user-id>` 是动态 flag 名，cobra/pflag 不支持。改成可重复的固定 flag `--bind` 加值。

### 三段语法 / IAM 反查

`--bind` 的值有两种形态：

| 形态 | 解析 | 行为 |
|---|---|---|
| `seat-001=83144215` | UserName 留空 | service 层自动调 **Volc IAM `ListUsers`**（直连 `iam.volcengineapi.com`），分页扫账号下全部子用户，按 ID 匹配 UserName |
| `seat-001=83144215:ivan` | UserName 显式 | 跳过 IAM 反查，直接喂给 `GrantSeats` |

服务端 `GrantSeats` 对 `{SeatID, UserID, UserName}` 三字段强校验（**UserName 必须非空**），所以 CLI 必须先拿到 UserName。IAM 反查是默认路径；如果 STS 凭据缺 `iam:ListUsers` 权限或想避开 IAM round-trip，用 `:user-name` escape hatch。

IAM 反查是**一次 list-all + 内存 map 缓存**，同一次 seat-assign 调用里多个 binding 复用同一份缓存（不会每个 binding 都 N 次往返）。

## 返回值

```json
{
  "plan": "agent-plan-team",
  "project_name": "default",
  "success_count": 2,
  "failed_count": 0,
  "success": [
    {
      "seat_id": "seat-001",
      "user_id": "83144215",
      "api_key_sid": "apikey-...",
      "api_key": "sk-..."
    }
  ],
  "failed": []
}
```

| 字段 | 说明 |
|------|------|
| `success[].api_key` | 该席位绑定后的 APIKey 明文，立即给用户保存 / 同步给 harness |
| `success[].api_key_sid` | APIKey 稳定 ID，后续做轮换 / 删除时定位资源 |
| `failed[].reason` | 服务端 `FailedReason` 透传，典型：`BindCountLimitExceeded` |

## 失败语义

**部分失败**是合法终态：
- stdout 始终是完整 result（含 success + failed 两个数组）
- `failed_count > 0` 时 stderr 追加 `grant_seats_partial` envelope，exit code = 5（`ExitAPI`）
- `failed_count == 0` 时 exit 0

| 错误 | 原因 | 处理 |
|------|------|------|
| `--plan must be one of ...` | 拼错 / 用了个人版 plan | 用 `agent-plan-team` / `coding-plan-team` |
| `--bind ... must be in the form seat-id=user-id` | 格式错 | 严格 `seat-id=user-id` 或 `seat-id=user-id:user-name` |
| `--bind specifies SeatID %q more than once` | 同一 SeatID 在多个 `--bind` 里出现 | 服务端会拒，客户端先拦 |
| `resolve UserName for UserID=...: IAM ListUsers ...: AccessDenied` | STS 凭据缺 IAM 权限 | 走 escape hatch：`--bind seat-id=user-id:user-name` 显式指定 |
| `failed_count > 0` + `BindCountLimitExceeded` | 该子用户已经绑了 N 个席位（受限） | 先解绑（暂未实现 unbind 子命令）；frontend 走的是 `RevokeSeats` |

## 注意事项

- IAM 反查是**全 listing**（分页拉账号下所有 IAM 用户），只在第一个 binding 触发；同一调用里之后 lookup 走内存
- IAM 反查防御性兜底：超过 10000 条还没找到目标 ID 时报错（防止后端循环游标无限分页）
- 同一席位**不能同时**绑多个子用户（服务端约束）；要换人需先 unbind 再 assign（unbind = `RevokeSeats`，CLI 暂未暴露）
- `success[].api_key` 是**新生成的 APIKey 明文**，绑定即生效；第一次返回后 console / CLI 都拿不到原始明文（mask 形态），务必立即保存

## 参考

- [arkcli-plans](../SKILL.md) -- skill 概览
- [`plans team seat-list`](arkcli-plans-team-seat-list.md) -- 先列出空闲席位
- [`plans team rotate-apikey`](arkcli-plans-team-rotate-apikey.md) -- 已绑定后轮换 APIKey
- [arkcli-shared](../../arkcli-shared/SKILL.md)
