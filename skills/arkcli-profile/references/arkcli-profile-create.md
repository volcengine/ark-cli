# profile create 详细参考

> **前置**：先读 [`../SKILL.md`](../SKILL.md)。本文件只补 flag 细节、输出形态、错误码。

## 命令模板

```bash
# Platform 类（最常用，默认 region=cn-beijing / project=default）
arkcli profile create --type platform --set-default

# 显式 region + project
arkcli profile create --type platform --region cn-beijing --project myproject --set-default

# Agent Plan：自动 Detect 订阅状态
arkcli profile create --type agent-plan --set-default

# Agent Plan：强制 plan-tier（绕过 OpenTOP 后端可见性 bug）
arkcli profile create --type agent-plan --plan-tier medium --set-default

# Coding Plan
arkcli profile create --type coding-plan --plan-tier lite --set-default

# 完全 inline（CI 友好，所有必填字段都给）
arkcli profile create \
  --type platform \
  --region cn-beijing \
  --project default \
  --owner-trn "trn:iam::21xxxxxxx:user/myname" \
  --no-interactive \
  --set-default
```

## Flag 一览

| 参数 | 必填 | 说明 |
|------|------|------|
| `--type` | 是（除非 interactive） | `platform` / `agent-plan` / `coding-plan` |
| `--region` | 否 | 默认 `cn-beijing` |
| `--project` | 否 | 默认 `default`；0.1.16 起 Project 进 profile 切面 |
| `--name` | 否 | 默认 `{type}_{region}_{project}` 派生；显式传跳过派生 |
| `--default-api-key` | 否 | 显式指定 default API Key；不传时 fetcher 自动拉 active list 第一个 |
| `--owner-trn` | 否 | 火山 IAM trn；通常留空（SSO 登录场景下由 `sso.Switch` 从 IDToken `claims.Trn` 直接写入；0.1.16 起不再走 self-heal 回填） |
| `--plan-tier` | 否 | **仅 plan type 适用**；agent-plan: `small`/`medium`/`large`/`max`，coding-plan: `lite`/`pro`；传了等于 bypass `ListSubscribeTrade Detect`（绕过 OpenTOP 后端可见性 bug） |
| `--set-default` | 否 | 创建后立即设为 default profile |
| `--no-interactive` | 否 | 缺必填字段直接报错，不弹 promptui |

## 输出形态

```json
{
  "created": "platform_cn-beijing_default",
  "type": "platform",
  "region": "cn-beijing",
  "project": "default",
  "is_default": true,
  "api_key_count": 3
}
```

`api_key_count` 是 fetcher 拉到的 active 全 list 长度；写到 `profile.available_api_keys`。0 表示控制面没拉到 key（NotLogin / 权限不足），fetcher 会用 `.env` 缓存单 key 兜底，并打 stderr warn。

## Plan 类 Detect 失败的兜底

Agent Plan / Coding Plan 创建时会先调 `ListSubscribeTrade` 确认订阅状态：

```
Agent Plan 订阅已回收, 请前往 console 续费:
  https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement
  如已购买但 Detect 看不到, 加 --plan-tier=<small|medium|large|max> 绕过 (后端可见性问题)
```

**已知 OpenTOP 后端可见性 bug**：用户在 console 看得到订阅，但 server 返空 InfoList。这时 `--plan-tier=<tier>` 强制注入 PlanTier，跳过 Detect。

## 跟旧 `config init` 的差异

| 维度 | 旧 `config init` | 新 `profile create` |
|------|------------------|---------------------|
| `--type` | 没有此 flag，profile = (region × project) | 必填，明确区分 platform / agent-plan / coding-plan |
| `--plan-tier` | 不存在 | 新增（仅 plan 类） |
| `--owner-trn` | 不存在；trn 不绑 profile | 新增，profile schema v1 切面字段 |
| API Key 拉取 | 单 key（用户传的 `--api-key`） | 全 list fetch + masking + 选 default |
| Project 字段 | 仅 `.env` `VOLCENGINE_ARK_PROJECT_NAME` | 进 profile 切面，`.env` 仅兜底 |

老脚本里 `arkcli config init ...` 仍然能跑（标 deprecated，0.2.x 删除），但 Agent 不应再引导用户用。
