# 十四天历史峰值与容量证据

解析范围、固定十四天 start/end，并保留一分钟 step/range。基础层分片查询和去重，不把十四天压成大窗口后取平均。

```bash
arkcli doctor monitor capacity --endpoint ep-a,ep-b --start 2026-10-01T00:00:00Z --end 2026-10-15T00:00:00Z --step 1m --range 1m --format json
```

输出 `vmp.capacity.v1`，含逐 EP RPM、输入 TPM、输出 TPM、purchased completion TPM 的 collection evidence。峰值有时间戳，缺数据不会变成零，单位为 req/min 或 tokens/min。

`capacity` 会计算逐 EP 利用率：`peak_total_tpm = max(input TPM + output TPM)`（原始点可用时按时间戳对齐；不可用时退化为两侧峰值相加的上界），`purchased_tpm` 取 `token_type="completion"` 的最新观测，`utilization = peak_total_tpm / purchased_tpm`。`>=0.8` 为 `warning`，`>=1.0` 为 `critical`；配额缺失、覆盖不足或没有正 TPM 样本时 `inconclusive`。

`assumptions[]` 必须随报告保留：配额口径来自 completion 指标，若账号按其它口径计费，以控制台为准。可回答“观测到的历史峰值、何时出现、数据覆盖多少、逐 EP 利用率是多少”；不能仅凭该输出给出账单口径的配额余量或安全调额比例。不要将所有 EP 峰值求和称为项目同时峰值，也不要把 input TPM 单独除以 completion 配额。

需要看模型级免费额度余量时加 `--include-free-quota`：

```bash
arkcli doctor monitor capacity --endpoint ep-a,ep-b --window 336h --step 1m --range 1m --include-free-quota --format json
```

输出新增 `free_quota`：`status`（available/partial/unavailable）、`matched_endpoints`、`unmatched_endpoints`，以及按模型聚合的 `balances[]`（`model`、`total`、`consumed`、`remaining`、`scope`）；逐 EP 行也会带 `model` 和该模型的 `free_quota`。额度读取兼容 `InferenceFreeUsage` 与 `ResourcePackItems[Type=FreeInference]` 两种定价快照形态。口径是**模型级免费额度**（同模型的所有 EP 共享），不是单个 EP 独占余量；自定义模型 ID 没有对应的基础模型额度行时记为 unmatched。免费额度查询是 best-effort：pricing 或 EP 元数据失败时降级为 `free_quota.status=unavailable` 并保留容量结论，不会让整个命令失败。

若用户要安全调整配额，先读 当前安装的 `arkcli-quota` Skill（仅在当前产品提供该 Skill 时），按实际支持的配额池和 dry-run/plan 契约衔接；其它产品保留建议和所缺证据。建议需结合流量预测、余量策略、限流和明确可比的配额口径，不能用临时脚本绕开确认或计划校验。

产出：逐 EP 历史峰值及时间、coverage、缺片/空数据、配额证据、限制、建议补充的事实。TTFT 当前不可用；E2E 时延、吞吐量和配额是不同指标。
