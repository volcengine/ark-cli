# 异常下钻（vmp-incident-drilldown）

## 何时使用

放量盯盘、每日巡检或用户已经发现异常，需要继续定位到具体 EP、错误类型、时延或配额压力。

## 必填输入

- 异常 EP、时间点或时间窗。
- 已知指标和阈值。
- 是否允许读取 Trace/日志；可用的 request_id 或错误码。

## 执行

1. 先确认异常事实：用 `collect` 在窄窗口复现，保存 `evidence/anomaly.json`。不要直接沿用摘要里的聚合值。
2. 单 EP 诊断：按 `arkcli-doctor` 的真实入口执行 `arkcli doctor infer-endpoint <ep>`，检查状态、用量、错误率、配额压力。
3. 错误分布：先用原子下钻命令拿逐 EP 错误码 Top N、请求量和排查建议：

```bash
arkcli doctor monitor errors --endpoint ep-a --window 1h --top-codes 5 --format json > evidence/errors.json
```

`vmp.errors.v1` 的 `results[].codes[]` 含 `code`、`http_status_code`、`api`、`count`、`share`；`hints[]` 是按错误码规则生成的建议，不是根因结论。需要自定义分组（例如按 model/api 交叉）时，再用 `arkcli doctor metrics run --id request.error.distribution --param by=...`。错误率上升不等于限流，必须看错误码分布。
4. 异步/离线请求：显式查询 `async.gateway.*`；默认五项不含这些指标。
5. 需要 Trace/日志时，只使用宿主实际可用的工具；没有就保留为待补充证据。
6. 区分事实、假设和未知；不要从 EP 聚合指标推断某一次请求的根因。
7. 输出时间线、证据、假设、下一步；本 Skill 只读，不改配额、不切模型、不停止 EP。

## 判定

- `confirmed`：有可复查的指标/错误分布/诊断证据支持。
- `hypothesis`：有线索但缺少关键证据。
- `inconclusive`：数据不足、覆盖不足或无法关联单次请求。

## 报告模板

- 异常时间线与 EP
- 复现命令与证据路径
- 指标/错误码分布
- 诊断输出摘要
- 事实 / 假设 / 未知
- 建议的下一步与所需授权

## 失败处理

- 无法复现：说明窗口和口径，不宣布“已恢复”。
- 只有 request_id：按 `arkcli-doctor` 的边界说明无法反查单条日志，请求补充错误码/资源 ID/时间。
- 需要写操作：转对应能力 Skill，并先给出影响和确认点。

## 示例提示词

“ep-a 刚才错误率突然到 5%，帮我看是限流、超时还是模型错误，给出证据和下一步。”
