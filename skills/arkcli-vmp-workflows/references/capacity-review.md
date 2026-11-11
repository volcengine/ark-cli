# 容量复盘（vmp-capacity-review）

## 何时使用

用户要看过去十四天的分钟级峰值、覆盖率，或为模型切换/扩缩容准备容量证据。

## 必填输入

- EP 集合。
- 截止时间（固定 end）。
- 配额来源与口径（已购 completion TPM 等）。
- 是否需要与七天窗口做同口径对照。

## 执行

1. 读 [capacity](../../arkcli-vmp/references/capacity.md) 与 [collect](../../arkcli-vmp/references/collect.md)。
2. 固定十四天 `start/end`，保持 `--step 1m --range 1m`；十四天写作 `336h`，不要写成 `14d`。
3. 采集：

```bash
arkcli doctor monitor capacity --endpoint ep-a,ep-b --start 2026-10-01T00:00:00Z --end 2026-10-15T00:00:00Z --step 1m --range 1m --format json > evidence/capacity.json
```

4. 检查分片、coverage、缺片和空数据；十四天一分钟网格应为 20161 个点。两百万预期采样的上限意味着 14d/1m × 4 指标约支持 24 个 EP/批；更多 EP 分批，所有批次共享同一 start/end。不得为绕过限制扩大 range 或粗化 step。
5. 读取 `vmp.capacity.v1`：
   - `endpoints[].peak_input_tpm`、`peak_output_tpm`、`peak_total_tpm`（优先时间对齐求和，无法对齐时是上界）。
   - `purchased_tpm` 来自 `ark_api_proxy_purchased_provisioned_tpm{token_type="completion"}` 的最新观测。
   - `utilization = peak_total_tpm / purchased_tpm`；`>=0.8` 为 warning，`>=1.0` 为 critical。
   - `inconclusive` 表示配额缺失、覆盖不足或没有正 TPM 样本，不是健康。
6. 不同 EP 的峰值不能相加为项目同时峰值；需要项目同时峰值时，按同一时间戳聚合后单独计算。
7. 需要七天对照时，用同一 EP 集、同一 step/range、同一截止时间重新采集，不引用不同口径的历史结论。
8. 配额口径仍以控制台为准：`capacity` 的假设已写在 `assumptions[]`，报告必须原样保留；不要把 `utilization` 直接当成账单或安全调额比例。

## 判定

- 能回答：观测到的历史峰值、峰值时间、覆盖率、逐 EP 利用率，以及 `ok/warning/critical` 判定。
- 不能回答：账单口径的配额余量、项目同时峰值、TTFT、安全调额比例。
- 七天与十四天峰值不一致时，先确认同口径；CLI 不解释后端平滑或打点差异，把差异作为待核实事实。

## 报告模板

- 目标、EP 清单、截止时间
- 十四天区间、step/range、20161 点
- 逐 EP 峰值、峰值时间、coverage、缺片
- 逐 EP `purchased_tpm`、`peak_total_tpm`、`utilization`、`status` 与 reason
- 七天对照（同口径）
- `assumptions[]` 与待控制台核实的配额口径
- 证据路径与下一步核实项

## 失败处理

- 分片失败或 coverage 不足：保留已取得证据，标 `partial`，不宣布峰值。
- 配额缺失或口径不可比：`inconclusive`，列出缺的证据。
- 用户要求直接改配额：转 `arkcli-quota`（若当前产品提供），本 Skill 不执行写操作。

## 示例提示词

“看 ep-a、ep-b 最近十四天的一分钟峰值和覆盖率，判断容量够不够；不够的话给出还缺什么证据。”
