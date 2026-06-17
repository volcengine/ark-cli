# 创建精调任务

- 参考以下顺序完成创建；当用户提供信息不足时，查询并给出建议。
- 若用户已明确给出 model、version、训练数据、训练类型（--type）和超参数/默认策略，允许直接执行 create --dry-run 获取服务端预览；dry-run 失败或返回信息不足时，再按错误类型补查对应命令。
- 真实 create 必须等待用户确认。

## 1. 查询模型、训练方法、价格与部署能力

| 命令                                     | 何时用                        | 常用参数                                                    |
| :------------------------------------- | :------------------------- | :------------------------------------------------------ |
| `arkcli models search <keyword>`       | 找候选基础模型（是否支持精调的信息仅供参考）     | `--modality`、`--strict-filter`                          |
| `arkcli models versions <model>`       | 找基础模型下的模型版本（是否支持精调的信息仅供参考） | 无                                                       |
| `arkcli train finetune capability get` | 查某模型版本支持的训练方法（以此为准）        | `--model`、`--version`                                   |
| `arkcli train finetune pricing`        | 查训练单价                      | `--model`、`--type`                                      |
| `arkcli infer endpoint capability get` | 查询基础模型（model+version）、自定义模型（custom-model-id）支持的推理部署方式             | `--model`、`--version` 、`--model-id` |

这些命令的 flags 会随 CLI 演进；常规场景可直接按本 reference 使用，遇到报错或不确定再查 `--help`

根据用户目标搜索候选模型。选定模型后，查询：

```bash
arkcli models versions <model>
arkcli train finetune capability get --model <model> --version <version>
arkcli train finetune pricing --model <model> --type <type>
arkcli infer endpoint capability get --model <model> --version <version>
```

向用户说明：

- 该版本支持哪些训练方法。
- 选定训练方法的计价单位和价格。
- 选定版本及训练方法，训练后支持的部署方式。
  - 使用`--model`、`--version`或`--model-id`查询到的是基础模型的部署能力，为精调产物的潜在部署能力，不保证具备完全相同的能力。
  - 训练完成后的实际部署方式需要用 `--custom-model-id` 再查，并由部署 skill 继续处理。

默认训练类型与训练方法：

- 用户未指定训练类型（`--type`）时，默认 SFT。
- 用户未指定 LoRA 还是全量训练等训练方法时，默认 LoRA。
- 因此，在模型支持 LoRA 时，默认选择 LoRA 对应的 `--type`；若模型不支持 LoRA，再展示可用训练类型和训练方法并请用户选择。
- 用户明确要求全量训练时，使用该训练类型对应的全量训练 `--type`，并在继续前提示：当前 ArkCLI 还不支持对全量训练产物进行部署，训练完成后的部署需要到控制台完成。

若用户未指定模型，展示少量相关候选项并请用户选择，不要自行提交高成本任务。

## 2. 判断是否需要fallback到精调SDK

若 ArkCLI 可以完成，继续标准流程。若命中主 skill 的 SDK Fallback Gate，暂停并征求确认；确认后转交精调 SDK Skill，不再执行本 reference 的标准提交步骤。

精调SDK 是 fallback，不是默认入口\
当且仅当识别到 ArkCLI 精调训练相关命令无法完成，例如：

- 复杂强化学习任务
- 自定义job.YAML/job.py
- 自定义 grader / rollout plugin
- 自定义训练代码
- 识别到ArkCLI 当前版本没有对应能力时

则提示：

```
当前任务需要精调 SDK，ArkCLI 标准创建流程无法表达该配置。
是否现在自动安装精调 SDK 并继续？
```

用户确认后，参考[`references/ark-finetune-sdk.md`](references/ark-finetune-sdk.md)下载安装SDK后，完成指令

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
| `--train-dataset` + `--train-dataset-version`                | 已有训练数据集引用；本 skill 不创建数据集                                     |
| `--validation-file`                                          | 本地验证文件，可重复                                                   |
| `--validation-tos-uri`                                       | 已上传验证集 TOS URI                                               |
| `--validation-dataset-id` + `--validation-dataset-version`   | 已有验证数据集引用                                                    |
| `--validation-percentage`                                    | 从训练集中切分验证集；与显式验证集互斥                                          |
| `--hyperparameters`                                          | JSON 字符串或 `@file`；值按后端要求传字符串                                 |
| `--epochs`、`--lr`、`--lora-rank`、`--beta`                     | 旧式快捷参数；可用时会合并进 hyperparameters，快捷参数冲突时优先                     |
| `--max-invalid-records-ratio`、`--max-invalid-records-number` | 数据容错                                                         |
| `--shuffle-random-seed`                                      | 数据顺序控制：随机、不打乱或固定种子                                           |
| `--save-model-limit`                                         | 保留训练产物数量                                                     |
| `--enable-trajectory`                                        | RL 轨迹日志；需要任务和项目配置支持                                          |
| `--pipeline`                                                 | RL pipeline 配置文件；若需要 Python plugin 且 CLI 不能表达，走 SDK fallback |
| `--yes`                                                      | 跳过 CLI 确认；Agent 只能在用户二次确认/明确表示直接创建后添加                        |

## 3. 获取并校验训练数据

本期不创建或管理平台数据集。接受以下任一输入：

- 本地训练文件
- 已上传的 TOS URL

用户未提供数据时，请其提供训练集文件或现有数据引用。

### `dataset_schema` 映射

先从 `arkcli models finetune-config <model> <version> --type <type>` 读取 `dataset_schema`，再按下表理解它对应官方文档中的数据集格式。字段细节、样例和限制仍以[模型精调数据集格式说明](https://www.volcengine.com/docs/82379/1099461?lang=zh)为准。

| `dataset_schema`      | 数据集格式               |
| --------------------- | ------------------- |
| `PromptResponse`      | SFT，文本生成模型          |
| `ImageRecognitionSFT` | SFT，多模态模型；兼容文本生成模型  |
| `ImageGenerationSFT`  | SFT，图片生成模型          |
| `VideoGenerationSFT`  | SFT，视频生成模型          |
| `TextDPO`             | DPO，文本生成模型          |
| `ImageRecognitionDPO` | DPO，多模态模型；兼容文本生成模型  |
| `TextRL`              | 强化学习，文本生成模型         |
| `ImageRecognitionRL`  | 强化学习，多模态模型；兼容文本生成模型 |
| `Text`                | 继续预训练，多模态模型、文本生成模型  |

若 finetune-config 未返回 `dataset_schema`，不要猜测；直接使用 `dataset validate` 或请用户确认数据格式。

### 本地离线检查

本地文件先依据[模型精调数据集格式说明](https://www.volcengine.com/docs/82379/1099461?lang=zh)检查：

- 文件可读、编码正确、大小合理
- JSONL 每个非空行都是一个 JSON object
- 样本数和无效行数
- 文档要求的字段、类型和内容结构

使用结构化 JSON 解析器，不用正则表达式验证 JSON。报告准确样本数和失败行号；不要把本地检查描述为平台权威校验。

Token 数处理：

- 环境中存在与目标模型匹配的 tokenizer 时，可以给出本地估算并注明 tokenizer 和误差来源。
- 没有匹配 tokenizer 时，不要用字符数冒充精确 token 数；将 token 统计交给服务端 dry-run。

## 4. 查询并确认超参数

| 命令                              | 何时用                               | 常用参数                         |
| :------------------------------ | :-------------------------------- | :--------------------------- |
| `arkcli models finetune-config` | 确定模型版本和训练方法后查支持的超参和 `dataset_schema` | `<model> <version>`、`--type` |
| `arkcli train finetune create`  | 预览或创建任务                           | `--dry-run` 预览，确认后真实提交       |

打印参数名、默认值、范围或枚举、简短说明。不要依赖记忆填写字段名。

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
| `arkcli train finetune create` | 预览或创建任务 | `--dry-run` 预览，确认后真实提交 |

根据 `arkcli train finetune create --help` 组装命令，先执行 `--dry-run`。本地文件上传如需确认，先获得用户授权，再按 CLI 提示添加 `--yes`。

预览至少汇总：

- job 名称、模型、版本、训练方法
- 若采用默认：说明“未指定训练类型和训练方法，默认 SFT + LoRA”
- 若采用全量训练：再次提示“当前 ArkCLI 还不支持对全量训练产物进行部署，训练完成后的部署需要到控制台完成”
- 训练和验证数据来源
- 自定义超参、推荐超参及其余参数采用默认值的说明
- 服务端统计的样本或 token 信息
- 计价单位、单价和预估费用
- 数据容错、随机种子、产物数量等非默认配置

如果 dry-run 没有返回某个字段，明确说明“未提供”，不要自行补造。

## 6. 最终确认并创建

把完整预览呈现给用户，明确询问是否创建。只有用户确认后，才执行真实创建命令；非交互执行按 CLI 要求添加 `--yes`。

成功后返回：

- job id、名称和初始阶段
- 模型、版本、训练方法
- 关键数据与超参数摘要
- 控制台 URL（若 CLI 返回）
- 后续查询命令，例如 `arkcli train finetune get <job-id>` 或 `watch <job-id>`

不要在本流程中自动部署模型。
