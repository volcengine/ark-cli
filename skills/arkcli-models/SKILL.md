---
name: arkcli-models
version: 1.0.0
description: "arkcli 模型查询能力：列出、搜索、获取火山**公共基础模型**（foundation models）详情。优先使用产品命令 `arkcli models ...`，而不是直接调用 Raw API。注意：查询/管理账号下**自传或精调的自定义模型**（`cm-xxx`）走 arkcli-custommodel；本 skill 只覆盖公共基础模型目录。语音/TTS/ASR/播客/音色/实时语音交互模型只支持广场检索和选型说明，不要引导 +chat/+gen/+deploy/+code-example/usage/pricing/onboard/auth apikey。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli models --help"
---

# arkcli models

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、配置排查与命令选择顺序**
**CRITICAL — 所有 `models` 命令在执行之前，务必先用 Read 工具读取其对应的 reference 文档，禁止直接盲目调用命令。**

## 使用原则

- 模型相关需求优先使用 `arkcli models ...`
- 这些命令虽然是标准 CLI 类型，但实现入口仍然来自 `shortcuts/models/`
- 只有产品命令无法覆盖时，才回退到 [`../arkcli-api-explorer/SKILL.md`](../arkcli-api-explorer/SKILL.md)
- 本 skill 不是默认兜底入口；用户明确问模型查询、模型资产盘点、模型详情或为上游命令挑选模型时才进入
- **语音模型边界**：TTS / ASR / 播客 / 音色 / 实时语音交互等语音模型在 arkcli 中只支持广场检索与选型说明；`models search` 能搜到不代表可 `+deploy`、可 `+code-example`、可查 `usage` 或可查 `pricing`。除非用户另问官方文档，不要主动给控制台 / OpenAPI / SDK 等非 arkcli 接入步骤或链接。

## 适用场景

- 用户要搜索、筛选、对比方舟模型
- 用户要查看模型详情、版本、上下文、模态或 lifecycle 状态
- 用户要统计或列出"我的模型"、"自定义模型"、"最近创建的模型"
- 上游 `+chat` / `+gen` / `+deploy` 需要先确定可用模型名
- 用户问语音模型是否存在、有哪些语音模型、TTS/ASR 模型在广场叫什么：只回答广场可搜事实和当前 arkcli 不支持后续场景能力的边界

## 反唤起信号

- 用户只是要直接对话、生成图片/视频或部署 Endpoint：先转对应 skill，只在缺模型名时回来查模型
- 用户明确要求调用原始 Action、列出 OpenAPI 或构造底层 params：转 `arkcli-api-explorer`
- 用户是鉴权、profile、region、base URL 排障：转 `arkcli-auth` 或 `arkcli-config`
- 不要把 `arkcli models` 当默认入口；非模型问题不要先查模型

## 业务定位

- 本 skill 是其他业务链路的依赖 skill，不是终点 skill
- 它通常服务于三类上游目标：
  - 给 `+chat` / `+gen` 找正确模型
  - 给 `+deploy` 确认可部署模型
  - 给业务排障确认模型详情和版本
- 除非用户明确就是在做模型查询，否则查完模型后应回到原始任务
- **例外**：语音模型查询本身就是终点。查到 `doubao-seed-tts-*`、`doubao-seed-asr-*`、`seedasr-*`、播客、音色设计、实时语音交互等广场语音模型后，停在"可搜到但 arkcli 不支持调用/部署/示例/用量/费用"说明，不继续交给 `+deploy` / `+code-example` / `usage` / `pricing` / `onboard`，也不主动补非 arkcli 接入路径。

## 快速决策

### 第 0 步（硬闸门）：先读场景表，再选命令

**这是任何"按意图找模型"需求的第一动作，先于 `search` / `list` / `resources` 的命令选择。MUST 先用 Read 打开 [`references/arkcli-models-scenario-table.md`](references/arkcli-models-scenario-table.md) 判定意图是否命中某个场景标签，然后才决定跑什么命令。** 这张人工策展的 `场景标签 → 推荐模型` 表是**意图排序的最高权重信号，压过下方一切命名启发式 / modality / capability 过滤**。

**易被截胡的反例（必读）——下面这些"闻起来像硬指标"的意图，本质是场景标签，必须先查场景表，不要本能地跳到 `--capability` / `--modality` 降级线：**

| 用户这么说 | ❌ 别直接 | ✅ 第 0 步命中的场景标签 → 推荐 |
|---|---|---|
| "复杂推理 / 多步骤 / 效果最强的模型" | `search --capability thinking` | 复杂推理 / Agent 任务 → `doubao-seed-2-0-pro` |
| "做图片生成用哪个模型" | `resources list --modality image` 就收手 | 图片生成 → `doubao-seedream-5-0` |
| "做视频 / 角色扮演 / 字段抽取" | `search --modality ...` | 视频生成 / 角色扮演 / 信息抽取 → 查表 |

命中后按表里 JOIN 协议执行：取推荐模型族名词干 → `arkcli models search <族名词干>` 回左表校验事实 → 置顶推荐 + 补 2–3 备选。**表里的版本/full-id 只是起点，命中以 search 实时返回为准**（实测表 id 会陈旧）。

**只有这两种情况才跳过场景表、降级到下方 `search` + 命名启发式：**
- 用户给了**明确的数值/能力硬约束**（"200K 上下文"、"必须支持 functioncall"），且场景表没有对应标签；
- 用户**点名第三方 / 开源 / 历史模型**（qwen、glm、Seedream 4.5…）——不强行替换。

> 注意区分："复杂推理"是场景标签（查表 → pro），**不等于**用户给了 `thinking` 硬指标。前者走第 0 步，后者才走降级线。

**层级边界（重要，别串台）：场景表是"模型广场选型层"，回答"这个意图在广场上选哪个模型"，数据源永远是 `models search`（catalog）。** 它给的是 catalog 推荐模型名（如生图 → `doubao-seedream-5-0`）。

- "当前 profile 能不能调用、确切可调用 id 是什么"是**下游可用性层**的事，由 `resources list` → `models get` → `+gen` / `+chat` 负责（arkcli-gen 已实装），**不在本表职责内**。
- 因此：用户是**纯选型问题**（"查/选哪个模型"）→ 只走 `models search`，**不要**跑去查 `resources list` 把"选型"做成"查可用资源"。只有用户**真的要生成 / 调用**时，才下沉到 resources list 解析本 profile 的真实 id。

这是**意图排序线**专属。枚举 / 盘点 / 统计是另一条线，走 `list`，不挂场景表重排。

### 命令选择

- 用户只知道模糊模型名 / 想按意图找（"最强生视频"、"200K 上下文 LLM"、"支持 thinking 的模型"）：用 `search` —— 它做关键词模糊 + modality/context/capability 结构化过滤
- 需要按 modality / 分页参数枚举、或精确名匹配：用 `list`
- 用户问"我的模型"、"自定义模型"、"最近创建了多少"、"列出来"、"统计数量"：这是模型资产盘点，不是找候选模型；先读 [`references/arkcli-models-list.md`](references/arkcli-models-list.md)，用 `arkcli models list --page-all` 拉取后做客户端过滤，不要跳到 Raw API Explorer
- 已经有明确模型 ID：用 `get`
- 只是为 `+chat` / `+gen` / `+deploy` 找模型：查到后立即回原任务
- 用户**主动**要开通某个基础模型（"先把 doubao-seed-1-6-flash 开通好"、"我想试用 fast-infer 子服务"、"dry-run 看看请求合不合法"）：用 `activate`，先读 [`references/arkcli-models-activate.md`](references/arkcli-models-activate.md)；如果用户只是要 deploy / 创建端点，由 deploy / infer-create 自行触发隐式开通即可，不要先单独 activate

## Agent 快速执行顺序

1. 先确认是否已登录；不确定时先看 `arkcli auth status`
2. 用户描述带意图（modality / 数值容量 / capability）时，用 `arkcli models search` + 对应 flag（`--modality`、`--min-context-window`、`--capability`...）
3. 用户只知道模糊名称时，仍用 `arkcli models search <keyword>`（默认返回全部命中，无分页）
4. 需要按 modality 全量枚举或翻页统计时，用 `arkcli models list`
5. 用户问"最近 N 天创建的自定义模型/我的模型"时，用 `arkcli models list --page-all --sort-by CreateTime --sort-order Desc --format json`，再在本地按 `create_time`、`model_type` / `customization_type` / `source_type` 等字段过滤；不要因为没有时间过滤 flag 就改探 `arkcli api --list`
6. 已有明确模型 ID，需要详情时，用 `arkcli models get`
7. 用户明确要开通模型（主动激活，不是为了立刻 deploy）时，用 `arkcli models activate <name> [--sub-services ...]`；CI 场景加 `--yes`，校验场景加 `--dry-run`
8. 查到 / 激活完模型后，回到发起它的上游链路：`+chat` / `+gen` / `+deploy`

## 典型业务链路

### 0. 为数据面命令补全完整模型 ID（必做前置）

`+chat` / `+gen` 的 `--model` 必须是 `<name>-<primary_version>` 完整形式（或 Endpoint ID `ep-xxx`）。**直接传族名会触发 `InvalidEndpointOrModel.NotFound` 404**。

**`primary_version` 格式不固定**，**不要用正则自行判断"看起来像完整 ID"**。实际观察到的格式分布（~152 个模型中约半数非 6 位）：

| 格式 | 样例 full ID | 常见家族 |
|------|--------------|----------|
| `YYMMDD`（6 位日期） | `doubao-seed-1-6-251015` | doubao 系 |
| `YYYYMMDD`（8 位日期） | `qwen3-14b-20250429`、`glm-4-6-20250930` | qwen / glm 开源 |
| 限定前缀 + 日期 | `doubao-seed-2-0-code-preview-260215`、`doubao-seed-1-6-nano-unfiltered-250928` | 灰度 / 变体 |
| 纯路由字符串 | `kimi-k2-6-modelhub`、`open-source-models-default` | 外部接入 |
| 短数字 | `qwen3-235b-a22b-instruct-2507` | qwen 部分 |
| 空串 | `doubao-seed-tts-2-0`（full ID 就等于族名） | TTS / ASR 部分 |

**两条获取路径，按效率递减：**

- **路径 A（高效，复用已有结果）**：刚执行过 `arkcli models search <keyword>` 或 `arkcli models list` 时，返回 JSON 每个 item 自带 `primary_version` 字段，Agent 直接读取并拼接 `<name>-<primary_version>`，无需额外调用。**`search` 还会同步返回 `context_window` / `input_modalities` / `output_modalities` / `capabilities` 等结构化字段（来自 ArkModels enrich），下游可以直接基于此判断模型是否合适，省去再调 `get` 的成本**。
- **路径 B（单独查询）**：只知道裸名时，`arkcli models get <name> --transform 'primary_version'` 直接返回版本号

**注意**：`--transform` 输出带 JSON 双引号（如 `"251228"`），shell 拼接前必须 `tr -d '"'` 剥掉：

```bash
VER=$(arkcli models get doubao-seed-1-8 --transform 'primary_version' | tr -d '"')
# VER=251228  -> full ID: doubao-seed-1-8-251228
# VER=""（空串模型）-> 直接用族名: doubao-seed-tts-2-0
if [ -n "$VER" ]; then
  MODEL="doubao-seed-1-8-$VER"
else
  MODEL="doubao-seed-1-8"
fi
arkcli +chat --model "$MODEL" "你好"
```

### 1. 为试用找模型

`models search [--modality / --capability ...]` -> `+chat` / `+gen`

### 2. 为正式接入找模型

`models search [--min-context-window / --capability ...]` -> `+deploy`

### 3. 按硬指标筛选

```bash
# 找最新最强的 200K+ 文本模型，含 thinking 能力
arkcli models search --modality text --min-context-window 200000 --capability thinking --strict-filter

# 找视频生成模型
arkcli models search --modality video --strict-filter

# 找多模态 LLM
arkcli models search --multimodal --output-modality text --strict-filter
```

`--strict-filter` 强烈建议带上：默认行为是"缺数据保留"（避免误杀），加 strict 后只返回 100% 满足条件的模型。

## 常见降级

- 认证失败：转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
- profile / region 看起来不对：转 [`../arkcli-config/SKILL.md`](../arkcli-config/SKILL.md)
- 业务上只是为了给 `+chat`、`+gen`、`+deploy` 找模型名，应优先完成"查模型后回到原任务"，不要停留在 models 本身

## 命令一览

| 命令 | 说明 |
|------|------|
| `arkcli models search [keyword] [filters]` | **Agent 首选**：全量召回 + ArkModels enrich + modality/context/capability 结构化过滤 + 重排；返回字段含 `context_window` / `input_modalities` / `output_modalities` / `capabilities` |
| `arkcli models list` | 按 modality 全量枚举、翻页统计、模型资产盘点；轻量，不含 enrich |
| `arkcli models get <id> [version]` | 单个模型完整详情（聚合多 API，最重也最全）|

## 命名约定的 tier 启发式（降级兜底：仅在场景表未命中时启用）

**优先级口诀：场景表命中(★★★) > 命名启发式(★★，本节) > update_time。** 先走「快速决策 第 0 步」的场景表；只有场景表没覆盖该意图、或用户点名第三方/开源/历史模型时，才用本节的命名启发式排序。

`search` 加 filter 之后通常还会剩多个候选，命名里有规律的 tier 信号可以帮 agent 做最后一步排序。**这是启发式，不是硬规则**：硬指标（context_window / capabilities / modality）永远优先于命名。

### 代次（major version）— 优先级最高

```
2-x > 1-8 > 1-6 > 1-5 > ...
```

代次跳跃通常**强于 tier**。例：`doubao-seed-2-0-mini` 在多数任务上强于 `doubao-seed-1-6-pro`，因为基座模型代次差距大于尺寸档差距。

### tier（同代内的尺寸档）

```
pro  ≥  无后缀  >  lite  >  flash  >  mini  >  nano
```

- `pro`：旗舰，最大尺寸 / 最强能力
- 无后缀（如 `doubao-seed-1-8`）：主力档，通常 ≈ pro
- `lite`：成本/性能均衡
- `flash`：速度优化
- `mini` / `nano`：低延迟、高并发、最低成本

### specialty 后缀（不参与 tier 比较，按用途选）

```
-code        → 编程优化（Doubao-Seed-Code 等）
-thinking    → 思考能力强化（更长 reasoning）
-vision      → 视觉理解强化
-character   → 角色扮演 / 长旁白
-translation → 翻译专用
-tts / -voice → 语音合成
-asr         → 语音识别
```

specialty 模型在它的领域内**通常强于同代通用 pro**，但跨任务时不可移植。

### `primary_version` 比较的注意事项

- **同一 family 内**：日期越新越好（`260215 > 251228 > 251015`）
- **不同 family 之间**：**不可数值比较**。`qwen3-14b-20250429`（8 位）和 `doubao-seed-1-6-251015`（6 位）数值上前者大，但和"哪个强"无关
- 比较"哪个更新"统一用 `update_time` 字段（ISO 8601，可字典序比较），不要去 parse `primary_version`

### 决策口诀

```
filter (硬指标，必须满足)
   ↓
代次 (2-x > 1-x，跨代差距通常 > 同代 tier 差距)
   ↓
tier (pro > 无后缀 > lite > flash > mini > nano)
   ↓
update_time (同代同 tier 时，新者优先)
```

### 排序已经在 `search` 命令里实装

不用 agent 再做客户端 reorder —— `arkcli models search` 按"先 family 分组、后组内细排"的两阶段排序：

```
1. bucket             — keyword 可见性 (name 命中 > desc 命中 > 兜底 > hidden)
   ↓
2. 跨 family：       按 family 的"代表分"排（family 成员中的 max ctx → max update_time → family 名）
   ↓                  → 同 family 的所有成员保持在一起
3. 同 family 内部：  gen DESC → tier DESC → ctx DESC → update_time DESC → name ASC
```

**关键设计**：
- **family 分组先于细排**。doubao-seed 家族（max ctx=262144）整体排在 doubao-seed-character（max ctx=131072）和所有 ctx=null 家族之前。这避免了"同家族的 1-6-vision 比 1-8 还靠前"这种违反传递性的奇怪结果。
- **同 family 内 gen 优先于 ctx**。新版本即使 ArkModels 元数据还没补齐（`context_window=null`），也会浮在已有 ctx 数据的旧版本之前。例：`glm-5-1`（gen=501, ctx=null）排在 `glm-4-6`（gen=406, ctx=262144）之前；`kimi-k2-6`（gen=600, ctx=null）排在 `kimi-k2-5`（gen=500, ctx=262144）之前。
- **代次优先于 tier**。同 family `doubao-seed-2-0-mini`（gen=200, tier=50）排在 `doubao-seed-1-8`（gen=108, tier=80）之前 —— 跨代差距通常大于同代 tier 差距。
- **跨 family 不比 gen/tier**。pro/lite/mini 是 family 内部命名，跨 family 没有可比性 —— 由 family 代表分（max ctx + max time）决定家族先后。

被打了 `体验隐藏` / `推理隐藏` / `广场隐藏` 任一标签的旧模型自动沉底（不会被 list/search 屏蔽，但永远在结果末尾）。

> **说明**：`体验隐藏` / `推理隐藏` / `广场隐藏` 是火山方舟模型平台 `customized_tags` 中的旧版页面隐藏标签，与本 CLI 命令无关，arkcli 仅读取它们用于排序。

## 模型生命周期：Shutdown / Retiring / Published

ArkModels 给每个模型打 `lifecycle_status`，三种值，Search 处理方式不同：

| status | 含义 | search 默认行为 |
|--------|------|----------------|
| `Published` | 正常服务 | ✓ 显示 |
| `Retiring` | 正在下线（仍可调用，不建议新接入）| ✓ 显示，agent 应口头告诉用户 "X 正在下线，建议换 Y" |
| `Shutdown` | 已下线（调用必失败）| **❌ 默认过滤掉**（加 `--include-deprecated` 才回来）|

另外，`display_name` 含 `废弃` / `下线` / `已下架` / `deprecated` 关键词的模型也按 Shutdown 处理（兜底，因为有些遗留模型不在 ArkModels 元数据里，靠人工标记）。

**Agent 行为约定**：
- 不需要做"过滤掉 Shutdown"的客户端逻辑 —— search 默认已经做了
- 看到 `lifecycle_status="Retiring"` 的模型时，**主动提醒用户它正在下线**，并尝试在同 family 里找一个更新版本（用 `search` + 正确 keyword 即可）

## ❌ 反模式（agent 必读）

下面这些行为是错的 —— 命令选错会让 agent 拿不到 enrich 数据（context_window / modalities / capabilities）、得不到 hidden 沉底 / 加权重排，进而给出错误推荐。

- ❌ **不要用 `list` 找模型** —— 找模型（任何"哪个模型/找一个 X 模型"意图）一律走 `search`。`list` 不做关键词模糊、不做加权重排、不返回 enrich 字段。
- ❌ **不要用 `list --modality video` 选生视频模型** —— 用 `arkcli models search --modality video --strict-filter`，它结合 ArkModels 的 `output_modalities` 和 `task_types` 兜底，召回更准且能进一步组合 `--min-context-window` / `--capability` 过滤。
- ❌ **不要用 `list --modality text` 选 LLM** —— 同上，用 `search --modality text --strict-filter`。`list --modality text` 只看 foundation_model_tag.filter_domains 一层信号，会漏掉很多模型。
- ❌ **不要先 `list` 再 `get` 来验证 context_window 等参数** —— `search` 已经在结果里直接返回 `context_window` / `max_input_tokens` / `max_completion_tokens`，省掉一次 `get` 调用。
- ❌ **不要为了"看 5 条最热门"而 `search` 不传 keyword** —— 现在不传 keyword 是返回**全量 152 条**按 UpdateTime 降序，不再是策展热门。需要少量结果用 `--size 5`。
- ❌ **不要把 `lifecycle_status="Retiring"` 的模型推荐给用户做新接入** —— 这些虽然还能调，但 vendor 已经标记下线倒计时。看到 Retiring 候选时主动提示并搜更新版本。
- ❌ **不要在 `search` 上做客户端二次过滤来弥补 list 的不足** —— 直接用 `search` 自带的 `--modality` / `--min-context-window` / `--capability` flag。
- ❌ **不要为单个 model 信息直接 `get` 而不试 `search`** —— `search <name>` 一次返回所有候选 + enrich；只有需要计费/限流/详细能力描述时才用 `get`。
- ❌ **不要因为 `models list` 没有服务端时间过滤 flag 就去 `arkcli api --list` 探 Raw API** —— 先 `models list --page-all --format json`，再用本地 JSON 处理按 `create_time` 过滤。
- ❌ **不要告诉用户"CLI 没有此能力，请去控制台"**，除非已经确认 `arkcli models list --help` 当前版本确实没有可枚举输出，且本地 JSON 过滤也无法完成用户要的统计。

合法的 `list` 唯一用途：
- ✓ 按 `--modality` 做**全量穷举/审计**（不是为了"找最强"）
- ✓ 需要 `total_count` 这种统计
- ✓ `--name foo` 精确匹配（agent 几乎用不到，因为 search 也能命中）
- ✓ 盘点"我的/自定义/最近创建"这类资产清单：用 `--page-all` 全量拉取，再按字段和时间窗口做客户端过滤

## 参考

- [arkcli-chat](../arkcli-chat/SKILL.md) -- 查到模型后进入对话 / 推理链路
- [arkcli-gen](../arkcli-gen/SKILL.md) -- 查到图片/视频模型后进入生成链路
- [arkcli-deploy](../arkcli-deploy/SKILL.md) -- 查到模型后进入 Endpoint 部署链路
- [`references/arkcli-models-list.md`](references/arkcli-models-list.md)
- [`references/arkcli-models-get.md`](references/arkcli-models-get.md)
- [`references/arkcli-models-search.md`](references/arkcli-models-search.md)
- [`references/arkcli-models-scenario-table.md`](references/arkcli-models-scenario-table.md) -- 场景化推荐表（意图排序最高权重）+ JOIN 验证协议
- [`references/arkcli-models-activate.md`](references/arkcli-models-activate.md)
