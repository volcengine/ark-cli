---
name: stream-events
description: arkcli +chat --include-events 用法 reference, 流式模式下输出原始 SDK 事件 NDJSON, 供 autotest / agent 程序化消费。
---

# +chat Stream Events (--include-events)

流式模式下把每个 SDK 事件原样输出为一行 JSON (NDJSON), 而非人类可读的 `Thinking:` / `Response:` 文本。

## 何时使用

- **autotest** —— 自动化测试需要逐事件断言 (event type、usage、response id 等)
- **agent / 程序化消费** —— 上游程序需要结构化解析流式输出, 而非从 human-readable 文本里正则提取
- **调试** —— 想看服务端实际下发的每个事件类型和字段

## flag 速查

| flag | 用途 |
|---|---|
| `--include-events` | 流式模式专用; 加上后 stdout 输出 NDJSON (每行一个原始 SDK 事件 JSON) |

`--include-events` **必须配合 `--stream`** 使用; 非流式模式下该 flag 无效。

## 典型用法

### 最简: 流式 NDJSON

```bash
arkcli +chat "解释量子纠缠" --model ep-xxx --stream --include-events
```

输出示例 (每行一个独立 JSON):

```json
{"type":"response.created","response":{"id":"resp_abc123","object":"response","status":"in_progress",...}}
{"type":"response.reasoning_summary_text.delta","delta":"量子纠缠是..."}
{"type":"response.output_text.delta","delta":"简单来说,"}
{"type":"response.output_text.delta","delta":"两个粒子..."}
{"type":"response.completed","response":{"id":"resp_abc123","status":"completed","usage":{"input_tokens":10,"output_tokens":50,"total_tokens":60}}}
```

### 配合 jq 过滤特定事件

```bash
# 只看 delta 事件
arkcli +chat "写一首短诗" --model ep-xxx --stream --include-events | jq 'select(.type | contains("delta"))'

# 只看最终 completed 事件拿 usage
arkcli +chat "1+1等于几" --model ep-xxx --stream --include-events | jq 'select(.type == "response.completed")'

# 提取所有 delta 文本拼成完整回答
arkcli +chat "介绍北京" --model ep-xxx --stream --include-events | jq -r 'select(.type | contains("output_text.delta")) | .delta' | tr -d '\n'
```

### 与 --store + chat get 联动

```bash
# 流式跑完, 从 completed 事件拿 id
RESP_ID=$(arkcli +chat "记住我叫小明" --model ep-xxx --store --stream --include-events \
  | jq -r 'select(.type == "response.completed") | .response.id' | head -1)

# 后续用 chat get 查完整响应
arkcli chat get "$RESP_ID" --format json | jq '{id, content, usage}'
```

## 与 `--format json` 的区别

| 维度 | `--stream --include-events` | `--stream` (不带 include-events) | `--format json` (非流式) |
|---|---|---|---|
| 输出格式 | NDJSON (每行一个事件) | human-readable 文本 | 单个 JSON 对象 |
| 事件粒度 | 每个原始 SDK 事件 | 只输出 text delta | 汇总后输出 |
| 可解析性 | 高 (每行独立 JSON) | 低 (需正则) | 高 (单个 JSON) |
| 实时性 | 实时逐事件 | 实时逐文本段 | 等全部完成 |
| 用途 | autotest / agent 消费 | 人类阅读 | 结构化取完整结果 |

## ChatStreamDelta 事件类型

当 `--include-events` 启用时, service 层的 `ChatStreamDelta` 会携带以下元数据字段:

| 字段 | 说明 |
|---|---|
| `EventType` | SDK 事件类型 (如 `response.output_text.delta`) |
| `RawEvent` | 原始事件 JSON 字符串 (即 NDJSON 输出的每行内容) |
| `Content` | 文本增量 (仅 text delta 事件) |
| `ReasoningContent` | 推理增量 (仅 reasoning delta 事件) |
| `FunctionCallArgs` | 工具调用参数增量 (仅 function_call.arguments.delta 事件) |
| `FinalResponseID` | 响应 ID (response.created / response.completed) |
| `FinalUsage` | 最终用量 (response.completed) |
| `FinalContent` | 文本全文 (output_text.done) |
| `FinalReasoning` | 推理全文 (reasoning_summary_text.done) |
| `Incomplete` | 是否不完整 (response.incomplete) |

## 常见错误

| 现象 | 原因 |
|---|---|
| `--include-events` 没有任何输出 | 忘了加 `--stream`; 该 flag 只在流式模式生效 |
| 输出是纯文本而非 JSON | `--include-events` 没加上, 回退到 human-readable 模式 |
| NDJSON 行数比预期少 | 某些事件类型 (如 `response.output_item.added`) 在 PR-4 才识别; 当前版本只解码常见事件 |
| `jq` 解析报错 | 某些行可能不是有效 JSON (如 `[DONE]` 终止标记); 用 `jq -R 'fromjson?' ` 容错 |

## autotest 解锁清单

| 用例 | 是否解锁 |
|---|---|
| `Test_Stream_IncludeEvents` | ✅ |
| `Test_Stream_EventTypeCoverage` (11 类事件) | ✅ |
| `Test_Stream_NDJSONFormat` (每行独立 JSON) | ✅ |
| `Test_Stream_CompletedEventHasUsage` | ✅ |
| `Test_Stream_TextFormat` (流式入参可传; 回显走 --include-events) | ✅ |
