# +understand sub-skill 配方目录

> **前置条件：** 先读 [`arkcli-understand.md`](arkcli-understand.md)（命令/flag/返回值/错误）。

12 个 sub-skill = 12 条配方 `{模态, 默认模型, 内置 system prompt}`，跑在同一个 Responses 引擎上。下表的「期望输出形态」来自各 sub-skill 的内置 system prompt，是 agent 解析 `.content` 的依据。

通用规则：

- **默认模型别覆盖**（已验证可用）；要换 `--model` 须用完整版本化 ID 或 `ep-xxx`。
- `[prompt]` 位置参数是叠加在专家 prompt 之上的**用户附加要求**（指定字段、目标、语言、粒度等）。
- 想改输出格式/语言又保留专家能力：`--system-prompt-append`；想完全自定义：`--system-prompt-override`。
- JSON 类产出（grounding / gui / doc-extract）落在 `.content` 字符串里，需 `jq -r .content | jq .` 二次解析。

---

## image 模态（默认模型 `doubao-seed-1-6`）

### `image-caption` — 图片描述 / OCR
- **用途**：单图或多图的清晰、结构化描述与解析；用户要 OCR 时完整保留图中文字、不意译。
- **期望输出**：自然语言描述文本；OCR 场景为原样保留的文字。
- **示例**：
  ```bash
  arkcli +understand image-caption --input @photo.jpg "描述画面主体和场景"
  arkcli +understand image-caption --input @receipt.png "请 OCR，原样保留所有文字与数字"
  arkcli +understand image-caption --input @a.jpg --input @b.jpg "对比这两张图的差异"
  ```
- **自动路由**：省略 sub-skill 且首个 `--input` 是图片时，兜底到本 sub-skill。

### `image-grounding` — 视觉定位（bbox）
- **用途**：给定图片 + 目标描述，精准定位目标在图中的位置。
- **期望输出**：每个目标的 bbox 坐标 `(x1, y1, x2, y2)` + confidence；**图中不存在目标时输出空数组 `[]`**，不编造。
- **示例**：
  ```bash
  arkcli +understand image-grounding --input @scene.jpg "框出所有红色车辆"
  arkcli +understand image-grounding --input @ui.png "定位登录按钮" | jq -r .content | jq .
  ```

### `image-gui` — GUI 操作识别
- **用途**：给定界面截图 + 操作目标，输出可执行操作序列（GUI agent 用）。
- **期望输出**：JSON 数组，每步含操作类型、目标元素 bbox（若适用）、必要参数（如 type 的文本）。操作类型限 12 类：`click` / `double-click` / `right-click` / `tap` / `long-press` / `scroll` / `swipe` / `drag` / `type` / `key` / `wait` / `screenshot`。
- **示例**：
  ```bash
  arkcli +understand image-gui --input @screen.png "在搜索框输入 arkcli 并点击搜索"
  ```

---

## video 模态

### `video-summary` — 视频总结（默认 `doubao-seed-1-6`）
- **用途**：结合画面 + 音频，结构化总结视频。
- **期望输出**：按用户指定粒度——`overall`（整体摘要）/ `segment`（按时间段）/ `chapter`（按主题章节）；每段附关键时间点（`HH:mm:ss`）、画面要点、对话要点、整体主题。不编造未出现内容。
- **示例**：
  ```bash
  arkcli +understand video-summary --input @lecture.mp4 "按 chapter 总结，每章给起始时间点"
  ```

### `video-qa` — 视频问答（默认 `doubao-seed-1-6`）
- **用途**：按用户具体问题，结合画面 / 音频 / 时间轴精准回答。
- **期望输出**：优先直接答案，必要时附支撑证据（画面描述 / 对话片段 / 时间点）；视频中无相关信息时明确说「视频中未涉及」，不编造。
- **示例**：
  ```bash
  arkcli +understand video-qa --input @clip.mp4 "视频里一共出现了几个人？分别在做什么？"
  ```
- **自动路由**：省略 sub-skill 且首个 `--input` 是视频时，兜底到本 sub-skill。

### `vau` — 音视频联合理解（默认 `doubao-seed-2-0-lite`）
- **用途**：VAU（Video-Audio Understanding），深度分析视频中的声音元素（人声 / 音效 / 音乐）及其声学特征与叙事作用。模态登记为 video，**输入仍是视频文件**。
- **期望输出**：结构清晰、内容详实的音频分析报告。
- **示例**：
  ```bash
  arkcli +understand vau --input @film_clip.mp4 "分析这段视频的声音设计"
  ```
- **注意**：`vau` 不在视频的自动路由默认里（视频默认是 `video-qa`），需**显式**指定。

---

## audio 模态（默认模型 `doubao-seed-2-0-lite`）

> audio 走 base64 内联，**上限 25MB**（见 [`arkcli-understand.md`](arkcli-understand.md) 的大小约束）。

### `asr` — 语音转写
- **用途**：纯语音转文字（ASR）。
- **期望输出**：**仅**转写文本——无引导语、无解释、无 markdown 格式；音频不清晰 / 无语音时输出空字符串。
- **示例**：
  ```bash
  arkcli +understand asr --input @speech.mp3
  arkcli +understand asr --input @speech.mp3 | jq -r .content   # 直接拿转写正文
  ```
- **自动路由**：省略 sub-skill 且首个 `--input` 是音频时，兜底到本 sub-skill。

### `asr-align` — 字幕打轴
- **用途**：带时序的转写，输出字幕。
- **期望输出**：默认 **SRT** 格式（每条：序号 / 起止时间 `HH:MM:SS,mmm --> HH:MM:SS,mmm` / 文本三行，条目间空行）；用户给模板时按模板输出。
- **示例**：
  ```bash
  arkcli +understand asr-align --input @speech.mp3                       # 默认 SRT
  arkcli +understand asr-align --input @speech.mp3 "按 WebVTT 格式输出"   # 自定义模板
  ```

### `asr-speakers` — 多说话人转写
- **用途**：多人对话转写并标记说话人。
- **期望输出**：带说话人标签的转写——第一个人 `[spk0]`、第二个 `[spk1]`，依此类推。
- **示例**：
  ```bash
  arkcli +understand asr-speakers --input @meeting.wav
  ```

### `ast` — 语音翻译
- **用途**：把音频里的口语内容翻译成文本（AST）。
- **期望输出**：译文文本。内置 prompt 未固定目标语言，**用 `[prompt]` 指定**（如「译成英文」）。
- **示例**：
  ```bash
  arkcli +understand ast --input @speech.m4a "翻译成英文"
  ```

### `meeting-minutes` — 会议纪要
- **用途**：根据会议音频（内置 prompt 也支持会议视频语义），输出结构化纪要。
- **期望输出**：markdown 纪要，含：会议主题 / 参会人（按声纹标 `spk0` `spk1`...）/ 议程与讨论要点（按时间序、附时间戳）/ 决议事项（Action Items，含 owner / deadline 若可识别）/ 待跟进问题。保留关键引述，避免主观推断。
- **示例**：
  ```bash
  arkcli +understand meeting-minutes --input @standup.m4a
  ```

---

## file 模态（默认模型 `doubao-seed-1-6`）

### `doc-extract` — 文档字段抽取
- **用途**：从 PDF/文档抽取结构化信息。
- **期望输出**：按用户指定的 schema（字段名 / 含义）输出 JSON；字段在文档中不存在时填 `null`，不凭空生成；多页文档保留原文出现的页码 `page` 以便溯源。
- **示例**：
  ```bash
  arkcli +understand doc-extract --input @invoice.pdf "抽取：发票号 / 金额 / 开票日期 / 销售方"
  arkcli +understand doc-extract --input @contract.pdf "抽取甲方、乙方、合同金额、签署日期" | jq -r .content | jq .
  ```
- **自动路由**：省略 sub-skill 且首个 `--input` 是非图/视/音的文件（如 `.pdf`）时，兜底到本 sub-skill。
