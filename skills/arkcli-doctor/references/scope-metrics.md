# `arkcli doctor metrics`(具名 PromQL 查询)

> 这是 [`arkcli-doctor`](../SKILL.md) 的 **metrics domain reference**。当用户/Agent 需要**直接拿一条具体可观测指标的数值**(过去 1h 错误码 429 在 Endpoint X 上的累计次数 / 模型 Y 的 P99 时延 / Prompt Cache 命中率 / 异步任务成功率 / ...)而不是走"endpoint 体检 / model 体检"的整套诊断时,读这里。
>
> **CRITICAL — 开始前 MUST 先用 Read 工具读取**:
> - [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md)(认证闸门、命令选择顺序)
> - [`../SKILL.md`](../SKILL.md)(doctor 总入口与路径决策表)

## 它解决什么

`arkcli doctor metrics` 是一份**编入 binary 的具名 PromQL 查询编目**。每条 entry 是一个可参数化的查询(template + params),覆盖 Ark 全部 21 个语义指标族 / 54 个原子指标 / 5 种计算方式(增量、速率、比值、分位数、瞬时值)——一共 36 条。

跟 `doctor infer-endpoint` / `doctor model` 的关系:

| 想要……                                            | 跑哪条                                                          |
|--------------------------------------------------|----------------------------------------------------------------|
| 一屏看清 ep-xxx 整体情况(状态 + 用量 + 错误率 + 配额)| `arkcli doctor infer-endpoint ep-xxx`                          |
| 一屏看清 model X 跨 endpoint 整体情况               | `arkcli doctor model X`                                        |
| **拿一个具体可观测指标的数值** 给 LLM 推理           | `arkcli doctor metrics <id> --param ...`                       |
| 临时跑一条编目里没有的 PromQL                       | `arkcli doctor metrics raw --promql '...'`                     |

后两条基础设施跟前两条共用同一条 VMP 链路(`vmp:QueryRange`),只是 PromQL 模板从内置 catalog 取而非诊断 scope 内部 hard-code。

## 边界:不做什么

- **不做诊断决断**——只返指标数值,不打"红黄绿"健康判定。要红黄绿走 `infer-endpoint` / `model` scope。
- **不做 PromQL 合成器**——catalog 收的是已经被运维验证过的具名查询。任意 metric × 任意 calc 组合不放开。
- **不做面板**——输出是 series JSON / scalar 数值 / 表格,不试图复刻 timeseries 图。
- **不跨工作区**——一次调用 = 一个工作区 = 一条查询。

## 第一步:看清编目

```bash
# 列出 36 条具名查询(json 一行一 entry)
arkcli doctor metrics list

# 拿单条 entry 的完整元数据(template / params / 默认值 / 单位 / 聚合提示)
arkcli doctor metrics describe request.error.rate
```

## 第二步:执行查询

bare form: `arkcli doctor metrics <id> [--param key=val ...] [--filter label=value ...]`

```bash
# 拿过去 1h Endpoint X 的 RPM, 按 api 拆
arkcli doctor metrics request.rpm.by_api --param endpoint=ep-202602xxx

# 模型 Y 过去 6h 的 P99 时延, 按 api+le 维度
arkcli doctor metrics request.duration.quantile \
  --param model=doubao-pro-32k \
  --param quantile=0.99 \
  --window 6h

# 错误码分布: 把 endpoint 限定在 ep-x, 用 --filter 过出仅 429
arkcli doctor metrics request.error.distribution \
  --param endpoint=ep-x \
  --filter code=429

# scalar 形态 (给 agent 直接拿数值, 单位与聚合策略由 catalog.aggregation_hint 决定)
arkcli doctor metrics request.error.rate \
  --param endpoint=ep-x \
  --format scalar
# → {"query_id":"request.error.rate","value":0.0234,"unit":"ratio","aggregation":"mean", ...}

# 表格 (面向人, 每条 series 一行: latest / mean / max)
arkcli doctor metrics request.rpm.by_api --format table
```

公共 flag:

| flag                    | 默认     | 说明                                                                |
|-------------------------|---------|---------------------------------------------------------------------|
| `--param key=val`       | —       | 重复绑定 catalog 声明的参数; 缺省走 entry default                     |
| `--filter label=value`  | —       | 重复, 注入额外 `label=~"value"` 过滤; 模板需声明 `${@filters}` 占位符 |
| `--window <duration>`   | `1h`    | 查询窗口, Go duration 语法 (不接受 `7d`)                             |
| `--step <duration>`     | 自动    | step, 默认 window/60 clamp 到 [10s, 60s]                            |
| `--workspace-id <uuid>` | 自动    | VMP 工作区 ID, 缺省走 `ListTelemetryConfigs` 反查                     |
| `--format <fmt>`        | `json`  | `json` / `scalar` / `table`                                         |

## 第三步:逃生口

catalog 不够用 → `arkcli doctor metrics raw --promql '<expr>'`(必须引用至少一个 `ark_*` 指标,挡掉 `node_*` / `up{}` 等无关 metric)。

```bash
arkcli doctor metrics raw \
  --promql 'sum(rate(ark_api_proxy_request_total{ark_endpoint="ep-x"}[5m])) by (api)' \
  --window 30m
```

## 输出形态

无论哪种 format,顶层结构都包含:

- `query_id` —— catalog entry id (raw 路径不返此字段)
- `promql` —— **渲染后实际跑的 PromQL**, 必返。出问题时拷出来直接到 VMP 控制台对一遍。
- `window` —— `{start, end, step}` (RFC3339 + step 字面量)
- 数据字段 (三选一):
  - `series` —— format=json, 全 series 全 points (Prometheus matrix 形态)
  - `scalar` —— format=scalar, `{value, unit, aggregation}` 单值
  - `table` —— format=table, 每条 series `{tags, latest, mean, max}` 一行

错误形态:

- 参数错误 / 未知 entry id / `--filter` 在不支持的 entry 上 → 退出码 2,JSON `{ok:false, error:{type, message, hint}}`。`hint` 会给 "did you mean: <候选 id>" 之类的纠错。
- 工作区未绑定且未传 `--workspace-id` → 退出码 1,提示绑 VMP 或显式传 ID。
- 上游 VMP 错误(网络 / 鉴权 / 工作区不存在)→ 透传。
- 命中数据 0 → 退出码 0,数据字段为空数组 / 零值。Agent 据此分支即可。

## 36 条具名查询编目(按 namespace 分组)

> **注**:这份清单是 catalog 当前快照,catalog 可能在仓库 PR 后扩展。**任何时候用 `arkcli doctor metrics list` 或 `describe <id>` 取最新元数据,不要依赖本表手抄的参数列表**。

### A. 用户面 API 代理 (`ark_api_proxy_*`)

| ID                              | calc_kind            | 单位        | 含义                                              |
|---------------------------------|----------------------|-------------|--------------------------------------------------|
| `request.rpm.by_api`            | increase_sum         | req/min     | 接口 RPM, 按 api 拆                               |
| `request.tpm.input`             | rate                 | tokens/sec  | 输入 Token 速率                                   |
| `request.tpm.output`            | rate                 | tokens/sec  | 输出 Token 速率                                   |
| `request.error.rate`            | ratio                | ratio       | 错误率 (code!~Success 占比)                       |
| `request.error.distribution`    | increase_sum         | count       | 错误码累计 (按 code, http_status_code, api 拆)    |
| `request.duration.quantile`     | histogram_quantile   | seconds     | 请求耗时分位数 (默认 P99)                          |
| `request.duration.avg`          | ratio                | seconds     | 请求平均耗时                                       |
| `request.token.input.avg`       | ratio                | tokens      | 平均输入 Token 数                                 |
| `request.token.output.avg`      | ratio                | tokens      | 平均输出 Token 数                                 |
| `request.token.detail`          | increase_sum         | tokens      | Token 细分(按 token_type 拆,常配 --filter 过滤)    |
| `request.fallback.count`        | increase_sum         | count       | 兜底模型触发次数                                  |

### B. Prompt Cache

| ID                          | calc_kind | 单位       | 含义                                  |
|-----------------------------|-----------|------------|--------------------------------------|
| `cache.hit.rate`            | ratio     | ratio      | hit_tokens / (hit_tokens + missed)   |
| `cache.hit.tokens`          | rate      | tokens/sec | 命中 Token 速率                       |
| `cache.missed.tokens`       | rate      | tokens/sec | 未命中 Token 速率                     |

### C. 流式吐字间隔

| ID                                       | calc_kind          | 单位     | 含义                                |
|------------------------------------------|--------------------|----------|------------------------------------|
| `request.stream.per_token.quantile`      | histogram_quantile | seconds  | 流式吐字间隔分位数 (排除首 token)    |
| `request.stream.per_token.avg`           | ratio              | seconds  | 平均流式吐字间隔                     |

### D. Realtime API

| ID                                      | calc_kind | 单位        | 含义                          |
|-----------------------------------------|-----------|-------------|------------------------------|
| `realtime.connections`                  | gauge     | connections | 在线连接数                    |
| `realtime.connection.duration.avg`      | ratio     | seconds     | 平均连接时长                  |
| `realtime.request.error.rate`           | ratio     | ratio       | Realtime 请求错误率           |

### E. Quota

| ID                       | calc_kind | 单位 | 含义                |
|--------------------------|-----------|------|---------------------|
| `quota.purchased.tpm`    | gauge     | tpm  | 已购 PTU/TPM 配额    |

### F. 异步网关 (`ark_user_ark_async_gateway_*`)

| ID                                          | calc_kind            | 单位      | 含义                              |
|---------------------------------------------|----------------------|-----------|----------------------------------|
| `async.gateway.request.count`               | increase_sum         | count     | 网关请求数 (按 action) — 该 metric 每个 label 组合是单点 series, rate() 全 0, 改用 increase() |
| `async.gateway.task.duration.quantile`      | histogram_quantile   | seconds   | 任务总耗时分位数 (仅成功任务)       |
| `async.gateway.task.duration.avg`           | ratio                | seconds   | 任务平均耗时                       |
| `async.gateway.task.success_rate`           | ratio                | ratio     | succeed / (succeed+failed+expired)|
| `async.gateway.task.failure.distribution`   | increase_sum         | count     | 失败原因累计 (按 volc_error_code) |
| `async.gateway.task.cancelled.count`        | increase_sum         | count     | 取消任务数                        |
| `async.gateway.task.expired.count`          | increase_sum         | count     | 过期任务数                        |
| `async.gateway.task.pending`                | gauge                | tasks     | 当前 pending 任务数               |
| `async.gateway.task.queued`                 | increase_sum         | tasks     | 排队任务累计 (按 api 拆)           |
| `async.gateway.task.queue_wait.quantile`    | histogram_quantile   | seconds   | 排队等待时间分位数                 |
| `async.gateway.concurrent_requests`         | gauge                | requests  | 并发请求数                         |
| `async.gateway.webhook.rate`                | rate                 | calls/sec | webhook 调用速率 (按 callback_url) |

### G. 内容生成 v2 (`ark_user_ark_content_generation_v2_*`)

| ID                                              | calc_kind            | 单位     | 含义                                 |
|------------------------------------------------|----------------------|----------|-------------------------------------|
| `content.gen.request.count`                    | increase_sum         | count    | 请求数 (按 model_service_type) — 该 metric 每个 label 组合是单点 series, rate() 全 0, 改用 increase() |
| `content.gen.request.error.rate`               | ratio                | ratio    | 错误率 (http_status_code!=200)      |
| `content.gen.request.duration.quantile`        | histogram_quantile   | seconds  | 请求耗时分位数                       |
| `content.gen.image.distribution.by_resolution` | increase_sum         | images   | 图像生成分辨率累计                   |

## Agent 路由提示

用户问什么 → 跑哪条:

- **"过去 N 时间错误率多少 / 错了多少次"** → `request.error.rate` (比例) 或 `request.error.distribution` (按错误码计数)
- **"P99 / P95 时延是多少"** → `request.duration.quantile` (`--param quantile=0.99`)
- **"流式吐字快不快"** → `request.stream.per_token.quantile` 或 `.avg`
- **"接入点 / 模型现在并发多少 / RPM 多少"** → `request.rpm.by_api` (近 1m 增量) 或 `async.gateway.concurrent_requests` (异步类)
- **"Prompt Cache 命中率"** → `cache.hit.rate`
- **"是不是触发了限流 (429 / RateLimitExceeded)"** → `request.error.distribution --filter code=429`
- **"异步任务成功率怎么样"** → `async.gateway.task.success_rate`
- **"图像生成都生成了哪些分辨率"** → `content.gen.image.distribution.by_resolution`
- **catalog 不覆盖的边角查询** → `arkcli doctor metrics raw --promql '...'`,记得 PromQL 必须引用 `ark_*` metric

## 设计深读

- 完整设计文档 [`docs/doctor-metrics.md`](../../../docs/doctor-metrics.md):catalog schema、参数白名单类型、`${@filters}` 占位符语义、安全/转义模型、与诊断 scope 的边界。
- catalog 源 [`internal/service/doctor/metrics/catalog.yaml`](../../../internal/service/doctor/metrics/catalog.yaml):36 条 entry 的 PromQL 模板与参数声明,运维 PR 即可扩展。
