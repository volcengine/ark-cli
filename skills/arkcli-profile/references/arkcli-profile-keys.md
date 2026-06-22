# profile keys 详细参考

> **前置**：先读 [`../SKILL.md`](../SKILL.md)。

`profile keys` 是 0.1.16 引入的 API Key 管理子树，跟 `arkcli auth apikey`（交互式选 key）**不是同一个东西**：

- `profile keys list/use/refresh` 操作的是 **profile.yaml 里的 available_api_keys 列表 + default_api_key 选择**
- `arkcli auth apikey` 操作的是 **`.env` 里的 `VOLCENGINE_ARK_API_KEY`（identity-level 凭证）**

## 命令模板

```bash
# 列当前 profile 的 default + available_api_keys（masked）
arkcli profile keys list --format json
arkcli profile keys list --profile platform_cn-beijing_default --format json

# 切 default key（必须 ∈ available list）
arkcli profile keys use 392xxxxdab0

# 重拉控制面 ListApiKeys → 更新 available list（写 profile.yaml）
arkcli profile keys refresh
arkcli profile keys refresh --profile platform_cn-beijing_default
```

## `keys refresh` 关键行为

1. **用 target profile 的身份打控制面**（codex P0-A 修正）
   - `--profile X` 时内部 `Factory.RebuildForProfile(X)` 重建 invoker
   - 不会再用 active=A 的 token 打控制面后写到 B
2. **用 `FetchActiveRaw` 拿 full plaintext**（codex P1-B 修正）
   - `apikeyservice.List` 返回的 `Items[].Key` 是 mask 字符串（`392****dab0`）
   - 直接写到 `available_api_keys` 会让数据面 SDK 拿 mask 当 Bearer → server 报 `API key format is incorrect`
   - `FetchActiveRaw` 内部 List + 对每个 active item 调 `GetRawApiKey` 拿真 UUID
3. **graceful fallback**：控制面拉 key 失败 (登录态/STS 过期、上游临时不可用) → 用 `.env` 缓存单 key 兜底
   - stderr 打 warn：`控制面拉 API Key 列表失败 (NotLogin), 用 .env 缓存单 key 兜底`
   - `profile.available_api_keys` 仅含 1 项，恢复后再 refresh

## `keys refresh` 的防御用途（后端 key 变更 → 自愈）

refresh 本质是「**以后端为 SSOT、单向把当前 profile 的 key 同步下来**」，所以它是「**后端 key 被改 / 轮换 → 本地 profile.yaml 变旧 → 业务命令报鉴权错**」这类故障的**第一线自愈动作**：比 `auth login` 重登、`auth apikey` 交互重选都轻，一条非交互命令即可，适合 agent 自动自愈（agent 完整恢复流程见 [`../../arkcli-auth/references/auth-modes.md`](../../arkcli-auth/references/auth-modes.md) 的「API Key 错误恢复」playbook）。

**但 refresh 不是万能 —— 它能/不能自愈什么：**

| 症状 | refresh 能否自愈 | 正确动作 |
|---|---|---|
| `InvalidApiKey` / `Unauthorized` / `AuthenticationError`（key 旧 / 被轮换，且没主动换过 key） | ✅ 能（后端有有效 key 可拉） | `keys refresh` → 重试原命令 |
| key 来自 `ARK_API_KEY` env / `--api-key` flag | ❌ 不能（覆盖优先级高于 profile.yaml，refresh 写盘不生效） | 先去掉覆盖，再 refresh |
| `AccessDenied` 针对某资源（key 缺权限） | ❌ 不能（同账号同权限，refresh 后还是同样被拒） | console 给 key 加权限 / 建带权限新 key |
| SSO / 身份过期（refresh 自己的控制面调用也 401） | ❌ 不能 | `arkcli auth login volc-sso` 重登 |
| team 档无 Running 席位 | ❌ 不能（refresh 直接报错） | 查席位 / 套餐（`arkcli-plans`） |

**refresh ≠ rotate**：refresh 同步「已被改的」key；主动「换一把 / 废弃泄露的 key」走 `arkcli plans personal|team rotate-apikey`。**单 profile 范围**：只治当前 / `--profile` 那条，不影响其它 profile。

## 输出形态

### `keys list`

```json
{
  "profile": "platform_cn-beijing_default",
  "default_api_key": "392****dab0",
  "available_api_keys": ["392****dab0", "abc****1234", "..."]
}
```

### `keys use`

```json
{
  "profile": "platform_cn-beijing_default",
  "new_default": "abc****1234"
}
```

### `keys refresh`

```json
{
  "profile": "platform_cn-beijing_default",
  "default_api_key": "392****dab0",
  "available_api_keys": ["...", "..."],
  "added_count": 2,
  "removed_count": 0,
  "refresh_status": "ok"
}
```

stderr 同步打：
```
[arkcli] profile "platform_cn-beijing_default" keys refreshed: +2 -0 (total 5)
```
