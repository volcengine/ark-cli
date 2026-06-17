# arkcli models custommodel get

详情 / 轮询 status。

## Usage

```bash
arkcli models custommodel get <id> [flags]
```

## Examples

```bash
# 完整详情
arkcli models custommodel get cm-xxxxx

# 只取关键字段（避免大块 JSON）
arkcli models custommodel get cm-xxxxx --transform id,name,active_endpoints,artifact_types

# 仅查 status（轮询用）
arkcli models custommodel get cm-xxxxx --transform status
```

## Flags

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<id>` | 是 | string | 自定义模型 ID `cm-xxxxx` |
| `--transform` | 否 | string list | 字段白名单（支持逗号分隔 `id,name` 或重复传），可选 `id` / `name` / `status` / `foundation_model` 等 CustomModel 字段，也可选 `active_endpoints` / `artifact_types`（注：本命令的 `--transform` 是字段子集，**不是**全局 GJSON 表达式） |

## `--transform` 说明

`custommodel get` 的 `--transform` 是本命令自己的字段白名单，用来减少额外查询和输出体积。它不是根命令的 GJSON 风格 `--transform`，不支持 `items.0.id` 这类嵌套路径表达式。

常用字段：

- `id,name,status,foundation_model`：只看基础身份和状态，不查额外接口
- `artifact_types`：额外查询可部署产物，返回 `endpoint_supported_methods` 和 `supported_inference_types`
- `active_endpoints`：额外查询当前引用该自定义模型的 endpoint

## Output

返回 JSON 详情。常见字段：`id` / `name` / `description` / `status` / `base_model` / `customization_type` / `source` / `active_endpoints`（当前引用本模型的 endpoint 列表）/ `artifact_types`（权重产物类型）/ `create_time` / `update_time`。

**注意**：删除前用 `--transform active_endpoints` 检查；非空说明仍有 endpoint 引用，删了会破坏推理链路。
