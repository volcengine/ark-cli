# models search

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

**Agent 优先**的模型搜索：召回 = 全量目录；enrich = ArkModels（context window / 模态 / capability）；filter = 结构化条件；rank = 4 桶 + context_window 加权 + 时序。**没有分页概念**，默认返回全部命中。

只查询公共基础模型目录；Plan 支持清单用 `plans model-list --plan <plan>`，账号自定义资产用 `custommodel list/get`，当前身份可调用资源用 `resources list`。无命中先判资源类型，不跨 Profile 穷举，也不把当前公共目录未找到说成账号内不存在。

## 空结果、规范化与证据冲突

1. 用户给精确模型 ID/显式版本时先 `models get` 原值；不要把完整版本串当搜索关键词。模型名中的空格、点号或展示形式可作为检索线索规范化为连字符，但真正用于代码/部署的 ID 必须来自本轮 `name` 与版本字段，不凭字符串替换制造新 ID。
2. 模糊查询用带引号的家族/名称关键词；先校验 `--capability` 的真实枚举，缓存条件使用 `--cache-type`。成功的 `items: []` 仅代表当前组合无命中；非零退出、缺 `.items`、JSON 损坏和截断都不是空结果。
3. 在同一身份、资源范围内，先修正关键词；仍为空时可逐次只放宽一个**偏好**条件并说明变化，保留用户硬约束。放宽硬条件、身份、Project 或费用范围必须先询问。规范化后同一关键词加一次 `--include-deprecated` 可判断是否被下线过滤；不能在多个家族间盲猜。
4. `--modality text` 表示文本**输出**，不是文本输入。图像/视频任务用对应输出模态或 `--input-modality text`；放宽后只剩 embedding/router/非目标任务的模型时，报告没有合适候选，不能把它们包装成任务推荐。
5. 仅命中其他规格/变体时，分别报告“原请求未命中”和“当前候选”，不得静默替代。精确 get 报 NotFound 后，可按这一流程查候选，再用真实候选校正一次 get；仍失败则列尝试值与范围，保留未知，不循环。
6. 只用本次输出陈述上下文、模态、生命周期和能力。缺失/null 是未知，不是 false。特定 API、参数支持或精确版本需读 [get](arkcli-models-get.md)；search 与 get 不一致且版本/范围未对齐时先核对，已明确同一 Name/Version/范围则直接保留冲突，不重复要求对齐或重查。分别标明来源，不取并集拼成新结论。`supported_params` 是正向证据，未列字段的运行时行为须查对应数据面契约或在授权后验证。
7. 工具预览截断时先读其保存的完整文件；没有完整产物才把**同一只读查询**重新保存到私有临时目录。stdout/stderr 分开、检查退出码，再由 jq 提取字段；不吞 stderr，不逐行截断 JSON，也不因捕获失败直接终止可恢复的查询。

## 未定模型时的实时候选模板

开通/部署前的候选只来自本轮结果。已有完整结果就本地投影，不重查；初次可取 10 项留出退役过滤余量。下面命令按共享协议补全调用归因环境：

```bash
candidate_dir=$(mktemp -d)
arkcli models search --size 10 --format json > "$candidate_dir/models.json" 2> "$candidate_dir/models.stderr"
# 先检查上条退出码和 stderr；仅成功且 JSON 完整时继续
jq -e '.items | type == "array"' "$candidate_dir/models.json"
jq '[.items[]
  | select(.lifecycle_status != "Shutdown" and .lifecycle_status != "Retiring")
  | {display_name,name,primary_version,lifecycle_status,input_modalities,output_modalities,task_types,create_time,update_time}]
  | sort_by([(.create_time != null), (.create_time // "")]) | reverse' "$candidate_dir/models.json"
```

模糊家族场景在 `search` 中加用户关键词。候选已知状态/关键字段缺失时标未知，不能把模板的过滤结果说成全都 Published 或可调用；需要新接入时仅推荐已证实未退役且满足目标任务者。

无具体偏好时展示 5–8 个有证据的代表，优先保留不同文本、多模态输入、图像、视频任务选项；不是强凑名额：不足就如实列实际数，指定任务时不混入无关模态。`label` 保留真实 `display_name`，重复名在说明中附 `name/version` 区分，不改版本、不额外造“其他”选项。开通/部署候选默认按真实 `create_time` 倒序；用户明确要求最近更新时才用 `update_time`，缺时间标未知，不猜时间或用 context 排序冒充创建时间排序。候选选择回到 owning Skill；选模型不等于批准开通或创建。

## 命令

```bash
# 基础
arkcli models search                       # 全量按 UpdateTime 降序
arkcli models search doubao                # 关键词（也可 --keyword doubao）
arkcli models search doubao --size 5       # 显式截断到 5 条

# 模态过滤（D3 必备）
arkcli models search --modality video      # 视频生成模型
arkcli models search --modality text       # 文本输出模型
arkcli models search --modality image      # 图片生成模型
arkcli models search --input-modality text,image --output-modality text  # VQA 类
arkcli models search --multimodal          # 输入支持多模态的模型

# 数值容量过滤
arkcli models search --min-context-window 200000
arkcli models search --max-context-window 1000000
arkcli models search --min-input-tokens 100000
arkcli models search --min-output-tokens 8192

# 能力过滤（可重复，AND 关系）
arkcli models search --capability thinking
arkcli models search --capability mcp --capability functioncall

# 缓存类别过滤（可重复，AND 关系）
arkcli models search --cache-type implicit_cache
arkcli models search --cache-type session_cache --cache-type prefix_cache

# 复合查询：找最强 200K+ 思考型 LLM
arkcli models search --modality text --min-context-window 200000 --capability thinking --strict-filter

# 缓存控制
arkcli models search --refresh-cache       # 强制同步刷新一次 ArkModels 元数据
```

## 参数

| 参数 | 类型 | 说明 |
|------|------|------|
| `[keyword]` / `--keyword` | string | 关键词。匹配 name / display_name / short_name / description / introduction 任一字段（小写不敏感）|
| `--size` | int | **不传 = 不截断**（agent 默认拿全量）；显式传 N 则截断到 N |
| `--modality` | string | 粗粒度别名：`text` / `image` / `video` / `audio`（输出模态简写）|
| `--input-modality` | string | 显式输入模态过滤，逗号分隔（如 `text,image,video`）|
| `--output-modality` | string | 显式输出模态过滤 |
| `--multimodal` | bool | 仅返回输入模态数 ≥ 2 的模型 |
| `--min-context-window` | int | 最小 context_window（tokens）|
| `--max-context-window` | int | 最大 context_window |
| `--min-input-tokens` | int | 最小 max_input_tokens |
| `--min-output-tokens` | int | 最小 max_completion_tokens |
| `--capability` | string（可重复）| 必备能力，多个为 AND：`thinking` / `mcp` / `functioncall` / `web_browsing` / `knowledge_base` / `response_format` / `reasoning_effort` |
| `--cache-type` | string（可重复）| 必备缓存类别，多个为 AND：`explicit_cache` / `implicit_cache` / `session_cache` / `prefix_cache` |
| `--strict-filter` | bool | 缺 enrich 数据的模型直接排除（默认保留，避免误杀）|
| `--include-deprecated` | bool | 包含 `lifecycle_status=Shutdown` 或 `display_name` 含"废弃/下线"的模型（**默认过滤**，只保留可调用模型）|
| `--refresh-cache` | bool | 同步刷新 ArkModels 元数据 cache（默认 stale-while-revalidate 异步刷新）|

## 工作流程

```
search [keyword] [filters]
   │
   1️⃣  ListFoundationModels(全量, sort=UpdateTime DESC)
   │   ↓ 152 个候选
   2️⃣  EnrichWithMetadata
   │   ├─ 读 cache: ~/.arkcli/cache/<profile>/<region>/<project>/arkmodels-meta.json
   │   ├─ hit  → 立即返回 + fork detached 子进程异步刷新
   │   └─ miss → 同步 ArkModels(IsPrimaryVersionOnly:true) → 写 cache
   │       (无 SSO 时 enrich 静默跳过，filter 自动 no-op)
   │   ↓ 各 item 补 context_window / input_modalities / output_modalities / capabilities
   3️⃣  Modality 兜底：output_modalities 为空时按 task_types 推导
   │   (TextToVideo → out=[video], VisualQA → in=[text,image] out=[text], ...)
   4️⃣  关键词过滤（小写子串，匹配 name+display+short+description+intro）
   5️⃣  结构化过滤（modality / context / tokens / capability / cache_type，AND）
   6️⃣  4 桶重排 + (context_window desc, update_time desc, name asc) tie-break
   7️⃣  --size N 截断（默认不截断）
```

## 返回字段

每条 item：

```json
{
  "name": "doubao-seed-2-0-pro",
  "display_name": "Doubao-Seed-2.0-Pro",
  "primary_version": "260215",
  "update_time": "2026-03-02T...",
  "foundation_model_tag": { "filter_task_types": [...], "task_types": [...] },

  // —— enrichment（来自 ArkModels；SSO 失败时为 null）——
  "context_window": 262144,
  "max_input_tokens": 229376,
  "max_completion_tokens": 65536,
  "input_modalities": ["text", "image", "video"],
  "output_modalities": ["text"],
  "capabilities": {
    "thinking": true, "mcp": true, "functioncall": true,
    "web_browsing": true, "knowledge_base": true,
    "response_format": false, "reasoning_effort": true
  },
  "cache_types": ["implicit_cache", "session_cache", "prefix_cache"]
}
```

`capabilities.caching` 是旧版兼容字段，只表示 `caching.support`（即
`explicit_cache`），不要再用于缓存能力判断。模型支持哪些缓存机制以
`cache_types` 为准。`+chat --caching` 是发请求时的运行参数，与这里的模型能力字段无关。

## 重要行为

### 缺数据怎么办（StrictFilter）
- **默认 `--strict-filter=false`**：`--modality video` 时，**没有** modality 数据的模型也会**保留**（agent 看到候选但置信度低）
- **`--strict-filter`**：缺数据视为不满足，直接排除（agent 拿到的 100% 是 ground truth）

模态查询通常加 `--strict-filter` 更准；数值查询（context window）保持默认（缺数据不杀）通常更合理。

`--cache-type` 是精确过滤：只有 `cache_types` 明确包含所请求类别的模型才会返回；缺少缓存元数据的模型不会作为候选保留。

### 召回不再有 top-K 限制
不像旧版 fuzzy 接口默认 9 条，本命令默认返回**全部命中**。如需限制，传 `--size N`。

### Cache 与 SSO
- ArkModels 是 Console BFF 接口，需要 SSO 凭证
- 无 SSO（AK/SK only）→ enrich 静默跳过，模型仍能返回，filter 自动 no-op
- Cache 路径按 `(identity/account, profile, region, project)` 分；缺少权威 identity 时不启用持久缓存，避免跨账号复用
- 默认 TTL 为 5 分钟，可通过 `ARKCLI_MODEL_CACHE_TTL` 调整；TTL 内只读本地缓存
- TTL 过期后 stale-while-revalidate：当前调用先返回旧缓存，并由 detached 子进程刷新（`arkcli models _refresh-cache`）；同 scope 冷启动/刷新使用跨进程 single-flight
- `auth logout` 与 fresh 账号切换会清除当前产品的全部模型缓存
- 所有 ArkModels Action 都经过 transport 级账号隔离节流，最小间隔 250ms（约 4 QPS，为后端 5 QPS 留余量）

### 关键词匹配范围
- name / display_name / short_name（命中 → 重排第 1 桶）
- description / display_description / introduction（命中 → 第 2 桶）
- 子串匹配，**小写不敏感**
- **不做语义/同义词扩展**

### 重排细节
| 桶 | 条件 |
|---|------|
| 1 | 非 hidden + name 命中 keyword |
| 2 | 非 hidden + description 命中 keyword |
| 3 | 非 hidden + 无 keyword 命中（兜底）|
| 4 | hidden（`customized_tags` 含 `体验隐藏`/`推理隐藏`/`广场隐藏`，是火山方舟平台旧版页面隐藏标签，与本 CLI 命令无关）|

桶内：`context_window` 大者优先 → `update_time` 新者优先 → `name` 字典序

## 常见错误

| 错误 | 原因 | 处理 |
|------|------|------|
| `--modality video` 返回大量噪声 | 默认 strict-filter=false，缺数据模型未排除 | 加 `--strict-filter` |
| 数值过滤后结果意外 | 部分模型 ArkModels 没有 context_window | 接受默认保留（注意 ctx 为 null 的可能不准）|
| 没有 thinking 模型 | --capability 默认非 strict，可能被噪声淹没 | 加 `--strict-filter` |
| enrich 字段全为 null | 未 SSO 登录 | 运行 `arkcli auth login volc-sso` 重新建立 Volc 身份 |
| 召回明显少于预期 | cache 旧；可能 ArkModels 暂时不可用 | 加 `--refresh-cache` 强制同步刷新 |

## 与 `models list` / `models get` 的分工

| 用例 | 推荐 |
|------|------|
| 用户给模糊关键词，找候选 | **`search`** ✓ |
| 找最新发布的模型（time-sensitive） | **`search`**（结果按 update_time DESC）|
| 按 modality / context / capability / cache type 找 | **`search`** + 对应 flag |
| 全量枚举所有模型 | `search` 无 keyword（152 条全量）|
| 按 name 收窄 | `list --name foo`（子串）或 `search "foo"`；精确身份验证用 `get` |
| 拿单个模型的完整详情（计费、限流、能力位详细描述）| `get <name>` |

## 参考

- [arkcli-models](../SKILL.md) — models 全部命令
- [arkcli-shared](../../arkcli-shared/SKILL.md) — 认证和全局参数
