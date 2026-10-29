---
name: arkcli-understand
version: 1.1.0
description: "arkcli +understand：对用户提供的图片、音频、视频或文档完成转写、抽取、字幕、定位等 12 类专项产出，包含 Agent Plan 个人版/企业版原生 WebSocket ASR 和字幕打轴。仅解释参数或排查已有多模态请求报错时转 arkcli-chat；实际理解任务支持临时凭证与本地 dry-run，开放对话走 +chat，生成走 +gen。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli +understand --help"
---

# arkcli +understand

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、API Key 错误恢复、配置排查与命令选择顺序。**
**CRITICAL — 执行 `+understand` 之前，MUST 先用 Read 工具读取 [`references/arkcli-understand.md`](references/arkcli-understand.md)（命令/flag/返回值/错误）与 [`references/sub-skills.md`](references/sub-skills.md)（12 个 sub-skill 的用途与期望输出形态）。禁止凭印象拼命令。**

## 原生语音识别分支

Agent Plan 个人版/企业版的 Seed ASR 2.0 使用 `+understand asr` 或 `asr-align`，
显式指定 `--audio-protocol ws-bidirectional`（流式结果）或 `ws-unidirectional`
（完整音频发送后返回结果）。也可通过 `--model doubao-seed-asr-2.0` 选择原生
双向 WS。执行前读取 [`references/arkcli-understand-asr.md`](references/arkcli-understand-asr.md)。
该路径使用专属/席位 Key、原生 ASR 参数和真实时间戳，不使用 Responses 配方、
Endpoint 或临时 Key/Base URL。仅本地 preview 时直接 `--dry-run`，不先查 auth。
下文的 Responses 引擎、返回值、25MB 内联限制和 recipe fallback 模型描述适用于
`--audio-protocol responses`（默认），不会覆盖显式原生 ASR 选择。

## 先区分理解任务与调用问题

| 本轮要交付什么 | 执行边界 |
| --- | --- |
| 对具体素材转写、抽取、翻译或定位 | 使用本 Skill 的配方；缺少素材时补齐输入，不能凭空生成素材或理解结果 |
| 解释音频/多模态参数、比较接口，或排查已有请求的字段报错 | 读取 [`../arkcli-chat/SKILL.md`](../arkcli-chat/SKILL.md) 的请求证据与条件修复；本轮交付诊断说明，到此结束，不进入后文业务准备 |
| 同时要求解释错误并实际处理素材 | 分开记录两项交付；只在已提供素材、调用目标和实际执行授权的范围内处理，不用新请求的结果替代原错误证据 |

`+understand` 的默认配方实现走 Responses API，只能说明这个 CLI 命令的行为，不能说明用户原来的音频 SDK、Chat 请求或网关也走 Responses。单凭 `unknown field` 不能判定实际接口；原 URL、模型版本或请求体缺失时保留未知，按 Chat Skill 给出条件分支。仅有“音频理解”“转写参数”等字样不会启动配方执行。

参数诊断无需先准备认证、解析默认资源或选择下表的 fallback 模型。用户未提供原模型时，不自行选另一个模型查询上限并据此解释原请求；已知模型的只读能力查询也须在本次允许的工具范围内。只有本轮确实要求实际处理素材或授权复现，才进入后文执行顺序。

## 核心概念

- 默认理解配方复用一个引擎：底层引擎就是 `+chat` 用的数据面 Responses API；每个 sub-skill 只是一条**配方** `{模态, fallback 模型, 内置 system prompt}`。Platform Profile 省略 `--model` 时必须使用当前 Profile 的 `Resources.Text.Default` Endpoint；只有 Plan 类 Profile 才使用 recipe fallback 模型。
- 命令形态：`arkcli +understand <sub-skill> --input @file [prompt]`。
  - `args[0]` **命中注册表**（12 个之一）→ 当作显式 sub-skill，其余位置参数当 prompt 叠加在内置 prompt 之上。
  - `args[0]` **不命中** → 整段位置参数都当 prompt，服务端按**首个 `--input` 的文件模态自动路由**到该模态的默认 sub-skill。
- **必须有 `--input`**：sub-skill 是「对某个文件做理解」，没有 `--input` 会直接 `missing_input` 报错。纯 prompt 无法推导配方。
- 返回值与 `+chat` 完全一致：arkcli **扁平** schema `{id, model, content, reasoning_content, usage}`，**不是** Responses 原生 `output[].content[].text` 嵌套。详见 [`references/arkcli-understand.md`](references/arkcli-understand.md) 的「返回值」段。
- 多模态上传：image/video/doc 由 SDK `file://` preprocessor 自动走 Files API；**audio 是特例**——内联为 base64 data URL，**上限 25MB**。详见 reference。
- 用户临时提供 `--api-key` / `--base-url` / Endpoint 时，MUST 读取 [`../arkcli-shared/references/execution-context.md`](../arkcli-shared/references/execution-context.md)。不要根据 Key 文本或 Endpoint 名称猜数据面/工作流。
- `--dry-run` 只在本地解析配方、当前 Profile 已持久化的默认 Endpoint、显式模型和输入引用，输出统一 `preview.v1`；不读取在线 Endpoint/模型元数据、不上传文件、不调用 Responses API、不产生模型用量、不创建 response id 或持久化响应。在线才能补齐的值必须标为 `unresolved`，不能把预览当服务端 validation。

## 快速决策（understand vs chat vs gen）

| 用户意图 | 走哪个 |
|---|---|
| 有**明确产出形态**的多模态理解任务（转写 / 翻译 / 字幕 / 框选定位 / GUI 操作 / 字段抽取 / 分章节总结 / 多说话人 / 会议纪要） | **`+understand`**（命中某个 sub-skill） |
| 开放式带图/视频对话、追问、推理、需要多轮接续（`--store`/`--previous-response-id`）、需要 tools（web_search/function）或 `--text-format json_schema` 严格 JSON | [`../arkcli-chat/SKILL.md`](../arkcli-chat/SKILL.md) |
| **生成**图片 / 视频（不是理解已有素材） | [`../arkcli-gen/SKILL.md`](../arkcli-gen/SKILL.md) |

一句话判据：**要求实际处理素材且产出对应某个 sub-skill → 用 `+understand`；仅询问参数或报错 → Chat 诊断；开放对话用 `+chat`，生成用 `+gen`。**

## Agent 快速执行顺序

1. 判断用户的理解任务命中哪个 sub-skill（见下表）。命中就显式带上 sub-skill 名，不要让它走自动路由去猜。
2. 用户只给 `ep-...` → `arkcli resources resolve <ep-id> --format json`；只有用户任务属于明确理解产出，且 `supported_workflows` 包含 `understand`，才走本 skill。
3. 用户给了临时 Key/Base URL/Endpoint → 按共享 execution-context 组合规则决定参数，不先切 profile。
4. 没有完整 stateless 上下文时过认证闸门：不确定登录态先 `arkcli auth status`；鉴权失败按 [`../arkcli-auth/references/auth-modes.md`](../arkcli-auth/references/auth-modes.md) 的「API Key 模式的错误恢复」处理，**不要原地重试**。
5. 备好输入：`--input @<file>`（可多次）。本地文件用 `@` 前缀；远程用 `https://` / `tos://` URL。
6. **默认不传 `--model`**：Platform Profile 会使用 `Resources.Text.Default` 中已部署的 `ep-xxx`，与 `+chat` 保持一致；Plan 类 Profile 才使用 sub-skill 的 recipe fallback 模型。只有用户明确指定其他资源时才传完整版本化 ID 或 `ep-xxx`。
7. 需要逐段输出加 `--stream`；想微调任务指令用 `--system-prompt-append "..."`，或 `--system-prompt-override "..."`。
8. 跑完回到用户原始目标，不要停在中间产物上。

## sub-skill 速查表（12 个 / 4 模态）

> 下表是 Plan 类 Profile 使用的 recipe fallback 模型。Platform Profile 不使用这些裸模型名，而是使用当前 Profile 的 `Resources.Text.Default` Endpoint。确需覆盖时见上文执行顺序。

| sub-skill | 模态 | Plan recipe fallback 模型 | 一句话用途 |
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
| `arkcli +understand image-caption --input @x.jpg "描述图片" --dry-run --format json` | 返回 recipe / 模型 / 请求摘要的 `preview.v1`；零网络、不上传、不推理 |

## 详细文档

- 全部 flag、`@file`/audio 上传机制、返回值扁平 schema、错误码、raw API 兜底、与 `+chat` 的关系 → [`references/arkcli-understand.md`](references/arkcli-understand.md)
- 12 个 sub-skill 的逐条用途、**期望输出形态**、示例命令 → [`references/sub-skills.md`](references/sub-skills.md)
- 最小评估/回归用例 → [`references/evals.md`](references/evals.md)
