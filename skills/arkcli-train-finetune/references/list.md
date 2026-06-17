# 查询精调任务列表

本 reference 只处理多任务列表和筛选。用户指定了一个 job id 时改读 [`manage.md`](manage.md)。

## 命令与常用过滤

```bash
arkcli train finetune list [flags]
```

| 参数 | 说明 |
|---|---|
| `--name` | 按任务名模糊过滤 |
| `--phase` | 按阶段过滤，可重复 |
| `--customization-type` | 按后端训练类型过滤，可重复；不确定枚举时查 `--help` |
| `--create-time-after`、`--create-time-before` | 按创建时间过滤，使用 RFC3339 |
| `--page-size`、`--page-number` | 手动分页 |
| `--page-all`、`--page-limit` | 自动分页；必须设置合理上限 |
| `--sort-by`、`--sort-order` | 排序 |
| `--transform` | 提取单字段；复杂投影用 JSON 工具处理 |

## 流程

1. 查看当前版本参数：

```bash
arkcli train finetune list --help
```

2. 根据用户条件组合过滤：

- 名称
- 任务阶段
- 训练方法
- 创建时间范围
- 分页和排序

3. 用户要求“全部”时使用当前 CLI 支持的自动分页参数，并设置合理页数上限。不要无界抓取。

4. 输出精简表格或摘要，优先包含：

- job id
- 名称
- 训练方法
- 当前阶段
- 创建时间或更新时间

不要为列表中的每个任务自动调用 `get`。只有用户要求详情，或列表字段不足以回答问题时，才查询对应 job。

## 命令骨架

```bash
arkcli train finetune list
arkcli train finetune list --name <keyword>
arkcli train finetune list --phase <phase>
arkcli train finetune list \
  --create-time-after <RFC3339> \
  --create-time-before <RFC3339>
```

阶段和训练方法枚举必须从当前 `--help` 获取，不要使用静态清单。

## 时间处理

- 用户使用“今天、昨天、最近一周”等相对时间时，按当前 profile/用户时区换算为明确的 RFC3339 边界。
- 在结果中写出实际使用的绝对时间范围，避免时区歧义。

## 空结果

空结果时说明实际过滤条件，并建议放宽最可能过严的一项。不要把空列表解释为没有权限，除非 CLI 明确返回鉴权或权限错误。
