# 告警开通（vmp-alert-provision）

本工作流按产品分两条告警路径，开始前先确认产品。指标证据仍来自 `arkcli doctor monitor`；告警后端不同，不能互串。

## 标品 CN：VMP 告警

适用范围：**标品 CN 的 VMP 告警**。

命令契约与写协议见 [`../../arkcli-vmp/references/alert.md`](../../arkcli-vmp/references/alert.md)；能力分工见 [`../../arkcli-vmp/references/capability-model.md`](../../arkcli-vmp/references/capability-model.md)。

### 能力现状

- 只读：`arkcli doctor monitor alert list` / `alert templates` / `alert template --template-id` 可用。
- 写入：`arkcli api vmp.create_alerting_rule_group` / `update_alerting_rule_group` / `delete_alerting_rule_group`、`tag_resources` / `untag_resources`、通知策略与联系组 action 可用，但**没有 shortcut**。
- 仍不可验证：通知实际投递（OpenAPI 不暴露投递记录）；CLI 没有内建调度。
- 单次配置优先引导控制台；批量 / 幂等编排才走 CLI。

### 开通规格

执行前先把规格填完整，缺项标为待确认，不要自造 PromQL 或阈值：

| 字段 | 内容 |
|---|---|
| 范围 | profile、产品（必须是标品 CN）、workspace、EP 集合 |
| 指标 | 模板 ID / 版本、指标 ID、单位、统计窗口 |
| 阈值 | 触发值、比较方向、持续时间（模板默认值可覆盖时写明） |
| 缺数据策略 | 缺数据是否触发、连续多久 |
| 恢复条件 | 恢复阈值与持续时间 |
| 通知 | 接收人/群、通知渠道、通知策略、联系组 |
| 标签 | `ark_endpoint_<id>`（控制台按标签反查） |
| 验证 | 规则回读、启用状态、通知绑定；投递测试注明由控制台完成 |

### 执行

1. 确认产品与 workspace；未绑定时先解决 telemetry 绑定，不要直接写告警。
2. `arkcli doctor monitor alert templates` 选模板，`alert template --template-id` 读回版本、PromQL 与默认阈值。
3. 生成 `vmp.create_alerting_rule_group` 参数：`Items[].Enable=true`，`Variables` 绑定 endpoint，`Tags` 打 `ark_endpoint_<id>`。
4. 需要通知时：`vmp.list_notify_policies` 查现有策略，必要时 `vmp.create_contact_group` + `vmp.create_notify_policy`（Levels 按 P0/P1/P2 升序）。
5. 展示完整 `--params`、目标 workspace 和影响；用户确认后执行。
6. 回读：`alert list` 核对规则组、启用状态、tags；`vmp.list_notify_policies` 核对绑定。
7. 把规格、命令、回读证据写入 `run.json` 与 `report.md`。

### 判定

- `planned`：规格完整，但未执行写入。
- `created`：写入成功且 `alert list` 回读到规则组。
- `verified`：规则、启用状态、通知绑定三项都有回读证据。**投递验证不在 CLI 判定范围内**，报告里单列为"待控制台/真实告警验证"。
- 任何一项缺失，不得升级为 `verified`。

### 批量编排示例

先 `alert list` 按 `ark_endpoint_<id>` 标签查重：已存在则走 `update_alerting_rule_group`，不要重复创建。确认参数后逐 EP 执行：

```bash
for ep in ep-a ep-b; do
  arkcli api vmp.create_alerting_rule_group --params "$(jq -n \
    --arg ep "$ep" --arg ws "$WS" --arg tpl "$TPL" --arg ver "$VER" '{
      Name: ("ark_alerting_rules_" + $ep),
      WorkspaceId: $ws,
      AlertingRuleTemplateId: $tpl,
      AlertingRuleTemplateVersion: $ver,
      Items: [{Id: 1, Enable: true}],
      Variables: [{Id: 1, Values: [$ep]}],
      Tags: [{Key: ("ark_endpoint_" + $ep), Value: "true"}]
    }')"
done
arkcli doctor monitor alert list --workspace-id "$WS" --format json > evidence/alerts-after.json
```

批量脚本不是天然幂等：每轮执行前查重，执行后回读；任何一步失败都停止并报告已创建/已更新的范围。

### 示例提示词

“给 ep-a、ep-b 配一个错误率超过 5% 持续 5 分钟就飞书通知我的告警。”

