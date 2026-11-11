# 放量盯盘（vmp-rollout-watch）

## 何时使用

用户要在一段时间内盯住一组 EP，按阈值判断放量是否继续，并在异常时收到通知。

## 必填输入

- EP 集合（1–50/批；更多分批，同一 start/end）。
- 观察时长、轮询间隔、每轮 window/step/range。
- 指标与阈值（例如错误率 > 1%、E2E P99 > 2s、输出 TPM > 配额 80%）。
- 接收方（飞书群/个人/文档）和是否授权发送。

缺阈值时只报告变化，不判定通过。缺接收方时只落本地报告。

## 执行

1. 读 [`../arkcli-vmp/SKILL.md`](../../arkcli-vmp/SKILL.md)、[watch](../../arkcli-vmp/references/watch.md) 与 [collect](../../arkcli-vmp/references/collect.md)。
2. 解析 EP 和身份；保存清单来源与解析时间。
3. 基线：

```bash
arkcli doctor monitor snapshot --endpoint ep-a,ep-b --window 10m --range 1m --format json
```

保存为 `evidence/baseline.json`。

4. 启动观察。把用户阈值直接翻译成 `--alert-on`，让 CLI 在每轮采样后做确定性判断；事件流里只有状态切换才出 `alert` 记录：

```bash
arkcli doctor monitor watch --endpoint ep-a,ep-b --window 5m --range 1m --step 15s --interval 30s --duration 30m --summary-only --alert-on 'error_rate>0.01,latency>2.0' --alert-for 3 --format jsonl
```

保存为 `evidence/watch.jsonl`。每条 sample 带 sequence/status/collection；`type=alert` 带 endpoint/metric/value/threshold/consecutive_samples/state（firing/resolved）；end 为 completed/cancelled/failed，并带 `active_alerts`。结束时仍有 firing 告警，命令以非零退出，编排层据此触发通知或停止放量；退出码非零时要先读 `active_alerts` 和事件流，不要直接归因于采集失败。

EP 集合若来自“我的/某模型/名称前缀”，用 `--mine/--model/--name-regex/--status` 隐式解析（与 `--endpoint` 互斥）；解析结果在 `scope` 里，启动时只解析一次。阈值语义见 [watch](../../arkcli-vmp/references/watch.md)。

## 负载与安全

watch 每个 sample 发 `metrics × 时间分片` 次查询，EP 数量不放大查询数。默认 `--max-qps 2`；`--interval` 不得低于 `metrics / max-qps`（5 个指标、2 QPS → 至少 2.5s，建议取整到 3s），低于下限直接报错，不会先打满再限流。每轮采集完成后才等待 interval，不会并发堆积。先用 `--plan` 看每轮和总查询数：

```bash
arkcli doctor monitor watch --endpoint ep-a,ep-b --interval 30s --duration 30m --plan
```

429/5xx 有界退避；连续失败按 partial/failed 记录，不要把限流误判成业务异常。多进程/多用户同时盯盘时，各自遵守自己的 `--max-qps`，不要对同一工作区开很多 1s 间隔的 watch。

5. 每轮按判据检查。异常样本记录 sequence、时间、EP、指标、值、阈值、coverage 和 query。
6. 异常 EP 用 `arkcli doctor infer-endpoint <ep>` 或 `arkcli doctor metrics` 下钻；不要用同一份聚合值反复推断。
7. 观察结束后检查 end 记录；`completed` 只表示时间结束，不表示健康。
8. 生成 `report.md`，按 [飞书集成](feishu.md) 推送。

## 判定

- `pass`：所有轮次、所有 EP 的指标都在阈值内，且 coverage 足够；零流量不能算 pass。
- `violation`：任一 EP 在有效样本上越过阈值；记录首次越界时间和持续轮次。
- `inconclusive`：缺 end、partial/no_data、coverage 不足、阈值口径不明确或流量不足。

## 报告模板

- 目标与判据
- EP 清单与解析时间
- 基线（窗口、流量、覆盖率）
- 观察范围与轮次
- 逐 EP 结果（pass/violation/inconclusive）
- 首次异常时间线、指标、值、阈值、证据路径
- 诊断与建议
- 副作用回读（消息 ID/文档 token）

## 失败处理

- watch 没有 end：报告截断，保留已有样本，不宣称持续监控。
- 推送失败：保留 `report.md`，重试一次；仍失败则标记 `side_effect_failed`，不写“已通知”。
- 用户中途改阈值：新开 run，不覆盖旧证据。

## 示例提示词

“盯住 ep-a、ep-b 半小时，每 30 秒看一次最近 5 分钟，错误率超过 1% 或 P99 超过 2 秒就推到这个群。”
