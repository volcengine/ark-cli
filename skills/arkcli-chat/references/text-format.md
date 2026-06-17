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

## 典型用法

### json_object (最简, 让模型出 JSON)

```bash
arkcli +chat "草莓什么颜色? 用 JSON 回答" --model ep-xxx \
  --text-format json_object
# {"color":"red"}
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
# {"color":"red","hex":"#FF3333"}
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

## 常见错误

| 现象 | 原因 |
|---|---|
| `--text-format=json_schema requires --text-schema <path>` | json_schema 模式忘加 --text-schema |
| `read --text-schema "X": no such file or directory` | 路径写错或权限不足 |
| `unsupported text.format.type "yaml"` | format 取值不是 text / json_object / json_schema |
| `text.format.schema is required when type=json_schema` | 走 raw API 时漏传 schema |
| 模型输出不符合 schema | 试 `--text-strict` 强约束, 或检查 prompt 是否清楚要求结构化 |

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
