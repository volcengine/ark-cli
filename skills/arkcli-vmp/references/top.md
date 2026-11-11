# 明确范围内的峰值排行

```bash
arkcli doctor monitor top --endpoint ep-a,ep-b --metric rpm --limit 5 --window 1h --format json
```

metric 可选 rpm、error_rate、latency、tpm（input）、output_tpm；limit 1–50。返回 vmp.ranking.v1，order=observed_peak_desc，items 为先排序后截取的结果，evidence 保留全范围采集。相同峰值按 EP ID 稳定排序。空数据排后，无数值不会填零。

只按已解析的 EP 排名；不支持隐式全工作区、model/api 分组或把不同 EP 的峰值求和当成同时峰值。latency 是 E2E P99 观测点的峰值，并非窗口整体 P99。
