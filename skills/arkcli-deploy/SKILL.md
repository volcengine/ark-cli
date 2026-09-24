---
name: arkcli-deploy
version: 1.4.9
description: "arkcli +deploy：普通创建推理接入点（Endpoint）的统一首选入口。用户说『创建/新建/create 一个 endpoint/接入点』或『部署/上线/deploy 某模型』时优先走这里；**但脚本化 / CI / 无护栏 / 原始 raw CRUD 创建是唯一例外，必须改走 `arkcli-infer-endpoint`，不能由本 skill 截获**。只有不命中该例外的普通创建，才在模型尚未选择、只给品牌/家族名，或当前轮只查询候选时进入本 skill 完成模型澄清；候选要用实时 search 返回的 `name` 与 `primary_version` 组合成可直接传给 `--model` 的完整 ID。对**已有** Endpoint 做获取/列表/启停/更新等全生命周期管理也走 arkcli-infer-endpoint；本 skill 只负责带产品护栏的一键创建。创建成功后会自动把多语言调用示例渲染到 ./ark-examples/<ep-id>/。反触发：TTS/ASR/语音模型不能 +deploy，只能转 models search 说明广场可搜但 arkcli 不支持 Endpoint 创建。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli +deploy --help"
---

# arkcli +deploy

**CRITICAL — 路由例外必须先于任何命令：用户明确要求脚本化 / CI / 无护栏 / 原始 raw CRUD 创建 Endpoint 时，立即读取 [`../arkcli-infer-endpoint/SKILL.md`](../arkcli-infer-endpoint/SKILL.md) 并由它接管；在完成交接前禁止认证检查、模型查询或其他命令。**

**前置：** 先用 Read 读 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md) 获取共享认证/配置/写操作守卫规则。

## 创建意图中的模型澄清

只要用户的最终目标仍是“创建 / 新建 / 部署 Endpoint”，即使尚未给出完整模型 ID，也必须留在 `arkcli-deploy` 工作流；`arkcli-models` 此时只是临时调用的只读候选查询能力，不能把创建任务改路由成纯模型发现。

候选必须来自**本轮** ArkCLI 的实时结构化输出，禁止从模型记忆、示例或旧版本号补全。

**查询预算是当前用户回合恰好一次 Bash 调用。** 在执行前一次性确定 keyword、过滤条件和 `--size`。`models search` 的 keyword 是单个 catalog 子串，不要把多个概念拼成带空格的短语：先选最能缩小范围的一个 ASCII token，其余条件放到同一次调用的 flags 或 `jq` 本地过滤。例如“豆包代码模型”用 keyword `code`，再在同一管道按 `name` 的 `doubao|seed` 过滤；“生图”用 `seedream`。不确定稳定 token 时宁可省略 keyword 并使用已有结构化 filter，禁止把未翻译的中文短语直接提交后再换词重试。若担心输出过长，把字段投影直接放进同一条管道，例如
`arkcli models search "<keyword>" --size 0 --format json | jq -c '{items: [(.items // [])[] | {name, primary_version, lifecycle_status, input_modalities, output_modalities}]}'`。
这次调用无论成功、空结果、截断还是失败，都不得换关键词、调大分页、重跑同一命令或为了重新格式化再调用一次 ArkCLI；只能使用已捕获的 stdout，信息不足就如实停止。

1. **完全没给模型**：执行一次有界查询，例如 `arkcli models search --size 10 --format json`。若用户已说明用途或模态，把对应的 keyword / `--modality` 加进同一次查询。
2. **只给品牌、系列或家族名**（例如 “Doubao”“Seed 2”）：执行一次 `arkcli models search <keyword> --size 10 --format json`；家族名不是可直接传给 `--model` 的完整 ID。
3. 从同一次返回的 `items` 中读取 `name`、`primary_version`、`lifecycle_status` 和模态等已有字段。优先保留 `lifecycle_status=Published` 且名称不含 `internal` / `test` 的候选；状态缺失或非 Published 时只能如实标为未核实，不能称为“可部署”。完整模型 ID 按模型查询契约确定：`primary_version` 非空时使用返回值精确拼成 `<name>-<primary_version>`，为空时才使用 `<name>`；这是结构化字段组合，不是从名称或日期规律猜版本。不要自行给 `name` 拼 `pro`、`lite` 或任何未返回的版本后缀。
4. **0 个可用候选**：说明本轮没有查到，并请用户补充用途、模态或关键词；不要猜一个继续。
5. **1 个可用候选**：复述本轮返回的完整 ID，请用户确认；若用户已要求本轮只查询，则停在这里。
6. **多个可用候选**：列出精简候选及完整 ID，请用户明确选择；若候选仍过多，按用户用途缩小范围，不能擅自选第一项。

澄清阶段的收敛边界：本回合第一次 `search` 返回后，**禁止再次执行 `models search`，也不要对候选循环执行 `models get`、价格查询或其他详情调用**。只有用户选定单个候选后又明确要求比较某个缺失属性，才在后续回合追加一次针对性查询。模型尚未唯一确定前，禁止执行 `+deploy`、`infer endpoint create` 或 Raw API 创建。

Endpoint 名称不应阻塞上述只读候选查询；模型选定后，再收集名称并确认最终创建参数。

**新增 flag `--set-default <modality>`**: 部署成功后自动把新 endpoint 设为 active profile 该 modality (`text` / `image` / `video`) 的默认资源。仅在真实部署成功且用户明确传 modality 时生效；失败仅 stderr warn，不阻断部署主流程。详见 [`../arkcli-shared/references/profile-defaults.md`](../arkcli-shared/references/profile-defaults.md)。

**写操作 + 计费**：`+deploy` 创建在线推理 Endpoint 是真实写操作，会产生计费资源。该工作流依赖在线探测，**不支持 `--dry-run`**；执行前必须与用户显式确认最终参数。

**模型开通是独立计费写动作，非交互环境一律不自动开通**：若目标基础模型尚未开通，`+deploy` 会触发"开通模型"（账号级计费写）。**在 agent / CI / 管道这类非 TTY 环境，开通被硬拒——`--yes` 也不放行**（`--yes` 只在真人交互终端里用于跳过 `[y/N]`）。命中时 CLI 返回 `model_activation_required` + console 链接，你**必须结束本轮、把"开通（计费）"这件事连同链接交还给真人**，由真人在交互终端确认或在网页 console 开通。**严禁自己补 `--yes` / `echo Y` / 设 `ARKCLI_ALLOW_HEADLESS_ACTIVATION` 替用户开通**——你打印一句"请确认"然后同一轮自己加 `--yes` 跑掉，等于没问。`ARKCLI_ALLOW_HEADLESS_ACTIVATION=1` 只留给真·无人值守自动化（CI 流水线），不是 agent 该设的。

**实名前置（开通类硬闸门）**：`+deploy` 会触发开通模型，**第一步**先 `arkcli auth status` 读 `volc_sso.identity.verified`——`false` 即停并把实名页 `https://console.volcengine.com/user/authentication/detail/` 贴给用户、暂停等待（详见 [`../arkcli-auth/references/realname-gate.md`](../arkcli-auth/references/realname-gate.md)）。**不要**先试 `+deploy`、撞到 model-id 无效 / 模型未开通报错再回头查实名——那些报错会把你带偏。

**语音模型硬边界**：TTS / ASR / 配音 / 朗读 / 播客 / 音色设计 / 实时语音交互，或模型名命中 `doubao-seed-tts-*`、`doubao-seed-asr-*`、`seedasr-*` 时，**不要执行 `+deploy`**。这些模型在 arkcli 当前只支持 [`models search`](../arkcli-models/SKILL.md) 做广场发现；不能创建 Endpoint，也不能通过开通模型绕过。

## 示例代码：`+deploy` 创建后自动渲染 + 独立命令 `+code-example`

两条路径都可用，都走 OpenTOP `OpenGetSampleCode`：

- **`+deploy` 创建成功后自动渲染**：把该接入点的多语言调用示例写到 `./ark-examples/<ep-id>/`（用 ep-id 当作可直接调用的 model id，示例直接调你刚建的接入点），并在 stderr 打一行落盘摘要。best-effort —— 取不到示例只软提示，不影响 Endpoint 创建/启动。
- **独立命令 `arkcli +code-example`**：按基础模型名/版本单独生成示例，转 [`../arkcli-code-example/SKILL.md`](../arkcli-code-example/SKILL.md)；注意它按 model-version 提供，部分版本后端无 group 会返回 not found，缺失时再降级到 `ark-examples/` 静态示例或方舟控制台示例代码页。

## 子命令穷举（只有这一个）

| 调用 | 说明 |
|------|------|
| `arkcli +deploy --name <ep-name> --model <model-id> [...]` | 创建 Endpoint；执行即真实创建 |

> ⚠️ **没有** `arkcli deploy ...` / `arkcli endpoint create` / `arkcli +deploy create` 等子命令。整个能力就是一个 `+deploy` 命令加 flag。

## 反幻觉清单

- `--name`、`--model` 必填
- 模型版本、价格与额度只能引用本轮只读命令返回的结构化结果；**不得编造模型版本或价格**。查询失败或未登录时，明确标为“尚未核实”，不要用“典型价格”或猜测值代替
- **查询失败后不得补全依赖字段**：认证、模型、价格或资源查询任一失败时，保留对应字段为“尚未核实”，并明确指出失败来源；不得继续断言具体金额、空闲计费策略、精确模型版本或“没有现存 Endpoint”
- 检查已有 Endpoint 时使用 `arkcli resources list --format json` 后按返回字段筛选；**不得添加未记录的 `--filter`** 或其他 `resources list --help` 中不存在的 flag
- JSON 类 flag（`--rate-limit` / `--moderation` / `--intelligent-router` / `--tags` 等）字段名一律 **PascalCase**：`Rpm`、`Tpm`、`Strategy`、`Mode`，不是 `rpm`/`tpm`
- 独立 `+code-example` 的 flag 是 **`--model`（基础模型名/带版本 id）+ `--language`**，不是 `--endpoint-id`：`arkcli +code-example --model <id> --language python`（细节见 [`../arkcli-code-example/SKILL.md`](../arkcli-code-example/SKILL.md)）
- `+deploy` 创建成功后会**自动**把示例渲染到 `./ark-examples/<ep-id>/`（按 ep-id）；想按基础模型名另出一份则跑 `+code-example`
- 模型未开通时 `+deploy` 的开通在**非 TTY 下被硬拒、`--yes` 也不放行**；**禁止自己补 `--yes` / `echo Y` / 设 `ARKCLI_ALLOW_HEADLESS_ACTIVATION`**，必须把开通（计费）交还真人在终端 / console 处理
- 语音模型（TTS / ASR / 播客 / 音色 / 实时语音交互）广场可搜不等于可部署；命中这类模型时停在 `arkcli models search <keyword>`，不要给 `+deploy` 命令

## 路由判断

- 用户要创建 / 部署 Endpoint，但模型缺失或只有品牌 / 家族名 → **仍路由到本 `arkcli-deploy` skill**，按上节执行一次实时只读查询并让用户选择；不要直接创建
- 用户已有模型 ID + 想正式部署 → 复述 `model/name/region`；确认后执行 `arkcli +deploy --name <ep> --model <id>`
- 用户要部署 / 接入语音模型，或模型名看起来是 `*-tts-*` / `*-asr-*` / `seedasr-*` → 转 [`arkcli-models`](../arkcli-models/SKILL.md) 说明"只支持广场检索，不支持 Endpoint 创建"
- 用户传入自定义模型 ID（`cm-xxxxx`）时，真实创建前会先查是否已有引用该自定义模型且状态为 `Running` 的 Endpoint；若有则直接复用并输出已有 `endpoint-id`，不会再创建第二个计费资源。该在线复用决策也是 `+deploy` 无法提供可靠离线 Client Preview 的原因之一
- 用户语气紧急要求"立刻创建" → **不要跳过确认**，复述 `model/name/region`
- 已通过 `arkcli infer endpoint create` 拿到 `Id` → **不要**再 `+deploy` 创建第二个，转 `arkcli-infer-endpoint`；要调用示例转 `arkcli-code-example`（按模型名生成）

## 反触发（路由到别处，附完整命令避免下游幻觉）

| 用户意图 | 路由到 | 完整示范命令 |
|---------|--------|------------|
| 只想试模型效果 / 一次性生成 | `arkcli-chat` / `arkcli-gen` | `arkcli +chat --model <id> '...'` 或 `arkcli +gen --model <id> '...'` |
| 要某模型的调用示例 | `arkcli-code-example` | `arkcli +code-example --model <model-id> --language python`（按模型名/版本生成；缺失版本降级到静态示例或控制台示例页） |
| 只想发现 / 对比模型，尚无创建 Endpoint 意图 | `arkcli-models` | `arkcli models search <keyword>` 或 `arkcli models list` |
| 语音模型部署 / TTS 接入点 / ASR Endpoint | `arkcli-models` | `arkcli models search <keyword>`（只做广场发现；当前不支持 Endpoint 创建） |
| 401 / 鉴权失败 | `arkcli-auth` | `arkcli auth status`，必要时 `arkcli auth login` |
| profile / region / project 不符预期 | `arkcli-config` | 身份摘要用 `arkcli auth whoami --format json`；默认资源用 `arkcli resources list --format json`。显式 Profile 管理才使用可能同步 Key 的 `profile show/list` |
| 脚本化 / CI / 需要精细控制每个参数、跳过护栏 | `arkcli-infer-endpoint` | `arkcli infer endpoint create --model <id> --name <ep>` |

## 典型链路

1. **从模型选择到正式接入**：`arkcli auth status` → `arkcli models search/get` → 复述并确认最终模型、名称、计费影响 → `arkcli +deploy ...`（自定义模型若已有 Running Endpoint 会复用）
2. **从试用切换到正式接入**：`arkcli +chat` / `arkcli +gen` 验证效果 → 复述最终参数并确认 → 真实创建
3. **创建后做调用集成**：`+deploy` 已把示例自动写到 `./ark-examples/<ep-id>/`，直接用即可；想按基础模型名另出一份跑 `arkcli +code-example --model <model-id> --language <lang>`（部分版本无示例时降级到 `ark-examples/` 静态示例或控制台示例页）

详细 flag、JSON 字段示例、错误码见 [`references/arkcli-deploy.md`](references/arkcli-deploy.md)。

## 参考

- [arkcli-chat](../arkcli-chat/SKILL.md) -- 快速对话试用，不创建 Endpoint
- [arkcli-gen](../arkcli-gen/SKILL.md) -- 图片/视频一步生成
- [arkcli-models](../arkcli-models/SKILL.md) -- 部署前确认模型 ID
- [arkcli-code-example](../arkcli-code-example/SKILL.md) -- 已有 endpoint 时生成调用代码
- [arkcli-shared](../arkcli-shared/SKILL.md) -- 认证与全局参数
