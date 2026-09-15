# Events 与 Chat

## Event / Chat 体验

先按用户是否需要看到 Agent 回复选择入口：

| 场景 | 推荐入口 | 收尾要求 |
| ---- | ---- | ---- |
| 创建新 Session 并直接对话 | `arkcli +new session <agent-id> --message "..."` | CLI 自动监听到终态并输出回复 |
| 已有 Session 发送问题并等待回复 | `agent session events send <session-id> --type user.message --text "..." --stream` | 建立 SSE，等待 `agent.message` 和 `idle`/其他终态 |
| 大 payload（约 50KB 以上）或长耗时调研/报告任务 | `agent session events send <session-id> ... --poll` | CLI 只轮询 `events list`，不建立 stream；更长任务仍可不带 `--stream/--poll`，由调用方分步轮询 |
| 只投递事件，不等待 Agent | `agent session events send ...` | 仅适用于用户明确要求异步投递或脚本写入 |
| 观察已有 Session 的实时事件 | `arkcli +tail <session-id>` | 持续读取到终态或明确错误 |

### Event Deltas

数据面 `/events/stream` 支持按查询参数请求低延迟增量帧。CLI 的实时入口默认发送：

```text
event_deltas=agent.message&event_deltas=agent.thinking
```

对这两类事件，服务端可能依次返回：

1. `event_start`：声明本次完整事件的 `id/type`；
2. 多个 `event_delta`：增量文本位于 `delta.content[].text`，通过 `event_id` 关联到 `event_start`；
3. 最终的完整 `agent.message` 或 `agent.thinking` 事件。

`+tail`、`events send --stream`、`+new session` 和 `+iterate` 的 pretty 输出会把增量显示成 `[agent delta #N]` / `[thinking delta #N]`，最终完整事件只显示一次完成标记，避免重复刷屏。`--raw` 保留原始 `event_start`、`event_delta` 和最终事件，适合调试或落盘；`agent session events stream` 默认输出原始 NDJSON。

如果服务端或兼容环境不接受增量查询，使用 `--no-event-deltas` 回退到完整事件：

```bash
arkcli +tail <session-id> --no-event-deltas
arkcli agent session events send <session-id> --text "..." --stream --no-event-deltas
arkcli agent session events stream <session-id> --no-event-deltas
```

`events list` 是历史回查接口，不发送 `event_deltas`；断线补偿会继续按 cursor 拉取已落库的完整事件。增量帧与最终事件的重复投递由 CLI 按事件类型、事件 id 和增量内容去重。

**硬规则：用户说“询问 Agent”“让 Agent 分析/回复/执行”时，不能把 write-only 的 `events send` 当作完整流程。短请求使用 `--stream`；大 payload 或长任务优先使用 `--poll`，更长任务使用不带 `--stream/--poll` 的 send，再用 `events list --after` 轮询或 `+tail` 读取回复。**

- `agent session list --agent-id <agent-id>` 按 Agent 过滤；CLI 会按 `ListSessionsForTop` 契约发送 `AgentIds` 数组。
- Session 列表使用页码分页：`--page` 对应 `PageNumber`，`--limit` 对应 `PageSize`。需要遍历时使用全局 `--page-all`，并通过 `--page-limit` 控制最多页数。

- `events send` 主动发事件。常用：

```bash
arkcli agent session events send <session-id> --type user.message --text "帮我分析这个数据" --stream
```

- `+new session` 和 `+iterate` 的等待链路使用内部 event channel：先消费 `/events/stream`，连接结束或传输失败后按最后一个 event id 调 `/events?order=asc&limit=100&after=...` 补偿，再重连；stream 与 list 的重复事件按 id 去重。只有 idle、requires_action、failed 或 terminated 等终态才结束，耗尽重连次数会返回明确错误，不会把断流当成成功。
- `+tail` 的 pretty/raw 输出也使用同一个 channel，因此断流后的 list 补偿事件会沿用相同的输出格式；pretty 模式默认将 user/thinking 等噪声文本压缩展示（普通事件最多 2000 字符，user/thinking 最多 320 字符），保留 event id 和省略数量；需要完整内容使用 `--raw` 或 `--max-event-chars 0`。底层 `agent session events stream` 保持原始流式接口语义。

- `events send` 是写入接口，返回成功只表示事件已接收，不表示 Agent 已完成回复。AI Agent 编排已有 session 的对话时，必须继续执行：`events stream`/`+tail` 持续读取到 `session.status_idle`、`requires_action`、`failed` 或明确错误；需要最终 session 元数据时，再调用 `agent session get <session-id> --format json`。
- 如果希望底层 send 命令直接完成一次对话，可加 `--stream`：CLI 会从发送响应提取最后一个 event id，建立 SSE 并进入可靠 channel，输出 Agent 回复并等待终态；`--raw` 可保留原始事件输出。没有 `--stream` 时仍是 write-only，适合脚本只提交事件。`--wait` 保留为兼容别名。
- `--stream` 适合短、边界明确的请求；它先在前台消费事件 stream，默认最长 120 秒（CLI 本地默认值，可用 `--wait-timeout` 覆盖）。stream 等待超时后会自动切换为基于 cursor 的 events list polling，默认再持续 120 秒，可用 `--wait-fallback-timeout` 覆盖；两阶段都超时才返回带 cursor 的非 0 错误。显式 `--stream` 即使 payload 约 50KB 以上也会建立 SSE；只有旧 `--wait` 兼容模式可能直接返回。长任务仍优先使用 `--poll`。
- `--poll` 会在发送成功后只轮询 `/events` 列表，不建立 `/events/stream` 长连接；它仍会占用当前 CLI 进程直到终态，但更适合规避 stream 连接不稳定的长请求。默认每 2 秒轮询，可用 `--poll-interval` 调整。
- `events send --stream` 和 `+new session` 也支持 `--max-event-chars` 控制 pretty 输出；该参数只影响终端展示，不改变发送给 Agent 的原始 payload。`--raw` 仍输出完整事件，适合机器留档。
- 长任务推荐“send + cursor 轮询”：先发送并保留返回中的最后一个 event id，再每隔 2-5 秒执行 `agent session events list <session-id> --after <event-id> --order asc --limit 100 --format json`；读取新增 `agent.message`、`tool`、`thinking` 等事件，直到 `idle`、`requires_action`、`failed`、`terminated` 或 `archived`。需要实时人类可读输出时改用 `arkcli +tail <session-id>`。
- 轮询单次超时或连接中断时，可以保留同一个 cursor 重试；遇到鉴权、权限、参数、开通状态或其他明确业务错误时立即停止。轮询达到总超时时间后必须报错并保留最后状态，不要报告为 Agent 已完成。
- 推荐的长任务流程：
  1. **单独执行** `arkcli agent session events send <session-id> --type user.message --text "..." --poll`；如果任务可能超过当前 shell/tool 超时时间，则不要传 `--poll`，改为不带等待的 send。
  2. **单独执行解析步骤**，从发送响应提取 event cursor；如果响应没有 cursor，改用发送时间或最近一条 user event 作为回查起点，并避免从 session 历史开头读取。
  3. **单独执行并重复轮询** `arkcli agent session events list <session-id> --after <cursor> --order asc --limit 100 --format json`，每轮处理新增事件并更新 cursor；发现终态后停止。
  4. **单独执行** `arkcli agent session get <session-id> --format json`，需要最终 Session 元数据时再调用。
- 不要写成下面这种串联命令：`SID=$(...); arkcli session create ...; arkcli events send "$SID" ...; arkcli events list "$SID" ...`。如果 create 卡住，后面的 send/list 根本不会执行，也无法判断卡点。
- `session create` 或 `events send` 出现超时后，不要立即重试写请求：先单独执行 `session get/list` 或 `events list` 判断第一次请求是否已生效。只有确认未生效，或服务端提供幂等键并已使用时，才可重试。
- `session get` 只返回 Session 状态、配置和关联信息，不包含 Agent 的完整自然语言回复；回复必须从 `agent.message` 等事件中读取。不要用 `session get` 替代 events stream。
- 新建 session 并直接聊天时，优先使用 `arkcli +new session <agent-id> --environment-id <env-id> --message "..."`：CLI 会创建 session、发送 `user.message` 并等待数据面事件流；仍应从输出中的 session id 和 agent 事件确认结果。若使用底层 `agent session create`，必须自行完成上述 send/stream/get 收尾。

- `--events` / `--file` 可以提供多条事件。由于当前线上数据面接口实际只接受单条 event，CLI 会按数组顺序逐条发送；这保证顺序，但不保证原子性。返回结果会标记 `transport=serial`、`atomic=false` 和 `sent` 数量。中途失败时不会自动重试，错误会指出失败的 `event[index]` 以及已经发送的数量。
- 不要把这种 CLI 兼容行为理解成服务端已经支持原子 batch；如果业务必须原子提交，应等待服务端支持多事件请求。
- 多事件可以包含不同 `type`，CLI 不会按类型重排。例如先发送 `user.message` 再发送 `user.interrupt`，服务端会先启动消息对应的 turn，再处理 interrupt；回查时应看到原始 `user.message`、`user.interrupt`，以及被中断的 model request。`user.interrupt` 要在存在 active turn 时发送；对 idle session 发送会产生 `session.error: no active turn to cancel`。

- `events list` 拉历史，支持 `--after/--before` event cursor，`--since/--created-after/--created-before` 时间过滤，`--type` 可重复或逗号分隔。加全局 `--page-all` 后默认 `limit=100`，沿响应 `next_page -> page` 拉取并合并 `data`。
- `threads list` 同样支持全局 `--page-all` 和 `next_page -> page`；`resources list` 当前没有分页契约，不要伪造 page 参数。
- `events stream` 输出机器友好的 SSE data / NDJSON 行。
- 已知 session-id 时优先使用主动压缩命令：`arkcli agent session compact <session-id> [--instructions <text|@file>] [--session-thread-id <id>]`。它先确认 Session 为 idle，再发送前端同款 slash envelope，并从发送响应最后一个 event id 之后等待。正常的 `session.thread_status_idle(stop_reason=end_turn)` 与 `session.status_idle(stop_reason=end_turn)` 表示手动 compact 命令执行完成；`agent.thread_context_compacted` 是可选的“实际发生压缩”确认，后端在没有可压缩消息时可能不返回它。失败/终止事件或超时返回非 0。
- `session compact` 默认 SSE 等待 120 秒，`--poll` 改用 events list 轮询，`--raw`/`--format jsonl` 输出原始事件。`--dry-run` 是零网络 Workflow Client Preview，会展示 idle 预检、XML 转义后的 payload、cursor 等待、双 idle 成功谓词和可选 compacted 确认事件。
- REPL 的 `/compact <instructions>` 会把整段 instructions XML 转义后传入 `<command-args>`；`/clear` 语义不变。不能只凭 slash message 发送成功报告命令完成，也不能在仅有 idle、没有 `agent.thread_context_compacted` 时声称已确认实际折叠了历史消息。
- `+tail` 输出人类可读短行，默认归类 `[user]`、`[agent]`、`[thinking]`、`[tool]`、`[tool_result]`、`[model]`、`[status]`、`[action]`、`[error]`、`[outcome]`。机器读取用 `+tail --raw`。
- `+debug <session-id>` 会在 `event_type_count` 外额外返回 `event_delta_count` 和 `event_delta_by_type`，用于确认服务端是否实际返回增量帧；`recent` 中的增量和最终事件也按同一关联状态聚合展示。
- `+chat <prompt>` 保留为 Responses API 快速对话，不进入 Managed Agent。
- `+new session` 是 Managed Agent session 入口（PRD 早期写作 `+chat` 的地方都按这个命令理解）：
  - `arkcli +new session`：打开交互选择器，先 `ListSessions`，可选择已有 session 进入 REPL；也可选 `[新建]` 后 `ListAgents`、`ListEnvironments`，再 `CreateSession` 进入 REPL。
  - `arkcli +new session <agent-id> --environment-id <env-id>`：直达创建新 session，再发送首条消息或进入 REPL。
  - TTY 下进入 REPL。
  - 非 TTY stdin 或 `--message` 是 one-shot。
  - `/exit` 退出，`/interrupt` 发 `user.interrupt`。
  - `/compact` 和 `/clear` 会按前端同样的 slash envelope 发送 `user.message`；这是协议触发入口，不要仅凭发送成功判断后端已完成操作。手动 `/compact` 以正常 thread idle + session idle 收口；`agent.thread_context_compacted` 仅在服务端实际压缩时提供更强确认，可能因没有可压缩消息而缺失。`/clear` 的具体结果以服务端返回事件为准。二者都等待本轮完成后继续 REPL。
  - `/allow [tool_use_id]`、`/deny [tool_use_id] [reason]` 发送 tool confirmation。
- 已知 `session-id` 的脚本/非交互场景不要用选择器，改用 `arkcli agent session events send <session-id>`、`arkcli agent session events stream <session-id>` 或 `arkcli +tail <session-id>`。
- 面向 AI 工具调用的固定收尾顺序。短请求需要 Agent 回复时优先使用下面的单命令路径：
  1. `agent session events send <session-id> --type user.message --text "..." --stream`。
  2. 从输出读取 `agent.message` 等回复事件，并确认 `idle`、`requires_action`、`failed` 或其他终态。
  3. 需要状态、版本、环境或标题等最终元数据时，再执行 `agent session get <session-id> --format json`。
- 如果必须拆开发送和监听，则使用以下固定收尾顺序：
  1. 单独执行 `agent session create` 或取得已有 `<session-id>`，成功后立即保存 ID。
  2. 单独执行 `agent session events send <session-id> --type user.message --text "..."`。
  3. 单独执行 `agent session events stream <session-id>` 或 `+tail <session-id>`，读取 Agent 回复并等待本轮终态。
  4. 单独执行 `agent session get <session-id> --format json`，需要状态、版本、环境或标题等最终元数据时执行。
- 重试顺序：网络超时/连接中断/429/5xx 可有限重试并指数退避；4xx 参数、鉴权、权限、开通和业务状态错误直接返回。写请求结果未知时先查后重试，读请求可以直接重试。
- `+new session` 选择器里的 `[全部]` 会用已知状态集合重新拉取 session，尽量包含 terminated/archived；如果后端新增状态枚举，需要同步更新。
- `+new session` 选择器里的 `[起新 agent]` 目前只提示转去 `arkcli +new-agent "..."`，不会在选择器内嵌自然语言建 agent 流程。
- `events send` 支持多模态便捷参数：
  - `--image file-xxx` 直接发送图片 file source；`--image @./a.png` 或 `--image ./a.png` 会先上传 Files API、等待 active，再发送图片。
  - `--document file-xxx` 直接发送文档/PDF file source；`--document @./a.pdf` 或 `--document ./a.pdf` 会先上传 Files API、等待 active，再发送文档。
  - 图片 content schema：`{type: image, source: {type: file, file_id: ...}}`。
  - 文档/PDF content schema：`{type: document, source: {type: file, file_id: ...}}`。
  - `--image/--document` 不能和 `--event/--events` 混用；手写事件时直接在 `--event/--events` 里按上述 schema 传完整 content。
- Event 关联字段会在 CLI 发送前统一校验，typed flags、`--event`、`--events`、`--file` 行为一致：
  - `user.custom_tool_result` 必须带 `custom_tool_use_id`（typed flag 为 `--custom-tool-use-id`）。它不是所有 event 的全局必填字段，但对 custom tool result 是语义必填，否则结果无法关联到具体的 custom tool 调用。
  - 旧版服务端曾经错误放行缺少该字段的事件，并将它落成空字符串；不要依赖这种兼容行为。当前 CLI 会在请求前 fail-fast，避免产生孤儿事件；服务端也应拒绝该请求。
  - CLI 只校验字段非空，ID 是否存在、是否属于当前 session、是否仍待处理由服务端判断。
  - `user.tool_confirmation` 必须带 `tool_use_id`，且 `result` 只能是 `allow` 或 `deny`。
  - `user.tool_result` 只允许用于 `self_hosted` environment。实际发送时 CLI 会通过 `GetSession -> GetEnvironment` 预检 `Config.Type`；`--dry-run` 不发起该读取，也不证明环境满足约束，最终仍由真实执行与服务端校验。
  - raw payload 可使用数据面 snake_case；兼容的 PascalCase 别名会在发送前归一化成 snake_case。

### Custom tool result

`user.custom_tool_result` 用于回传某次 custom tool 调用的结果，必须提供对应的 `custom_tool_use_id`：

```bash
arkcli agent session events send <session-id> \
  --type user.custom_tool_result \
  --custom-tool-use-id <custom-tool-use-id> \
  --content "tool output" \
  --format json
```

等价的结构化事件：

```json
{
  "type": "user.custom_tool_result",
  "custom_tool_use_id": "ctu-xxx",
  "content": [{"type": "text", "text": "tool output"}]
}
```

缺少或传空 `custom_tool_use_id` 时，CLI 返回 validation 错误且不会调用数据面 API。不要用 `user.tool_confirmation` 的 `tool_use_id` 代替；两者属于不同 event 类型。

## 示例

```bash
arkcli agent session list --agent-id <agent-id> --page 1 --limit 20 --format json
arkcli --page-all --page-limit 10 agent session list --agent-id <agent-id> --limit 100 --format json
arkcli agent session events send <session-id> --type user.message --text "<one small test task>" --format json
arkcli agent session events send <session-id> --type user.message --text "<one small task>" --stream
arkcli agent session events send <session-id> --events '[{"type":"user.message","content":[{"type":"text","text":"先执行一个短任务"}]},{"type":"user.interrupt"}]' --format json
arkcli agent session events send <session-id> --text "看这张图" --image file-xxx --format json
arkcli agent session events send <session-id> --text "总结这个 PDF" --document @./report.pdf --format json
arkcli agent session events list <session-id> --limit 20 --format json
arkcli agent session compact <session-id> --instructions @./compact-instructions.txt --format json
arkcli agent session compact <session-id> --poll --timeout 180 --raw
arkcli +new session
arkcli +tail <session-id> --session-thread-id <thread-id>
arkcli +new session <agent-id> --environment-id <env-id> --message "帮我分析这个数据" --format json
```
