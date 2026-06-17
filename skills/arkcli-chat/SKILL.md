---
name: arkcli-chat
version: 1.1.0
description: "arkcli +chat：通过数据面 Responses API 快速对话/推理，支持多模态（@file 本地图片、视频、音频、通用文件）、流式输出、system instructions、采样调节（temperature/top-p/max-output-tokens）、reasoning effort 调节、--store + previous-response-id 多轮接续。当用户需要与模型即时对话、问答、推理、做带图/视频/音频的开放式多模态对话与追问、调温度/采样、控制思考强度、或多轮接续对话时使用。注意：有明确产出形态的多模态理解（语音转写、文档字段抽取、字幕打轴、bbox 框选定位、视频分章节总结等）走 arkcli-understand；本 skill 只做开放式对话/推理。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli +chat --help"
---

# arkcli +chat

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、配置排查与命令选择顺序**
**CRITICAL — `+chat` 在执行之前，务必先用 Read 工具读取 [`references/arkcli-chat.md`](references/arkcli-chat.md)，禁止直接盲目调用命令。**

## 核心概念

- `--model` 缺省时 CLI 自动 fallback 到 active profile 的 `resources.text.default`；如果用户显式传 `--model X` 且 `X` ≠ 该默认，按 [`../arkcli-shared/references/profile-defaults.md`](../arkcli-shared/references/profile-defaults.md) "Default 漂移检测与 promote nudge" 询问用户是否 promote
- `+chat` 是数据面 Responses API（`POST /responses`）的高层封装：一次请求即返回助手文本。
- 支持**多模态**：用 `--input @photo.jpg` 把本地文件随请求上传，模型可看图/看视频/听音频后回答。图片(`.jpg/.png/.webp/...`)、视频(`.mp4/.mov/...`)、音频(`.mp3/.wav/.m4a/...`)、通用文件按扩展名自动分流。
- 支持**流式**：`--stream` 模式逐段输出，先输出推理（thinking），再输出正式回答（response）。
- 支持**进度提示**：非流式调用在 5s 后开始向 stderr 输出 `arkcli +chat: still running… elapsed Xs` 心跳行（每 10s 一次），避免长调用看不到任何输出；脚本场景可用 `--no-progress` 关闭。stdout 不受影响。
- 支持**system instructions**：`--instructions "..."` 注入系统级指令。
- 支持**采样调节**：`--temperature` / `--top-p` / `--max-output-tokens`。
- 支持**思考强度**：`--reasoning-effort minimal|low|medium|high`（仅在支持 reasoning 的模型上生效）。
- 支持**多轮接续**：`--store` 持久化本次响应，下一次用 `--previous-response-id <id>` 接续。
- 支持**Tools**：`--tools web_search`（简单语法糖）或 `--tools-file tools.json`（function 等完整形态），配合 `--tool-choice auto|required|none` 与 `--max-tool-calls`。详见 [`references/tools.md`](references/tools.md)。
- 支持**对账面命令**：`arkcli chat get/delete/list-input-items <response-id>`，对 `--store` 过的 response 做 CRUD，详见 [`references/chat-meta.md`](references/chat-meta.md)。
- 支持**缓存与思考**：`--caching enabled|disabled`（配 `--cache-prefix`）控制服务端 prompt cache；`--thinking auto|enabled|disabled` 控制思考阶段；`--expire-at <epoch_sec>` 给 stored response 加过期。详见 [`references/caching-thinking.md`](references/caching-thinking.md)。
- 支持**流式事件**：`--stream --include-events` 输出原始 SDK 事件 NDJSON（每行一个 JSON），供 autotest / agent 程序化消费。详见 [`references/stream-events.md`](references/stream-events.md)。
- **输出 schema（务必先看）**：`+chat` 返回是 arkcli **扁平** schema（`{id, model, content, reasoning_content, usage, ...}`），**不是** Responses API 原生 `output[].content[].text` 嵌套；助手文本直接用 `.content` 取。`--format` 不会切换 shape。详见 [`references/arkcli-chat.md`](references/arkcli-chat.md) 的「返回值」段。

## 快速决策

**走 `+chat` 的判据（三条全满足）：**
1. 用户带图/视频/音频但意图是**开放对话/问答/推理/感想/评论**（非固定产出形态）
2. 不能映射到 `arkcli-understand` 12 个子技能之一
3. 或用户需要 `--store` / `--previous-response-id` 多轮接续

**转 `arkcli-understand` 的判据（任一满足即转）：**
- 用户要「转写/语音转文字/语音识别/ASR」→ understand
- 用户要「字幕/打轴/SRT」→ understand
- 用户要「多人对话/标说话人/会议转写」→ understand
- 用户要「OCR/识别文字/识别图片文字/发票金额」→ understand
- 用户要「框出来/标框/bbox/视觉定位」→ understand
- 用户要「PDF 字段提取/文档提取/抽字段/合同提取」→ understand
- 用户要「视频总结/分段/章节总结」→ understand
- 用户要「视频问答/视频内容提问」→ understand

一句话：**有 @file 且有明确产出形态 → understand；带图聊天/追问/开放感想 → chat**。

- 用户要生成图片/视频（非对话）：转 [`../arkcli-gen/SKILL.md`](../arkcli-gen/SKILL.md)。

## Agent 快速执行顺序

1. 用户目标是对话/推理/多模态理解时，优先用 `arkcli +chat`。
2. 不确定认证状态时，先看 `arkcli auth status`；未登录/无 API Key 转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)。
3. 不确定模型名时，先转 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md)。
4. **`--model` 必须是 `<name>-<primary_version>` 完整形式**（或 Endpoint ID `ep-xxx`）。`primary_version` 格式不固定：6 位日期、8 位日期、带限定前缀、短数字、甚至空串都有（详见 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md) 链路 0 的完整表格）——**不要用正则自行猜测"看起来是否完整"**。若用户只给了族名或不确定是否完整，先查 `primary_version` 再拼：刚 `models search/list` 过就直接复用返回里的字段，否则 `arkcli models get <name> --transform 'primary_version' | tr -d '"'`（`--transform` 输出带引号，必须剥掉）。跳过会直接 404 `InvalidEndpointOrModel.NotFound`。
5. 需要流式输出时加 `--stream`；需要多模态时加 `--input @<file>`（可多次）。

## 常见降级

- 模型名不确定：先 `models search`。
- endpoint ID 要用 `+chat`：直接传 `--model ep-xxx`（endpoint 本身已决定模态，无需额外 flag）。
- 鉴权失败：转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)。

## 命令一览

| 命令 | 说明 |
|------|------|
| `arkcli +chat --model <id> "<prompt>"` | 最简用法：纯文本对话 |
| `arkcli +chat --model <id> --stream "<prompt>"` | 流式输出（thinking + response 两段） |
| `arkcli +chat --model <id> --instructions "你是简洁助手" "<prompt>"` | 系统级指令 |
| `arkcli +chat --model <id> --temperature 0.2 --max-output-tokens 256 "<prompt>"` | 采样调节 |
| `arkcli +chat --model <id> --reasoning-effort high "<prompt>"` | 提高思考强度 |
| `arkcli +chat --model <id> --input @file.jpg "<prompt>"` | 多模态（本地文件，支持图/视/音） |
| `arkcli +chat --model <id> --input @a.jpg --input @b.jpg "<prompt>"` | 多文件 |
| `arkcli +chat --model <id> --store "<prompt>"` 拿到 id 后再 `--previous-response-id <id> "<下一句>"` | 持久化 + 多轮接续 |
| `arkcli +chat --model <id> --tools web_search --tool-choice auto "<prompt>"` | Tools: 联网检索 |
| `arkcli +chat --model <id> --tools-file tools.json --tool-choice required "<prompt>"` | Tools: 自定义 function |
| `arkcli chat get <response-id>` | 拿回 store 过的 response（含 function_calls） |
| `arkcli chat list-input-items <response-id> --order desc --limit 5` | 列出输入项（多轮历史） |
| `arkcli chat delete <response-id>` | 删除 store 过的 response |
| `arkcli +chat --model <id> --caching enabled --store "<prompt>"` | 启用 prompt cache + 持久化 |
| `arkcli +chat --model <id> --thinking disabled --max-output-tokens 100 "<prompt>"` | 关思考压短输出 |
| `arkcli +chat --model <id> --text-format json_object "<prompt>"` | 强制模型出合法 JSON |
| `arkcli +chat --model <id> --text-format json_schema --text-schema schema.json --text-strict "<prompt>"` | 用 JSON Schema 强约束 shape |
| `arkcli +chat --model <id> --stream --include-events "<prompt>"` | 流式 NDJSON（每行一个 SDK 事件 JSON） |

## 详细文档

- `+chat` 的所有参数、返回值、错误码、多模态文件自动上传机制等见 [`references/arkcli-chat.md`](references/arkcli-chat.md)。
- `+chat` 的 Tools 能力（`--tools / --tools-file / --tool-choice / --max-tool-calls`，含 function 与 web_search）见 [`references/tools.md`](references/tools.md)。
- `chat get / chat delete / chat list-input-items` 三个对账面命令见 [`references/chat-meta.md`](references/chat-meta.md)。这些命令操作的 response 必须是 `+chat --store` 过的。
- `+chat` 的 `--caching / --cache-prefix / --thinking / --expire-at` 用法、回显字段与 autotest 对应见 [`references/caching-thinking.md`](references/caching-thinking.md)。
- `+chat` 的 `--text-format / --text-schema / --text-schema-name / --text-strict` 用法见 [`references/text-format.md`](references/text-format.md)。
- `+chat` 的 `--stream --include-events` NDJSON 流式事件输出见 [`references/stream-events.md`](references/stream-events.md)。
