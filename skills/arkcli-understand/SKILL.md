---
name: arkcli-understand
version: 1.0.0
description: "arkcli +understand：多模态理解工作流，通过 12 个任务型 sub-skill（4 模态：image/video/audio/file）在数据面 Responses API 引擎上做专项理解。覆盖图片描述/OCR、视觉定位（bbox grounding）、GUI 操作识别、PDF/文档字段抽取、视频总结/视频问答、音视频联合理解、语音转写（ASR）、语音翻译（AST）、SRT 字幕打轴、多说话人转写、会议纪要。每个 sub-skill 自带专家 system prompt 与默认模型。当用户要『转写这段录音 / 把图里目标框出来 / 抽取这份 PDF 的字段 / 总结这段视频 / 生成字幕 / 识别会议纪要 / 描述这张图』这类有明确产出形态的多模态理解任务时使用；开放式带图对话用 +chat，生成图/视频用 +gen。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli +understand --help"
---

# arkcli +understand

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、API Key 错误恢复、配置排查与命令选择顺序。**
**CRITICAL — 执行 `+understand` 之前，MUST 先用 Read 工具读取 [`references/arkcli-understand.md`](references/arkcli-understand.md)（命令/flag/返回值/错误）与 [`references/sub-skills.md`](references/sub-skills.md)（12 个 sub-skill 的用途与期望输出形态）。禁止凭印象拼命令。**

## 核心概念

- `+understand` **不是 12 套实现，而是 1 个引擎 + 一层语义**：底层引擎就是 `+chat` 用的数据面 Responses API；每个 sub-skill 只是一条**配方** `{模态, 默认模型, 内置 system prompt}`。sub-skill 之间的唯一差异，是注入到 Responses `instructions` 字段的那段专家 system prompt。
- 命令形态：`arkcli +understand <sub-skill> --input @file [prompt]`。
  - `args[0]` **命中注册表**（12 个之一）→ 当作显式 sub-skill，其余位置参数当 prompt 叠加在内置 prompt 之上。
  - `args[0]` **不命中** → 整段位置参数都当 prompt，服务端按**首个 `--input` 的文件模态自动路由**到该模态的默认 sub-skill。
- **必须有 `--input`**：sub-skill 是「对某个文件做理解」，没有 `--input` 会直接 `missing_input` 报错。纯 prompt 无法推导配方。
- 返回值与 `+chat` 完全一致：arkcli **扁平** schema `{id, model, content, reasoning_content, usage}`，**不是** Responses 原生 `output[].content[].text` 嵌套。详见 [`references/arkcli-understand.md`](references/arkcli-understand.md) 的「返回值」段。
- 多模态上传：image/video/doc 由 SDK `file://` preprocessor 自动走 Files API；**audio 是特例**——内联为 base64 data URL，**上限 25MB**。详见 reference。

## 快速决策（understand vs chat vs gen）

| 用户意图 | 走哪个 |
|---|---|
| 有**明确产出形态**的多模态理解任务（转写 / 翻译 / 字幕 / 框选定位 / GUI 操作 / 字段抽取 / 分章节总结 / 多说话人 / 会议纪要） | **`+understand`**（命中某个 sub-skill） |
| 开放式带图/视频对话、追问、推理、需要多轮接续（`--store`/`--previous-response-id`）、需要 tools（web_search/function）或 `--text-format json_schema` 严格 JSON | [`../arkcli-chat/SKILL.md`](../arkcli-chat/SKILL.md) |
| **生成**图片 / 视频（不是理解已有素材） | [`../arkcli-gen/SKILL.md`](../arkcli-gen/SKILL.md) |

一句话判据：**任务能映射到下面某个 sub-skill 名 → 用 `+understand`；否则开放对话用 `+chat`，生成用 `+gen`。**

## Agent 快速执行顺序

1. 判断用户的理解任务命中哪个 sub-skill（见下表）。命中就显式带上 sub-skill 名，不要让它走自动路由去猜。
2. 过认证闸门：不确定登录态先 `arkcli auth status`；数据面调用靠 ARK API Key（`Authorization: Bearer`），鉴权失败按 [`../arkcli-auth/references/auth-modes.md`](../arkcli-auth/references/auth-modes.md) 的「API Key 模式的错误恢复」走 `arkcli auth apikey`，**不要原地重试**。
3. 备好输入：`--input @<file>`（可多次）。本地文件用 `@` 前缀；远程用 `https://` / `tos://` URL。
4. **默认不传 `--model`**：每个 sub-skill 自带已验证可用的默认模型（见下表）。只有在用户明确要换模型时才传 `--model`，且必须是完整版本化 ID（`<name>-<primary_version>`）或 `ep-xxx`——拿不准先转 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md) 补版本号。
5. 需要逐段输出加 `--stream`；想微调任务指令用 `--system-prompt-append "..."`（追加），或 `--system-prompt-override "..."`（整体替换内置 prompt）。
6. 跑完回到用户原始目标，不要停在中间产物上。

## sub-skill 速查表（12 个 / 4 模态）

> 默认模型均为已验证可用的已知好值（真机或同模态机制等价覆盖），**默认不要覆盖**；确需换模型见上文「Agent 快速执行顺序」第 4 条。

| sub-skill | 模态 | 默认模型 | 一句话用途 |
|---|---|---|---|
| `image-caption` | image | `doubao-seed-1-6` | 单图/多图描述、OCR（保留原文，不意译） |
| `image-grounding` | image | `doubao-seed-1-6` | 视觉定位：输出目标 bbox `(x1,y1,x2,y2)` + confidence |
| `image-gui` | image | `doubao-seed-1-6` | GUI 截图 → 可执行操作序列（12 类操作 JSON 数组） |
| `doc-extract` | file | `doubao-seed-1-6` | PDF/文档按 schema 抽取结构化 JSON（缺失填 null，保留 page） |
| `video-summary` | video | `doubao-seed-1-6` | 视频总结：overall / segment / chapter + 关键时间点 |
| `video-qa` | video | `doubao-seed-1-6` | 视频问答：结合画面+音频+时间轴精准回答 |
| `vau` | video | `doubao-seed-2-0-lite` | 音视频联合理解：分析视频内声音元素，出音频分析报告 |
| `asr` | audio | `doubao-seed-2-0-lite` | 语音转写：仅输出纯文本，无任何前后缀/格式 |
| `asr-align` | audio | `doubao-seed-2-0-lite` | 字幕打轴：默认 SRT 格式（序号+起止时间+文本） |
| `asr-speakers` | audio | `doubao-seed-2-0-lite` | 多说话人转写：标 `[spk0]` / `[spk1]` ... |
| `ast` | audio | `doubao-seed-2-0-lite` | 语音翻译：把音频里的话译成文本 |
| `meeting-minutes` | audio | `doubao-seed-2-0-lite` | 会议纪要：结构化 markdown（主题/参会人/要点/决议/待跟进） |

逐条的期望输出形态与示例命令见 [`references/sub-skills.md`](references/sub-skills.md)。

## 省略 sub-skill 时的自动路由

`args[0]` 不命中注册表时，按**首个 `--input` 的文件扩展名**推断模态并兜底到该模态的默认 sub-skill：

| 推断模态 | 兜底 sub-skill |
|---|---|
| image（`.jpg/.png/.webp/...`） | `image-caption` |
| video（`.mp4/.mov/...`） | `video-qa` |
| audio（`.mp3/.wav/.m4a/...`） | `asr` |
| file（其他扩展名） | `doc-extract` |

- **没有 `--input` 时无法自动路由**（纯 prompt 推不出配方）→ 报 `missing_sub_skill`。
- 目前**只按文件模态兜底，不做 prompt 关键词级路由**（例如说「框出」不会自动选 `image-grounding`）。Agent 想要非默认 sub-skill（如 grounding / gui / summary / 字幕 / 翻译）就**显式写出 sub-skill 名**，别依赖自动路由。

## 命令一览

| 命令 | 说明 |
|------|------|
| `arkcli +understand image-caption --input @photo.jpg "描述这张图"` | 图片描述 |
| `arkcli +understand image-caption --input @doc.png "OCR，原样保留文字"` | OCR |
| `arkcli +understand image-grounding --input @scene.jpg "框出所有红色car"` | 视觉定位 bbox |
| `arkcli +understand doc-extract --input @invoice.pdf "抽取 发票号/金额/日期"` | 文档字段抽取 |
| `arkcli +understand video-summary --input @clip.mp4 "按 chapter 总结"` | 视频分章节总结 |
| `arkcli +understand video-qa --input @clip.mp4 "视频里出现了几辆车？"` | 视频问答 |
| `arkcli +understand asr --input @speech.mp3` | 语音转写（无需 prompt） |
| `arkcli +understand asr-align --input @speech.mp3` | 生成 SRT 字幕 |
| `arkcli +understand asr-speakers --input @meeting.wav` | 多说话人转写 |
| `arkcli +understand ast --input @speech.m4a "译成英文"` | 语音翻译 |
| `arkcli +understand meeting-minutes --input @meeting.m4a` | 会议纪要 |
| `arkcli +understand "<prompt>" --input @photo.jpg` | 省略 sub-skill → 按模态自动路由（此处 → image-caption） |
| `arkcli +understand asr --input @speech.mp3 --stream` | 流式逐段输出 |
| `arkcli +understand image-caption --input @x.jpg --system-prompt-append "用英文回答"` | 在内置 prompt 后追加指令 |

## 详细文档

- 全部 flag、`@file`/audio 上传机制、返回值扁平 schema、错误码、raw API 兜底、与 `+chat` 的关系 → [`references/arkcli-understand.md`](references/arkcli-understand.md)
- 12 个 sub-skill 的逐条用途、**期望输出形态**、示例命令 → [`references/sub-skills.md`](references/sub-skills.md)
- 最小评估/回归用例 → [`references/evals.md`](references/evals.md)
