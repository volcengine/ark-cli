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

## 参数证据与查询冲突

- `supported_params` 是当前版本的参数目录：已列出且 `support=true` 才能据此说明支持，并继续核对 `type/min/max/enum/required`；明确 `support=false` 要如实报告。**未列出不等于服务端一定拒绝、忽略或采用某个默认值**，也不能证明运行时会怎样处理这个参数。
- 参数目录、`api_support` 和真实调用是三类证据。回答 Responses/Chat/图像/视频 API 是否支持时，读取目标 API 对应项；不能拿另一 API 的支持、`thinking` 能力或模型名称代替。项缺失/未知就说明未确认。
- `search` 与 `get` 不一致时先核对已有证据中的模型名、精确版本、查询范围及缓存/告警信息。题设或本轮结果已明确同名、同版本、同范围时，直接报告该条件下仍存在冲突，不再以“先对齐版本”推迟回答，也不为重复确认已知事实再查询。分别列出“search 返回什么 / get 返回什么”，保留来源，不取并集、取较乐观的一项或静默覆盖。主版本的 search 元数据不能反证另一个显式版本的 get。
- 对“未列出的参数能否传、会被忽略吗、默认值是多少”，应转对应执行 Skill 核对 CLI/SDK 契约或官方文档；确需运行时验证时先说明会产生的调用/费用，并限于用户授权的同模型、同 API、同身份。没有验证就保持未知，不为答一个目录问题擅自发起推理。

## NotFound / 空结果的定位

精确 ID 优先直接 `get`；失败不能立即总结为模型不存在。先分清错误与正常空结果，再核对用户原始 ID、版本和 CLI 已支持的 DisplayName 归一化；不要手工用六位日期正则拆 ID。用户给的是 `cm-*` / `ep-*` / Plan 调用别名时，分别转自定义模型、接入点或套餐资源查询，不在公共目录反复试。

纯目录排查确需搜索时，用规范名称/族名做一次有界 `models search`，每次只调整一个有依据的名称或过滤条件；若用户查询历史/退役模型，核对 `--include-deprecated`，不要把默认隐藏解释成从未存在。候选不唯一时让用户选，不能自动换成名称相似的模型；最终说明查过的范围和仍未知的部分。创建/部署路径服从 owning Skill 的有界查询与完整捕获恢复流程，不能以“只查一次”为由丢弃可恢复的截断结果，也不借此循环 search/get。

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
