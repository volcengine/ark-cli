# 告警：读取、写入与调度边界

适用范围：**标品 CN 的 VMP 告警**。分工见 [能力模型](capability-model.md)。

## 只读命令（shortcut）

```bash
# 列工作区下的告警规则组；不传 --workspace-id 时只读解析 telemetry 绑定
arkcli doctor monitor alert list [--workspace-id ws-xxx] [--template-id tpl-xxx] [--page-size 50]

# 列出 CN 区域内置告警模板目录（本地常量，不访问后端）
arkcli doctor monitor alert templates

# 查看单个模板详情：PromQL、默认阈值、版本
arkcli doctor monitor alert template --template-id b2876e3d-c52b-41d6-9b41-1b020fc7b2aa
```

`alert list` 输出 `workspace_id` / `total` / `items[]`；每条规则组含 `id`、`name`、`template_id`、`template_version`、`notify_policy_id`、`tags[]`（`ark_endpoint_<id>` 是控制台反查 endpoint 告警的约定 key）、`rules[]`（`enabled` 与各档 `levels`）。未绑定工作区时命令报 `no VMP workspace bound to ark telemetry`，不静默返回空。

## 写入（原始 action，无 shortcut）

写操作不挂 shortcut，避免在 CLI 复刻控制台表单校验；需要批量 / 幂等编排时用 `arkcli api`：

| 动作 | 用途 |
|---|---|
| `vmp.create_alerting_rule_group` | 从模板创建规则组（替换 `Variables` / `Tags`） |
| `vmp.update_alerting_rule_group` | 全量更新；关闭告警把 `Items[].Enable` 置 `false` |
| `vmp.delete_alerting_rule_group` | 从 VMP 删除规则组（不可恢复） |
| `vmp.tag_resources` / `vmp.untag_resources` | 打 / 摘 `ark_endpoint_<id>` 标签 |
| `vmp.list_notify_policies` / `vmp.create_notify_policy` / `vmp.update_notify_policy` | 通知策略读写 |
| `vmp.list_contact_groups` / `vmp.list_contacts` / `vmp.create_contact_group` | 联系组与联系人 |
| `vmp.list_notify_group_policies` | 取默认聚合策略 ID |
| `vmp.get_alerting_rule_template` / `vmp.list_alerting_rule_groups` | 写前读模板、写后回读 |

写协议：

1. 先 `alert templates` / `alert template` 确认模板 ID 与版本；不要自造 PromQL 或阈值。
2. 按模板 `Items` 生成参数，`Variables` 绑定 endpoint，`Tags` 打 `ark_endpoint_<id>`；`Items[].Enable` 显式传 `true`。
3. 执行前展示完整 `--params`、目标 workspace 和影响（新建 / 更新 / 删除），取得用户确认。
4. 执行 `arkcli api vmp.<action> --params '<json>'`。
5. 回读：`alert list` 核对规则组、启用状态与 tags；`arkcli api vmp.list_notify_policies` 核对通知绑定。
6. 只有回读通过才能说"已创建 / 已生效"。

示例（创建规则组）：

```bash
arkcli api vmp.create_alerting_rule_group --params '{
  "Name": "ark_alerting_rules_ep-a",
  "WorkspaceId": "ws-xxx",
  "AlertingRuleTemplateId": "b2876e3d-c52b-41d6-9b41-1b020fc7b2aa",
  "AlertingRuleTemplateVersion": "<从 alert template 读回>",
  "Items": [{"Id": 1, "Enable": true}],
  "Variables": [{"Id": 1, "Values": ["ep-xxx"]}],
  "Tags": [{"Key": "ark_endpoint_ep-xxx", "Value": "true"}]
}'
```

## 边界

- 单次配置优先走控制台（表单、模板校验、影响预览更完整）；CLI 面向批量与可重复编排。
- 通知实际投递无法通过 OpenAPI 验证；只能验证规则存在、启用状态与通知策略绑定。投递测试需在控制台或真实告警中完成。
- 不把 `watch` 当常驻告警；CLI 没有内建调度。定时巡检由宿主编排（见 `arkcli-vmp-workflows` 的 scheduler 参考）。
