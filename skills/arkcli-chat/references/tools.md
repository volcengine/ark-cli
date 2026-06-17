---
name: tools
description: arkcli +chat Tools 用法 reference, 含 web_search 简单语法糖、function 自定义工具的 JSON 文件协议、tool_choice / max_tool_calls 调节。
---

# +chat Tools

PR-1 内置 `function` 与 `web_search` 两种 tool; `mcp` / `knowledge_search` 留 PR-5b。

## 何时使用

- 用户问"今天有什么 AI 新闻" → `web_search`
- 用户提供函数 schema 让模型调 → `function`
- 用户问"帮我查个会议纪要" / "用我们的知识库回答" → 当前 PR 不支持, 需提示用户走 `arkcli api arkruntime.create_responses --params '...'` (raw API) 或等 PR-5b。

## 4 个新 flag

| flag | 类型 | 用途 |
|---|---|---|
| `--tools <type>` | 可重复 string | 简单语法糖, 只支持单字段 `{"type":"<type>"}` 形态。当前实测能用的只有 `--tools web_search`。 |
| `--tools-file <path>` | string | 读取一个 JSON 文件, 内容是一个 tool object 数组, 完整 SDK 形态; function 必须走这条路。 |
| `--tool-choice <mode>` | string | 工具调用策略: `auto` / `required` / `none`。不传则使用服务端默认。 |
| `--max-tool-calls <int>` | int | 单次会话最大 tool 调用次数, 防止失控。**仅对 built-in tools 生效**（如 `web_search`）；与 function tool（`--tools-file` 或 `--tools function:*`）组合时 arkcli 会在客户端就拦截，原因是服务端 `max_tool_calls only supported by build-in tools now`。 |

`--tools` 与 `--tools-file` 可同时存在, 会按"sugar 在前, 文件在后"合并。

## web_search (最简)

```bash
arkcli +chat "帮我查今天与AI相关的几条重要新闻, 给出要点和来源" \
  --model ep-xxx \
  --tools web_search \
  --tool-choice auto \
  --max-tool-calls 3
```

## function (需要 JSON 文件)

`tools.json`:

```json
[
  {
    "type": "function",
    "name": "query_weather",
    "description": "查询某个城市的天气, 返回 JSON",
    "parameters": {
      "type": "object",
      "properties": {
        "city": {"type": "string", "description": "城市名, 如 Beijing"}
      },
      "required": ["city"]
    }
  }
]
```

调用:

```bash
arkcli +chat "北京今天天气怎么样?" \
  --model ep-xxx \
  --tools-file tools.json \
  --tool-choice required
```

`function` object 的字段 (与 SDK `responses.ToolFunction` 对齐):

| 字段 | 必填 | 说明 |
|---|---|---|
| `type` | ✅ | 必须是 `"function"` |
| `name` | ✅ | 函数名, 模型用它来识别 tool |
| `description` | 推荐 | 给模型看的函数说明; 写得越清楚, 模型越能正确触发 |
| `parameters` | 推荐 | JSON Schema 描述参数; 直接嵌套写, arkcli 会内部序列化成 SDK Bytes |
| `strict` | 可选 bool | 严格模式: 强制参数符合 schema |

## tool_choice 三种模式

| mode | 含义 |
|---|---|
| `auto` | 服务端默认, 让模型自行决定要不要调 tool |
| `required` | 强制必须至少调一次 tool |
| `none` | 即使传了 tools 也禁止模型调用 |

复杂的 function-specific tool_choice (`{"type":"function","name":"query_weather"}`) 在后续 PR 实现, 当前请用 `required` 加唯一一个 tool 等效达成。

## 输出形态

### 非流式 (`+chat ...` 不带 `--stream`)

`ResponsesResult` 多了 `function_calls` 字段:

```json
{
  "id": "resp_...",
  "model": "...",
  "content": "...",                  // 模型最终文本; tool 触发时可能为空
  "reasoning_content": "...",
  "usage": { ... },
  "function_calls": [
    {
      "type": "function_tool_call",
      "id": "call_...",
      "name": "query_weather",
      "arguments": "{\"city\":\"Beijing\"}",
      ...
    }
  ]
}
```

`function_calls[]` 是原始 SDK 形态透传; 不同 tool type 字段不同 (`function_tool_call` / `function_web_search` / 等), 用 `jq '.function_calls[].type'` 区分。

### 流式 (`--stream`)

stdout 上会出现 `[Tool args] {"city":"Beijing"}` 这样的标签行, 在 `Response:` 之前 / 之后插入, 取决于事件顺序。`ChatStreamDelta.FunctionCallArgs` 的所有 chunk 拼起来就是完整 args JSON。

## 错误处理

| 现象 | 原因 |
|---|---|
| `tool[N] type "X" not supported in this build` | 用了 PR-5b 才支持的 tool type (mcp/knowledge_search/image_process); 走 raw API 或等 PR-5b |
| `function tool requires 'name'` | function 缺 name 字段 |
| `unsupported tool_choice "X"` | tool_choice 不是 auto/required/none |
| 模型回复但没触发 tool | tool description 写得不够清楚, 或 tool_choice=none, 或问题与 tool 不相关 |
| 模型反复调用同一 tool | 加 `--max-tool-calls` 限制, 或在 tool description 里说明何时停止 |

## 与 chat get 联动

加了 `--store` 之后, 即使流式跑完, 后面也能用 `chat get <id>` 拿到完整 `function_calls` 列表; `chat list-input-items <id>` 还能看到完整的 tool 调用历史 (含 `function_tool_call_output` 这种结果项)。
