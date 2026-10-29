# 自动上下文压缩配置

用于创建 / 更新 Agent，以及创建 Session 时覆盖 Agent 的压缩策略。它只保存配置，
不立即压缩当前上下文；立即压缩已有会话用 `agent session compact`，见
[会话交互](events-chat.md)。未要求调整压缩策略时，不主动填入配置。

## 参数

`ContextManagement` 是策略数组，目前支持空数组或一个 `compaction` 策略。

| 字段 | 含义 |
| --- | --- |
| `Type` | 必填，当前为 `compaction` |
| `CompactThreshold` | 可选，触发自动压缩的完整上下文 token 阈值，整数；不是压缩后长度。后端要求至少 50000；Ark 模型取得上下文窗口时还限制为窗口的 80%（向下取整），BYOK 不做该上限校验 |
| `Instructions` | 可选，给压缩过程的补充指令，例如保留任务目标、关键结论和未完成事项；省略或空字符串使用内置指令。后端限制为最多 100000 个字符，非空时不能全为空白；不沿用前端输入框的 400 字符限制 |

CLI 不查询模型 metadata 来拦截写入，也不钳制阈值。策略类型、数量和阈值合法性由服务端判断，错误应原样报告。

| 输入 | 行为 |
| --- | --- |
| 创建时省略字段或传 `null` | 不提供自定义配置，由服务端使用默认策略 |
| 更新 / Session override 时省略或 `null` | 保留现有 / 所选 Agent 的配置 |
| `[]` | 清除自定义策略，回到默认行为；**不是禁用自动压缩** |
| `[{"Type":"compaction"}]` | 替换为无自定义阈值和指令的策略，不合并旧字段 |
| 非空数组 | 整体替换策略；只填指令时不继承旧的自定义阈值 |

响应可能补齐展示用默认策略与阈值；当前后端动态配置的兜底值为 192000，但不是所有模型的固定适用值。
不要将这个数写死到创建请求中，也不要用回显的默认值判断用户曾显式配置过。

## 创建与更新 Agent

先按 [Agent 创建](agent.md) 确定模型与资源。下列占位 ID 必须替换为真实 ID。

```bash
arkcli agent agent create --name my-agent --model <model-id> \
  --context-management '[{"Type":"compaction","CompactThreshold":64000,"Instructions":"保留任务目标、关键结论和未完成事项"}]'
arkcli agent agent update <agent-id> --context-management '[]'
```

`--context-management` 接受 JSON/YAML **数组或 null**，也支持 `@文件`；也可在 `--file` / stdin 的完整请求中提供
`ContextManagement`。显式 flag 优先于文件，包括 `null`：它不回退到文件配置，创建时使用默认行为，更新时保留服务端原配置；`[]` 才是清除自定义策略。空字符串不是 `null`，会报输入错误。支持 PascalCase、camelCase、snake_case 字段，
指令文本原样保留。`+new-agent`、`+iterate` 同样支持该 flag；`+new-agent --fork` 保留源 Agent 返回的配置，
可显式传 `[]` 清除。`+iterate` 的此 flag 更新 Agent 本身，不仅作用于新 Session。

## 创建 Session 时覆盖

```bash
arkcli agent session create --agent-id <agent-id> --environment-id <env-id> \
  --agent-overrides '{"ContextManagement":[{"Type":"compaction","CompactThreshold":64000,"Instructions":"保留代码路径和测试结果"}]}'
```

也可用 `--agent-overrides @override.json`，或完整请求的 `AgentWithOverrides.ContextManagement`。
这只影响新 Session，不更新原 Agent，不追溯修改已有 Session。清除继承的自定义策略使用
`--agent-overrides '{"ContextManagement":[]}'`。不要把配置放到 Session 请求顶层。

## 验证

创建 / 更新后用 `agent agent get <agent-id> --format json`；创建 Session 后用
`agent session get <session-id> --format json` 检查最终 Agent 配置。回读一致只能证明配置被接收，
不能证明已经触发压缩；触发发生在后续模型请求的上下文达到阈值时，应另查运行事件。
`--dry-run` 只验证本地请求构造，不证明服务端已部署该字段或接受参数。
