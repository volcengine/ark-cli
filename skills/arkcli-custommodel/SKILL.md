---
name: arkcli-custommodel
version: 1.0.1
description: "arkcli 自定义模型仓库管理：从 TOS 导入自定义模型、查询/筛选自定义模型、查看详情、改名、删除、查询可用量化模式、量化已就绪的模型。任何提到自定义模型 ID（`cm-*`）的管理、部署准备，或要求用 `cm-*` 直接对话/推理/试效果的边界判断，都必须使用本 skill。注意：查询火山**公共基础模型**（doubao 等 foundation models）走 arkcli-models；本 skill 只管账号下的自定义模型仓库。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli models custommodel --help"
---

# arkcli models custommodel

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、配置排查与共享安全规则**
**CRITICAL — 所有 `models custommodel` 命令在执行之前，务必先用 Read 工具读取其对应的 reference 文档，禁止直接盲目调用命令。**
**CRITICAL — 写操作（upload / update / delete / quantize）必须先确认用户意图。删除前必须确认是否还有 endpoint 引用。**

## 守卫与使用原则

- 自定义模型相关需求优先使用 `arkcli models custommodel ...`
- 这些命令虽然是标准 CLI 类型，但实现入口仍然来自 `shortcuts/models/`
- 只有产品命令无法覆盖时，才回退到 [`../arkcli-api-explorer/SKILL.md`](../arkcli-api-explorer/SKILL.md)
- 本 skill 不是基础模型查询入口；基础模型目录查询转 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md)
- 写操作和异步任务必须把影响范围、轮询方式和后续动作串起来，不要停在单条命令

## 适用场景

- 把训练好/微调好的权重从 veTOS 导入到 ARK 自定义模型
- 查询账号下已有的自定义模型（"我的自定义模型有哪些 / 状态如何"）
- 查看自定义模型详情、产物形态、活跃 endpoint 引用
- 修改自定义模型展示名或描述
- 删除不再使用的自定义模型
- 把已 `ready` 的自定义模型量化，准备给 `+deploy` 当目标

## 反唤起信号

- 找官方基础模型 → 用 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md) 的 `search/list/get`
- 直接调用自定义模型推理 → 必须先 `+deploy`，再走 `+chat` / `+gen`
- 触发模型微调任务（customization job 本身）→ 转 [`../arkcli-train-finetune/SKILL.md`](../arkcli-train-finetune/SKILL.md)
- **从精调任务的 step（`global_step_N`）注册成 `cm-`**(=「导出训练产物」)→ 转 [`../arkcli-train-finetune/SKILL.md`](../arkcli-train-finetune/SKILL.md) 的 `arkcli train finetune artifacts list / export`，**不要**用本 skill 的 `upload`(那是给"用户自己的 TOS 文件"用的,后端 Action `UploadModel`;mcj 输出走 `CreateCustomModel`,完全不同的 API)
- 已经拿到 endpoint-id 后想管理 endpoint → 转 [`../arkcli-infer-endpoint/SKILL.md`](../arkcli-infer-endpoint/SKILL.md)

## `cm-*` 直接推理边界

- 用户要求“用 `cm-*` 直接对话/推理/试效果”时也必须加载本 skill。明确说明 `cm-*` 是自定义模型资源 ID，不能直接传给 `+chat` / `+gen`；推理前需要单独获得或部署 Endpoint。
- 该请求本身不授权部署、查询账号下 Endpoint 或发起推理。未经用户继续授权，不执行 `arkcli +deploy`、`arkcli +chat`、`arkcli infer endpoint list`，也不拼接 `jq` 等扫描方案。
- 只说明边界和下一步选择；用户明确要求继续部署后，才转 [`../arkcli-deploy/SKILL.md`](../arkcli-deploy/SKILL.md) 并遵守其确认流程。

## 核心概念

- 本 skill 统一把 `arkcli models custommodel ...` 管理的资源称为**自定义模型**（CustomModel，ID 形如 `cm-xxxxx`）；它与 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md) 中 `search/list/get` 操作的官方基础模型（FoundationModel）是**两套独立资源**
- 自定义模型来源有两类：
  - `import` —— 用户从 TOS 上传权重导入(本 skill `upload` 命令,走 `UploadModel` API)
  - `customization` —— 通过模型微调任务产出(走 [`../arkcli-train-finetune/SKILL.md`](../arkcli-train-finetune/SKILL.md) 的 `train finetune artifacts export`,**底层是另一个 OpenAPI Action `CreateCustomModel`**,跟 `upload` 不互通)
- 生命周期状态机：`preparation → processing → ready`（成功）或 `failed`；导出场景另有 `exporting` / `exportfailed`
- 量化是单独流程：先 `available-quantizations <id>` 查可用模式，并查看 `supported_inference_types_by_quantization` 预判每种量化方式支持的部署/付费形态；再 `quantize <id> --quantization <mode>` 提交量化任务，结果是一个**独立的新 cm-xxxxx**。源模型、量化结果模型、最终部署出来的 endpoint 是三类不同资源，不能混用 ID
- 自定义模型 ID（`cm-xxxxx`）**不是** `<name>-<primary_version>` 形式，**不能直接作为 `+chat` / `+gen` 的 `--model`**；必须先通过 `arkcli +deploy` 获得 endpoint，拿 `ep-xxx` 才能调用推理。若该自定义模型已有 Running Endpoint，`+deploy` 会直接复用已有 endpoint

## 快速决策

- 用户问"我的自定义模型"：读 [`references/arkcli-custommodel-list.md`](references/arkcli-custommodel-list.md)，用 `arkcli models custommodel list --mine`
- 用户要按名称 / ID / base 模型 display name 模糊查找：读 [`references/arkcli-custommodel-list.md`](references/arkcli-custommodel-list.md)，用 `list --search <kw>`
- 用户已有 `cm-xxxxx` 并要状态 / 详情 / endpoint 引用 / artifact types：读 [`references/arkcli-custommodel-get.md`](references/arkcli-custommodel-get.md)，用 `get`
- 用户要从 TOS 导入权重：读 [`references/arkcli-custommodel-upload.md`](references/arkcli-custommodel-upload.md)。用户已有 `tos://...` 时确认 TOS URI、base model 和名称；用户还没有 TOS URI 时先引导开通/上传 TOS，不要直接跑 `upload`
- 用户要改名 / 改描述：读 [`references/arkcli-custommodel-update.md`](references/arkcli-custommodel-update.md)，确认意图后用 `update`
- 用户要删除：读 [`references/arkcli-custommodel-delete.md`](references/arkcli-custommodel-delete.md)，先查 `active_endpoints`，确认后再执行
- 用户要量化：先读 [`references/arkcli-custommodel-available-quantizations.md`](references/arkcli-custommodel-available-quantizations.md)，再读 [`references/arkcli-custommodel-quantize.md`](references/arkcli-custommodel-quantize.md)

## 列表参数与输出契约

- `custommodel list --sort-order` 只接受小写 `asc` / `desc`。收到其他值时应把它视为本地参数错误，不要尝试调用接口或改成其他大小写后盲目重试。
- 使用 `--format table` 或 `--format csv` 时，每个 `result.items[]` 模型是一行；不要把分页响应根对象或 `result` map 当成模型记录。需要保留完整分页元数据时使用 JSON/YAML。

## 部署前的自定义模型目标澄清

仅当用户的最终目标是“把我的自定义模型部署成 Endpoint”、但没有给出唯一 `cm-*` 时执行本节；用户已明确给出 `cm-*` 时跳过。

1. 先执行同范围只读查询：`arkcli models custommodel list --mine --statuses ready --page-all --page-delay 500 --format json`。保留退出码与 stderr，不能把认证、权限或网络失败当作空清单；认证/权限错误按 shared 处理，不扫描凭证或自动切身份。
2. 候选只能来自本轮完整结构化结果。输出被宿主截断时，先读取工具保存的完整文件；确实没有完整捕获时，允许把同一只读查询的 stdout 重新落盘，stderr 单独保留，再用 `jq` / `sed` 检查 `result.items` 与分页信息。该同范围补查一轮最多执行一次，不循环重试。`jq` 空结果或字段缺失时先核对当前 reference 与原始 JSON，不盲猜字段、不凭记忆补 `cm-*`。分页未取全、补查仍失败或关键字段仍缺失时，说明具体边界并停止；拿全候选前不出选择题，更不执行部署。
3. 按 0 / 1 / N 收敛：
   - 0 个：停止，提示用户先 upload、量化或完成精调产物导出。
   - 1 个：复述该模型的 `id / name / foundation_model / create_time`，将它作为唯一目标转交 `arkcli-deploy`。
   - N 个：使用宿主提供的结构化选择能力，把每个候选的 `id / name / foundation_model / create_time` 直接列给用户选择；不要在通用 Skill 中写死某个宿主的工具名，也不要额外写死 `Other` 选项。
4. 唯一目标确定前，不执行 `arkcli +deploy`、`arkcli infer endpoint create` 或 Raw API。目标选择完成后转 `arkcli-deploy`，并继续遵守其写操作确认；选择模型本身不等于授权部署。

## Agent 快速执行顺序

1. 不确定认证状态时，先 `arkcli auth status`
2. "我的自定义模型"语义：直接 `custommodel list --mine`，**不要套 shared 的 Tags 默认过滤**（custommodel 服务端原生支持 `--mine`）
3. 上传前必填三项：`--name` / `--base-model <foundation-model-id>` / `--tos tos://<bucket>/<prefix>`；缺任一会被服务端拒
4. `upload` / `quantize` 是异步任务：返回后用 `custommodel get <id>` 轮询 status，**不要原地循环 upload/quantize**
5. `quantize` 前必跑 `available-quantizations <id>`：不同 base model 支持的量化模式不同，盲传服务端会拒；若用户关心 token / 模型单元等部署形态，优先看返回里的 `supported_inference_types_by_quantization`
6. `quantize --dry-run` 只输出本地 `preview.v1`，不会调用 `CreateQuantizedCustomModel`；`steps[].payload` 只描述真实请求字段，不应出现后端 `DryRun`。它不是服务端校验，核对后仍需确认再执行真实量化
7. `delete` / `update` / `quantize` 是写操作，执行前向用户复述影响范围
8. `delete` 默认会走 [Y/N] 二次确认；`--yes` 表示跳过本地二确，`--dry-run` 表示只预览不删除。只有用户已经明确确认删除目标和影响范围后，agent 才能把 `--yes` 加到命令里
9. `get --transform` 是 `custommodel get` 自己的字段白名单，不是全局 GJSON 表达式；要查嵌套路径时不要把它当作全局 `--transform`

## 典型业务链路

### 1. 从 TOS 上传新自定义模型

```
auth status → custommodel upload --name X --base-model <fm-id> --tos tos://b/p
            → custommodel get <id>  （轮询直到 status=ready）
            → custommodel get <id> --transform 'artifact_types'  （看产物形态）
```

### 2. 量化已就绪的自定义模型

```
custommodel get <id>  （确认 status=ready）
        → custommodel available-quantizations <id>  （看支持哪些 mode）
        → custommodel quantize <id> --quantization <mode> --dry-run
        → 用户确认量化目标和影响范围
        → custommodel quantize <id> --quantization <mode>
        → custommodel get <new-id>  （量化结果是新 cm-xxxxx，再次轮询）
```

### 3. 准备给 `+deploy` 当目标

```
custommodel list --mine --statuses ready --page-all --page-delay 500 --format json
        → 按 0 / 1 / N 澄清唯一 cm-xxxxx
        → +deploy --model cm-xxxxx ...   （若已有 Running Endpoint 会复用；详见 ../arkcli-deploy/SKILL.md）
```

### 4. 清理不再使用的自定义模型

```
custommodel get <id> --transform 'active_endpoints'  （确认无 endpoint 引用）
        → custommodel delete <id> --dry-run
        → custommodel delete <id>  （交互二确）或 custommodel delete <id> --yes
```

## 反模式（agent 必读）

- 不要用 `arkcli models search` / `list` 找自定义模型 —— 那两条只走 FoundationModel 目录，自传模型一律不在里面。要找自传模型用 `custommodel list --search <kw>` 或 `--mine`
- 不要在 `upload` 之后立刻 `quantize` —— upload 是异步任务，status 经历 `preparation → processing → ready`；先 `custommodel get <id>` 确认 ready，再走 `available-quantizations` → `quantize`
- 不要给 `quantize` 传一个 `available-quantizations` 没列的 mode —— 不同 base model 支持的量化集合不同，盲传必失败。先 `available-quantizations <id>`，从返回里挑
- 不要把 `cm-xxxxx` 直接传给 `+chat` / `+gen` 的 `--model` —— 自定义模型必须先通过 `+deploy` 获得 endpoint（`ep-xxx`）才能推理调用；`+deploy` 可能复用已有 Running Endpoint
- 不要在多个 ready 自定义模型中自行挑一个部署 —— 先按“部署前的自定义模型目标澄清”让用户从本轮实时结果中选择
- 不要为了自动化主动补 `--yes` —— 没 `--yes` 时 CLI 会走 [Y/N] 二确；只有用户已经确认删除 `cm-xxxxx` 且知道 endpoint 引用风险时才带
- 不要在 "我的" 语义下走 shared 的 Tags 客户端过滤 —— `custommodel list --mine` 是服务端原生过滤，更准也更省请求
- 不要密集刷 `get` 来轮询 status —— 推荐间隔 ≥ 10s，否则会被限流

## 命令一览

| 命令 | 说明 |
|------|------|
| `arkcli models custommodel list` | 翻页 + 多维过滤 |
| `arkcli models custommodel get <id>` | 详情 / 轮询 status |
| `arkcli models custommodel upload` | 从 TOS 导入（异步） |
| `arkcli models custommodel update <id>` | 改名 / 改描述 |
| `arkcli models custommodel delete <id> [--yes] [--dry-run]` | 删除（破坏性，不可逆）；默认二确，`--yes` 跳过，`--dry-run` 预览 |
| `arkcli models custommodel available-quantizations <id>` | 查可用量化模式（quantize 前必跑） |
| `arkcli models custommodel quantize <id> --quantization <mode> [--dry-run]` | 量化（异步）；`--dry-run` 仅本地预览，不调用后端 |

## 常见降级

- 认证失败：转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
- profile / region 看起来不对：转 [`../arkcli-config/SKILL.md`](../arkcli-config/SKILL.md)
- 找的是基础模型不是自定义模型：转 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md)
- 拿到自定义模型后要部署：转 [`../arkcli-deploy/SKILL.md`](../arkcli-deploy/SKILL.md)
- 产品命令未覆盖（例如想直接触发 customization job）：转 [`../arkcli-api-explorer/SKILL.md`](../arkcli-api-explorer/SKILL.md)

## 参考

- [`references/arkcli-custommodel-list.md`](references/arkcli-custommodel-list.md)
- [`references/arkcli-custommodel-get.md`](references/arkcli-custommodel-get.md)
- [`references/arkcli-custommodel-upload.md`](references/arkcli-custommodel-upload.md)
- [`references/arkcli-custommodel-update.md`](references/arkcli-custommodel-update.md)
- [`references/arkcli-custommodel-delete.md`](references/arkcli-custommodel-delete.md)
- [`references/arkcli-custommodel-available-quantizations.md`](references/arkcli-custommodel-available-quantizations.md)
- [`references/arkcli-custommodel-quantize.md`](references/arkcli-custommodel-quantize.md)
- [`references/evals.md`](references/evals.md)
- [arkcli-models](../arkcli-models/SKILL.md) — 基础模型查询（与本 skill 是两套资源）
- [arkcli-deploy](../arkcli-deploy/SKILL.md) — 拿到 ready 的自定义模型后部署成 endpoint
- [arkcli-shared](../arkcli-shared/SKILL.md) — 认证、全局 flag、安全规则
