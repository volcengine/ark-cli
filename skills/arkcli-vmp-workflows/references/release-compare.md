# 发布前后对比（vmp-release-compare）

## 何时使用

用户要在一次发布、模型切换或配置变更前后，比较同一组 EP 的指标是否退化。

## 必填输入

- EP 集合与发布时刻 T。
- 前后窗口长度（等长、不重叠）。
- 指标、step、range。
- 判定口径：哪些变化算退化，流量差异如何处理。

## 执行

1. 读 [collect](../../arkcli-vmp/references/collect.md) 与 [指标语义](../../arkcli-vmp/SKILL.md#能力边界)。
2. 冻结两个区间：发布前 `[T-L, T)`；发布后从 `T + range` 开始，避免统计回看跨越边界。两段等长、不重叠。
3. 优先使用原子对比命令，不要在 Agent 侧重复实现 diff：

```bash
arkcli doctor monitor compare \
  --endpoint ep-a,ep-b \
  --metric rpm,error_rate,latency \
  --start <post-start> --end <post-end> \
  --baseline-start <pre-start> --baseline-end <pre-end> \
  --step 1m --range 1m \
  --format json > evidence/compare.json
```

只给当前窗口时，baseline 默认取紧邻的前一个等长窗口：

```bash
arkcli doctor monitor compare --endpoint ep-a,ep-b --metric rpm,error_rate,latency \
  --window 30m --step 1m --range 1m --format json > evidence/compare.json
```

4. 判定阈值用显式参数固定并随输出回显：

| Flag | 默认 | 含义 |
|---|---|---|
| `--min-coverage` | `0.8` | 两侧覆盖率低于该值时不可判定 |
| `--max-error-rate-delta` | `0.01` | 错误率允许的绝对增幅（1 个百分点） |
| `--max-latency-delta-pct` | `20` | 时延均值允许的相对增幅 |
| `--min-rpm-ratio` | `0.5` | 流量低于基线的该比例时判为流量塌陷 |

5. 读取 `vmp.compare.v1`：`status`、`thresholds`、`assumptions`、`baseline`/`current` 证据、逐 EP `metrics[]` 的 `sample_mean`/`peak`/`coverage`/`delta`/`delta_pct`/`verdict`/`reason`。
6. 退出码即判定：`pass` 为 0；`violation`、`inconclusive`、`no_data`、`failed` 均为非零，但 JSON 已先输出。CI 可直接用退出码卡发布。

## 判定

- `pass`：两侧证据充分、覆盖率达标，且没有超过阈值的退化。
- `violation`：在有效样本上出现明确退化（错误率、时延或流量塌陷）。
- `inconclusive`：区间不等长、口径不一致、覆盖不足、缺少样本或流量不可比。
- `no_data` / `failed`：窗口内没有可用数据或全部查询失败；不得写成“发布无影响”。

## 口径限制

- `sample_mean` 是采样点的算术平均，不是按请求加权的错误率，也不是全时段 pooled P99。
- 错误率阈值是绝对比例差（`0.01` = 1 个百分点），不是相对变化。
- `--min-rpm-ratio` 只是流量塌陷守卫，不知道流量下降是否是发布预期；判 violation 前向用户确认。
- `compare` 不比较 `quota.purchased.tpm` 之外的业务配额口径；需要自定义规则时，用两次 `collect` 保存原始证据后在 Agent 侧计算，并在报告里写明公式。

## 报告模板

- 发布时刻、前后区间、EP 清单
- 指标、step/range、单位、阈值
- 逐 EP 前后 `sample_mean`/峰值与变化、覆盖率、流量
- `verdict` 与 reason、首次越界项
- 证据路径（`evidence/compare.json`）、限制与下一步

## 失败处理

- 发布后区间早于 `T + range`：无效对比，重选区间。
- 两段 EP 集或指标不一致：无效对比，重新采集。
- 只有一段有数据：标 `inconclusive`，不写“发布无影响”。
- 覆盖不足：先扩大窗口或降低 step，不要直接放行。

## 示例提示词

“比较 ep-a、ep-b 在 14:00 发布前后各 30 分钟，错误率和 P99 有没有退化，生成飞书报告。”
