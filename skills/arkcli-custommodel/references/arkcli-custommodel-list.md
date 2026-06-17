# arkcli models custommodel list

翻页 + 多维过滤。

## Usage

```bash
arkcli models custommodel list [flags]
```

## Examples

```bash
# 列出全部（默认分页）
arkcli models custommodel list

# 只看自己的
arkcli models custommodel list --mine

# 按状态过滤（支持逗号分隔或重复传）
arkcli models custommodel list --statuses ready
arkcli models custommodel list --statuses ready,processing

# 按训练类型过滤
arkcli models custommodel list --customization-types lora,sft

# 按 base 模型名 / ID 过滤
arkcli models custommodel list --base-models doubao-seed-1-6
arkcli models custommodel list --base-model-ids fm-xxxxx

# 模糊搜索（命中 name / id / base 模型 display name）
arkcli models custommodel list --search my-finetune

# 分页与排序
arkcli models custommodel list --page 1,20 --sort-by CreateTime --sort-order Desc

# 自动翻页
arkcli models custommodel list --mine --page-all --page-delay 500
```

## Flags

> 所有 `string list` 类型的多值过滤项都支持两种等价写法：逗号分隔（`--statuses ready,processing`）或重复传（`--statuses ready --statuses processing`），结果一致。

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--mine` | 否 | bool | 只返回当前 SSO 子用户创建的自定义模型 |
| `--statuses` | 否 | string list | `preparation` / `processing` / `ready` / `failed` / `exporting` / `exportfailed` |
| `--customization-types` | 否 | string list | `lora` / `sft` / `dpolora` / `dpo` / `grpolora` / `grpo` / `ppo` / `opdlora` / `opd` / `pretrain` |
| `--supported-customization-types` | 否 | string list | 按"基座模型支持哪些训练类型"反向过滤 |
| `--base-models` | 否 | string list | 按基座模型名过滤 |
| `--base-model-ids` | 否 | string list | 按基座模型 ID 过滤 |
| `--sources` | 否 | string list | `import` / `customization` |
| `--source-jobs` | 否 | string list | 按模型微调任务 ID 过滤 |
| `--search` | 否 | string | 模糊搜索 name / id / base 模型 display name |
| `--page` | 否 | string | 分页表达式 `<number,size>`，例：`1,10`（不接受空格） |
| `--sort-by` | 否 | string | 排序字段，默认 `CreateTime` |
| `--sort-order` | 否 | string | 排序方向，默认 `Desc` |

## Output

返回 JSON 分页结果，每个 item 通常含 `id` (cm-xxxxx) / `name` / `status` / `base_model` / `customization_type` / `source` / `create_time` / `update_time`。

