# models list

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

分页列出火山引擎 ARK 平台的模型清单，支持按模态筛选和排序。它适合做全量枚举、统计和资产盘点；如果用户是在找"哪个模型适合某任务"，仍优先使用 `arkcli models search`。

## 命令

```bash
# 列出所有模型（默认分页）
arkcli models list

# 按模态筛选
arkcli models list --modality text

# 按名称精确筛选
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
| `--name` | 否 | string | 按模型名精确筛选 |
| `--page-number` | 否 | int | 页码（>=1） |
| `--page-size` | 否 | int | 每页数量 |
| `--sort-by` | 否 | string | 排序字段，如 `UpdateTime`、`CreateTime` |
| `--sort-order` | 否 | string | 排序方向，合法值：`Asc` / `Desc`（首字母大写） |

## 自定义模型 / 最近创建统计

当用户问"我的自定义模型"、"最近 7 天建了多少个自定义模型"、"列出来"时，这属于资产盘点，应使用 `models list`，不要切到 `arkcli api --list` 探 Raw API。

推荐流程：

1. 先确认认证状态：`arkcli auth status`
2. 拉完整清单：`arkcli models list --page-all --sort-by CreateTime --sort-order Desc --format json`
3. 如果当前版本提供自定义模型类型 flag，优先使用该 flag；如果没有，就在本地 JSON 中按可用字段过滤，例如 `model_type`、`type`、`customization_type`、`source_type`、`customized_tags`、`create_time`
4. 时间窗口没有服务端 flag 时，在本地按 `create_time` 做日期比较

示例：

```bash
arkcli models list --page-all --sort-by CreateTime --sort-order Desc --format json > /tmp/ark-models.json
python3 - <<'PY'
import json
from datetime import datetime, timedelta, timezone

data = json.load(open("/tmp/ark-models.json"))
items = data.get("items") or []
cutoff = datetime.now(timezone.utc) - timedelta(days=7)

def parse_time(value):
    if not value:
        return None
    value = value.replace("Z", "+00:00")
    try:
        return datetime.fromisoformat(value)
    except ValueError:
        return None

def is_custom(item):
    haystack = " ".join(str(item.get(k, "")) for k in (
        "model_type",
        "type",
        "customization_type",
        "source_type",
    ))
    tags = item.get("customized_tags") or []
    haystack += " " + " ".join(str(tag) for tag in tags)
    return "custom" in haystack.lower() or "customization" in haystack.lower()

matched = []
for item in items:
    created = parse_time(item.get("create_time") or item.get("CreateTime"))
    if created and created.tzinfo is None:
        created = created.replace(tzinfo=timezone.utc)
    if created and created >= cutoff and is_custom(item):
        matched.append(item)

print(json.dumps({
    "count": len(matched),
    "items": matched,
}, ensure_ascii=False, indent=2))
PY
```

`--transform` 只适合轻量提取和计数，例如 `--transform 'items.#'`、`--transform 'items.#.name'`。当前 transform 不是完整 jq，不支持日期运算或 `?(@.create_time)` 这类谓词；涉及时间窗口时用 `--format json` 后接 `python3` / `jq` 做客户端过滤。

## 返回值

JSON 格式的分页结果，顶层字段包括 `page_number`、`page_size`、`total_count`、`items`；每个 item 通常包含 `name`、`display_name`、`primary_version`、`access_type`、`foundation_model_tag` 等字段。

## 常见错误

| 错误 | 原因 | 处理方式 |
|------|------|---------|
| 空结果 | `--name` 精确匹配无命中 | 改用 `arkcli models search` 做模糊搜索 |
| 认证失败 | 未登录或凭证过期 | 按 profile tenant 选择：火山 `arkcli auth login volc-sso` |

## 注意事项

- `--name` 是精确匹配，如需模糊搜索请用 `arkcli models search`
- 结合 `--transform` 可提取特定字段，如 `--transform 'items.0.name'`
- 统计/盘点需求不要因为缺少某个服务端 filter 就退到 Raw API Explorer；优先 `--page-all --format json` 后本地过滤

## 参考

- [arkcli-models](../SKILL.md) -- models 全部命令
- [arkcli-shared](../../arkcli-shared/SKILL.md) -- 认证和全局参数
