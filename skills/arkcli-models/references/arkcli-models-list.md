# models list

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

分页列出火山引擎 ARK 平台的公共基础模型清单，支持按模态筛选和排序。它适合做公共目录的全量枚举和统计；如果用户是在找"哪个模型适合某任务"，仍优先使用 `arkcli models search`。

## 命令

```bash
# 列出所有模型（默认分页）
arkcli models list

# 按模态筛选
arkcli models list --modality text

# 按名称做不区分大小写的子字符串筛选
arkcli models list --name doubao

# 分页控制
arkcli models list --page-size 20 --page-number 2

# 排序（sort-order 必须是 Asc/Desc，首字母大写）
arkcli models list --modality text --sort-by UpdateTime --sort-order Desc

# 拉取完整清单，供本地脚本统计/过滤
arkcli models list --page-all --sort-by CreateTime --sort-order Desc --format json
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--modality` | 否 | string | 按模态筛选：`text` / `image` / `video` / `audio` / `embed` |
| `--name` | 否 | string | 按模型名做不区分大小写的子字符串筛选 |
| `--page-number` | 否 | int | 页码（>=1） |
| `--page-size` | 否 | int | 每页数量 |
| `--sort-by` | 否 | string | 排序字段，如 `UpdateTime`、`CreateTime` |
| `--sort-order` | 否 | string | 排序方向，合法值：`Asc` / `Desc`（首字母大写） |

## 公共目录与账号资产边界

`models list` 枚举公共基础模型目录，不是账号自定义模型或已部署 Endpoint 清单。

- “我的自定义模型 / 最近 7 天创建了多少自定义模型”转 [arkcli-custommodel](../../arkcli-custommodel/SKILL.md)，使用 `arkcli models custommodel list --mine --page-all --format json`；用户明确要求账号全部资产时才去掉本人范围。统计所有创建记录时不额外加 `--ready` 丢掉未就绪项。
- 最近创建统计基于该资产清单实际返回的创建时间，在完整、未截断结果上按用户时区/时间窗口过滤；字段缺失、无法解析或分页不完整就披露未知/部分统计。不得从公共目录的 `model_type`、名称、tags 或 `CreateTime` 猜账号资产归属。
- 已部署接入点用 `infer endpoint list`；当前 Profile 可调用资源用 `resources list`。它们和公共目录不是同一集合。
- Plan 支持哪些模型转 `plans model-list --plan <plan>`；目录中找不到套餐路由别名，不证明套餐不能使用该别名。
- `models list` 缺少某能力字段，不证明支持或不支持；能力问题用 `models search/get` 的精确模型/版本元数据。公共目录零条不证明账号没有自定义模型或 Endpoint。

`--transform` 只适合字段提取（如 `items.#`、`items.#.name`），不是完整 jq，不支持日期运算。公共目录统计也应先保留完整 JSON，再本地投影/过滤，不用截断后的样本冒充总量。

## 返回值


JSON 格式的分页结果，顶层字段包括 `page_number`、`page_size`、`total_count`、`items`；每个 item 通常包含 `name`、`display_name`、`primary_version`、`access_type`、`foundation_model_tag` 等字段。

## 常见错误

| 错误 | 原因 | 处理方式 |
|------|------|---------|
| 空结果 | `--name` 子字符串筛选无命中 | 核对名称片段与查询范围，必要时用 `models search` |
| 认证失败 | 未登录或凭证过期 | 运行 `arkcli auth login volc-sso` 重新建立 Volc 身份 |

## 注意事项

- `--name` 是不区分大小写的子字符串筛选；精确详情使用 `models get <id>`
- 结合 `--transform` 可提取特定字段，如 `--transform 'items.0.name'`
- 先选对公共目录/自定义资产/Endpoint/Plan 的产品命令，再考虑分页和本地过滤；不要因缺少 filter 就探 Raw API 或改查另一类资源

## 参考

- [arkcli-models](../SKILL.md) -- models 全部命令
- [arkcli-shared](../../arkcli-shared/SKILL.md) -- 认证和全局参数
