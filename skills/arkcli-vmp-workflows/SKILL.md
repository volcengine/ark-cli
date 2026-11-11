---
name: arkcli-vmp-workflows
version: 1.0.0
description: 编排 ARK 可观测工作流：放量盯盘、发布前后对比、容量复盘、每日巡检和异常下钻。用户说“盯放量并推飞书”“每天巡检发日报”“发布前后对比出报告”“异常了帮我下钻”时使用。组合 arkcli-vmp 的证据与飞书/看板/宿主调度；采集保持只读，告警写入是显式写操作（确认 + 回读），推送、定时和留档由本 Skill 执行。
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli doctor monitor --help"
---

# ARK 可观测工作流编排

**CRITICAL — 开始前 MUST 先读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md) 与 [`../arkcli-vmp/SKILL.md`](../arkcli-vmp/SKILL.md)。** 前者是认证与安全协议，后者是能力面命令契约和证据解释。本 Skill 只负责编排、判定、推送、留档和完成判据，不重算指标。

## 这一层解决什么

能力面回答“数据是什么”；编排面回答“这个任务现在算完成了吗、下一步做什么、结果发给谁”。同一个能力面证据可以被多个工作流复用；工作流不固化成 CLI 命令，阈值、接收方、节奏和分支在 Skill/宿主侧调整。

## 工作流路由

| 用户目标 | 工作流 | 参考 |
|---|---|---|
| 放量/压测期间盯住一组 EP，按阈值判断并推送 | vmp-rollout-watch | [放量盯盘](references/rollout-watch.md) |
| 发布前后是否退化，生成对比报告 | vmp-release-compare | [发布对比](references/release-compare.md) |
| 过去两周峰值、覆盖率、容量证据 | vmp-capacity-review | [容量复盘](references/capacity-review.md) |
| 每天定时巡检并发日报 | vmp-daily-digest | [每日巡检](references/daily-digest.md) |
| 已发现异常，继续定位到 EP/错误/时延 | vmp-incident-drilldown | [异常下钻](references/incident-drilldown.md) |
| 想配置告警或飞书通知 | vmp-alert-provision | [告警开通](references/alert-provision.md) |

只查一次指标、快照或排行，用 `arkcli-vmp`；已有明确故障、只做诊断，用 `arkcli-doctor`。不要为了走工作流而重复采集。

## 产品路由（开始前必须确认）

先确认 profile 对应的产品，再选择告警路径。盯盘、发布对比、容量复盘、每日巡检和异常下钻的指标证据都走 `arkcli doctor monitor`（各产品指标后端一致）；告警后端不同，不能互串。

| 产品 | 告警后端 | 告警路径 |
|---|---|---|
| 标品 CN | VMP 告警 | `arkcli doctor monitor alert` 只读 + `arkcli api vmp.*` 写入，见 [告警开通](references/alert-provision.md) |

无法确认产品时先问用户，不要默认按标品 VMP 写告警。


## 统一执行协议

每个工作流按同一顺序执行：

1. **确定任务与判据。** 明确产品、EP 集合、时间范围、指标、阈值、接收方和授权范围。用户没给阈值时只做描述性观察，不宣布通过。
2. **解析范围。** 用户给 EP 时直接使用；说“全项目”时先按 `arkcli-vmp` 的清单流程读取控制面全部分页，保存清单来源与解析时间。单批 1–50 个 EP，多批共享同一 start/end。
3. **冻结参数。** 固定 profile、产品、window/start/end、step、range、interval、duration、指标和单位。工作流运行中不改口径；要改就新开 run。
4. **采集并保存证据。** 原始 JSON/JSONL 写入本次 run 的 `evidence/`，不要覆盖历史文件。检查 collection.status、逐 EP status/coverage 和 queries。
5. **判定。** 只用能力面输出计算；结论分为 `pass` / `violation` / `inconclusive`。`partial`、`no_data`、低覆盖率和真实零流量都不能写成健康。
6. **执行副作用。** 推送、写文档、写 Base/Sheet 前必须展示目标、内容摘要和接收方；用户确认后执行。副作用要幂等并回读。
7. **留档并报告。** 写 `run.json` 和 `report.md`，报告包含目标、范围、证据路径、判定、限制、副作用回读结果和下一步。

## 运行记录与副作用契约

每个 run 使用独立目录：

```text
runs/<workflow>/<run-id>/
  run.json          # arkcli.workflow.run.v1
  report.md         # 人读报告
  evidence/         # 原始 JSON/JSONL
```

`run.json` 的最小字段见 [运行记录](references/run-record.md)。副作用统一走适配器：

- `notify`：发飞书消息/卡片；参数为 channel、target、title、summary、evidence links、idempotency key。
- `document`：创建或更新飞书文档；回读 document token。
- `table`：写入多维表格/电子表格；回读 record id 或 range。
- `schedule`：注册宿主定时任务；回读 job id、时区和下次运行时间。

飞书适配器见 [飞书集成](references/feishu.md)；宿主调度见 [定时执行](references/scheduler.md)。没有对应工具或权限时，输出可复制的消息/文档草稿，并明确“尚未发送”。

## 完成判据

一个工作流只有同时满足以下条件才算完成：

- 范围和口径已冻结，EP 清单可追溯；
- 证据文件存在，查询失败和缺口已解释；
- 判定与证据一致，不可判定时给出原因；
- 副作用已回读，或明确标记为未执行；
- `run.json` 的 status 与报告一致。

## 反唤起

- 单纯查用量/账单：`arkcli-usage` / `arkcli-billing`。
- 已有明确错误码或 EP 故障：`arkcli-doctor`。
- 只做一次采集/快照/排行：`arkcli-vmp`。
- 创建/更新 Endpoint、配额、模型：对应能力 Skill；本 Skill 不执行这些写操作。
- 告警写入：单次配置优先引导控制台；批量 / 幂等编排才走 `arkcli api vmp.*`，且必须确认参数、写后回读。投递验证与调度仍由控制台 / 宿主负责。

## 能力边界

- 本 Skill 不新增 CLI 命令。采集与告警读取只读；告警写入是显式写操作，只通过 `arkcli api vmp.*` 原始 action 提供，必须"先展示参数、再执行、后回读"。
- 告警能力只覆盖标品 CN 的 VMP 告警；其他产品或环境的告警后端不同，不要用 VMP 告警 API 代替，遇到时说明边界并给出对应控制台路径。
- 定时执行依赖宿主，不是 CLI 内建调度；进程退出后 watch 不会继续。
- 飞书集成需要机器人和目标群/文档权限；没有权限时保留本地报告。
- 通知实际投递无法通过 OpenAPI 验证；只能验证规则存在、启用状态与通知策略绑定。投递测试需在控制台或真实告警中完成。
- 本 Skill 面向国内站与飞书；TT 等其他产品形态暂不随包分发。
