# 失败模式与降级策略

判据统一看结构化错误里的 `error.type`，以及命令成功时的 `next_action`。**不要用
错误文案做字符串匹配**，那会随语言切换而失效。

## 零命中与搜索链路故障

| 现象 | 结构化信号 | 处置 |
|---|---|---|
| 搜索返回空 `results` | `next_action` 给出目录降级指引 | 改跑 `arkcli docs list --query "<原查询>"`；过滤为空再去掉 `--query` 看完整目录。确认目录里也没有，才说「公开文档未覆盖」 |
| 搜索接口未注册 / 路由不通 | `error.type=api`，`error.hint` 带目录降级指引 | 报告 request_id 与「搜索后端未就绪」，**同时**给出 `docs list --query` + `docs get` 的可用路径 |
| 搜索后端 5xx | `error.type=service_unavailable`，带同一 `hint` | 有界重试（≤2 次），仍失败就按上一行处置 |
| 限流 | `error.type=rate_limit` | 退避后有界重试；持续失败就报告并降级到目录浏览 |
| 凭证缺失 / 过期 | `error.type=auth`，或统一闸门的 `not configured` / `not logged in` | 停止业务重试，转 `arkcli-auth`；已有登录授权时走同一环境的两段式 `--no-browser`，需要授权码时等待用户，登录后恢复原请求。不擅改配置、不索要检索后端密钥 |

核心判断：**搜索不可用 ≠ 文档不可读。** `get` / `list` 读的是产品 CDN 上的已发布
产物，和搜索是两条链路。任何时候都不要因为搜索失败就告诉用户「查不到相关文档」。

以上 `--query` 降级仅适用于公共后端。MCP 零命中或搜索失败时使用普通
`docs list`，再 `docs get`；如果也失败，就报告 MCP 不可用。目录筛选续页必须
保留原 query、snapshot 和 next_offset，不能仅复制 snapshot。

续页的 `network` 错误（包括 HTTP 404）不能直接证明 CDN 或 offset 有问题。
先将实际 `--snapshot` 与首个成功响应逐字比较；不同则从原 JSON 提取后纠正。
不要删除 snapshot 保留 offset 重跑。只有原值一致仍失败，才报告这次读取失败；
没有更多证据时不声称是「已知 CDN 故障」。

## 读取类错误

| `error.type` | 含义 | 处置 |
|---|---|---|
| `not_found` | 这篇文档不在当前发布目录里（已下线、改路径，或搜索索引落后） | 用 `docs list` 找当前路径，或换一条搜索结果。**绝不能把 `snippet` 当全文回答。** 拿不到正文就明说读不到 |
| `snapshot_expired` | 文档重新发布、目录有文档下线，或**搜索快照过期/与查询不匹配**（搜索快照约 5 分钟） | 丢弃已读的局部结果从头重来：读取去掉 `--snapshot` 且 `--chunk-start` 回 0；搜索去掉 `--snapshot` 和 `--offset` 回第一页。不要猜 offset，不要拼不同批次 |
| `section_not_found` | 锚点 ID 或标题不存在。**最常见的原因是拿了搜索结果的锚点**：整页命中时它是文档 H1 的 slug，而已发布大纲不含 H1 | 先跑 `--outline` 看真实 ID，再选一节；或直接读整页。也核对是不是把整个 URL 当了选择器、或自己解码过锚点。**不要编锚点、不要反复重试、不会静默回退全文** |
| `section_ambiguous` | 标题重名 | 改用唯一的 `#锚点ID` |
| `section_unavailable` | 已发布标题层级与 Markdown 正文对不上（常见于自定义 DSL 密集页） | 说明该页不支持按章节读，改读全文并按块拼接 |
| `validation` | 参数超界、选择器冲突、MCP 能力差异、未配置文档后端 | 按 commands.md 的能力表修正。公共续读保留 snapshot；MCP 从头用 chunk 参数读取，不拼接其他后端或版本的片段 |
| `invalid_response` | 搜索结果或目录的渠道 / 版本 / 数据结构不符，正文校验失败，或标题元数据无效 | 当作后端异常上报，不循环重试，不要降级去读未校验内容 |

公共后端会自动驱逐未通过校验的单条缓存，并回源一次；新下载的无效内容不会
写入缓存。仍报 `invalid_response` 时无需手工清缓存；上游修复后正常重跑即可。

## 自定义 DSL 造成的退化

已发布 Markdown 保留文档站自定义标签（`<Card>` / `<Columns>` / `<ColumnsItem>` /
`<Tabs>` / `<Tab>` / `<Note>` / `<Tip>` / `<Warning>` / `<Danger>` / `<APILink>` /
`<Attachment>` / `<RenderMd>` / `<span id>`，以及尾部带 `{target="_self"}`
属性块的内联链接）。CLI 不降级它们。

后果与处置：

- **`--section` 可能失效**：自定义标签里的标题不进 Markdown AST，会触发
  `section_unavailable`。→ 改读全文。
- **分块边界可能切断 DSL 块**：单个超大块会被强制拆分。→ 顺序拼接多个 `content`
  再解读，不要在半个标签里判断内容完整性。
- **不要把标签当内容**：只引用标签内部的文字与围栏代码块。给用户的答案里不应出现
  `<Card>` / `<Tab>` 这类标记。

## 有界重试规则

- 网络类（`network`、`service_unavailable`、`rate_limit`）：最多 2 次退避重试。
- `auth`、`not_found`、`validation`、`invalid_response`、`section_*`：**一次都不要重试**，它们不会因为
  再试一次而变化。
- 任何情况下都不要对同一个缺失路径循环 `get`，也不要靠猜 hash / CDN URL 去读已下线
  内容。
