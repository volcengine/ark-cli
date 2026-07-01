---
name: arkcli-gen
version: 2.0.0
description: "火山方舟 Ark 图片/视频生成入口：用户要生图、画图、生成图片/视频、图生图、图生视频、参考图/视频/音频生成，或明确使用 seedream/seedance 创作新内容时使用 arkcli +gen。+gen 按当前 profile 的可用资源与模型 supported_params 生成；图片同步返回，视频提交后返回 task_id/status，需用 --wait 或 arkcli gen get/list 轮询并下载结果。反触发：用户问失败原因、失败率、模型健康、持续慢/超时/限流、配额是否够、内容审核/敏感内容/PolicyViolation 拦截怎么解决时，走 arkcli-doctor 的 doctor model 或 doctor error，不走本 skill。单条视频 queued/还在跑通常是异步任务未完成，仍用 gen get/list 查结果；跨任务统计性失败或持续异常才转 doctor。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli +gen --help"
---

# arkcli 生成工作流（+gen）

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)（认证闸门、模型查找回退、共享安全规则）。**

**CRITICAL — 这是一段三步工作流，不是单条命令。生成图/视频 MUST 按 `Step 1 → Step 2 → Step 3` 顺序执行。禁止跳过 Step 1/2 直接 `+gen`：会因模型名形态不对（404）或传了模型不支持的参数而失败。执行前务必读 [`references/arkcli-gen.md`](references/arkcli-gen.md)。**

## 为什么是工作流（核心，先理解再执行）

用户说"生成一个视频/一张图"，本质是**三件独立的事，必须按序**：

```
① 当前 profile 能用哪些模型   ── 不同 profile 模型来源完全不同(EP vs 模型名)
② 该模型支持哪些参数          ── 不查就传参 = 瞎猜 = 被校验拒/被后端拒
③ 按可用参数真去生成
```

把这三步压成"直接 `+gen` 猜一条命令"，正是失败之源：模型名形态不对会 404，参数模型不支持会被拒。

## 适用场景

- "生成一张图" / "文生图" / "画一个 X"
- "生成一个视频" / "文生视频"
- 图生图 / image-edit / 加参考图；图生视频(I2V)；参考视频(R2V)；参考音频
- "用这张图当首帧生成视频" / "保持这个参考视频的运动"

## 工作流总览

```
用户意图: "生成 X"
  │
  ▼ Step 1【强制】列当前 profile 可用模型  ──► arkcli resources list --modality image|video
  │     platform    → 列 EP (ep-xxx)           ┐
  │     agent-plan  → 列视觉模型名              ├─ 选一个，记为 $MODEL
  │     coding-plan → 列 EP (借道 platform)     ┘
  │
  ▼ Step 2【强制·EP 除外】查 $MODEL 可用参数  ──► arkcli models get $MODEL --transform supported_params
  │     模型名 + 有 sp → **只能**用列出的参数，取值落 min/max/enum 内
  │     模型名 + sp 空(没配,如 2.0/1.5-pro) → +gen 自动套 modality 兜底默认(video 720p/5s, image 2048)
  │     EP(ep-xxx)            → 跳过, 不强填(背后能力未知), 服务端裁决
  │
  ▼ Step 3 据可用参数生成  ──► arkcli +gen --model $MODEL [Step2 允许的参数] "prompt"
  │
  ▼ Step 4【结果处理】
        视频 = 异步：返回 task_id + status=queued(**不是失败!**) → arkcli gen get <task_id> 轮询;轮到 succeeded 自动下载到本地(local_path);要同步阻塞加 --wait
        图片 = 同步：直接返回 output_url + local_path
```

## Step 1【强制】列出当前 profile 可用的模型

```bash
# 按目标模态列；输出 items[].id 就是可作 --model 的候选
arkcli resources list --modality video   # 或 image
```

- **平台差异（resources list 已自动按 profile 分流，你只管读 items）**：
  - `platform` profile → items 是**推理接入点 EP**（`ep-xxx`），每个 EP 内部绑定一个模型
  - `agent-plan` profile → items 是**视觉模型名**（如 `doubao-seedance-2.0-fast`）
  - `coding-plan` profile → 自身不含视觉模型，`+gen image/video` 会自动借道 platform 数据面，**`--model` 必须显式传一个 platform 上的 EP**
- `is_default: true` 标记的是该模态当前默认；用户没指定时优先用它
- **选定一个 id，记为 `$MODEL`，贯穿 Step 2/3**
- 用户已明确给了模型/EP 时，仍建议 `resources list` 核对它在当前 profile 可用；若与默认不同，按 [`../arkcli-shared/references/profile-defaults.md`](../arkcli-shared/references/profile-defaults.md) "Default 漂移检测与 promote nudge" 处理

## Step 2【强制·EP 除外】查 $MODEL 的可用参数

```bash
arkcli models get "$MODEL" --transform supported_params
```

- **`$MODEL` 是模型名**：拿到该模型的 `supported_params` 清单（每项含 `name / type / support / min / max / enum / required`）。
  - > **MUST：Step 3 只能使用这里 `support=true` 的参数，且取值必须落在 `min/max/enum` 范围内。** 不在清单里的参数（或 `support=false`）传了会被 `+gen` 拒绝。
  - **可直接用 Step 1 `resources list` 给的 id**（点号 / display 形态如 `doubao-seedance-2.0-fast` 都行）：`models get` 会自动按 DisplayName 归一化到规范连字符 name，无需手动转。极个别仍报 `not found` 才用 `arkcli models search <族名>` 核对名字。
  - 查到模型但 `supported_params` 为空 / `null`（很多模型没配，如 `doubao-seedance-2-0` / `doubao-seedance-1-5-pro`）→ **不用手动补参数**：`+gen` 会自动用内置 modality 兜底默认（video: `resolution=720p` / `duration=5` / `ratio=adaptive`；image: `size=2048x2048`）填充你没指定的参数。直接进 Step 3。
- **`$MODEL` 是 EP（`ep-xxx`）**：跳过本步。EP 查不到 supported_params 是正常的；且 `+gen` **不会**对 EP 套兜底默认（EP 背后模型可能支持更高能力，强填会误降级），直接 degrade-open 由服务端裁决。

## Step 3 据可用参数生成

```bash
# 文生图 / 文生视频
arkcli +gen --model "$MODEL" "<prompt>"

# 带 Step 2 确认过的参数（示例：视频 1080p + 优先级 9，前提是 supported_params 列了它们）
arkcli +gen --model "$MODEL" --resolution 1080p --priority 9 "<prompt>"

# 图生图 / 图生视频 / 参考素材：--input 可重复
arkcli +gen --model "$MODEL" --input @ref.jpg "<prompt>"
```

- 参数全集、多模态 `--input` 规则、新增 `--n/--priority/--wait` 见 [`references/arkcli-gen.md`](references/arkcli-gen.md)
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

## 快速决策

- 用户要一步到位出图/视频 → 走本工作流（Step 1→2→3）
- 用户还没定模型 → Step 1 `resources list` 列候选；模型族不确定 → 转 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md)
- 图生图 / 参考素材 → Step 3 加 `--input @<file>`（可重复）
- 视频生成后"没看到视频" → 多半是异步 `queued`，用 `arkcli gen get <task_id> --open` 轮询；轮到 `succeeded` 那次会自动下载到本地（看返回的 `local_path`）并弹出成品，别重提
- **给真人出图/视频默认加 `--open`** → 你是 agent（非 TTY），不加用户只能看到路径、看不到成品；只有"别打开/脚本里/批量 >4 张"才省略或 `--no-open`

## 进阶 flag 自然语言触发词表

| 用户怎么说 | 对应 flag / 命令 |
|---|---|
| "生成完直接打开/帮我打开看看/出来就弹给我" | `arkcli +gen --open`（强制用系统默认程序打开；默认在交互终端已自动打开） |
| "别自动打开/不要弹窗/我在脚本里跑别开" | `arkcli +gen --no-open`（强制不打开） |
| "预览/别真发/只看参数/dry run/试跑/先看一下" | `arkcli +gen --dry-run`（不会真正生成） |
| "强制执行/跳过校验/我知道不支持但想试一下" | `arkcli +gen --force` |
| "连贯多张/按顺序/统一风格/4格漫画/连续图片" | `arkcli +gen --sequential` |
| "我之前的任务/生成历史/任务列表/任务状态" | `arkcli gen list`（列出所有异步生成任务） |
| "那个任务跑完没/查进度/查状态" | `arkcli gen get <task_id>`

## 命令一览

| 命令 | 角色 |
|------|------|
| `arkcli resources list --modality image\|video` | **Step 1** — 当前 profile 可用模型/EP |
| [`arkcli models get <model> --transform supported_params`](../arkcli-models/SKILL.md) | **Step 2** — 查模型可用参数 |
| [`arkcli +gen`](references/arkcli-gen.md) | **Step 3** — 按可用参数生成 |
| [`arkcli +gen --stream`](references/image-stream.md) | 图片任务流式 NDJSON 输出 |
| [`arkcli gen get <task-id>`](references/gen-meta.md) | **Step 4** — 轮询/查询异步视频任务 |
| [`arkcli gen list`](references/gen-meta.md) | 列出/过滤异步生成任务 |
| [`arkcli gen delete <task-id>`](references/gen-meta.md) | 删除异步生成任务 |

## 常见降级

- 模型名报 `not found` → `models get` 已自动归一化点号/display 形态，仍报多半是名字真写错了，用 `arkcli models search <族名>` 核对
- 参数被拒（`param_not_supported`）→ 回到 Step 2 看 `supported_params`，只用列出的；确需强制可加 `+gen --force` 跳过校验（服务端仍有最终裁决）
- **内容被审核拦截**（`ContentRiskBlocked` / `*SensitiveContentDetected` / 命中敏感 / 版权）→ 不是参数问题、`--force` 也绕不过；调整 prompt / 输入素材里的敏感内容后重试。要结构化的拦截原因 + 修复指引，转 [`../arkcli-doctor/SKILL.md`](../arkcli-doctor/SKILL.md) 的 `arkcli doctor error <code>`（生视频拦截 5 个 subtype 全覆盖）
- 鉴权错误 → 转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)

## 参考

- [arkcli-shared](../arkcli-shared/SKILL.md) — 认证和全局参数（必读）
- [arkcli-models](../arkcli-models/SKILL.md) — Step 2 模型查询/`supported_params` 详解
- [references/arkcli-gen.md](references/arkcli-gen.md) — `+gen` 全参数 + 多模态 + 异步语义
