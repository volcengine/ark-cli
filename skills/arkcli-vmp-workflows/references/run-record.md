# 运行记录（arkcli.workflow.run.v1）

每个工作流 run 写一份 `run.json`，作为报告、调度和复盘的事实源。

```json
{
  "schema": "arkcli.workflow.run.v1",
  "workflow": "vmp-rollout-watch",
  "run_id": "20261108T153000+0800-abc123",
  "started_at": "2026-11-08T15:30:00+08:00",
  "finished_at": "2026-11-08T16:00:00+08:00",
  "status": "completed",
  "profile": "prod",
  "product": "volc",
  "scope": {
    "endpoints": ["ep-a", "ep-b"],
    "inventory_source": "user-provided",
    "resolved_at": "2026-11-08T15:29:50+08:00"
  },
  "window": {
    "start": "2026-11-08T15:30:00+08:00",
    "end": "2026-11-08T16:00:00+08:00",
    "step": "15s",
    "range": "1m"
  },
  "criteria": {
    "error_rate": {"op": ">", "value": 0.01},
    "latency_p99_ms": {"op": ">", "value": 2000}
  },
  "evidence": [
    "evidence/baseline.json",
    "evidence/watch.jsonl"
  ],
  "result": {
    "verdict": "violation",
    "summary": "ep-b error rate crossed 1% at 15:47:30+08:00"
  },
  "side_effects": [
    {
      "type": "feishu_message",
      "target": "oc_xxx",
      "idempotency_key": "vmp-rollout-20261108",
      "result": {"ok": true, "message_id": "om_xxx"},
      "verified": true
    }
  ],
  "errors": [],
  "next_steps": ["check ep-b upstream latency", "decide whether to pause rollout"]
}
```

## 字段规则

- `status`：`completed` / `partial` / `failed` / `cancelled`。
- `verdict`：`pass` / `violation` / `inconclusive` / `confirmed` / `hypothesis` / `planned`。
- `evidence`：相对 run 目录的路径；不得引用不存在的文件。
- `side_effects`：没有执行就写空数组；失败写 `verified=false` 和错误。
- `errors`：保留原始错误摘要，不只写“失败”。
- `run_id`：时间戳 + 随机后缀；同一任务重跑生成新 ID，不覆盖旧目录。
