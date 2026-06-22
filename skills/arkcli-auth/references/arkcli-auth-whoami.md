# auth whoami

## 用途

返回当前认证态的用户身份。区别于 `auth status`：

| | `auth status` | `auth whoami` |
|---|---|---|
| 关注点 | 凭证（token / api_key / aksk）健康度 | 身份（你是谁） |
| 输出 | 嵌套（按凭证类型分组） | 扁平（脚本友好） |
| 主要消费方 | 排障 | Skill / 脚本过滤"我的 xxx" |

## 标准命令

```bash
arkcli auth whoami
arkcli auth whoami --transform user_id   # 直接拿 IAM 用户 ID
arkcli auth whoami --transform name      # 直接拿用户名
```

## 输出契约（SSO 已登录）

```json
{
  "logged_in": true,
  "auth_method": "sso",
  "name": "alice",
  "user_id": "10000001",
  "account_id": "2000000001",
  "trn": "trn:iam::2000000001:user/alice",
  "is_root": false,
  "project_name": "default",
  "region": "cn-beijing"
}
```

字段语义：

- `user_id`：IAM 用户 ID。在 ARK 各资源里以多种形态出现：
  - 带 `sys:ark:createdBy` Tag 的资源（如 infer endpoints、自定义模型）：拼成 `IAMUser/<user_id>/<name>` 作为 Tag value
  - API Key：火山走 `Filter.UserId` / 响应里的 `UserId` 字段
  - usage / 已天然按账号 scope 的资源：不需要这个字段
  ⚠️ 不存在跨资源通用的"我的 xxx"过滤公式 —— 具体怎么用 `user_id` 由各资源 skill 的 reference 决定。详见 [`identity-resolution.md`](identity-resolution.md) 的决策顺序与速查表。
- `account_id`：火山账号 ID（JWT `sub` 声明）。同一账号下不同 IAM 用户共享。
- `trn`：完整 IAM TRN，例如 `trn:iam::<account_id>:user/<name>` 或 `trn:iam::<account_id>:root`。
- `is_root`：是否根账号身份。仅在能解出 `trn` 时出现。
- `project_name` / `region`：当前生效的 project / region（已应用 flag / 环境变量 / `.env` 优先级解析）。

## 边角场景

### SSO 已过期但 .env 还有 token

```json
{
  "logged_in": false,
  "auth_method": "none",
  "sso_expired": true,
  "name": "alice",
  "user_id": "10000001",
  "account_id": "2000000001",
  "trn": "trn:iam::2000000001:user/alice",
  "is_root": false,
  "project_name": "default",
  "region": "cn-beijing",
  "hint": "SSO token expired; run `arkcli auth login volc-sso` to refresh"
}
```

身份字段仍然能从过期 token 解出来（用于"我曾经是谁"），但 `logged_in=false` 提示无法发起远端调用。Agent 命中 `sso_expired=true` 应按当前 profile tenant 引导用户重新登录（火山：`arkcli auth login volc-sso`）。

### AK/SK 模式

```json
{
  "logged_in": true,
  "auth_method": "aksk",
  "project_name": "default",
  "region": "cn-beijing",
  "hint": "AK/SK mode cannot resolve IAM user identity; run `arkcli auth login volc-sso` to identify yourself"
}
```

AK/SK 模式拿不到 IAM 用户身份（无 IDToken）。"我的 xxx"语义在这种状态下无法满足，应引导 SSO 登录。

### 完全未登录

```json
{
  "logged_in": false,
  "auth_method": "none",
  "hint": "run `arkcli auth login` or `arkcli auth login volc-sso`"
}
```

## Agent 使用要点

- **不要解析 `~/.arkcli/.env` 里的 JWT**：whoami 已经做了这件事，直接读 `user_id` 就好
- **不要把 `account_id` 当成 IAM 用户 ID**：两者语义完全不同，`account_id` 是账号级别（同账号下多个 IAM 用户共享），`user_id` 是 IAM 子用户级别
- **`user_id` 怎么用，看具体资源 skill 的 reference**：whoami 只负责"告诉你你是谁"，不预设过滤公式。endpoint 走 Tag、API Key 火山走 `Filter.UserId`，机制不同 —— 详见 [`identity-resolution.md`](identity-resolution.md) 的速查表
- 看到 `sso_expired=true` 或 `auth_method=aksk` 但缺 `user_id` 时，先回到 [`../SKILL.md`](../SKILL.md) 走登录流程，再回原任务

## 与 status 的关系

- `auth status` 在 SSO 路径下也会暴露 `volc_sso.identity` 嵌套子对象（同字段集），但更适合排障语境
- 脚本和 Skill 优先使用 `auth whoami`：扁平 + 稳定 schema + 支持 `--transform`
