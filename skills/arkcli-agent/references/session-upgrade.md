# 已有 Session 运行配置升级

用户要让**已有 Session** 使用另一个模型、同 Agent 的新版本或新运行配置时，使用 `agent session upgrade`；不要用 `agent agent update` 后假设老会话自动跟随，也不要用 `+iterate` 重建会话。只改标题/标签仍走 `session update`；创建时一次性 override 见 `session-files.md`。

## 命令与输入

配置或更换模型运行参数前，先查[模型参数 metadata](model-config.md)。查询结果用于说明可选值，不替代升级基线规则，也不新增客户端业务校验。

```bash
arkcli agent session upgrade <session-id> \
  --model <selected-model-id> \
  --thinking enabled \
  --reasoning-effort low \
  --dry-run --format json
```

```bash
arkcli agent session upgrade <session-id> \
  --agent-version 4 \
  --agent '{skills: [], mcp_servers: []}' \
  --environment '{config: {env: {LOG_LEVEL: debug}, packages: {pip: [pytest]}}}' \
  --compact \
  --dry-run --format json
```

| 参数 | 用法 |
| --- | --- |
| `<session-id>` | 必填；保持原 Session ID，不创建新会话 |
| `--file <path>` | 本地 JSON/YAML 完整请求文件；不自动读 stdin，不支持 `--file -` |
| `--agent <object>` | JSON/YAML 或 `@file`；支持 `type/id/version/model/system/tools/mcp_servers/skills/multiagent/display_name` |
| `--environment <object>` | JSON/YAML 或 `@file`；支持 `type/id/config`；config 的 cloud 字段见下文 |
| `--agent-version <int>` | 覆盖 `agent.version`；不传或 0 保留当前快照，不会选择最新版本 |
| `--model <id或object>` | ID 只覆盖 `agent.model.id`；对象或 `@file` 替换请求中的 model 对象 |
| `--speed/--thinking/--reasoning-effort/--service-tier` | 覆盖 `agent.model` 同名 snake_case 字段；显式空字符串会保留在请求中 |
| `--vault-ids <array>` | JSON/YAML 数组或 `@file`；刷新既有绑定，不是新增/解绑 Vault |
| `--initial-events <array>` | JSON/YAML 数组或 `@file`；可省略或为 `[]`；非空时后端仅接受一条固定的 `/compact` 事件 |
| `--compact[=false]` | true 构造固定 `/compact` initial event；false 显式设置 `initial_events: []`；与 `--initial-events` 互斥 |

升级是数据面 `POST /api/v3/sessions/{id}/upgrades`，请求统一用 **snake_case**，不套用 TOP CamelCase 或创建时的 `AgentWithOverrides`。CLI 仅在 type 缺失时补 `agent_with_upgrades` / `environment_with_upgrades`；显式 type 原样交给后端。id 可省略，后端沿用当前绑定；如传入，只能是当前 Session 的 Agent/Environment ID，不能用来切换绑定。

优先级：先加载 `--file`；`--agent/--environment` 整块覆盖文件中的对应对象；`--agent-version` 和模型独立 flags 再覆盖具体字段；事件/Vault flags 覆盖对应顶层数组。JSON/YAML 中未提供、`null`、`[]`、空字符串不会被 CLI 互相转换。除了补缺失的 type，CLI 不做在线查询、自动合并快照或业务白名单校验，服务端错误原样报告；但不要因此编造接口字段。

## 配置语义

- 后端要求 `agent/environment/vault_ids` 至少提供一个；`initial_events` 本身不是升级目标。upgrade 当前不支持 `title/tags/resources/checkpoint_id/environment_id` 等顶层字段。
- 不传 version 时，以当前 Session Agent 快照为基线，未传子字段保留。指定正版本时，以该已发布版本配置为新基线，再应用本次显式覆写；旧 Session overrides 不保证自动保留。用户要“最新版”时先查询真实版本号，不发送字符串 `latest`，也不省略版本来表示最新版。
- `tools/mcp_servers/skills` 非空数组整组替换，`[]` 清空，未提供或 `null` 保留基线；没有默认工具或 Skill 自动注入。
- `agent.model` 支持 `id/speed/thinking/reasoning_effort/service_tier/provider/protocol/base_url/headers`。**这是升级接口的契约，与创建 override 的五字段过滤不同**，CLI 不过滤用户显式传入的模型接入字段。模型 ID 改变后，模型解析及最终 protocol/defaults 由服务端决定。选择参数必须以目标模型 metadata 为准，不沿用旧模型能力假设；不要默认注入 provider、URL、headers 或 speed。
- cloud `config.networking` 整块替换；`config.env` 按 key 合并，本次同 key 覆盖旧值；`config.packages` 按包名合并，本次同名包替换旧 spec。空 env/packages 不会清空历史项。`setup_script` 显式值替换，空字符串可传。只给 environment id 不代表从 Environment 资源拉取最新配置；应明确提供期望 config。
- self_hosted Session 可以升级 Agent；不支持 cloud/self_hosted 类型互切，不应发送自托管 Sandbox 的 networking/packages/env/setup_script。用户 TOS 配置不可通过升级修改。
- `vault_ids` 必须与当前绑定集合一致（忽略顺序/重复）；`[]` 仅适用于原本没有绑定的 Session。不发送凭证密文；不要把刷新误称为更换绑定。

## 执行与结果核对

1. 用 `session get` 核对目标、当前 Agent/环境快照及状态。后端要求 idle；已终止、升级中或上一轮未正常 end_turn 等情况会拒绝。CLI 不额外抢先判断这些业务状态，也不重试业务错误。
2. 生成 leaf-local `upgrade ... --dry-run --format json`。这是零网络计划，不证明模型可用或资源状态合法。复述本次变化及可能重建 Sandbox/压缩上下文的影响，得到明确确认后再去掉 `--dry-run` 执行；此命令没有 `--yes`。
3. 真实执行只提交一次，返回后端 Session 对象，不返回独立 upgrade task，也不等待后台结束。`status=upgrading` 只表示受理/处理中；不要汇报“模型已生效”。不编造 `/upgrades/{upgrade_id}` 查询命令。
4. 后续单独 `session get` 回读，等待离开 upgrading，并检查模型 ID、版本、运行参数和目标配置。idle 单独不能证明升级成功；配置不符或异常时继续查询 events / `+debug`，不自动重发 upgrade。超时或响应不明确时，同样先回查，避免重复升级。
5. 升级期间不要发消息、删除、打断或重复升级。若本次显式启用 compact，不要再独立发送第二次 compact。需要验证推理时，在配置回读确认后，取得用户测试授权再发最小消息。

API Key 和项目来自当前产品/profile，显式 `--base-url` 需要配对显式 API Key。Preview 会对标准 Authorization/API Key 字段脱敏，但不要把任意环境变量或自定义 header 的值当成已全面脱敏；不要在对话、命令日志或报告中展示真实密钥。
