# 定时执行

## 定位

CLI 没有内建调度器。定时巡检由宿主（cron、CI、飞书任务、Agent 宿主）触发；本文件定义任务契约，避免每个工作流各写一套调度逻辑。

## 任务规格

```json
{
  "schema": "arkcli.workflow.job.v1",
  "workflow": "vmp-daily-digest",
  "profile": "prod",
  "product": "volc",
  "scope": {"type": "project", "endpoint_source": "control-plane"},
  "lookback": "24h",
  "thresholds": {"error_rate": 0.05, "p99_ms": 2000},
  "timezone": "Asia/Shanghai",
  "schedule": "0 10 * * *",
  "recipients": ["oc_xxx"],
  "idempotency_key": "vmp-daily-{date}",
  "max_runtime": "20m",
  "on_failure": {"notify": ["ou_xxx"], "keep_evidence": true}
}
```

## 执行约定

1. 每次运行创建独立 `runs/<workflow>/<run-id>/`，写 `run.json`。
2. 同一 workflow + profile + 日期只允许一个成功 run；用锁文件或幂等键去重。
3. 超时、查询失败、推送失败分别记录；不要用一个 `failed` 覆盖所有原因。
4. 失败时保留证据并通知失败接收人；不要发送“全量正常”的摘要。
5. 时区、下次运行时间、job id 必须回读并记录；没有回读就不算注册成功。

## cron 示例

```cron
0 10 * * * cd /path/to/workdir && flock -n runs/vmp-daily.lock ./run-vmp-daily.sh >> runs/cron.log 2>&1
```

`run-vmp-daily.sh` 是宿主包装脚本，职责固定为：读取 job.json → 调用 `arkcli doctor monitor` → 写 run.json/report.md → 调用飞书适配器 → 按退出码通知。脚本不复制采集逻辑，也不在 CLI 外拼 PromQL：

```bash
#!/usr/bin/env bash
set -euo pipefail
run_id="$(date +%Y%m%dT%H%M%S)-$(openssl rand -hex 3)"
dir="runs/vmp-daily/$run_id"; mkdir -p "$dir/evidence"
eps="$(jq -r '.scope.endpoints | join(",")' job.json)"
window="$(jq -r '.lookback' job.json)"
# 1) 采集：失败也保留已写出的 JSON
arkcli doctor monitor snapshot --endpoint "$eps" --window "$window" --range 1m --format json > "$dir/evidence/snapshot.json"
arkcli doctor monitor errors   --endpoint "$eps" --window "$window" --top-codes 5 --format json > "$dir/evidence/errors.json"
# 2) 判定与报告：由 Agent 或 jq 生成 $dir/report.md；禁止把 partial/no_data 写成正常
# 3) 推送：幂等键 = workflow + date；发送后回读 message_id
lark-cli im +messages-send --chat-id "$(jq -r '.recipients[0]' job.json)" \
  --markdown @"$dir/report.md" --idempotency-key "vmp-daily-$(date +%F)"
# 4) 留档：run.json 写入 status/verdict/evidence/side_effects
```

## 失败处理

- 清单解析失败：不发送摘要，标 `failed`，通知失败接收人。
- 部分批次失败：标 `partial`，发送时明确缺失范围。
- 推送失败：数据结论保留，标 `side_effect_failed`。
- 同一日期重复触发：幂等键命中则跳过发送，不重复写文档；`flock` 防止并发重入。

## 边界

- 本文件不承诺 ArkCLI 内建调度、任务 ID 或运行历史。
- 飞书任务/日历只作为宿主触发入口；其权限和可用性单独验证。
