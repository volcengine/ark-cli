# 创建精调任务

- 参考以下顺序完成创建；当用户提供信息不足时，查询并给出建议。
- 若参数已明确，允许直接执行 `create --estimate` 获取在线 Token 与费用估算；估算失败或信息不足时，再按错误类型补查。
- 真实 create 必须等待用户确认。

## 1. 查询模型、训练方法、价格与部署能力

| 命令                                     | 何时用                        | 常用参数                                                    |
| :------------------------------------- | :------------------------- | :------------------------------------------------------ |
| `arkcli models search <keyword>`       | 找候选基础模型（是否支持精调的信息仅供参考）     | `--modality`、`--strict-filter`                          |
| `arkcli models versions <model>`       | 找基础模型下的模型版本（是否支持精调的信息仅供参考） | 无                                                       |
| `arkcli train finetune capability get` | 查某模型版本支持的训练方法（以此为准）        | `--model`、`--version`                                   |
| `arkcli train finetune pricing`        | 按 Token 或实例计费查询训练价格          | `--model`、`--model-version`、`--type`、`--billing-method` |
| `arkcli infer endpoint capability get` | 查询基础模型（model+version）、自定义模型（custom-model-id）支持的推理部署方式             | `--model`、`--version` 、`--model-id` |

这些命令的 flags 会随 CLI 演进；常规场景可直接按本 reference 使用，遇到报错或不确定再查 `--help`

根据用户目标搜索候选模型。选定模型后，查询：

```bash
arkcli models versions <model>
arkcli train finetune capability get --model <model> --version <version>
arkcli train finetune pricing --model <model> --type <type>
# token 是默认计费方式；不要改用通用 arkcli pricing models
# 实例计费必须使用精确版本和训练方法；自定义超参需与后续创建保持一致
arkcli train finetune pricing --model <model> --model-version <version> --type <type> \
  --billing-method instance --hyperparameters '{"epoch":"3"}'
arkcli infer endpoint capability get --model <model> --version <version>
```

向用户说明：

- 该版本支持哪些训练方法。
- 选定训练方法的计价单位和价格。
- `token` 返回按 Token 计费项；`instance` 会先校验该模型版本和训练方法支持 Instance，
  再返回资源模板、机型单价、角色数量对应的小时价及整体小时价范围。只有
  `price_complete=true` 时价格范围才是完整的；否则必须报告 `missing_flavor_ids`，不得按 0 元解释。
- 选定版本及训练方法，训练后支持的部署方式。
  - 使用`--model`、`--version`或`--model-id`查询到的是基础模型的部署能力，为精调产物的潜在部署能力，不保证具备完全相同的能力。
  - 训练完成后的实际部署方式需要用 `--custom-model-id` 再查，并由部署 skill 继续处理。

默认训练类型与训练方法：

- 用户未指定训练类型（`--type`）时，默认 SFT。
- 用户未指定 LoRA 还是全量训练等训练方法时，默认 LoRA。
- 因此，在模型支持 LoRA 时，默认选择 LoRA 对应的 `--type`；若模型不支持 LoRA，再展示可用训练类型和训练方法并请用户选择。
- 用户明确要求全量训练时，使用该训练类型对应的全量训练 `--type`，并在继续前提示：当前 ArkCLI 还不支持对全量训练产物进行部署，训练完成后的部署需要到控制台完成。

若用户未指定模型，展示少量相关候选项并请用户选择，不要自行提交高成本任务。

## 2. ArkCLI 能力回退

精调 SDK 是 fallback，不是默认入口。

先检查当前版本的 `arkcli train finetune`、`arkcli models finetune-config` 和相关 `--help`。ArkCLI 能完整表达用户配置时继续标准流程；不能表达复杂强化学习、自定义 job、grader、rollout plugin 或自定义训练代码时，暂停并询问：

> 当前配置超出 ArkCLI 标准创建能力。是否转交扩展精调流程继续处理？

只有用户明确确认后，才读取 [`ark-finetune-sdk.md`](ark-finetune-sdk.md)，并停止执行本 reference 的后续创建步骤。安装、鉴权、配置和提交均由该文档负责；用户拒绝时不要安装或提交。

## 常用创建参数

检查用户要求能否由当前 ArkCLI flags 完整表达：

```bash
arkcli train finetune --help
arkcli train finetune create --help
```

<br />

`arkcli train finetune create` 的高频参数：

| 参数                                                           | 说明                                                           |
| ------------------------------------------------------------ | ------------------------------------------------------------ |
| `--name`                                                     | 任务名称                                                         |
| `--description`                                              | 任务描述                                                         |
| `--model` + `--model-version`                                | 基于基础模型训练                                                     |
| `--model-id`                                                 | 基于已有 custom model 继续训练；与 `--model/--model-version` 互斥        |
| `--type`                                                     | 训练类型；未指定时按“默认 SFT + LoRA”处理                                  |
| `--train-file`                                               | 本地训练文件，可重复                                                   |
| `--train-tos-uri`                                            | 已上传训练数据 TOS URI                                              |
| `--train-dataset <dataset-id>:<version-id>`                  | 已有训练 Dataset；旧的独立 `--train-dataset-version` 仍兼容                    |
| `--train-path <json>`                                       | 普通训练 Dataset，可重复；每项设置 `dataset_id`、`dataset_version_id`，最多再设置 `multiplier`/`sample_count` 之一 |
| `--preset-dataset <json>`                                   | 模型支持的预置数据集，可重复；每项设置 `dataset_version_id`，并在 `inject_multiplier`/`inject_sample_count` 中二选一 |
| `--validation-file`                                          | 本地验证文件，可重复                                                   |
| `--validation-tos-uri`                                       | 已上传验证集 TOS URI                                               |
| `--validation-dataset <dataset-id>:<version-id>`             | 已有验证 Dataset；旧的独立版本参数仍兼容                                    |
| `--validation-percentage`                                    | 从训练集中切分验证集；与显式验证集互斥                                          |
| `--hyperparameters`                                          | JSON 字符串或 `@file`；值按后端要求传字符串                                 |
| `--epochs`、`--lr`、`--lora-rank`、`--beta`                     | 旧式快捷参数；可用时会合并进 hyperparameters，快捷参数冲突时优先                     |
| `--max-invalid-records-ratio`、`--max-invalid-records-number` | 数据容错                                                         |
| `--shuffle-random-seed`                                      | 数据顺序控制：随机、不打乱或固定种子                                           |
| `--save-model-limit`                                         | 保留训练产物数量                                                     |
| `--enable-trajectory`                                        | RL 轨迹日志；需要任务和项目配置支持                                          |
| `--pipeline`                                                 | RL pipeline 配置文件；若需要 Python plugin 且 CLI 不能表达，按“ArkCLI 能力回退”处理 |
| `--yes`                                                      | 跳过 CLI 确认；Agent 只能在用户二次确认/明确表示直接创建后添加                        |

## 3. 获取并校验训练数据

本 skill 不创建或维护平台 Dataset；只有独立 Dataset 生命周期操作或不涉及精调任务的独立校验才转 [`../../arkcli-datasets/SKILL.md`](../../arkcli-datasets/SKILL.md)。精调任务创建或预检所需的本地/TOS 数据校验继续留在本流程，直接使用下方 `arkcli dataset validate`。创建任务可接受：

- 本地训练文件
- 已上传的 TOS URL
- 平台 Dataset 的 `<dataset-id>:<version-id>` 引用
- 模型配置明确支持的 preset Dataset

同一训练集或验证集只能选择一种普通数据来源；Dataset 与对应的 TOS/本地文件参数互斥。单个普通 Dataset 默认 `Multiplier=1`；需要重复引用、倍率或采样数时使用可重复的 `--train-path`。每项最多设置 `multiplier`/`sample_count` 之一，均不设置时仍默认 `Multiplier=1`：

```bash
--train-path '{"dataset_id":"ds-...","dataset_version_id":"dsv-...","multiplier":2}'
--train-path '{"dataset_id":"ds-...","dataset_version_id":"dsv-...","sample_count":500}'
```

`--train-path` 不能和 `--train-dataset`、训练 TOS 或本地文件混用。preset 使用可重复 JSON，`inject_multiplier` 与 `inject_sample_count` 也必须且只能设置一个：

```bash
--preset-dataset '{"dataset_version_id":"dsv-...","inject_sample_count":100}'
```

CLI 会根据精确模型版本和训练类型检查 Dataset schema，并拒绝模型配置未声明支持的 preset。

用户未提供数据时，请其提供训练集文件或现有数据引用。

### 使用 CLI 校验数据

本 skill 内直接使用 `arkcli dataset validate` 校验本地文件或 TOS 数据，无需为此创建 Dataset 或切换到数据集管理流程。先确认模型、精确版本及实际训练 `--type`，与后续创建保持一致；查看当前命令支持的参数：

```bash
arkcli dataset validate --help

# 本地 JSONL；多个文件重复 --local
arkcli dataset validate --model <model> --model-version <version> --type <type> \
  --local <train.jsonl>

# 已有 TOS 数据；多个 URI 重复 --tos-uri
arkcli dataset validate --model <model> --model-version <version> --type <type> \
  --tos-uri tos://<bucket>/<path>
```

- CLI 自动读取模型配置中的 `dataset_schema`，无需 Agent 先查格式文档、维护静态 schema 映射或自写校验器。训练集和显式验证集都要覆盖；可按数据集分别执行，便于归属结果。
- `--local` 与 `--tos-uri` 二选一。本地文件会上传并触发服务端校验，不是离线检查；执行前按主 skill 的上传确认规则处理。已有上传授权时不重复询问。
- TOS URI 必须可列举且包含 `.jsonl` 对象；同一次调用的多个 URI 必须属于同一 bucket，不能使用 `ds-ds/*` 上传暂存路径。以当前 `--help` 为准。
- 当前命令不直接接收 `ds-*/dsv-*`、preset 或 `--model-id`。已有 Dataset/preset 继续走创建流程的 schema/能力检查；需要内容校验时使用可确认来源的本地文件或受支持 TOS URI。自定义模型续训不能猜测基础模型映射，未确认校验上下文时说明能力缺口。

### 校验结果与失败处理

以 CLI 返回的服务端校验任务结果为准，汇总实际校验的数据来源、模型/版本/训练类型、成功或失败状态，以及返回的错误摘要、行号或统计信息；未返回的字段不补造，不输出完整样本。

- 所有目标数据校验成功后，才报告数据校验通过。`--dry-run` 只预览执行计划，不访问网络、不上传、不做真实校验；参数解析成功或 `create --estimate` 成功也不能替代数据校验结果。
- 校验失败时，根据返回的错误定位；需要理解字段要求或修复数据时，再读取火山方舟[模型精调数据集格式说明](https://www.volcengine.com/docs/82379/1099461?lang=zh)对应章节。修复后对相同模型配置重新执行 `dataset validate`。
- 命令不可用、鉴权/网络失败、模型未提供 schema 或校验未完成时，明确“未完成校验”及原因，不将基础检查、文档比对或用户确认格式当作校验通过。只请求补充缺失信息，不自动安装/升级 CLI 或提交训练。

Token 数处理：

- 环境中存在与目标模型匹配的 tokenizer 时，可以给出本地估算并注明 tokenizer 和误差来源。
- 没有匹配 tokenizer 时，不要用字符数冒充精确 token 数；将 token 统计交给在线 `--estimate`。

## 4. 查询并确认超参数

| 命令                              | 何时用                               | 常用参数                         |
| :------------------------------ | :-------------------------------- | :--------------------------- |
| `arkcli models finetune-config` | 确定模型版本和训练方法后查支持的超参和 `dataset_schema` | `<model> <version>`、`--type` |
| `arkcli train finetune create`  | 估算或创建任务                           | `--estimate` 在线估算，确认后真实提交       |

打印参数名、默认值、范围或枚举、简短说明。不要依赖记忆填写字段名。

显式传入 `--type` 时，`models finetune-config` 会使用与 `train finetune capability get` 相同的权威 `FinetuneTypes`，校验精确的模型版本是否支持该训练方法。不支持时会在读取配置 schema 前返回 validation error；不要继续 estimate 或创建任务。

- 用户要求自定义超参时，请其确认覆盖值。
- 根据用户选择的模型、偏好、数据量、效果指标、日志等信息，分析并帮助用户配置超参；如果没有更优配置，使用默认值。
- 拒绝超出 schema 的值

通用训练配置优先从 finetune-config 的 schema 读取默认值和范围。本 reference 只记稳定语义：

- 训练轮数、学习率、batch size、LoRA rank/alpha/dropout、DPO beta、RL 步数等字段名可能随训练方法不同而变化。
- `save_model_limit` 决定保留多少个训练产物；默认值和上限以当前 CLI/API 为准。
- 数据容错和 shuffle seed 属于数据配置，不是模型超参；预览时和超参分开展示。

## 5. 创建预览 校验配置、数据、Token与费用

| 命令                             | 何时用     | 常用参数                   |
| :----------------------------- | :------ | :--------------------- |
| `arkcli train finetune create` | 估算或创建任务 | `--estimate` 在线估算，确认后真实提交 |

根据 `arkcli train finetune create --help` 组装命令，先执行 `--estimate`。这是在线估算而非 Client Preview；本地文件上传如需确认，先获得用户授权。

预览至少汇总：

- job 名称、模型、版本、训练方法
- 若采用默认：说明“未指定训练类型和训练方法，默认 SFT + LoRA”
- 若采用全量训练：再次提示“当前 ArkCLI 还不支持对全量训练产物进行部署，训练完成后的部署需要到控制台完成”
- 训练和验证数据来源、`dataset validate` 的实际结果及未覆盖项
- 自定义超参、推荐超参及其余参数采用默认值的说明
- 服务端统计的样本或 token 信息
- 计价单位、单价和预估费用
- 数据容错、随机种子、产物数量等非默认配置

如果 estimate 没有返回某个字段，明确说明“未提供”，不要自行补造。

## 6. 最终确认并创建

- 认证通过后生成并在本次创建流程中复用 `ARKCLI_SKILL_FLOW_ID=ftf_<ULID>`。
- 按 shared 单命令前缀，在第一条创建业务命令前执行
  `arkcli train finetune _report-activity --action create_flow_enter`。
- 仅在 `dataset validate` 或其他明确的服务端数据校验覆盖本次目标数据并成功后执行 `--action data_validation_success`；普通参数解析、文档比对、Client Preview `--dry-run` 和仅有 Token/费用估算结果均不上报。

把完整预览呈现给用户，明确询问是否创建。只有用户确认后，才执行真实创建命令；非交互执行按 CLI 要求添加 `--yes`。

成功后返回：

- job id、名称和初始阶段
- 模型、版本、训练方法
- 关键数据与超参数摘要
- 控制台 URL（若 CLI 返回）
- 后续查询命令，例如 `arkcli train finetune get <job-id>` 或 `watch <job-id>`

不要在本流程中自动部署模型。
