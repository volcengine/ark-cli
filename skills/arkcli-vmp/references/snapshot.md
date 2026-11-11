# 逐 Endpoint 快照

```bash
arkcli doctor monitor snapshot --endpoint ep-a,ep-b --window 5m --range 1m --format json
```

snapshot 是默认五个指标的 summary-only collection。它固定各查询的时间范围并保留每个 EP，输出含峰值时间、coverage 和 queries。阅读 [采集契约](collect.md)，检查状态后再作判断。available 不是健康状态；无流量无法验证时延或错误率。
