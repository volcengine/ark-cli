# 精调产物优选、导出与部署

本流程覆盖两种入口：

- 用户指定一个 `mcj-...`，要求根据训练效果选择、导出并部署产物。
- 用户已指定一个或多个 `cm-...`，要求判断部署方式并部署。

开始前除 shared skill 外，按阶段读取：

- custom model 查询、量化和可部署版本准备：[`../../arkcli-custommodel/SKILL.md`](../../arkcli-custommodel/SKILL.md)
- 创建在线推理 Endpoint：[`../../arkcli-deploy/SKILL.md`](../../arkcli-deploy/SKILL.md)
- 用户要求底层 Endpoint 配置或后续管理时：[`../../arkcli-infer-endpoint/SKILL.md`](../../arkcli-infer-endpoint/SKILL.md)

不要复制或凭记忆执行这些 skill 的命令。进入对应阶段时读取其要求的 reference，并遵守写操作、计费和二次确认规则。

## 命令速查

| 命令 | 何时用 | 常用参数 |
|---|---|---|
| `arkcli train finetune metrics <mcj-id>` | 选择最佳 step | `--metric`、`--from-step`、`--to-step`、`--output` |
| `arkcli train finetune artifacts list <mcj-id>` | 列产物 | `--format table` 给人选择，`--lite` 跳过 cm 名称二查 |
| `arkcli train finetune artifacts export <mcj-id>` | 导出产物为 `cm-*` | `--artifact-name` 可重复，`--custom-model-name-prefix` 可选 |
| `arkcli models custommodel get <cm-id>` | 查 custom model 详情 | `--transform artifact_types,active_endpoints` |
| `arkcli models custommodel available-quantizations <cm-id>` | 查是否能创建新可部署版本 | 量化前必查 |
| `arkcli models custommodel quantize <cm-id>` | 创建量化新版本 | `--quantization`，异步返回新 `cm-*` |
| `arkcli infer endpoint capability get --custom-model-id <cm-id>` | 查实际部署能力 | 部署前查询 |
| `arkcli +deploy` | 创建或复用 Endpoint | 先 `--dry-run`，确认后真实执行 |

## A. 从 MCJ 选择最佳 Step

用户直接提供 `cm-...` 时跳到 C。

### 1. 查询任务与可用指标

```bash
arkcli train finetune get <mcj-id>
arkcli train finetune metrics <mcj-id>
```

先从 `get` 确认训练方法和任务状态。任务尚未产生有效指标或产物时，说明当前阶段并停止导出流程。

若任务是全量训练，停止 CLI 导出部署流程并提示：当前 ArkCLI 还不支持对全量训练产物进行部署，需要到控制台完成。不要继续执行 `artifacts export`、custom model 部署或 `+deploy`。

第一次调用 `metrics` 不指定指标，用返回的 `available_metrics` 获取真实名称。大结果使用 `--output` 写入文件，再以结构化 JSON 工具计算，不要人工扫描长序列。

### 2. 选择用于排名的效果指标

按以下优先级选择，名称必须来自 `available_metrics`：

| 任务类型 | 第一优先 | 无第一优先时 |
|---|---|---|
| 非 RL | test/eval/validation 范围的 loss，取最小值 | train 范围的 loss，取最小值 |
| RL | test/eval/validation 范围的 `final_rewards`，取最大值 | train 范围的 `final_rewards`，取最大值 |

匹配规则：

- `test`、`eval`、`validation` 都视为评测范围，优先级相同；存在多个候选时展示候选名并说明最终使用哪个。
- RL 优先匹配名称为或末段为 `final_rewards` 的指标。
- 不要把 accuracy、普通 reward、grad norm 或其他指标静默替代上述主指标。
- 若首选和 fallback 指标都不存在，列出真实可用指标并请用户指定，不要自行猜测。

记录以下信息：

- 实际使用的 metric 名称及选择理由
- 最优值
- 对应的一个或多个 step
- 指标属于评测集还是训练集

多个 step 具有相同最优值时全部保留。若用户指定 top-k，则按指标方向排序并记录 top-k；不要默认把邻近 step 当作多个“最优 step”。

## B. 将最佳 Step 映射到 Artifact 并导出

### 1. 查询产物

```bash
arkcli train finetune artifacts list <mcj-id>
```

从 artifact 名称或结构化字段提取其 step。只对能够可靠解析 step 的产物做自动映射；无法解析时展示列表让用户选择。

对每个最佳 step：

1. 优先选择 step 完全相同的 artifact。
2. 没有完全相同时，计算与每个可导出 artifact step 的绝对距离。
3. 选择距离最小的 artifact；距离相同则保留全部并列候选。
4. 明确展示 `metric step → artifact step → 距离`，不得把最近产物描述成精确命中。

合并不同最佳 step 映射出的重复 artifact，避免重复导出。

### 2. 用户 Check

导出前展示：

- MCJ 和训练方法
- 使用的效果指标、最优值和最佳 step
- 推荐 artifact 及映射距离
- artifact 当前 export status
- 已导出的 custom model id（若有）

要求用户确认最终导出的一个或多个 artifact。用户可以增删候选。

### 3. 导出

仅对用户确认且尚未导出的 artifact 执行：

```bash
arkcli train finetune artifacts export <mcj-id> \
  --artifact-name <artifact-1> \
  --artifact-name <artifact-2>
```

需要命名前缀时先确认，再使用当前 `--help` 支持的参数。已导出的 artifact 直接复用现有 `cm-...`，不要重复 export。

说明 export 是把 MCJ 产物注册为模型仓库中的 custom model，不是下载权重。记录每个 artifact 对应的 custom model id 和状态。

导出状态常见值：

- `Available`：可导出。
- `Exported`：已导出，直接复用 `custom_model_id`。
- `Exporting`：导出中，稍后查询。
- `ExportFailed` / `Expired`：不可直接部署，展示状态并请用户选择下一步。

## C. 查询 Custom Model 与可部署版本

对用户指定或刚导出的每个 `cm-...`，读取并遵循 `arkcli-custommodel` skill。

至少查询：

```bash
arkcli models custommodel get <cm-id>
arkcli models custommodel get <cm-id> --transform artifact_types,active_endpoints
arkcli infer endpoint capability get --custom-model-id <cm-id>
```

汇总：

- custom model 名称、来源、状态和基础模型
- 当前已有 Endpoint
- 当前模型直接支持的部署方式
- 是否存在通过量化等方式创建新版本后可支持的其他部署方式

若用户目标部署方式当前不支持，再按 `arkcli-custommodel` skill：

1. 查询 `available-quantizations` 及各量化方式预期支持的推理形态。
2. 向用户说明量化对精度、性能和资源形态的影响。
3. 用户确认后创建新的量化 `cm-...`。
4. 轮询新模型至 `ready`。
5. 对新 `cm-...` 重新查询详情和实际部署能力。

不要因为存在量化选项就自动创建新版本。源模型与量化模型是不同资源，必须保留清晰映射。

## D. 选择并部署

向用户展示一个部署候选表，至少包含：

- cm id 和名称
- 来源 artifact/step（已知时）
- 对应最优 metric 和值（已知时）
- 原始或量化版本
- 当前支持的部署方式
- 已有 Running Endpoint（若有）

询问用户：

1. 要部署哪一个或多个 `cm-...`
2. 每个模型采用哪种部署方式
3. Endpoint 名称及该方式必需的配置

用户选择后读取并遵循 `arkcli-deploy` skill：

1. 对每个目标分别执行 `--dry-run`。
2. 展示将创建或复用的 Endpoint、部署方式和计费影响。
3. 获得最终确认后真实部署。
4. 记录每个 `cm-... → ep-...` 映射及 Endpoint 状态。

多个模型部署属于多个计费资源，必须让用户看清目标数量。不要把用户对 artifact export 的确认视为 Endpoint 部署确认。
