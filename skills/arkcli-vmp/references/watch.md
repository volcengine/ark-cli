# 多 Endpoint 放量观察

1. 明确 EP 清单、观察时长和判据（例如用户给定的错误率或 E2E P99 上限）。没有判据时只报告变化，不擅自判定通过。
2. 用 snapshot/collect 保存放量前基线。需要严格对比时固定 start/end。说明基线流量、覆盖率和统计窗口。
3. 启动有界观察，立即首轮，之后每轮完成后等待 interval；duration 限制整个会话，window 不随 interval 改变。

```bash
arkcli doctor monitor snapshot --endpoint ep-a,ep-b --window 10m --range 1m --format json
arkcli doctor monitor watch --endpoint ep-a,ep-b --window 5m --range 1m --step 15s --interval 30s --duration 30m --summary-only --format jsonl
```

EP 清单也可以用 `--mine/--model/--name-regex/--status` 隐式解析，语义与 [采集契约](collect.md) 一致；范围只在启动时解析一次，watch 运行中不刷新。

## 阈值告警

`--alert-on` 在每轮采样后做确定性阈值判断，输出事件而非轮询全量日志：

```bash
arkcli doctor monitor watch --endpoint ep-a,ep-b --window 5m --range 1m --interval 30s --duration 30m --alert-on 'error_rate>0.05,latency>2.5' --alert-for 3 --format jsonl
```

- 规则格式 `metric<op>threshold`，逗号分隔；op 支持 `>` `>=` `<` `<=` `==`。
- metric 别名：`rpm`、`error_rate`、`latency`、`tpm`、`output_tpm`、`coverage`。`coverage` 读证据覆盖率；带 `_delta` 后缀（如 `rpm_delta<-100`）时与上一轮采样比较，首轮跳过。
- `--alert-for N`（1–60，默认 1）要求连续 N 轮超阈值才触发，避免单点抖动。
- 事件 `type=alert`（state=firing/resolved）只在状态切换时输出一次；`type=sample` 仍照常输出。end 记录的 `active_alerts` 是结束时未恢复的告警数。
- 观察结束仍有未恢复告警时进程以非零退出，便于编排层判定；`resolved` 后正常结束则退出码为 0。
- 告警判断基于每轮已采集的证据，不额外发查询；阈值未命中不代表服务健康，仍需按判据检查 coverage 和流量。

将输出保存为 `observation.jsonl`。每行 `vmp.watch.v1`，type=sample 包含 sequence/status/collection；type=end 表明 completed/cancelled/failed。completed 仅表示观察时间已结束，不是健康通过。取消时仍有 end，先前样本已保留，进程返回取消错误；未完成的当前查询不当作完整样本。若写文件/输出失败，终止记录也可能缺失，此时报告截断证据。

4. 根据用户判据逐 EP 检查样本。发现异常时记录时间、指标、阈值、覆盖率和对应查询，按 arkcli-doctor 的现有能力诊断该 EP。不要拿低覆盖率或零流量当作通过。刷新间隔和统计窗口重叠，所以相邻样本并非独立事件。
5. 报告基线、观察区间、是否完整结束、异常样本、诊断结果、未判定项和建议。没有 end、发生取消、数据缺失或无有效流量时说明限制；不宣称“已经持续监控”或“告警已生效”。
