---
name: chat-meta
description: arkcli chat get / delete / list-input-items 三个标准 CRUD 命令的 reference, 与 +chat 工作流配套使用以管理已 store 的对话。
---

# chat get / delete / list-input-items

这是 `+chat` 的"对账面"——`+chat` 负责发起对话并持久化；这三个命令负责事后查询、删除、查看输入历史。

## 何时使用

- `chat get` —— 已经有 response_id, 想拿回完整对话内容（含 reasoning, function calls, usage）。
- `chat delete` —— 删掉之前 `--store` 留下的 response 记录。
- `chat list-input-items` —— 想看某条 response 的 input 列表（适用于多轮接续后回溯历史）。

## 前置条件

- 这三个命令操作的 response **必须是 `+chat --store` 持久化过的**。无 `--store` 的 response 在服务端不留底, `chat get` 会返回 `InvalidParameter.PreviousResponseNotFound`。
- 鉴权与 `+chat` 一致, 见 `../arkcli-auth/SKILL.md`。

## 命令速查

```bash
# 1. 创建对话并持久化, 拿到 response id
RID=$(arkcli +chat "草莓什么颜色？" --model ep-xxx --store --format json | jq -r .id)

# 2. 查询
arkcli chat get "$RID"

# 3. 查询时让服务端展开 input_image 的 base64 (实测对调试图片输入有用)
arkcli chat get "$RID" --include image_url

# 4. 列出 input items (倒序最新 5 条)
arkcli chat list-input-items "$RID" --order desc --limit 5

# 5. 翻页 (拿到 first_id / last_id 后接着拉)
arkcli chat list-input-items "$RID" --after <last_id_from_previous_page> --limit 20

# 6. 删除
arkcli chat delete "$RID"
```

## flag 说明

### chat get

| flag | 说明 |
|---|---|
| `--include` | 让服务端在响应里展开扩展字段, 可重复。可选值（CLI 本地校验）: `image_url` / `audio_url` / `encrypted_content`。**警告**：SDK 与生产后端在 include 枚举字符串上当前存在 skew，部分账号 / 接入点对所有三个值都会回 `Invalid value for 'include'`；如遇此错，去掉 `--include` 即可（响应主体内容仍会正常返回）。 |

### chat list-input-items

| flag | 说明 |
|---|---|
| `--after <item_id>` | 游标分页: 列出该 item 之后的内容 |
| `--before <item_id>` | 游标分页: 列出该 item 之前的内容 |
| `--limit <int>` | 单页大小, 1-100, 服务端默认 20 |
| `--order asc\|desc` | 排序方向, 服务端默认 desc |
| `--include` | 同 `chat get` |

### chat delete

无额外 flag。

## 输出形态

- `chat get` 复用 `+chat` 的 `ResponsesResult` 结构: `{id, model, content, reasoning_content, usage, function_calls?}`。
- `chat list-input-items` 返回 `{object, data:[…], first_id, last_id, has_more}`, `data[]` 是原始 SDK 形态的 InputItem JSON (16 路 oneof, 不做 arkcli 侧 mirror); 用 `jq '.data[].type'` 看每项是 `input_message` / `function_tool_call` / `function_web_search` / `reasoning` / 等中的哪一种。
- `chat delete` 返回 `{id, deleted:true}`; 服务端无响应体, 这里是 arkcli 合成的。

## 常见错误

| 现象 | 原因 |
|---|---|
| `InvalidParameter.PreviousResponseNotFound` | response_id 拼错, 或创建时未 `--store` |
| `chat get/delete/list-input-items` 报 `missing response id` | positional arg 没传 |
| `tool/function call` 在 `data[]` 没看到 | 创建时未传 `--tools`; 或模型未触发 tool, 见 `references/tools.md` |
