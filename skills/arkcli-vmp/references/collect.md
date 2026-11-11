# 可信采集与前后对比

`collect` 为明确 EP 集合返回 `vmp.collection.v1`。先解析用户资源，把示例 ID 替换为实际 ID。

```bash
arkcli doctor monitor collect --endpoint ep-a,ep-b --window 30m --step 1m --range 1m --format json
arkcli doctor monitor collect --endpoint ep-a --start 2026-10-01T00:00:00Z --end 2026-10-15T00:00:00Z --step 1m --range 1m --summary-only --format json
arkcli doctor monitor collect --endpoint ep-a --metric request.rpm.by_api --metric request.duration.quantile --window 1h --range 5m --format json
arkcli doctor monitor collect --endpoint ep-a --metric request.rpm.by_api,request.tpm.input --window 1h --range 5m --format json
```

可用 catalog ID：`request.rpm.by_api`、`request.error.rate`、`request.duration.quantile`、`request.tpm.input`、`request.tpm.output`、`quota.purchased.tpm`。最后一项需显式选择。`snapshot` 返回相同契约的默认五项摘要。

## 隐式 Endpoint 范围

不想逐个列 ID 时，用 `--mine`、`--model`、`--name-regex`、`--status` 让 CLI 在控制面解析 EP 集合；四个条件可组合（AND），与 `--endpoint` 互斥：

```bash
arkcli doctor monitor collect --mine --model doubao-seed-1.6 --status running --window 30m
arkcli doctor monitor snapshot --model doubao-seed-1.6 --name-regex '^prod-' --format json
```

- 解析在命令开始时执行一次（watch 不会逐轮刷新），结果写入输出的 `scope`：`source`、`requested`、`resolved`、`resolved_at`；`resolved` 即本次实际查询的 EP 清单。
- `--mine` 优先使用服务端 `sys:ark:createdBy` 标签过滤；无法取得 ARK IAM UserID 的降级身份会退回客户端用户名后缀匹配，标签缺失或零匹配时直接报错，不会退化成全项目扫描。
- `--status` 接受 `Abnormal/Running/Scheduling/Stopped`，逗号分隔。
- 解析结果按 ID 排序；超过 50 个时截断，`scope.truncated=true`、`scope.skipped[]` 记录未纳入的 EP。匹配为零时报错，不静默查空。
- 范围一旦解析，报告里必须引用 `scope` 和时间；不要把它当作“项目全量”。

限制：1–50 个明确 EP ID，最大 744h，step 1s–1h，range 1m–1h，均需整秒；最多两百万个预期采样。多批保持相同 start/end。start/end 均向下对齐到 Unix epoch 的 step 网格（VMP 长跨度返回 epoch 对齐采样，短跨度沿用请求起点；统一网格避免丢点），实际范围以 window 为准。每片最多 360 个 step，边界重叠去重，不扩大统计窗口。PromQL 中固定保留 Endpoint 维度，直方图保留 le，TPM 已转成 tokens/min，RPM 按 range 归一化。

## 负载与安全

查询数按 `metric × 时间分片` 计算，EP 只在 PromQL 的 `endpoint=~"^(...)$"` 里聚合，所以 30 个 EP 和 1 个 EP 的查询数相同。执行前先看计划：

```bash
arkcli doctor monitor collect --endpoint ep-a,ep-b --window 30m --plan
```

`--plan` 只输出 `vmp.query_plan.v1`（queries、max_samples、estimated_seconds），不访问 VMP。默认 `--max-qps 2` 做客户端限速；429/5xx 按 1s、2s 有界退避，最多 3 次，其它错误立即失败。长跨度（例如 744h × 1m）会分成 56 片/指标，先用 `--plan` 估算，必要时降 step 或缩范围。不要把 `--max-qps` 当成绕过工作区配额的开关。

保存结果时可使用 shell 重定向到新的任务文件。不要覆盖之前的证据。输出字段：

- `workspace_id`、`endpoints`、`window`、`range`：实际查询范围。
- `results[]`：逐 EP、逐指标的 `unit`、`status`、`coverage`、`observed_points`、`expected_points`、`peak{ts,value}`、`latest`、`sample_mean`、可选 `points`。
- `queries[]`：每片实际 PromQL、时间、状态和错误。缺片返回 partial；所有查询失败会先输出证据再以非零退出。
- 空数据没有 peak/latest/sample_mean。`--summary-only` 只省略 points，保留证据和峰值时间。

前后对比：优先用 `arkcli doctor monitor compare`，它冻结两个等长、不重叠窗口，逐 EP 计算 `sample_mean`/峰值的 delta 与 `pass|violation|inconclusive` 判定；`--min-coverage`、`--max-error-rate-delta`、`--max-latency-delta-pct`、`--min-rpm-ratio` 随输出回显。为避免 range 回看跨越发布边界，发布后区间应从“发布时间 + range”开始。只有需要自定义规则（例如按 API 交叉、TPM 专项）时才用两次 `collect` 保存原始证据后在 Agent 侧计算。错误码分布用 `arkcli doctor monitor errors`，不要自己拼 `increase_pure`。不要把不同 EP 的 ratio 平均成项目错误率，也不要把一分钟峰值当作单请求值。缺数据的维度结论为不可判定。

报告至少包含：目标、profile 范围、EP 清单来源和解析时间、两个区间、step/range、指标单位、覆盖率、峰值时间、变化、原始证据路径、判断限制。
