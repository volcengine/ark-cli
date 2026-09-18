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

JSON 格式的模型详情，聚合自多个底层 API，包含模型名、版本、能力、定价、限流等信息。缓存能力看顶层 `cache_types`：可能包含 `explicit_cache` / `implicit_cache` / `session_cache` / `prefix_cache`；`capabilities.caching` 仅为旧版兼容字段，不用于判断全部缓存类别。

`supported_params` 按当前详情的精确模型版本补充。主版本可复用本地新鲜 ArkModels 元数据缓存；缓存不可用、显式非主版本或 `--no-cache` 时，CLI 改用精确版本的 `ListModelMetaDatas` 查询。`models get` 不会为了补该字段新增 ArkModels 网络请求，避免触碰其低 QPS 限制。`--no-cache` 仍会绕过 CardView 与 ArkModels 元数据缓存。

- 字段缺失或空数组：该版本当前没有可用参数目录，`--transform supported_params` 可能显示 `null`。
- 上游字段存在但 JSON 损坏：CLI 在 stderr 输出带模型名和版本的 `warn: model supported_params enrichment failed: ...`，stdout 仍返回其余模型详情。

## 价格与权益

价格来自新价格服务，输出为 `pricing.model_name`、`pricing.prices` 和可选 `pricing.dimension_attributes`。旧 `pricing.charge_items` / `pricing.multi_charge_items` 已删除，脚本和 Agent 必须迁移，不能继续按旧 `type` 筛选。

```bash
# 查看价格及原有开通状态、权益
arkcli models get doubao-seed-2-0-pro-260215 --format json --transform pricing

# 只查看统一价格数组（不包含免费额度或开通状态）
arkcli models get doubao-seed-2-0-pro-260215 --format json --transform pricing.prices
```

| 字段 | 消费规则 |
|------|----------|
| `service_type` | 区分 `infer`、`fast-infer`、`flex-infer`、`batch-infer`、`infer-storage`、`finetuned-infer`、`finetuned`；训练与推理不能混选 |
| `label` | 计费项标识，必须结合服务与适用条件判断；不等同于旧 `type` |
| `dimensions` | 完整适用条件，含上下文档位、时段、分辨率等；`ranges` 可提供范围描述，空串维度值也有意义 |
| `usage_unit` / `unit_code` / `usage_period` | 用量单位、价格单位及可选周期；原样读取，不固定按千 Token 换算，不猜币种 |
| `price` / `original_price` | 可空数字；`null` 是缺价，`0` 是该条件下零价，不能互相替代 |
| `discount_price_start_time` / `discount_price_end_time` | 可选折扣时间；说明适用时间，不自行补值 |
| `group_index` | 本次响应内的原始分组下标，不是跨请求稳定 ID；不同组不能直接合并 |
| `dimension_attributes` | 模型级维度属性；保留了未被价格行引用的属性，可用于理解范围 |

- 先按业务服务、计费项和全部适用条件筛选，再展示价格；多条仍匹配时列出差异，不能默认取第一条或最低价。未知维度无法解释时说明条件不明，不估算总价。
- `pricing.prices: []` 表示当前查询没有价格，不证明模型免费、未开通或不存在。价格请求失败时命令返回错误，不回退旧价。
- `pricing.state`、`inference_free_usage`、`resource_pack_items`、`sub_services` 及其他原有非价格字段继续由旧权益接口提供，字段路径不变；权益缺失不能按零额度或未开通解释。
- 价格按详情解析后的基础模型名查询；`--version` 不会变成价格请求条件。当前使用模型广场默认维度，地域为 global，不能把 profile 的物理地域当作价格维度。
- 本次变更只适用于 `models get`；`pricing models`、精调查价和估价仍使用各自的现有契约，不能将这里的新字段套用到它们的输出。


## 常见错误

| 错误 | 原因 | 处理方式 |
|------|------|---------|
| 模型不存在 | ID 拼写错误或模型已下线 | 用 `arkcli models search` 确认模型名 |
| 认证失败 | 未登录或凭证过期 | 运行 `arkcli auth login volc-sso` 重新建立 Volc 身份 |

## 注意事项

- `id` 和 `version` 都支持位置参数和 flag 两种传入方式
- 该命令会聚合多个底层 API，可能比 `list` 稍慢

## 守卫

- 认证失败先回到 `arkcli auth status`，再运行 `arkcli auth login volc-sso`
- 不确定模型 ID 时，先用 `arkcli models search` 或 `arkcli models list --name` 确认，避免反复调用不存在的 ID
- 只读查询优先，不要在模型详情排障中切换 profile 或修改本地配置，除非用户明确确认

## 参考

- [arkcli-models](../SKILL.md) -- models 全部命令
- [arkcli-shared](../../arkcli-shared/SKILL.md) -- 认证和全局参数
