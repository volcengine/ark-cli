# models search

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

**Agent 优先**的模型搜索：召回 = 全量目录；enrich = ArkModels（context window / 模态 / capability）；filter = 结构化条件；rank = 4 桶 + context_window 加权 + 时序。**没有分页概念**，默认返回全部命中。

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
| `--capability` | string（可重复）| 必备能力，多个为 AND：`thinking` / `mcp` / `functioncall` / `web_browsing` / `knowledge_base` / `caching` / `response_format` / `reasoning_effort` |
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
   5️⃣  结构化过滤（modality / context / tokens / capability，AND）
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
    "web_browsing": true, "knowledge_base": true, "caching": true,
    "response_format": false, "reasoning_effort": true
  }
}
```

## 重要行为

### 缺数据怎么办（StrictFilter）
- **默认 `--strict-filter=false`**：`--modality video` 时，**没有** modality 数据的模型也会**保留**（agent 看到候选但置信度低）
- **`--strict-filter`**：缺数据视为不满足，直接排除（agent 拿到的 100% 是 ground truth）

模态查询通常加 `--strict-filter` 更准；数值查询（context window）保持默认（缺数据不杀）通常更合理。

### 召回不再有 top-K 限制
不像旧版 fuzzy 接口默认 9 条，本命令默认返回**全部命中**。如需限制，传 `--size N`。

### Cache 与 SSO
- ArkModels 是 Console BFF 接口，需要 SSO 凭证
- 无 SSO（AK/SK only）→ enrich 静默跳过，模型仍能返回，filter 自动 no-op
- Cache 路径按 `(profile, region, project)` 分 → 多账号/多 project 不串
- TTL 概念不存在：**任何 search 调用都会触发后台异步刷新**，下次调用永远是新鲜数据
- 启动时 cache 命中即返回，刷新由 detached 子进程执行（`arkcli models _refresh-cache`）

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
| enrich 字段全为 null | 未 SSO 登录 | 按 profile tenant 选择：火山 `arkcli auth login volc-sso` |
| 召回明显少于预期 | cache 旧；可能 ArkModels 暂时不可用 | 加 `--refresh-cache` 强制同步刷新 |

## 与 `models list` / `models get` 的分工

| 用例 | 推荐 |
|------|------|
| 用户给模糊关键词，找候选 | **`search`** ✓ |
| 找最新发布的模型（time-sensitive） | **`search`**（结果按 update_time DESC）|
| 按 modality / context / capability 找 | **`search`** + 对应 flag |
| 全量枚举所有模型 | `search` 无 keyword（152 条全量）|
| 按 name 精确匹配 | `list --name foo`（精确）或 `search foo`（含 fuzzy）|
| 拿单个模型的完整详情（计费、限流、能力位详细描述）| `get <name>` |

## 参考

- [arkcli-models](../SKILL.md) — models 全部命令
- [arkcli-shared](../../arkcli-shared/SKILL.md) — 认证和全局参数
