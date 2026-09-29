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

# 团队版：当前身份必须已绑定对应的 Running 席位；不手填档位或普通池 Key
arkcli profile create --type agent-plan-team --no-interactive
arkcli profile create --type coding-plan-team --no-interactive

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
| `--type` | 是（除非 interactive） | `platform` / `agent-plan` / `coding-plan` / `agent-plan-team` / `coding-plan-team` |
| `--region` | 否 | 默认 `cn-beijing` |
| `--project` | 否 | 默认 `default`；0.1.16 起 Project 进 profile 切面 |
| `--name` | 否 | 默认 `{type}_{region}_{project}` 派生；显式传跳过派生 |
| `--default-api-key` | 否 | 显式指定绑定的 default API Key；默认按 type 取普通池、Agent Plan 专属池或团队席位 Key，不可跨池替代 |
| `--owner-trn` | 否 | 火山 IAM trn；通常留空（SSO 登录场景下由 `sso.Switch` 从 IDToken `claims.Trn` 直接写入；0.1.16 起不再走 self-heal 回填） |
| `--plan-tier` | 否 | **仅个人版手动声明**：agent-plan 取 `small`/`medium`/`large`/`max`，coding-plan 取 `lite`/`pro`；跳过个人订阅 Detect，不绕过权限/协议/余额。团队版仍查询真实席位，并以其档位为准 |
| `--set-default` | 否 | 创建后立即设为 default profile |
| `--no-interactive` | 否 | 缺必填字段直接报错，不进入交互；不是跳过远端资格校验 |

这些是创建配置的写命令，不是查订阅的探针。先确认创建目标和影响；`--set-default` 仅在用户要求创建后同时切换时使用，不能因为示例带了就自动添加。

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

## 个人版 Detect 与团队版席位校验

Agent Plan / Coding Plan **个人版**创建时会先调 `ListSubscribeTrade` 确认订阅状态：

```
Agent Plan 订阅已回收, 请前往 console 续费:
  https://console.volcengine.com/ark/region:ark+cn-beijing/openManagement
  如已购买但 Detect 看不到, 加 --plan-tier=<small|medium|large|max> 绕过 (后端可见性问题)
```

控制台可见但 Detect 返回空可能是可见性问题，也可能是身份/范围不同，不能仅凭空列表断言后端 bug。只有用户明确确认实际个人档位和手动声明意图时才用 `--plan-tier=<tier>`；创建成功仍不证明数据面可调用。

团队版独立处理：

| type | 检测与凭证 | 文本模型清单 |
| --- | --- | --- |
| `agent-plan-team` | `GetSeatInfo` 的 Agent Plan 企业 Scene；真实 `SeatID`、Running 状态、`BizInfo` 与席位 Key | Agent Plan 清单 |
| `coding-plan-team` | `GetSeatInfo` 的 Coding Plan Scene；同样要求当前身份的 Running 席位 | Coding Plan 清单，需要当前 AccountID |

- 缺席位、非 Running、身份不符或接口错误时停止；“团队已购买”不等于“当前子用户已绑定席位”。需要分配席位时转 Plans，并另行确认具体目标。
- `--plan-tier` 不能制造席位、绕过 `GetSeatInfo` 或恢复已过期席位；不要改成个人版、普通池 Key 或另一个账号重试。
- 团队 Key 来自该席位，不能把 `auth apikey` 成功或普通池非空当团队凭证已就绪；反馈只报脱敏信息。

## 跟旧 `config init` 的差异

| 维度 | 旧 `config init` | 新 `profile create` |
|------|------------------|---------------------|
| `--type` | 没有此 flag，profile = (region × project) | 明确区分 platform 与四类 Plan（个人/团队） |
| `--plan-tier` | 不存在 | 个人 Plan 可手动声明；团队版仍以真实 Running 席位为准 |
| `--owner-trn` | 不存在；trn 不绑 profile | 新增，profile schema v1 切面字段 |
| API Key 拉取 | 单 key（用户传的 `--api-key`） | 全 list fetch + masking + 选 default |
| Project 字段 | 仅 `.env` `VOLCENGINE_ARK_PROJECT_NAME` | 进 profile 切面，`.env` 仅兜底 |

老脚本里 `arkcli config init ...` 仍然能跑（标 deprecated，0.2.x 删除），但 Agent 不应再引导用户用。
