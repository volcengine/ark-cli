---
name: text-format
description: arkcli +chat --text-format 用法 reference, 让模型按指定格式 (text / json_object / json_schema) 输出, 配合 --text-schema 强约束 JSON Schema。
---

# +chat Text Format

让模型按结构化格式输出, 三种模式覆盖从"自由文本"到"严格 JSON Schema"的完整谱系。

## 何时使用

- **text** —— 默认; 自由文本
- **json_object** —— 让模型输出合法 JSON, 不约束 shape; 适合简单"返回一个 JSON" 场景
- **json_schema** —— 严格 JSON Schema; agent / 下游程序需要稳定 shape 时必用 (减少解析失败)

## flag 速查

| flag | 用途 |
|---|---|
| `--text-format` | text \| json_object \| json_schema |
| `--text-schema <path>` | JSON Schema 文件路径; json_schema 模式必填, 其它模式忽略 |
| `--text-schema-name` | schema 命名; 服务端 echo 时显示; 默认 `arkcli_response` |
| `--text-strict` | 强约束开关; 只在 json_schema 模式生效 |

启用 `--text-strict` 后，CLI 同时承担客户端验收：

1. 请求前本地编译 Schema；无效 Schema 直接报 validation error，不发网络请求。
2. 只接受状态为 completed 的响应；incomplete（包括 token 截断）按 `invalid_response` 非零退出。
3. `content` 必须是直接 JSON，不能带 Markdown code fence 或额外解释文字。
4. JSON 必须通过同一份 Schema 校验；不匹配时按 `invalid_response` 非零退出。

`--stream --text-strict` 会先缓冲事件，到 terminal event 后完成上述校验，再按原顺序输出；校验失败时不向 stdout 输出半截 JSON。未启用 strict 的流式和非流式行为保持不变。

## 典型用法

### json_object (最简, 让模型出 JSON)

```bash
arkcli +chat "草莓什么颜色? 用 JSON 回答" --model ep-xxx \
  --text-format json_object
# CLI 返回 JSON 包装；其中 .content 字符串的内容例如 {"color":"red"}
```

### json_schema (强约束 shape)

```bash
cat > schema.json <<'JSON'
{
  "type": "object",
  "properties": {
    "color": {"type": "string", "description": "颜色名称"},
    "hex":   {"type": "string", "pattern": "^#[0-9A-Fa-f]{6}$"}
  },
  "required": ["color", "hex"]
}
JSON

arkcli +chat "草莓什么颜色? 给出颜色名和 hex" --model ep-xxx \
  --text-format json_schema --text-schema schema.json --text-strict
# CLI 返回 JSON 包装；其中 .content 字符串的内容例如 {"color":"red","hex":"#FF3333"}
```

### 接续多轮 + json_schema

```bash
RID=$(arkcli +chat "草莓什么颜色?" --model ep --store \
  --text-format json_schema --text-schema schema.json --text-strict \
  --format json | jq -r .id)

arkcli +chat "苹果呢? 同样格式" --model ep --store \
  --previous-response-id "$RID" \
  --text-format json_schema --text-schema schema.json --text-strict
```

## 输出形态

非流式 (`+chat ...` 不带 `--stream`) 时, `ResponsesResult` 多了 `text_format` 回显:

```json
{
  "id": "resp_...",
  "model": "...",
  "content": "{\"color\":\"red\",\"hex\":\"#FF3333\"}",
  "usage": { ... },
  "text_format": "json_schema"
}
```

`text_format` 是服务端实际应用的格式, 用 `chat get $RID` 也能查回来 (autotest jsonschema_test 在 chat get 时断言这个字段)。

## 保存结构化正文

`--text-format` 约束模型正文，不会去掉 CLI 的 JSON 包装。用户要 JSON 文件时，
交付的是 `.content` 的原文，不是整个响应、推理文本或 Agent 重新组织的答案。

执行前确定本次已准入的 `MODEL`、用户的 `PROMPT`、对应的 `schema.json` 和交付路径。
以下为非流式 strict 示例，假设用户要求保存到新的 `result.json`；文件已存在时先处理
路径/覆盖意图，不先产生用量。按用户任务补充其他已核对的参数，保留本次 Profile 上下文。

```bash
test ! -e result.json && test ! -L result.json &&
chat_response_file="$(mktemp ./arkcli-chat-response.XXXXXX)" &&
arkcli +chat --model "$MODEL" --format json \
  --text-format json_schema --text-schema schema.json --text-strict \
  "$PROMPT" > "$chat_response_file" &&
(set -C; jq -ej 'select(.status == "completed") | .content | select(type == "string" and length > 0)' \
  "$chat_response_file" > result.json) &&
jq empty result.json
```

这里只有一次模型请求。`jq -j` 原样写出正文、不额外加换行；末行只验证 JSON 语法，
不重写正文，避免重新序列化改变空白或大整数。局部 `set -C` 拒绝覆盖请求期间新出现的
目标文件或符号链接，不改变后续 shell 设置。strict 的 Schema/完整性验收仍由 CLI
负责；任一步失败都不能宣称交付成功。普通文本保存不需要 JSON 语法检查。
本地提取、路径或写入失败时保留捕获文件并修复本地步骤，不再调用模型来“重新保存”。
从该文件也可读取 `.id` / `.usage`，不为本地备份额外开启 `--store` 或调用 `chat get`。
用户明确要求下一轮/重新生成时才安排对应新请求，不把这一保存流程当成多轮限制。

## 常见错误

| 现象 | 原因 |
|---|---|
| `--text-format=json_schema requires --text-schema <path>` | json_schema 模式忘加 --text-schema |
| `read --text-schema "X": no such file or directory` | 路径写错或权限不足 |
| `unsupported text.format.type "yaml"` | format 取值不是 text / json_object / json_schema |
| `text.format.schema is required when type=json_schema` | 走 raw API 时漏传 schema |
| `invalid --text-schema for --text-strict` | Schema 不是合法、可编译的 JSON Schema；请求尚未发出 |
| `invalid_response` 且提示 `not completed` | 严格响应被 token cap 等原因截断；提高 `--max-output-tokens` 或缩短输出 |
| `invalid_response` 且提示 `not a direct JSON value` | 返回内容含 Markdown fence、解释文字或非法 JSON |
| `invalid_response` 且提示 `did not match --text-schema` | 返回 JSON 不符合 Schema；检查 Schema、模型能力与 prompt |

## 与 raw API 等价

`+chat --text-format json_schema --text-schema f.json --text-strict --text-schema-name color` 等价于:

```bash
arkcli api arkruntime.create_responses --params '{
  "model":"ep-xxx",
  "input":"草莓什么颜色?",
  "text":{
    "format":{
      "type":"json_schema",
      "schema":{...},
      "name":"color",
      "strict":true
    }
  }
}'
```

## autotest 解锁清单

| 用例 | 是否解锁 |
|---|---|
| `responseapi/jsonschema/TestJson1..3` | ✅ |
| `responseapi/jsonschema/TestJsonCache1..3` | ✅ |
| `responseapi/jsonschema/TestJsonSchema1..3` | ✅ |
| `responseapi/jsonschema/TestJsonSchemaCache1..3` | ✅ |
| `Test_ResponseCreate_TextFormat` (含 text + json_object) | ✅ |
| `Test_Stream_TextFormat` | ✅ (流式入参可传; 回显走 PR-4 --include-events) |
