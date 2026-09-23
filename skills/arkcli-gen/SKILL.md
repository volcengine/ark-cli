---
name: arkcli-gen
version: 2.1.4
description: "火山方舟 Ark 图片/视频生成入口：支持 profile 默认资源与临时 API Key/Base URL/Endpoint；显式 Endpoint 不受当前 plan profile 误导。图片同步返回，视频异步轮询。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli +gen --help"
---

# arkcli 生成工作流（+gen）

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)（认证闸门、模型查找回退、共享安全规则）。**

**CRITICAL — 真实生成是工作流，不是猜一条命令：按 `Step 1 → Step 2 → 必要时 Step 2.5 → Step 3` 执行。用户只要求 `--dry-run` 时例外：全程本地，不先跑在线 Resources / Models / Usage 准入；在线未知项保留 `unresolved`。执行前务必读 [`references/arkcli-gen.md`](references/arkcli-gen.md)。**

**CRITICAL — 用户显式给出 API Key / Base URL / Endpoint 时，MUST 先读 [`../arkcli-shared/references/execution-context.md`](../arkcli-shared/references/execution-context.md)。显式 Endpoint 的权威元数据优先于当前 profile。**

**火山额外约束：不要因为 active profile 是 Agent/Coding Plan 就把用户给出的 Endpoint 当套餐模型调用。**

## Agent 快速执行顺序

先把用户要求拆成可验收项，再按资源、能力、整批预算、执行、成品检查推进。执行前读取
[`references/intent-and-validation.md`](references/intent-and-validation.md)：它约定不完整意图的默认值、
参数组合、凭证错误分流和成品验收。参数表是候选能力，不代表每个模型都支持。
用户已明确要图片或视频时，生成命令显式带 `--modality image|video`；不要让名称解析覆盖用户意图。
只使用用户本轮提供或明确授权复用的素材；不得从历史目录、旧任务或相似文件名擅自加入额外参考图/视频。

### 成品验收前：宿主视觉输入准入

生成模型能出图，不代表驱动当前 Agent 的宿主模型能看图。**Read 工具描述说支持图片，
也不是当前宿主模型的视觉能力证明。** 当前会话没有可靠的视觉能力声明或已验证的兼容性时，
按“未准入”处理：**不要原生 Read 图片、视频或抽出的帧，也不要用成品试探能否读取。**
本 Skill、参考说明和 JSON 等文本仍可正常 Read。

- 已有兼容且已授权的视觉工具，或同一主体下已准入、收费路径不变的视觉模型时，
  按对应 Chat/Understand Skill 分析实际媒体，只把文字结果返回宿主；不自动换宿主、Profile 或 Key。
- 没有这样的入口时，完成文件/解码/尺寸/时长等结构检查并交付已有成品，明确“视觉语义未验”；
  不声称看过或完全符合。不要为补验收擅自增加未授权的收费调用。
- 生成成功后宿主读图报错，是验收/交付失败；保留成品，不重新提交生成或轮转业务 Key。

## 反唤起信号

- 描述、分析现有图像而不是生成/编辑 → `arkcli-chat` 或有固定产出形态的 `arkcli-understand`。
- 只查资源、能力、用量 → 对应只读 Skill，不提交生成任务。

## 为什么是工作流（核心，先理解再执行）

用户说"生成一个视频/一张图"，本质是至少**三件独立的事，必须按序**；批量或多阶段任务还要先确认整批可完成：

```
① 本次资源从哪里来           ── 用户显式 Endpoint 优先；否则看当前 profile
② 该模型支持哪些参数          ── 不查就传参 = 瞎猜 = 被校验拒/被后端拒
②.5 多候选/多阶段额度与任务数 ── 先算完整批次，额度已耗尽就不启动半批任务
③ 按可用参数真去生成          ── 每个请求只提交一次并立即保存 task_id
```

把这三步压成"直接 `+gen` 猜一条命令"，正是失败之源：模型名形态不对会 404，参数模型不支持会被拒。

## 模态解析硬契约

`+gen` 的生产调用按以下固定优先级解析能力：

```text
explicit --modality > output_modalities > task types > unknown
```

- 直接传版本化模型 ID 时，读取 ArkModels 返回的 `output_modalities`；缺失时再读取 FoundationModel 的 `task_types` / `filter_task_types`。
- 传 `ep-*` 时，先读取 Endpoint 的 `ModelReference.FoundationModel(name, version)`，再精确匹配同版本模型的上述结构化元数据。
- **模型名与 DisplayName 只用于定位模型，不参与模态判断**。不要从 `seedream`、`seedance` 或任何国内/海外品牌前缀推断 image/video。
- 结构化元数据缺失或互相冲突时返回 `unknown`，提示用户显式传 `--modality image|video`；禁止静默猜测。
- `+gen --dry-run` 是纯本地 Client Preview：不读取 Endpoint/模型元数据、不调用
  生成 API，也不下载或打开文件。显式 `--modality` 最可靠；已知
  `seedream`/`seedance` 模型名可本地判断，其他模型或 Endpoint 必须显式传
  `--modality image|video`。在线才能补齐的执行上下文会以 `unresolved` 和
  `fidelity=partial` 明示。

## 适用场景

- "生成一张图" / "文生图" / "画一个 X"
- "生成一个视频" / "文生视频"
- 图生图 / image-edit / 加参考图；图生视频(I2V)；参考视频(R2V)；参考音频
- "用这张图当首帧生成视频" / "保持这个参考视频的运动"

## 工作流总览

```text
用户意图: "生成 X"
  │
  ▼ Step 1【强制】解析本次资源
  │     用户给 Endpoint → arkcli resources resolve <ep-id>
  │     未给 Endpoint   → arkcli resources list --modality image|video
  │
  │     当前 profile 可用资源：
  │     platform    → 列 EP (ep-xxx)           ┐
  │     agent-plan  → 列视觉模型名              ├─ 选一个，记为 $MODEL
  │     coding-plan → 列 EP (借道 platform)     ┘
  │
  ▼ Step 2【强制】查可用参数  ──► models get（EP 用 resolve 得到的绑定模型查能力）
  │     模型名 + 有 sp → **只能**用列出的参数，取值落 min/max/enum 内
  │     模型名 + sp 空(未配置或当前不可解析) → +gen 自动套 modality 兜底默认(video 720p/5s, image 2048)
  │     EP(ep-xxx)            → 可解析绑定则查精确版本；不可解析则说明未知，不猜支持
  │
  ▼ Step 2.5【批量/多阶段】额度预检 ──► plan/free-quota 快照；记录完整 create 数
  │
  ▼ Step 3 据可用参数生成  ──► arkcli +gen --model $MODEL [Step2 允许的参数] "prompt"
  │
  ▼ Step 4【结果处理】
        视频 = 异步：返回 task_id + status=queued(**不是失败!**) → arkcli gen get <task_id> 轮询;轮到 succeeded 自动下载到本地(local_path);要同步阻塞加 --wait
        图片 = 同步：直接返回 output_url + local_path
```


## Step 1【强制】解析显式 Endpoint，或列出 profile 可用资源

用户已经显式给出 Endpoint 时，不要先用 active profile 的模型池覆盖它：

```bash
arkcli resources resolve "$ENDPOINT" --format json
```

- 读取 `generation_modality` 决定 image/video；`image_or_video` 或 `unknown` 时再结合
  用户意图，必要时显式补 `--modality`。
- 读取 `resource_region`；Endpoint + 显式 API Key 且未给 Base URL 时，CLI 用该
  region 派生 platform Base URL。
- 不按 Endpoint ID 或绑定模型名称里的 `seedream` / `seedance` 子串猜模态。
- 显式 Endpoint + API Key 是临时调用，不切换 active profile，也不把值写回。
- 用户显式提供的 `ep-*` 是本次调用资源；不要忽略它后改用套餐 default，也不要把
  `resources resolve` 的位置参数误传成模型名。若该 Endpoint 已在本轮被用户授权用于
  后付费兜底，套餐额度不足时可回到这里重新做能力检查，但仍不修改 Profile/default。

用户未给显式 Endpoint 时，再按 profile 列资源：

```bash
# 按目标模态列；输出 items[].id 就是可作 --model 的候选
arkcli resources list --modality video   # 或 image
```

- **平台差异（resources list 已自动按 profile 分流，你只管读 items）**：
  - `platform` profile → items 是**推理接入点 EP**（`ep-xxx`），每个 EP 内部绑定一个模型
  - `agent-plan` / `agent-plan-team` → items 是**套餐视觉模型名**；使用对应个人/团队席位 Key
  - `coding-plan` / `coding-plan-team` → 无套餐内视觉模型；生成使用 **platform Endpoint + 后付费 API Key**。团队席位 Key 不能用于这个后付费请求
- `is_default: true` 标记的是该模态当前默认；用户没指定时优先用它
- 再核对 `invocable` / `required_overrides` / `data_plane` / `credential_kind`。默认或可见不等于当前凭证可调用；不要为生成自动切 Profile、轮转 Key 或修改 default。
- **选定一个 id，记为 `$MODEL`，贯穿 Step 2/3**
- 用户已明确给了模型名时，可用 `resources list` 核对当前 lane 的兼容性；用户明确给了 EP 时只先 `resources resolve`，不要再用列表/default 覆盖它。若模型与默认不同，按 [`../arkcli-shared/references/profile-defaults.md`](../arkcli-shared/references/profile-defaults.md) "Default 漂移检测与 promote nudge" 处理


## Step 2【强制】查 $MODEL 的可用参数

```bash
arkcli models get "$MODEL" --transform supported_params
```

- **`$MODEL` 是模型名**：拿到该模型的 `supported_params` 清单（每项含 `name / type / support / min / max / enum / required`）。
  - > **MUST：Step 3 只能使用这里 `support=true` 的参数，且取值必须落在 `min/max/enum` 范围内。** 不在清单里的参数（或 `support=false`）传了会被 `+gen` 拒绝。
  - **可直接使用 Step 1 选出的模型 id**（点号 / display 形态如 `doubao-seedance-2.0-fast` 都行）：`models get` 会自动按 DisplayName 归一化到规范连字符 name，无需手动转。极个别仍报 `not found` 才用 `arkcli models search <族名>` 核对名字。
  - 查到模型但 `supported_params` 为空 / `null` → 该版本未配置参数目录，或上游目录当前不可解析；若 stderr 有 `warn: model supported_params enrichment failed: ...`，保留该告警用于排障。**不要手动猜参数**：`+gen` 会自动用内置 modality 兜底默认（video: `resolution=720p` / `duration=5` / `ratio=adaptive`；image: `size=2048x2048`）填充你没指定的参数。直接进 Step 3。
- **`$MODEL` 是 EP（`ep-xxx`）**：不把 EP 本身交给 `models get`。使用 Step 1 的权威绑定：FoundationModel 查 `model_name` + `--version <model_version>`；CustomModel 只可用 `base_model_*` 查 lineage 能力，不能改写真实 `model_id` 或调用 EP。warning/歧义时说明能力未知，不根据名称猜测支持。
  - 能力查询所得模型 ID 仅用于查询；Step 3 仍传原 EP。CLI 当前不对 EP 强填模态兜底参数，也不替代服务端最终校验。

## Step 2.5【批量或多阶段任务】提交前检查完整预算

单个图片/视频请求不额外制造“试 Key”任务；批量候选、长视频拆段、续写链等会创建多个收费任务时，
必须在第一个 `+gen` 前列出总候选数、每个候选的阶段数、理论 create 总数以及阶段依赖。能用一个
原生 30 秒任务完成时，不要在模型/Endpoint 未核验前擅自拆成两个 15 秒任务。

对 Agent Plan / Agent Plan Team 的批量或多阶段视觉任务，读取当前 Profile 后执行额度快照：

```bash
arkcli usage plan --format json
arkcli usage balance --type free-quota --modality ComputerVision --page-all --format json
```

- 这两个结果是提交前快照，不是额度预占；只按当前选定 lane/model 的相关桶判断，不把另一个产品的额度混进来。
- 相关月/周/会话桶或模型免费额度已明确耗尽时，不启动只可能完成一半的批次。若用户本轮已明确授权
  某个后付费 Endpoint，则先 `resources resolve` 该 EP 并按它的精确绑定重走 Step 2；否则说明缺口并停止，
  不自动切 Profile、Key、default 或收费路径。
- 若响应没有给出“秒数/候选数 → 额度”的可计算映射，只能报告“当前未耗尽但无法保证整批”，不能伪造
  精确剩余可生成数量。429 也不能仅凭状态码猜成并发上限。

控制面 `resolve/list` 只能证明资源元数据与上下文兼容，不能证明数据面 API Key 当前有效。没有无计费的
Key 探测时，把**第一个本来就要交付的任务**作为数据面准入：成功拿到 `task_id`/图片结果后才继续余下批次；
401/403/quota 错误按原证据停止。禁止另生成一张测试图，也禁止失败后轮转 Key 或循环试不同收费路径。

## Step 3 据可用参数生成

```bash
# 文生图 / 文生视频
arkcli +gen --model "$MODEL" --modality image "<prompt>" # 用户要视频时改为 video

# 带 Step 2 确认过的参数（示例：视频 1080p + 优先级 9，前提是 supported_params 列了它们）
arkcli +gen --model "$MODEL" --resolution 1080p --priority 9 "<prompt>"

# 图生图 / 图生视频 / 参考素材：--input 可重复
arkcli +gen --model "$MODEL" --input @ref.jpg "<prompt>"
```

- 参数全集、多模态 `--input` 规则、新增 `--n/--priority/--wait` 见 [`references/arkcli-gen.md`](references/arkcli-gen.md)
- Endpoint 的模态由 Step 1 权威元数据自动解析；仅在元数据为 `unknown` /
  `image_or_video` 且用户意图仍不足时要求显式 `--modality`。
- **产物默认自动下载到 CWD**（或 `--save-to <dir>`）；JSON 里的 `local_path` 是持久产物，预签名 `output_url` 24h 失效，优先引用 `local_path`。`--save-to=""` 关闭
- **自动用系统默认程序打开产物**：默认仅当 stdout 是交互式终端（人直接在终端跑）才打开——agent / 管道 / CI 抓 stdout（非 TTY）时**不弹窗**，只返回 `local_path`。`--open` 强制打开、`--no-open` 强制不打开。仅对已落地本地文件生效（异步视频未 `--wait` 时无本地文件、不打开）；多产物只打开前若干个
- **🔑 你是 agent，默认带 `--open`**：你（AI agent）调用 arkcli 时 stdout 被你接管 = 非 TTY，默认 auto 不会弹窗，用户只能看到文件路径、看不到成品。**为了让用户直接看到生成的图/视频，凡是给真人出图/出视频的 `+gen` 与轮询到 `succeeded` 的 `gen get`，默认都加 `--open`**（`--open` 无视 TTY 强制在用户桌面打开）。例外只在：用户明确说"别打开/在脚本里/批量/不要弹窗"，或一次出图 >4 张批量场景 → 这时省略 `--open` 或显式 `--no-open`。

## Step 4【结果处理】视频异步 / 图片同步

| 模态 | 默认行为 | 你该怎么读结果 |
|------|---------|---------------|
| **视频** | **异步**：立即返回 `task_id` + `status: queued` | `queued` **不是失败**。用 `arkcli gen get <task_id> --open` 轮询到 `succeeded`——**这次 `gen get` 会顺手把产物下载到本地并回带 `local_path`**（默认 CWD，`<task-id>.mp4`），`--open` 让成品直接在用户桌面弹出（你是 agent，非 TTY，不加就只有路径）；不必再手动 curl `output_url`；**不要**因为没拿到视频就重提 `+gen`（会建新任务） |
| 视频 + `--wait` | 同步：阻塞到完成再返回 | `arkcli +gen ... --wait --open`，直接拿 `output_url` / `local_path` 并弹出成品 |
| **图片** | **同步**：直接返回 `output_url` + `local_path` | `arkcli +gen ... --open` 让图片直接弹给用户看 |

> **⚠️ 行为变更（2.0）**：视频任务默认已从"自动等待完成"改为"提交即返回 task_id"。需要旧的同步阻塞行为，显式加 `--wait`。

### 已有 task 的脚本轮询契约

`gen get --format json` 的 `status` 是对象，终态必须读 `.status.phase`，不是把整个 `.status` 与字符串比较。生成 shell 轮询脚本时必须遵守：

- 轮询阶段用 `arkcli gen get "$TASK_ID" --save-to="" --format json` 禁用自动下载，每轮只读状态。
- `PHASE=$(printf '%s' "$RESULT" | jq -r '.status.phase // empty')`，再对 `succeeded` / `failed` / `cancelled` 做显式分支。
- `succeeded` 时最多再执行一次带目标 `--save-to` 的 `gen get` 下载产物，然后立即 `break`；`failed` / `cancelled` 报告 `status.message` 或 `error` 后立即 `break`。
- `queued` / `running` 才 sleep 后继续；未知 phase 或 `gen get` 自身失败应停止并报错，不能当作 running 无限循环。
- 整个脚本只查已有 task，禁止在轮询或失败分支重新执行 `+gen`。

## 快速决策

- 用户要一步到位出图/视频 → 走本工作流（Step 1→2→3）
- 用户还没定模型 → Step 1 `resources list` 列当前 profile 候选；模型族不确定 → 转 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md)
- 图生图 / 参考素材 → Step 3 加 `--input @<file>`（可重复）
- 视频生成后"没看到视频" → 多半是异步 `queued`，用 `arkcli gen get <task_id> --open` 轮询；轮到 `succeeded` 那次会自动下载到本地（看返回的 `local_path`）并弹出成品，别重提
- **给真人出图/视频默认加 `--open`** → 你是 agent（非 TTY），不加用户只能看到路径、看不到成品；只有"别打开/脚本里/批量 >4 张"才省略或 `--no-open`
- 视频续写 → 先确认 `reference_video` 是服务端可访问 URL；本地 MP4 不能直接作为该 role 提交。ratio 逐值服从精确模型/EP 的 `supported_params`，不无条件强制 `adaptive`

## 进阶 flag 自然语言触发词表

| 用户怎么说 | 对应 flag / 命令 |
|---|---|
| "生成完直接打开/帮我打开看看/出来就弹给我" | `arkcli +gen --open`（强制用系统默认程序打开；默认在交互终端已自动打开） |
| "别自动打开/不要弹窗/我在脚本里跑别开" | `arkcli +gen --no-open`（强制不打开） |
| "预览/别真发/只看参数/dry run/试跑/先看一下" | `arkcli +gen ... --dry-run --format json`；核对 `steps`、`unresolved` 和 `fidelity`，不要把 partial 预览当作服务端校验 |
| "不要下载/只要 URL/不要保存到本地/关闭自动下载" | 命令显式加 `--save-to=""`；即使同时是 `--dry-run` 也要保留，以便预览能核对真实执行时的关闭下载意图 |
| "草稿/快速预览/越快越便宜/省钱先看" | 先区分“只看请求”和“真实生成低成本草稿”。前者用 `--dry-run`；后者仅在 `draft` 支持时加 `--draft=true`，不支持时说明限制，不把缩短时长冒充草稿模式 |
| "固定镜头/镜头不动/锁定相机/只拍光影变化" | 始终保留在 prompt；仅当 `camera_fixed` 支持该值时加 `--camera-fixed=true`。不支持时可用 prompt 表达视觉约束，但不能保证机械锁定，验收跨帧背景/镜头变化 |
| "不带水印/不要水印/关闭水印" | 查明支持后显式 `--watermark=false`；省略可能采用服务端默认值，裸 `--watermark` 表示 true |
| "强制执行/跳过校验/我知道不支持但想试一下" | `arkcli +gen --force` |
| "连贯多张/按顺序/统一风格/连续图片" | 先区分多张独立文件与一张多格图；多文件在能力支持时用 `--image-count N --sequential auto`，不可使用缺值的裸 `--sequential` |
| "我之前的任务/生成历史/任务列表/任务状态" | `arkcli gen list`（列出所有异步生成任务） |
| "那个任务跑完没/查进度/查状态" | `arkcli gen get <task_id>`

## 命令一览

| 命令 | 角色 |
|------|------|
| `arkcli resources list --modality image\|video` | **Step 1** — 当前 profile 可用模型/EP |
| `arkcli resources resolve <endpoint-id>` | **Step 1（显式 EP）** — 权威解析模态、工作流与 region |
| [`arkcli models get <model> --transform supported_params`](../arkcli-models/SKILL.md) | **Step 2** — 查模型可用参数 |
| [`arkcli +gen`](references/arkcli-gen.md) | **Step 3** — 按可用参数生成 |
| [`arkcli +gen --stream`](references/image-stream.md) | 图片任务流式 NDJSON 输出 |
| [`arkcli gen get <task-id>`](references/gen-meta.md) | **Step 4** — 轮询/查询异步视频任务 |
| [`arkcli gen list`](references/gen-meta.md) | 列出/过滤异步生成任务 |
| [`arkcli gen delete <task-id>`](references/gen-meta.md) | 删除异步生成任务 |

## 常见降级

- 模型名报 `not found` → `models get` 已自动归一化点号/display 形态，仍报多半是名字真写错了，用 `arkcli models search <族名>` 核对
- 参数被拒（`param_not_supported`）→ 对照 Step 2 的精确版本目录与实际参数；目录显示支持但 CLI 拒绝时保留冲突证据，不擅自换调用 ID、删用户硬要求或用 `--force` 绕过。只有用户明确要求跳过校验时才用 `--force`。
- **内容被审核拦截**（`ContentRiskBlocked` / `*SensitiveContentDetected` / 命中敏感 / 版权）→ 不是参数问题、`--force` 也绕不过；调整 prompt / 输入素材里的敏感内容后重试。要结构化的拦截原因 + 修复指引，转 [`../arkcli-doctor/SKILL.md`](../arkcli-doctor/SKILL.md) 的 `arkcli doctor error <code>`（生视频拦截 5 个 subtype 全覆盖）
- 鉴权错误 → 转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)

## 参考

- [arkcli-shared](../arkcli-shared/SKILL.md) — 认证和全局参数（必读）
- [arkcli-models](../arkcli-models/SKILL.md) — Step 2 模型查询/`supported_params` 详解
- [references/arkcli-gen.md](references/arkcli-gen.md) — `+gen` 全参数 + 多模态 + 异步语义
