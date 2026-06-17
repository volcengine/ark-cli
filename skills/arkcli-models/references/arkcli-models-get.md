# models get

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

查看模型详情，聚合多个底层模型元数据接口返回完整信息。

## 命令

```bash
# 位置参数方式（推荐）
arkcli models get doubao-seed-2-0-pro-260215

# 指定版本
arkcli models get doubao-seed-2-0-pro-260215 260215

# 显式 flag 方式
arkcli models get --id doubao-seed-2-0-pro-260215 --version 260215
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<id>` / `--id` | 是 | string | 模型标识符，如 `doubao-seed-2-0-pro-260215` |
| `[version]` / `--version` | 否 | string | 模型版本覆盖，如 `260215` |

## 返回值

JSON 格式的模型详情，聚合自多个底层 API，包含模型名、版本、能力、定价、限流等信息。

## 常见错误

| 错误 | 原因 | 处理方式 |
|------|------|---------|
| 模型不存在 | ID 拼写错误或模型已下线 | 用 `arkcli models search` 确认模型名 |
| 认证失败 | 未登录或凭证过期 | 按 profile tenant 选择：火山 `arkcli auth login volc-sso` |

## 注意事项

- `id` 和 `version` 都支持位置参数和 flag 两种传入方式
- 该命令会聚合多个底层 API，可能比 `list` 稍慢

## 守卫

- 认证失败先回到 `arkcli auth status`，再按 profile tenant 选择登录命令
- 不确定模型 ID 时，先用 `arkcli models search` 或 `arkcli models list --name` 确认，避免反复调用不存在的 ID
- 只读查询优先，不要在模型详情排障中切换 profile 或修改本地配置，除非用户明确确认

## 参考

- [arkcli-models](../SKILL.md) -- models 全部命令
- [arkcli-shared](../../arkcli-shared/SKILL.md) -- 认证和全局参数
