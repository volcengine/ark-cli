# +chat

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

通过数据面 Responses API（`POST /responses`）的高层封装。一条命令即返回助手回复，支持多模态、推理、流式、多轮。

## 命令

> **⚠️ `--model` 必须是 `<name>-<primary_version>` 完整形式（如 `doubao-seed-1-6-251015`、`qwen3-14b-20250429`）或 Endpoint ID（`ep-xxx`）。直接传族名（如 `doubao-seed-1-8`、`glm-5-1`）会 404 `InvalidEndpointOrModel.NotFound`。**
>
> `primary_version` 格式不固定——常见 6 位 `251015`，也可能是 8 位 `20250429`、带前缀 `preview-260215`、短数字 `2507`、甚至空串（此时 full ID 就是族名本身）。约半数模型非 6 位，详见 [`../../arkcli-models/SKILL.md`](../../arkcli-models/SKILL.md) 链路 0。**不要自行正则判断"是否像完整 ID"**。
>
> **调用前补全（必须）：** 若不确定 `--model` 是否已是完整形式，先执行：
>
> ```bash
> VER=$(arkcli models get <name> --transform 'primary_version' | tr -d '"')
> # --transform 输出带 JSON 双引号，必须 tr -d 剥掉；否则 $VER 会是 "251015" 导致拼接错误
> MODEL=${VER:+<name>-$VER}; MODEL=${MODEL:-<name>}   # VER 空串时退回 <name>
> arkcli +chat --model "$MODEL" "<prompt>"
> ```
>
> 若刚 `models search/list` 过，可直接复用返回里的 `primary_version` 字段，无需再调 `get`。

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
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<prompt>` | 是 | positional | 提示词（位置参数，放在命令最后） |
| `--model` | 否 | string | **完整版本化模型 ID**（如 `doubao-seed-1-6-251015`）或推理接入点 ID（如 `ep-xxx`）。仅传族名 `doubao-seed-1-6` 不稳定，且较新模型族会直接报 `InvalidEndpointOrModel.NotFound`。**0.1.16+ 可省略**：缺省时 fallback 到 active profile（按 `--profile > ARK_PROFILE > default` 解析）的 `Resources.Text.Default`，未设时报 hint 引导 `arkcli profile set-default --modality text <id>` |
| `--input` | 否 | string（可重复） | 文件引用，如 `@photo.jpg`；按扩展名分流到 image/video/audio/file ContentItem。重复传入即多文件 |
| `--stream` | 否 | bool | 流式输出（两段：`Thinking:` + `Response:`） |
| `--instructions` | 否 | string | 系统级 instructions，注入到本次 Responses 请求 |
| `--temperature` | 否 | float | 采样温度，如 `0.7` |
| `--top-p` | 否 | float | 核采样 top_p，如 `0.9` |
| `--max-output-tokens` | 否 | int | 助手回复 token 上限 |
| `--reasoning-effort` | 否 | string | 思考强度：`minimal` / `low` / `medium` / `high`（仅在支持 reasoning 的模型上生效） |
| `--store` | 否 | bool | 持久化本次响应，让后续 `--previous-response-id` 可以接续；不加 `--store` 则只保留极短窗口 |
| `--previous-response-id` | 否 | string | 上一轮响应 id，用于多轮对话接续 |

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
| `Error code: 404 - InvalidEndpointOrModel.NotFound` | `--model` 传的是模型族名（如 `doubao-seed-1-8`、`glm-5-1`），该族未注册族名别名 | 用 `arkcli models get <name> --transform 'primary_version'` 拿版本号，拼成 `<name>-<primary_version>` 再传 |

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
