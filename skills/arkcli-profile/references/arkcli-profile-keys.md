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
