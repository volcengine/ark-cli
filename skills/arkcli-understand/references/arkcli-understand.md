# +understand 命令参考

> **前置条件：** 先阅读 [`../../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md)（认证、API Key 错误恢复、全局参数、安全规则）与 [`sub-skills.md`](sub-skills.md)（12 个 sub-skill 的用途与期望输出形态）。

`+understand` 是数据面 Responses API（`POST /responses`）之上的「1 引擎 + 语义层」多模态理解工作流。它和 `+chat` 共用同一个 `CreateResponses` 引擎，差别只在：每个 sub-skill 预置了一段专家 system prompt 和一个默认模型，注入到本次请求的 `instructions` 字段。

## 命令形态

```bash
arkcli +understand <sub-skill> --input @file [prompt]   # 显式指定 sub-skill
arkcli +understand "<prompt>" --input @file             # 省略 sub-skill → 按 --input 模态自动路由
```

```bash
# 图片描述（最简）
arkcli +understand image-caption --input @photo.jpg "用中文描述画面主体"

# 视觉定位（输出 bbox）
arkcli +understand image-grounding --input @scene.jpg "框出图中所有行人"

# 文档抽取（按字段）
arkcli +understand doc-extract --input @invoice.pdf "抽取：发票号 / 金额 / 开票日期"

# 语音转写（无需 prompt，sub-skill 内置任务指令）
arkcli +understand asr --input @speech.mp3

# 生成 SRT 字幕
arkcli +understand asr-align --input @speech.mp3

# 视频分章节总结
arkcli +understand video-summary --input @clip.mp4 "按 chapter 输出，每章给时间点"

# 省略 sub-skill：按首个 --input 的模态自动路由（此处 .mp3 → asr）
arkcli +understand "把这段录音转成文字" --input @speech.mp3

# 流式逐段输出
arkcli +understand asr --input @speech.mp3 --stream

# 在内置 prompt 之后追加指令（不替换内置专家 prompt）
arkcli +understand image-caption --input @x.jpg --system-prompt-append "只用一句话，英文"

# 完整 JSON 输出
arkcli +understand asr --input @speech.mp3 --format json
```

## 位置参数解析（sub-skill vs prompt）

`+understand` 至少要 1 个位置参数（`MinArgs=1`），解析规则：

- `args[0]` 命中注册表（12 个 sub-skill 之一）→ `args[0]` 作 sub-skill，`args[1:]` 拼成 prompt（叠加在内置 prompt 之上，作为用户附加要求）。
- `args[0]` 不命中 → 整段 `args` 都当 prompt，服务端按首个 `--input` 的模态**自动路由**到默认 sub-skill（image→image-caption / video→video-qa / audio→asr / file→doc-extract）。
- 很多 sub-skill（如 `asr` / `asr-align`）的任务已由内置 prompt 完全定义，prompt 可省略——但**位置参数总数仍需 ≥1**，所以至少要给 sub-skill 名。

## 参数

> 严格对应 `arkcli +understand --help`。注意 `+understand` 的 flag 集合**比 `+chat` 小**：用 `--system-prompt-override/append` 取代 `+chat` 的 `--instructions`；**没有** `--tools` / `--tools-file` / `--tool-choice` / `--caching` / `--cache-prefix` / `--text-format` / `--text-schema` / `--previous-response-id` / `--include-events` / `--expire-at`。不要照搬 `+chat` 的这些 flag。

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<sub-skill>` | 是* | positional | 12 个之一（见 [`sub-skills.md`](sub-skills.md)）。省略时由首个 `--input` 模态自动路由，但位置参数总数仍需 ≥1 |
| `[prompt]` | 否 | positional | 附加在内置 system prompt 之上的用户要求（如抽取字段、回答的具体问题、输出语言） |
| `--input` | 是 | string（可重复） | 输入素材，如 `@photo.jpg`；按扩展名分流到 image/video/audio/file。**没有 `--input` 会 `missing_input` 报错** |
| `--model` | 否 | string | 覆盖 sub-skill 默认模型。**默认别传**——默认值已验证可用。要传须是完整版本化 ID（`<name>-<primary_version>`）或 `ep-xxx`，拿不准先查 [`../../arkcli-models/SKILL.md`](../../arkcli-models/SKILL.md) |
| `--stream` | 否 | bool | 流式输出（两段：`Thinking:` + `Response:`） |
| `--system-prompt-override` | 否 | string | **整体替换** sub-skill 内置 system prompt |
| `--system-prompt-append` | 否 | string | 在内置 system prompt **末尾追加**额外指令 |
| `--temperature` | 否 | float | 采样温度，如 `0.7` |
| `--top-p` | 否 | float | 核采样 top_p，如 `0.9` |
| `--max-output-tokens` | 否 | int | 回复 token 上限 |
| `--reasoning-effort` | 否 | string | 思考强度：`minimal` / `low` / `medium` / `high`（仅在支持 reasoning 的模型上生效） |
| `--thinking` | 否 | string | 内部思考：`auto` / `enabled` / `disabled` |
| `--store` | 否 | bool | 持久化本次响应（可后续用 `arkcli chat get <id>` 取回） |
| `--no-progress` | 否 | bool | 关闭非流式调用时 stderr 的心跳提示（脚本场景用） |

\* 至少要有 sub-skill 名或可路由的 `--input`，否则 `missing_sub_skill`。

全局 flag（`--profile` / `--api-key` / `--project-name` / `--region` / `--format json` / `--transform` / `--debug` 等）见 [`../../arkcli-shared/references/global-flags.md`](../../arkcli-shared/references/global-flags.md)。

## System prompt 解析链

每个 sub-skill 自带一段内置 system prompt（注入到 Responses `instructions` 字段）。解析顺序：

```
内置 prompt（按 sub-skill）
  → --system-prompt-override 非空时：整体替换为 override 内容
  → --system-prompt-append 非空时：在上一步结果末尾追加（中间空一行）
```

- 想**微调**任务（加语言要求、加输出约束）：用 `--system-prompt-append`，保留专家 prompt。
- 想**完全自定义**：用 `--system-prompt-override`，等于把 sub-skill 退化成「指定模型 + 自定义 prompt」的 `+chat`。
- `[prompt]` 位置参数是 **user message**，与 system prompt 是两层：sub-skill 定义「你是谁、怎么做」，prompt 给「这次具体要什么」。

## `@file` 上传机制（audio 是特例）

`--input @<path>` 按扩展名推断模态：

- `.jpg/.jpeg/.png/.webp/.gif/.bmp` → image
- `.mp4/.mov/.webm/.mkv/.avi` → video
- `.mp3/.wav/.m4a/.aac/.flac/.ogg` → audio
- 其他 → file（如 `.pdf`）

支持的引用形式：`@<path>`（本地，解析为绝对 `file://`）、`<path>`（无前缀，等价 `@<path>`）、`https://` / `http://`（远程 URL）、`file://` / `tos://`（已构造好的 URL）。

上传路径分两类：

- **image / video / doc**：SDK 的 `file://` preprocessor 在 `CreateResponses` 内部**自动**走 Files API（上传拿 `file_id`），arkcli 无需介入。
- **audio**：SDK 的 preprocessor **不处理 audio**，arkcli 在客户端层把本地音频**内联为 base64 data URL** 后再发请求。

### audio 大小约束

| 形式 | 上限 | 现状 |
|---|---|---|
| base64（`@本地音频`，内联 data URL） | **25MB** | 当前实现；**超出不拦截、不提醒**（产品决策）——大文件可能直接报错或被后端拒 |
| 公网 URL（`https://...`） | 25MB | 后端会下载（需国内可达） |
| Files API（file_id，512MB） | 512MB | ⏳ 待 SDK 升级后切换，当前不可用 |

> 给用户处理 **>25MB 音频**时：先提醒会超限，建议切片或降码率后再转；不要假装大文件一定能跑通。

## 返回值

> **与 `+chat` 一致：输出是 arkcli 扁平 schema，不是 Responses API 原生 `output[].content[].text` 嵌套。** 助手文本直接取 `.content`；`--format` 全局只接受 `json`，不会切回原生嵌套 shape。

**默认/`--format json` 输出**：

```json
{
  "id": "resp_021776932247884ea9dc72e4279a1799608cb64438f5f0f125439",
  "model": "doubao-seed-2-0-lite",
  "content": "...",
  "reasoning_content": "...",
  "usage": { "prompt_tokens": 88, "completion_tokens": 419, "total_tokens": 507 }
}
```

- `content`：sub-skill 的产出文本（转写结果 / bbox JSON / SRT / 抽取 JSON / 纪要 markdown 等，**已扁平为字符串**）。各 sub-skill 的具体形态见 [`sub-skills.md`](sub-skills.md)。
- `reasoning_content`：推理过程（仅 reasoning 模型非空）。
- `id`：本次响应 id；配 `--store` 后可 `arkcli chat get <id>` 取回。

**流式输出**（`--stream`）：

```
Thinking:
<推理增量...>

Response:
<正式产出增量...>
```

stream 模式直接打印、不输出 JSON，结束后换行。

### 用 jq 解析

```bash
# 取产出正文（转写文本 / JSON / SRT ...）
arkcli +understand asr --input @speech.mp3 | jq -r .content

# bbox / 抽取这类 JSON 产出：content 是「字符串形式的 JSON」，需二次 parse
arkcli +understand image-grounding --input @x.jpg "框出所有人" | jq -r .content | jq .

# 取 token 用量
arkcli +understand asr --input @speech.mp3 | jq .usage
```

> 切勿写 `jq '.output[].content[].text'`——已扁平化，没有该路径。模型按 prompt 产出的 JSON（bbox / 抽取结果）是 `.content` 里的**字符串**，要先 `jq -r .content` 取出再 `jq .` 二次解析。

## 常见错误

| 错误码 / 文案 | 原因 | 处理 |
|------|------|------|
| `missing_sub_skill`：`understand needs a sub-skill or an input file to auto-route` | 既没给 sub-skill，又没有可路由的 `--input` | 补 sub-skill 名，或补 `--input @file` 让它自动路由 |
| `unknown_sub_skill`：`unknown understand sub-skill "X"` | sub-skill 名拼错/不存在 | 用 12 个合法名之一（见 [`sub-skills.md`](sub-skills.md)） |
| `missing_input`：`<sub> requires an input file via --input` | 给了 sub-skill 但没 `--input` | 补 `--input @your-file` |
| `unroutable_input`：`cannot auto-route a "X" input` | 自动路由时模态无默认映射 | 显式写出 sub-skill 名 |
| `input file not found: <path>` | `@<path>` 不存在或是目录 | 核对路径；注意当前工作目录 |
| `ark runtime: API Key is required` | 未配置 API Key | `arkcli auth apikey` 或设 `ARK_API_KEY` |
| `Error code: 403 - {"code":"AccessDenied",...}` | 数据面鉴权/权限 | **不要重试**，按 [`../../arkcli-auth/references/auth-modes.md`](../../arkcli-auth/references/auth-modes.md)「API Key 模式的错误恢复」走 `arkcli auth apikey` |
| `Error code: 404 - InvalidEndpointOrModel.NotFound` | `--model` 传了不可解析的族名 | 改用完整 `<name>-<primary_version>` 或 `ep-xxx`，见 [`../../arkcli-models/SKILL.md`](../../arkcli-models/SKILL.md) |

## `arkcli api` 直接调用（raw 兜底）

`+understand` 背后的 Operation 就是 `+chat` 的 `arkruntime.create_responses` / `arkruntime.create_responses_stream`。sub-skill 的本质是「特定 `model` + 特定 `instructions`」，可手动复刻：

```bash
arkcli api arkruntime.create_responses --params '{
  "model": "doubao-seed-2-0-lite",
  "instructions": "<把对应 sub-skill 的内置 system prompt 贴进来>",
  "input": [{
    "role": "user",
    "content": [
      {"type": "input_audio", "audio_url": "file:///abs/path/speech.mp3"},
      {"type": "input_text", "text": "<可选的附加要求>"}
    ]
  }]
}'
```

> raw 路径下，audio 不再有 arkcli 的 base64 自动内联——需自己构造可被后端拉取的 URL 或 data URL。能用 `+understand` 就别走 raw。详见 [`../../arkcli-api-explorer/SKILL.md`](../../arkcli-api-explorer/SKILL.md)。

## 和其他命令的关系

- **[`arkcli +chat`](../../arkcli-chat/SKILL.md)**：开放式对话/推理、多轮接续（`--store` + `--previous-response-id`）、tools、`--text-format json_schema`。`+understand` 的产出若要继续追问，可 `--store` 后用 `+chat --previous-response-id <id>` 接续（同一引擎）。
- **[`arkcli +gen`](../../arkcli-gen/SKILL.md)**：图片/视频**生成**，不是理解已有素材。
- **[`arkcli models`](../../arkcli-models/SKILL.md)**：需要覆盖 `--model` 又拿不准完整版本号时先查。

## 安全与隐私

- `--input @<file>` 会把文件**完整上传** / 内联到请求中（image/video/doc 经 Files API，audio 经 base64）。敏感素材不要随便传；上传后服务端会保留一段时间。
- `--debug` 会把请求/响应（可能含 base64 音频、file URL）打到 stderr；外发日志前先 redact。
- `--store` 持久化的响应会留在服务端，按需用 `arkcli chat delete <id>` 清理。
