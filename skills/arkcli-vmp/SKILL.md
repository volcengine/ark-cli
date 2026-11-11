---
name: arkcli-vmp
description: 编排 ARK 可观测任务：多 Endpoint 放量观察、发布前后对比、十四天历史峰值和容量证据分析。用户说“盯着放量”“比较上线前后”“过去两周容量够不够”“找最忙接入点”时使用。组合资源清单、VMP 采集与诊断，输出可追溯证据和结论；告警写入走 raw action，自动调度尚未支持。
---

# ARK 可观测任务

先读取并遵循 `../arkcli-shared/SKILL.md`。

目标是完成用户的观察或分析任务。CLI 提供可靠查询、时间对齐、分片、统计和输出；由本 Skill 决定范围、执行顺序、条件分支、证据产物和完成标准。所有采集命令都在 `arkcli doctor monitor` 下；单 Endpoint 诊断仍用 `arkcli doctor infer-endpoint`。

## 适用场景与参考

| 用户目标 | 执行流程 | 参考 |
|---|---|---|
| 放量期间盯住多个 Endpoint | 明确资源和阈值 → 基线 → 有限时长观察 → 异常诊断 → 报告 | [放量观察](references/watch.md) |
| 发布前后是否退化 | 固定两个等长不重叠时间段 → `doctor monitor compare` → 检查 coverage/流量守卫 → 解释 pass/violation/inconclusive | [采集与对比](references/collect.md) |
| 过去两周峰值、容量是否足够 | 固定十四天范围 → 分片一分钟采集 → 峰值和缺口 → 核对配额口径 → 结论 | [历史容量](references/capacity.md) |
| 现在有什么异常/最忙资源 | 确定 EP 集合 → 摘要或排行 → 根据证据下钻 | [快照](references/snapshot.md)、[排行](references/top.md) |
| 配告警、每天定时检查 | 读模板与现有规则；写入走 raw action 并回读；定时检查转 workflow Skill | [告警与调度](references/alert.md) |

需要把证据推送到飞书、写入文档/看板、注册定时任务，或把多个观察/对比/巡检任务串成一条持续工作流时，转到 [`../arkcli-vmp-workflows/SKILL.md`](../arkcli-vmp-workflows/SKILL.md)。本 Skill 只负责一次性观察与分析。

## 反唤起信号

单纯查 Token 用量或账单分别用 arkcli-usage / arkcli-billing；已有明确故障、只需诊断时用 arkcli-doctor；创建或更新 Endpoint 用部署/Endpoint Skill。不要为了使用监控命令而扩大这些任务。

## 执行顺序与守卫

1. 确定 profile、明确的 Endpoint 集合、时间范围和用户判据。用户指定 EP 时直接使用。用户说“全项目”时，先读 `../arkcli-infer-endpoint/SKILL.md`，用控制面清单解析本 profile 可见的所有分页，再把明确 ID 传给采集；不要用共享 VMP workspace 代替项目范围。保存清单和解析时间。单批最多 50 个 EP，多批必须共享同一 start/end。
2. 需要核对身份范围时按 shared 分别执行 `arkcli auth status` / `arkcli auth whoami`，直接检查原始结果，保存脱敏后的产品、region、project 信息；字段不足时标明范围尚未核实。`profile show` 可能同步 Key 并回写配置，不能当作普通只读身份检查。查询只读解析已绑定的 workspace。未登录或凭证过期时转到 `../arkcli-auth/SKILL.md`，缺绑定或权限时按实际错误给出下一步；创建/绑定是另一个有副作用的操作，不假称“采集已完成”。
3. 选择固定时间或相对 window。历史分析和前后比较优先 start/end；固定时间需 RFC3339 整秒。`window` 是查询跨度，`step` 是取点间隔，`range` 是每点统计回看窗口，三者独立。十四天写 `336h`，不是 `14d`。用 [采集契约](references/collect.md)。
4. 将 JSON/JSONL 保存到本次任务目录，连同用户目标、资源清单、执行命令和结论报告。基础层负责所有统计、分片、去重和单位换算；不要临时写鉴权、PromQL 拼接或粗化峰值的脚本。
5. 检查 collection.status、每个结果的 status/coverage 和 queries。`available` 只表示请求网格上的有限采样齐全；它不证明源数据完整，也不表示健康。`partial`/`no_data`/`failed` 都不能填零。真实 RPM=0 只说明采样点观察到零流量；没有错误率/时延时不能据此宣称服务通过验证。
6. 指标异常时，先读 `../arkcli-doctor/SKILL.md`，按其真实入口诊断对应 EP。报告把观测事实、假设、诊断证据和处置建议分开。未经用户授权不要直接改配额、切模型或停止 EP；告警写入必须走 [告警](references/alert.md) 的"确认 + 回读"协议，不能只凭采集结果直接创建。
7. 完成时提供简短结论（达到判据 / 违反判据 / 不可判定），逐 EP 证据、时间、单位、覆盖率、峰值时间、缺口和下一步。阈值未提供时可做描述性观察，不自行宣布“放量通过”。

## 能力边界

- `collect` 默认采集 RPM、错误率、请求 E2E P99、输入/输出 TPM。**E2E P99 不是 TTFT**；当前没有经过核实的 TTFT 采集契约。
- 错误率是逐 EP 比例；sample_mean 是观测点平均值，不是全时段按请求加权的错误率。P99 的峰值是“每点 P99 的峰值”，不是整个十四天所有请求的 P99。
- `compare` 逐 EP 输出两个等长窗口的 `sample_mean`/峰值差异与 `pass|violation|inconclusive`；阈值随输出回显。`sample_mean` 不是按请求加权的错误率，P99 的均值不是全时段 pooled P99；`--min-rpm-ratio` 只是流量塌陷守卫，不代表流量下降一定是故障。
- `errors` 聚合逐 EP 错误码 Top N、请求量、错误率和规则化建议；`hints` 是建议不是根因。有流量但零错误是 `available` + `codes: []`；无流量才是 `no_data`。
- `capacity` 计算逐 EP TPM 利用率：`peak_total_tpm / purchased_tpm`，`>=80%` 为 warning、`>=100%` 为 critical；配额缺失或覆盖不足时为 `inconclusive`。`assumptions[]` 说明配额来自 completion 指标，账单口径以控制台为准。`--include-free-quota` 附带模型级免费额度余量（同模型 EP 共享，不是单 EP 独占）；pricing 不可用时降级为 `unavailable`，不影响容量结论。
- 所有 monitor 命令支持隐式范围 `--mine/--model/--name-regex/--status`（与 `--endpoint` 互斥）：启动时解析一次控制面清单，结果写入 `scope`；超过 50 个截断并记录 `skipped[]`，零匹配直接报错。`--mine` 在可取得 ARK IAM UserID 时走服务端 createdBy 标签过滤，降级身份走客户端用户名后缀匹配，不降级为全项目扫描。watch 不会在运行中刷新范围。
- `watch --alert-on` 在每轮采样后做确定性阈值判断，`--alert-for` 要求连续 N 轮越界；只在 firing/resolved 状态切换时输出事件，结束时仍有未恢复告警则非零退出。这是有界会话内的判断，不是常驻告警规则。
- watch 是有界前台观察，输出逐轮 JSONL 和 end 记录；退出或中断后不会继续运行。没有常驻告警、计划调度或恢复运行的承诺。
- 采集与告警读取是只读；`doctor monitor` 的 workspace 自动绑定和告警写入是显式写操作，必须展示影响并回读。告警写操作只通过 `arkcli api vmp.*` 原始 action 提供，不挂 shortcut。
- 告警能力只覆盖**标品 CN 的 VMP 告警**。详见 [能力模型](references/capability-model.md)。
- 不存在 `doctor monitor workspace list`、`--since`、`--output`、`--once`、`--threshold`。使用 `--workspace-id`、`--window`/`--start --end` 和全局 `--format`。只读采集没有 `--dry-run`。
