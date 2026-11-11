# 每日巡检（vmp-daily-digest）

## 何时使用

用户要每天固定时间巡检一组 EP 或整个项目，生成摘要并推送到飞书。

## 必填输入

- profile、产品、项目范围或明确 EP 集合。
- 回看窗口（例如过去 24h）、指标、异常阈值。
- 推送时间、时区、接收方（群/个人/文档/Base）。
- 失败时的接收人。

## 执行

1. 解析 EP 清单：用户给 EP 时直接用；说“全项目”时按 `arkcli-vmp` 的清单流程读取控制面全部分页，保存来源与解析时间。
2. 按批采集，所有批次共享同一 start/end：

```bash
arkcli doctor monitor snapshot --endpoint ep-a,ep-b --window 24h --range 1m --format json > evidence/snapshot.json
arkcli doctor monitor top --endpoint ep-a,ep-b --metric error_rate --limit 10 --window 24h --format json > evidence/top-error.json
arkcli doctor monitor errors --endpoint ep-a,ep-b --window 24h --top-codes 5 --format json > evidence/errors.json
```

- `snapshot` 给出逐 EP 的默认五项摘要（RPM、错误率、P99、输入/输出 TPM）。
- `top` 用于快速定位异常 EP。
- `errors` 给出逐 EP 错误码 Top N、错误率与规则化排查建议；错误率为零时 `codes` 为空、`status=available`，不是缺数据。
3. 需要 TPM 容量判断时追加：

```bash
arkcli doctor monitor capacity --endpoint ep-a,ep-b --window 24h --step 1m --range 1m --format json > evidence/capacity.json
```

`capacity` 的 `status` 为 `ok|warning|critical|inconclusive`；`inconclusive` 只表示配额或覆盖不足，不表示健康。
4. 按阈值筛选异常：错误率、E2E P99、TPM 利用率、零流量、coverage 不足。`available` 不是健康；零流量不能算通过。
5. 生成摘要：总览、异常 EP 列表、每项的时间/指标/值/阈值/证据、数据缺口、需要人工确认的项。
6. 推送：按 [飞书集成](feishu.md) 发 IM 或写文档/Base。幂等键使用 `date + profile + workflow`，发送后保存 message ID / document token 并回读。
7. 写 `run.json` 和 `report.md`；失败时保留本地证据并通知失败接收人。

## 最小可执行编排

宿主包装脚本只做“调用 CLI + 汇总 + 推送”，不复制采集逻辑：

```bash
#!/usr/bin/env bash
set -euo pipefail
run_id="$(date +%Y%m%dT%H%M%S)-$(openssl rand -hex 3)"
dir="runs/vmp-daily/$run_id"; mkdir -p "$dir/evidence"
date_key="$(date +%F)"
# 1) 证据：失败时保留已取得的 JSON，不覆盖历史
arkcli doctor monitor snapshot --endpoint "$EPS" --window 24h --range 1m --format json > "$dir/evidence/snapshot.json"
arkcli doctor monitor errors   --endpoint "$EPS" --window 24h --top-codes 5 --format json > "$dir/evidence/errors.json"
# 2) 报告：由 Agent 或 jq 生成 $dir/report.md，写入判定、异常项、coverage 和证据路径
# 3) 推送：幂等键在 1 小时内去重；发送后保存返回的 message_id
lark-cli im +messages-send --chat-id "$CHAT_ID" --markdown @"$dir/report.md" \
  --idempotency-key "vmp-daily-$date_key"
```

发送前必须向用户展示目标群、标题、摘要和证据链接；没有 `lark-cli` 时只保留 `report.md` 并标记“尚未发送”。

## 调度

本 Skill 不内建调度器。按 [定时执行](scheduler.md) 在宿主注册任务；任务只调用本工作流，不复制采集逻辑。

## 判定

- `pass`：范围内 EP 均有足够覆盖，且没有超过阈值的异常。
- `violation`：存在超过阈值的异常 EP。
- `inconclusive`：清单解析失败、覆盖不足、查询失败或阈值不明确。
- 发送失败不等于巡检失败：数据结论可以成立，但必须标 `side_effect_failed`。

## 报告模板

- 日期、profile、范围、解析时间
- 巡检窗口与指标
- 异常 EP 摘要表（EP、指标、值、阈值、时间、coverage、Top 错误码）
- 正常 EP 数量与零流量 EP
- 数据缺口与查询失败
- 证据链接/路径
- 推送回读（消息 ID、文档 token、Base record）

## 失败处理

- 清单只解析到部分分页：不发送“全量正常”，标 `partial` 并说明缺失范围。
- `errors` 为 `partial`：说明哪些查询失败，不把缺失错误码当成零错误。
- 推送失败：重试一次；仍失败则写本地报告并通知失败接收人。
- 同一天重复执行：用幂等键去重；不重复发相同摘要。

## 示例提示词

“每天 10 点巡检我名下所有 EP 过去 24 小时的错误率和 P99，有异常的推到群里，没异常也发一句摘要。”
