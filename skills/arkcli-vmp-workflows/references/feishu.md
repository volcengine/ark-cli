# 飞书集成

## 定位

飞书是编排面的副作用适配器，不是能力面。`arkcli` 只产出证据；本文件描述如何把证据变成消息、文档、Base/Sheet 记录，并回读结果。

## 前置检查

- 已安装并登录 `lark-cli`（或宿主提供等价的飞书工具）。
- 机器人在目标群；文档/Base/Sheet 有对应权限。
- 发送前展示目标、标题、摘要和证据链接，取得用户确认。
- 每次副作用使用幂等键，避免重复发送。

## 消息推送

```bash
lark-cli im +messages-send --chat-id oc_xxx --markdown @report.md --idempotency-key vmp-daily-2026-11-08
```

- `--chat-id` 与 `--user-id` 二选一；群聊用 `oc_`，个人用 `ou_`。
- 发送后保存返回的 message ID；失败时保留本地报告，重试一次。
- 没有 `lark-cli` 时，输出可直接粘贴的 Markdown，并标记“尚未发送”。

## 文档落盘

```bash
lark-cli docs +create --title "VMP 日报 2026-11-08" --content @report.md --doc-format markdown
```

- 已存在文档时用 `docs +update` 的 `block_replace`/`append`，不要整篇覆盖。
- 回读 document token 和 revision；报告里给出链接。
- 复杂表格用 DocxXML，普通摘要用 Markdown。

## 多维表格 / 电子表格

- 巡检记录、EP 清单与收藏优先用 Base：`lark-cli base +record-batch-create` / `+record-upsert`。
- 指标明细用 Sheet：`lark-cli sheets +cells-set` 或 `+csv-put`。
- 写入前确认 app token、table id、字段映射；写入后回读 record id 或 range。
- 具体参数以 `lark-cli skills read lark-base` / `lark-sheets` 为准，不在本文件复制完整命令目录。

## 副作用契约

```json
{
  "type": "feishu_message",
  "target": "oc_xxx",
  "title": "VMP 日报 2026-11-08",
  "idempotency_key": "vmp-daily-2026-11-08",
  "result": {"ok": true, "message_id": "om_xxx"},
  "verified": true
}
```

- `verified=true` 只在拿到平台返回 ID 或回读成功后设置。
- 发送失败、权限不足、目标不存在时，`verified=false`，并在报告里写明“未发送/未写入”。

## 失败处理

- 鉴权失败：转 `lark-shared` / `lark-cli auth login`，不要重试到封禁。
- 群不存在或无权限：请求用户提供正确的 chat_id 或邀请机器人。
- 文档写入部分成功：保留 revision 和已写入 block，报告 partial，不重复整篇覆盖。
- 网络错误：重试一次；仍失败则保留本地报告并通知失败接收人。
