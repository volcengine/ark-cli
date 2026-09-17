---
name: arkcli-chat
version: 1.2.5
description: "arkcli +chat：通过数据面 Responses API 快速对话/推理，支持多模态、流式、多轮、临时 API Key/Base URL/Endpoint 执行与无副作用 dry-run。当用户给出 Endpoint 但未说明工作流时，先用 resources resolve 识别候选；已经出现 Responses API capability/access 错误时，只读用 models get 核对精确模型的 api_support，不重试真实调用。有明确产出形态的多模态理解走 arkcli-understand。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli +chat --help"
---

# arkcli +chat

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、配置排查与命令选择顺序**
**CRITICAL — `+chat` 在执行之前，务必先用 Read 工具读取 [`references/arkcli-chat.md`](references/arkcli-chat.md)，禁止直接盲目调用命令。**
**CRITICAL — 一次用户要求的回答只发起一次真实 `+chat`。首次响应不完整、strict JSON 无效或不符合内容要求时，保留并报告该次失败，不自动重试或用第二次响应替换；本地提取、保存、校验失败也只能修复本地交付。只有用户明确要求新一轮或重新生成时才再次调用。**

## 核心概念

- `--model` 缺省时 CLI 自动 fallback 到 active profile 的 `resources.text.default`；用户显式传入不同值时按 [`../arkcli-shared/references/profile-defaults.md`](../arkcli-shared/references/profile-defaults.md) 做漂移提示。
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
- 支持**可验证的严格 JSON**：`--text-format json_schema --text-schema <file> --text-strict` 不只把约束传给服务端，还会在客户端确认响应完整、是直接 JSON 且符合 Schema；否则非零退出。严格流式会先缓冲，校验成功后才输出，避免泄出半截 JSON。详见 [`references/text-format.md`](references/text-format.md)。
- **输出 schema（务必先看）**：`+chat` 返回是 arkcli **扁平** schema（`{id, model, content, reasoning_content, usage, ...}`），**不是** Responses API 原生 `output[].content[].text` 嵌套；助手文本直接用 `.content` 取。`--format` 不会切换 shape。详见 [`references/arkcli-chat.md`](references/arkcli-chat.md) 的「返回值」段。
- 用户临时提供 `--api-key` / `--base-url` / Endpoint 时，MUST 读取 [`../arkcli-shared/references/execution-context.md`](../arkcli-shared/references/execution-context.md)。不要从 Key 文本猜套餐类型，也不要把临时值写入 profile。
- `--dry-run` 只在本地构造 `preview.v1` 请求摘要与无 secret 的执行上下文；不会读取在线元数据、刷新凭证、调用 Responses API、产生 token 用量或存储 response。在线依赖必须列为 `unresolved`。

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

先把用户要求记成可验收项：回答任务、指定模型/Profile、输入文件、输出格式、流式/多轮、
明确的参数值与交付文件。只预览请求时直接走本地 `--dry-run`，不先联网查认证或资源。

1. 用户只给 `ep-...` 且任务不明确 → `arkcli resources resolve <ep-id> --format json`；开放问答/追问才选择 `+chat`。
2. 用户给了临时 Key/Base URL/Endpoint → 按共享 execution-context 组合规则决定参数，不先切 profile。
3. 不确定认证状态且没有完整 stateless 上下文时，先看 `arkcli auth status`；未登录/无 API Key 转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)。
4. 不确定模型名时，先转 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md)。
5. 使用本次资源确认的调用 ID，不强制改成版本化名称。用户未指定模型时保留当前 text default 或省略 `--model`；合法套餐别名（如 `ark-code-latest`）不追加版本。用 `resources list --modality text` 核对 default、`invocable`、`required_overrides` 与凭证类型；可见不等于可调用。能力查询的规范 Name/Version 与计费调用 ID 分开，EP 始终保持 EP。只有未解析的模型族名才进一步查询，不凭字符串猜版本。
6. 需要流式输出时加 `--stream`；需要多模态时加 `--input @<file>`（可多次）。
7. 查模型实际支持的键、值和输入模态；不要把参数全集当成所有模型的能力。用户明确的 temperature、token 上限、thinking、store=false 等不得省略或换值；未指定时使用兼容默认，不额外开启联网、存储或工具执行。
8. 发起真实请求前先确定交付路径；要求保存时按 reference 的单次捕获流程，让调用 stdout 直接进入临时响应文件。用首次真实返回验收：普通结果取 `.content`，严格 JSON 检查完整性/Schema；多轮必须引用真实上轮 ID 并保持相同调用上下文。流式保留 NDJSON，确认成功终态后提取正式正文，不能把推理文本或 Agent 自己的回答充当模型产出。保存文件必须与该次正文逐字一致；任何本地失败都不得触发第二次模型请求。

认证准入只陈述已验证的事实：本地登录、Key 清单 Active 或资源可见均不能单独证明当前数据面请求可用。
实际调用报 401 时核对当前 lane 凭证，403/access 错误按下文只读核对能力和权限；
429 要区分限流与 quota exhausted。提醒可用候选，但不自动轮转 Key、切 Profile/收费路径或修改默认。
stored response 不存在/过期时准确报告，不能悄悄去掉 previous-response-id 重新对话后声称上下文接续成功。

## 常见降级

- 模型名不确定：先 `models search`。
- endpoint ID 要用 `+chat`：直接传 `--model ep-xxx`（endpoint 本身已决定模态，无需额外 flag）。
- Endpoint + 显式 API Key、未给 Base URL：CLI 会读取 Endpoint region 后派生 Base URL；无法取得权威 region 时再要求用户补 `--base-url`。
- 鉴权失败：转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)。

## Responses API capability/access 错误的只读核对

用户已经给出 `model does not have access to responses api` 类错误时，本轮是排障，不是「再试一次」：

1. 若已知精确、完整的版本化模型 ID，只执行 `arkcli models get <model-id> --format json`；禁止用 `models search` 的候选摘要代替单模型详情。
2. 在 `api_support` 数组中按 `name` / `key` / `path` 定位 Responses 项，以该项的 `supported` 为模型声明事实；不从模型名、lifecycle、tool 列表或其他 capability 反推。
3. `supported=true` 但实际调用报 access 错误：说明「模型声明支持，当前 Endpoint / 账号访问路径不可用」；若用户还给了 Endpoint ID，可再只读 `resources resolve` / `infer endpoint get` 核对绑定与状态。
4. `supported=false` 才能说模型目录声明不支持；Responses 项缺失则说明元数据不足，不做猜测。

全程禁止再次执行 `+chat`、自动 `models activate`、切 profile 或修改默认资源。

## 模型能力提示（stderr warn，不阻断）

`+chat` 真实执行时会在发请求前读一次 ArkModels 元数据，把模型差异以 `warn: ` 前缀打到 **stderr**；请求照常发出。识别要点：

1. **只警告、不阻断**：出现 `warn:` 行不代表请求失败。值为 `enabled`/`disabled`、别名表、声明取值集都只是「模型声明」，服务端与 SDK 解码器才是裁决方。
2. **不要把声明集当接受集**。实测：`glm-5-2` 声明 `reasoning_effort` 为 `[none, minimal, high, max]`，但服务端**拒绝** `none`/`minimal` 而**接受** `low`/`medium`（只出现在其 `mapping_config` 别名表里）。声明集与接受集可能近乎互补。
3. **只按实际存在的元数据提示**：缺少 reasoning 元数据时不得把其他 capability 当成 reasoning 声明；模型不在 ArkModels 中（如 `kimi-k3`）或查询失败时**静默跳过**，不产生任何 warn。
4. **完整版本 ID 必须原样执行**：可以用裸模型名查对应版本元数据，但不得把用户指定的旧版本自动改写成当前主版本。
5. **stderr 变化不影响 stdout 契约**：`--format json` 的 stdout 结构不变，`warn:` 行只进 stderr。需要稳定 JSON 时只解析 stdout，同时保留并检查 stderr，不得用 `2>/dev/null` 丢弃 warning。
6. **`--dry-run` 不读元数据**：Client Preview 保持零网络，不会出现任何 `warn:` 行。

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
| `arkcli +chat --model <id> --dry-run "<prompt>"` | 无副作用预演；不调用 Responses API |

## 详细文档

- `+chat` 的所有参数、返回值、错误码、多模态文件自动上传机制等见 [`references/arkcli-chat.md`](references/arkcli-chat.md)。
- `+chat` 的 Tools 能力（`--tools / --tools-file / --tool-choice / --max-tool-calls`，含 function 与 web_search）见 [`references/tools.md`](references/tools.md)。
- `chat get / chat delete / chat list-input-items` 三个对账面命令见 [`references/chat-meta.md`](references/chat-meta.md)。这些命令操作的 response 必须是 `+chat --store` 过的。
- `+chat` 的 `--caching / --cache-prefix / --thinking / --expire-at` 用法、回显字段与 autotest 对应见 [`references/caching-thinking.md`](references/caching-thinking.md)。
- `+chat` 的 `--text-format / --text-schema / --text-schema-name / --text-strict` 用法见 [`references/text-format.md`](references/text-format.md)。
- `+chat` 的 `--stream --include-events` NDJSON 流式事件输出见 [`references/stream-events.md`](references/stream-events.md)。
