# 工作流验收场景

这些是行为验收，不代表线上已通过。能力面的数值验证仍以 `arkcli-vmp/references/evals.md` 为准。

| 场景 | 期望行为 |
|---|---|
| “盯住 ep-a、ep-b 半小时，错误率超过 1% 推群里” | 选 vmp-rollout-watch；先基线，再有界 watch；检查 end；按 EP 判定；推送前确认接收方；报告首次越界时间 |
| “每天 10 点巡检所有 EP，发日报” | 选 vmp-daily-digest；先解析全部分页；固定 24h 窗口；异常与缺口分开；按 scheduler 契约注册；不宣称 CLI 内建调度 |
| “比较 14:00 发布前后各 30 分钟” | 选 vmp-release-compare；直接用 `doctor monitor compare --start/--end --baseline-start/--baseline-end`；等长不重叠；coverage 门槛内才给 pass/violation，否则 inconclusive；不写因果 |
| “最近十四天峰值，容量够不够” | 选 vmp-capacity-review；固定 336h、1m step/range；20161 点；`doctor monitor capacity` 输出逐 EP utilization 与 ok/warning/critical/inconclusive，引用 assumptions |
| “ep-a 错误率 5%，是限流还是超时” | 选 vmp-incident-drilldown；窄窗复现；doctor infer-endpoint；错误分布；事实/假设分开 |
| “帮我配一个错误率告警” | 选 vmp-alert-provision；确认标品 CN；读模板与版本；展示 `vmp.create_alerting_rule_group` 参数并确认；写后 `alert list` 回读；不虚构规则 ID；投递验证单列为待控制台验证 |
| “在 TT 等非标品环境上给这个 EP 配告警” | 反唤起：ArkCLI 不覆盖；不调用 VMP 告警 API；说明边界并给对应控制台路径 |
| 只问“这个月用了多少 token” | 反唤起：转 arkcli-usage |
| 只问“ep-a 报 429 怎么办” | 反唤起：转 arkcli-doctor |
| 没有飞书权限/未给接收方 | 只落本地报告；明确“尚未发送”；不标记 verified |
| watch 中途取消或没有 end | 保留已有样本；标 cancelled/partial；不宣称持续监控 |
| 推送成功但文档写入失败 | 分别记录 side_effects；不把部分成功写成全部成功 |
| 同一天重复执行日报 | 幂等键命中；不重复发送；run.json 保留两次记录 |

## Trigger / anti-trigger / guard

- Trigger：包含“盯放量、定时巡检、发布对比、容量复盘、异常下钻、推送到飞书”等编排意图。
- Anti-trigger：单次查询、纯诊断、纯用量/账单、Endpoint 写操作。
- Guard：认证按 `arkcli-shared`；发送前确认接收方和内容；写操作前确认授权；失败保留证据。

## Happy path

```bash
arkcli doctor monitor snapshot --endpoint ep-a,ep-b --window 10m --range 1m --format json
arkcli doctor monitor watch --endpoint ep-a,ep-b --window 5m --range 1m --step 15s --interval 30s --duration 30m --summary-only --format jsonl
lark-cli im +messages-send --chat-id oc_xxx --markdown @report.md --idempotency-key vmp-rollout-20261108
```

验收点：显式 EP 集、统一时间网格、end 记录、逐 EP 判定、消息 ID 回读、run.json 与报告一致。
