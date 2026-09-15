# +chat

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

通过数据面 Responses API（`POST /responses`）的高层封装。一条命令即返回助手回复，支持多模态、推理、流式、多轮。

## 命令

`--model` 使用本次 Profile/临时上下文兼容的调用 ID；模型族名、合法套餐别名、版本化 ID 与 EP 不可任意互换。
用户要用默认模型时，可以省略 `--model` 让 CLI 使用 text default；已确认合法的套餐别名
（如 `ark-code-latest`）保持原样，不因为 `models get` 返回了规范 Name/Version 就替换。
仅当用户给的是未解析族名时才查询权威模型详情及可调用资源，不正则猜版本，也不统一追加 primary_version。
下面是语法示例，不是所有模型都支持的固定组合；执行前按主 Skill 核对参数能力与实际值。

**先安排交付，再发请求：** 用户要求保存结果时，非流式调用先把完整 JSON stdout
捕获到本次新建的本地文件。成功后，从同一响应提取 `.content`、`.id`、`.usage`；
保存、检查格式、核对正文都在本地完成，不为这些步骤再次运行已经成功的 `+chat`。
下面的示例是不同任务的选项，不是为了取得多个字段逐条重跑的步骤。
结构化正文的捕获/保存示例见 [text-format.md](text-format.md#保存结构化正文)。
这不限制用户明确要求的多轮或重新生成；多轮每轮分别保留响应，引用真实上轮 ID。

```bash
# 纯文本对话
arkcli +chat --model doubao-seed-1-6-251015 "用三句话介绍你自己"

# 流式输出（先 thinking 后 response）
arkcli +chat --model doubao-seed-1-6-251015 --stream "解释量子纠缠"

# 系统级 instructions + 采样调节
arkcli +chat --model doubao-seed-1-6-251015 \
  --instructions "你是一个只用一句话回答的助手" \
  --temperature 0.2 --top-p 0.9 --max-output-tokens 128 \
  "为什么天是蓝的？"

# 提高思考强度（仅在支持 reasoning 的模型上生效）
arkcli +chat --model doubao-seed-1-6-251015 --reasoning-effort high \
  "证明：素数有无穷多个"

# 多模态（本地图片）
arkcli +chat --model doubao-seed-1-6-251015 --input @photo.jpg "描述这张图"

# 多模态（音频）
arkcli +chat --model doubao-seed-1-6-251015 --input @clip.mp3 "这段音频在说什么？"

# 多文件（对比、组合理解）
arkcli +chat --model doubao-seed-1-6-251015 --input @img1.jpg --input @img2.jpg "对比这两张图"

# 接入点 ID（endpoint）替代 model 名
arkcli +chat --model ep-20260416234150-zsd4v "hello"

# 多轮接续 — 必须先 --store 持久化第一轮
RESP_ID=$(arkcli +chat --model doubao-seed-1-6-251015 --store --format json "第一轮问题" \
  | jq -r .id)
arkcli +chat --model doubao-seed-1-6-251015 --previous-response-id "$RESP_ID" "第二轮问题"

# 完整 JSON 输出
arkcli +chat --model doubao-seed-1-6-251015 --format json "hello"

# Endpoint + 临时 API Key；未给 Base URL 时按 Endpoint 权威 region 派生
arkcli +chat --model ep-... --api-key '<temporary-key>' --dry-run "hello"
```

## 临时执行上下文与 dry-run

用户给出 Profile / API Key / Base URL / Endpoint 的任意组合时，先读
[`../../arkcli-shared/references/execution-context.md`](../../arkcli-shared/references/execution-context.md)。

- 显式 Base URL 必须同时显式提供 API Key，禁止把 profile Key 发往覆盖后的 URL。
- Endpoint + API Key 可以不传 Base URL：CLI 读取 Endpoint 权威 region 后派生。
- API Key + Base URL + Endpoint 是 stateless 单次调用，不切换/修改 profile。
- API Key 的 `ark-*` 文本形态不能用于判断它属于哪个套餐或数据面。
- `--dry-run` 会执行本地请求校验与上下文解析，输出无 secret 的
  `preview.v1`，但不会读取在线 Endpoint/模型元数据、调用 Responses API、产生
  token 用量或存储 response。在线依赖必须列为 `unresolved`。

用 `summary.would_send` 核对本次参数：采样、token 上限、`store`、工具调用上限、
`caching`、`thinking`、`expire_at`、`reasoning` 和合法 JSON 的 `text.format`。
显式 `0` / `false` 不等于省略；未指定的可选参数不要自行补值。旧的
`prompt` / `inputs` / `text_format` / `reasoning_effort` 别名保留，因此这不是可直接
发送的 wire JSON，不要把整个 `would_send` 复制为请求体。

预览维持旧准入行为：文件是否存在、strict Schema 是否可编译、非有限采样值是否
可发送，不由 `validated: true` 证明；不可 JSON 表达的 Schema/采样字段可能不显示。
真实调用仍按原检查处理。更不能把预览通过当成 Key 有效、套餐额度充足或模型支持
该组合；这些仍要走在线 resources 准入和实际响应检查。回归场景见 [evals.md](evals.md)。

不要在日志或回复中回显 API Key。示例中的 placeholder 必须由安全变量替换。

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<prompt>` | 是 | positional | 提示词（位置参数，放在命令最后） |
| `--model` | 否：可 fallback 到 active profile 的 `Resources.Text.Default` | string | 本次上下文允许的模型 ID、合法套餐别名或 `ep-*`；规范能力查询名不替换实际调用 ID。无默认时先 `resources list --modality text`，一次对话不自动修改 default。 |
| `--input` | 否 | string（可重复） | 文件引用，如 `@photo.jpg`；按扩展名分流到 image/video/audio/file ContentItem。重复传入即多文件 |
| `--stream` | 否 | bool | 流式输出（两段：`Thinking:` + `Response:`） |
| `--instructions` | 否 | string | 系统级 instructions，注入到本次 Responses 请求 |
| `--temperature` | 否 | float | 采样温度，如 `0.7` |
| `--top-p` | 否 | float | 核采样 top_p，如 `0.9` |
| `--max-output-tokens` | 否 | int | 助手回复 token 上限 |
| `--reasoning-effort` | 否 | string | 思考强度：`minimal` / `low` / `medium` / `high`（仅在支持 reasoning 的模型上生效） |
| `--store` | 否 | bool | 持久化本次响应，让后续 `--previous-response-id` 可以接续；不加 `--store` 则只保留极短窗口 |
| `--previous-response-id` | 否 | string | 上一轮响应 id，用于多轮对话接续 |

命令级 `--dry-run` 以及 `--profile` / `--api-key` / `--base-url` 见共享
[`global-flags.md`](../../arkcli-shared/references/global-flags.md)。

## @file 机制

`--input @<path>` 的处理链路：

1. arkcli 把 `@photo.jpg` 转为**绝对路径** `file:///Users/.../photo.jpg`
2. 按扩展名推断 mime：
   - `.jpg/.jpeg/.png/.webp/.gif/.bmp` → `input_image`
   - `.mp4/.mov/.webm/.mkv/.avi` → `input_video`
   - `.mp3/.wav/.m4a/.aac/.flac/.ogg` → `input_audio`
   - 其他扩展名 → `input_file`
3. 底层 SDK 自动执行"上传 → 获取 file_id → 用 file_id 替换 URL" 的流程，然后发请求
4. 用户全程无感

支持的引用形式：
- `@<path>` — 本地文件，路径解析到绝对路径
- `<path>`（无前缀）— 等价于 `@<path>`
- `https://...`、`http://...` — 远程 URL，直接传给模型
- `file://...`、`tos://...` — 已构造好的 URL

## 返回值

> **重要：`+chat` 输出已被 arkcli 扁平化为标量字段，不是 Responses API 原生 `output[].content[].text` 嵌套结构。** 助手文本直接以 `content` 字符串呈现（service 层把所有 `output_text` 片段拼接成单段）；推理文本同理走 `reasoning_content`。`--format` 全局只接受 `json`，**不**会切回原生嵌套 shape——按 `output[].content[].text` 写 `jq` 一定取不到。

**默认输出**（扁平 schema）：

```json
{
  "id": "resp_021776932247884ea9dc72e4279a1799608cb64438f5f0f125439",
  "model": "doubao-seed-1-6-250615",
  "content": "...",
  "reasoning_content": "...",
  "usage": {
    "prompt_tokens": 88,
    "completion_tokens": 419,
    "total_tokens": 507
  }
}
```

- `id`：本次响应 id，用于下一轮 `--previous-response-id`。
- `content`：助手的正式回答文本（**已扁平化为字符串**，由 service 拼接所有 `output_text` 片段）。
- `reasoning_content`：模型的推理过程文本（仅当模型支持 reasoning 时非空）。
- `usage`：token 用量。

**流式输出** (`--stream`)：

```
Thinking:
<推理过程增量文本...>

Response:
<正式回答增量文本...>
```

stream 模式下不输出 JSON，直接打印；结束后换行。

### 用 jq 解析（常用）

以下是独立调用的语法示例。若已有成功响应文件，把管道左侧替换为该文件输入，
例如 `jq -r .content response.json`；不要再次请求只为获取另一字段。

```bash
# 取助手正文
arkcli +chat --model <id> "<prompt>" | jq -r .content

# 取推理文本（仅 reasoning 模型非空）
arkcli +chat --model <id> --reasoning-effort high "<prompt>" | jq -r .reasoning_content

# 取 token 用量
arkcli +chat --model <id> "<prompt>" | jq .usage

# 取 response id（配合 --store 用来串多轮）
arkcli +chat --model <id> --store "<prompt>" | jq -r .id
```

> 切勿写 `jq '.output[].content[].text'`——`+chat` 已扁平化，没有该路径。

### 拿原生 SDK shape（兜底）

如必须按 Responses API 原生 `output[]` 嵌套结构解析，绕开 `+chat`、走 raw API explorer：

```bash
arkcli api arkruntime.create_responses --params '{
  "model": "<id>",
  "input": [
    {"role": "user", "content": [{"type": "input_text", "text": "你好"}]}
  ]
}'
```

返回的是 arkcli-side `CreateResponsesResponse`（已注册为 raw operation），保留 `output[]`、`usage`、`reasoning` 等更接近上游 SDK 的字段。详见 [`../../arkcli-api-explorer/SKILL.md`](../../arkcli-api-explorer/SKILL.md)。

## 常见错误

| 错误 | 原因 | 处理方式 |
|------|------|---------|
| `prompt is required when no --input is provided` | 纯文本对话必须提供 prompt | 补 prompt 位置参数 |
| `model is required` | 未传 `--model` | 补模型名或 endpoint ID |
| `input file not found: <path>` | `@<path>` 的文件不存在或为目录 | 核对路径；用相对路径时注意当前工作目录 |
| `ark runtime: API Key is required` | 未配置 API Key | 运行 `arkcli auth apikey` 或设置 `ARK_API_KEY` 环境变量 |
| `Error code: 400 - ... invalid scheme` | 传的 URL 后端不认（例如 `file://` 未被 SDK 上传成功） | 检查文件大小/权限；加 `--debug` 看 SDK 上传链路 |
| `Error code: 400 - ... model not found` | `--model` 名字或 endpoint ID 错误 | 用 `arkcli models search` / `arkcli infer endpoint list` 核对 |
| `Error code: 404 - InvalidEndpointOrModel.NotFound` | 不存在、不可见或与当前数据面不兼容；不能仅凭 404 认定缺版本 | 在同一 Profile/凭证下用 resources 和 models 核对，不自动追加版本、切计费路径或重开对话 |

## `arkcli api` 直接调用（raw）

`+chat` 背后的 Operation 是 `arkruntime.create_responses` 和 `arkruntime.create_responses_stream`，可以直接调：

```bash
# 纯文本（最简 shape）
arkcli api arkruntime.create_responses --params '{
  "model":"doubao-seed-1-6-251015",
  "input":"hello"
}'

# 带 role 的消息形（string 或 list 的 content 都支持）
arkcli api arkruntime.create_responses --params '{
  "model":"doubao-seed-1-6-251015",
  "input":[{"role":"user","content":"hello"}]
}'

# 多模态（list-form content）
arkcli api arkruntime.create_responses --params '{
  "model":"doubao-seed-1-6-251015",
  "input":[{
    "role":"user",
    "content":[
      {"type":"input_image","image_url":"file:///abs/path.jpg"},
      {"type":"input_text","text":"describe this"}
    ]
  }]
}'
```

注意：raw API 调用时，**本地文件 `file://` URL 依赖 SDK 的自动上传**，原始路径不能是相对路径。

## 和其他命令的关系

- **[`arkcli +gen`](../../arkcli-gen/SKILL.md)**：用于图片/视频**生成**，不是对话。
- **[`arkcli models`](../../arkcli-models/SKILL.md)**：用来查模型名/模态，找不到合适模型时先用它。

## 安全与隐私

- `--input @<file>` 会把文件**完整上传**到 ARK file service（通过 SDK 自动处理）。
- 敏感文件不要随便传；上传后的 file 会在服务端保留一段时间，不要依赖它"传完就消失"。
- 使用 `--debug` 时网络请求内容会打到 stderr；日志外发前请 redact API Key。
