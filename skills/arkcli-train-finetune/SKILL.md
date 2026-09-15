---
name: arkcli-train-finetune
description: 使用 ArkCLI 创建、查询和管理模型精调训练任务，并从训练指标选择最佳 step、导出训练产物为 custom model、衔接模型仓库与推理部署。任何包含精调任务 ID（`mcj-*`）的查询、查不到原因诊断、日志、trajectory、状态或生命周期操作都应使用本 skill；也适用于选择训练方法、查询精调价格和超参数、创建任务及导出部署。本 skill 不负责数据集管理。
---

# ArkCLI 精调训练

先读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，遵循认证、输出、安全和二次确认规则。

## 适用场景与能力边界

- 创建精调任务：读取 [`references/create.md`](references/create.md)
- 列出或筛选任务：读取 [`references/list.md`](references/list.md)
- 查询、观察或操作一个指定任务：读取 [`references/manage.md`](references/manage.md)
- 根据指标选择 step、导出产物并部署：读取 [`references/export-deploy.md`](references/export-deploy.md)
- 不直接管理数据集，但可以使用本地文件、TOS URL、`ds-*/dsv-*` 引用和模型支持的 preset；需要创建或维护 Dataset 时转 [`../arkcli-datasets/SKILL.md`](../arkcli-datasets/SKILL.md)。
- 普通训练 Dataset 默认使用 `--train-dataset`（`Multiplier=1`）；需要重复引用、倍率或采样数时改用可重复的 `--train-path`。每项最多设置 `multiplier` 或 `sample_count` 之一，均不设置时仍默认 `Multiplier=1`。preset 必须在 `inject_multiplier` 与 `inject_sample_count` 中二选一。
- 训练产物的指标分析和 artifact export 由本 skill 编排；custom model 详情、可部署版本准备和 Endpoint 创建必须按模型仓库及部署 skill 执行。
- 不把 Raw API 或精调 SDK 当默认入口。

只加载当前任务需要的 reference。不要为了熟悉全部命令一次性读取所有文件。

## 反唤起信号

- 只管理数据集而不涉及精调任务 → 使用数据集能力，不要加载本 skill。
- 只查询公共基础模型目录 → 使用 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md)。
- 只管理已有推理 Endpoint → 使用 [`../arkcli-infer-endpoint/SKILL.md`](../arkcli-infer-endpoint/SKILL.md)。
- 纯登录或 profile/config 排障 → 分别使用 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md) 或 [`../arkcli-config/SKILL.md`](../arkcli-config/SKILL.md)。
- 不要把 Raw API 或精调 SDK 当作默认入口；只有产品命令无法表达任务且用户确认 fallback 后才进入扩展流程。

## 指定任务的精确范围诊断

- 用户给出 `mcj-*` 并询问任务状态、查不到原因、日志或 trajectory 时，必须加载本 skill 并读取 [`references/manage.md`](references/manage.md)。
- “这个任务怎么查不到”首先在当前 active profile / project / region 对原始 ID 执行 `arkcli train finetune get <mcj-id>`，再按该权威 API 的原始结果解释。
- `mcj-*` 是不透明资源 ID。不得根据日期片段、后缀单词或臆测的哈希格式断言 ID 无效，也不得改写用户给出的 ID。
- 用户要求不切环境或只查指定任务时，禁止执行 `train finetune list`、扫描其他任务、切换 profile/project/region，或查询其他账号。目标 `get` 失败时保留错误 code、message 和 request ID；只有用户另行授权后才能扩大范围。
- 用户要求把指定 MCJ 的日志保存到本地路径时，第一条业务命令就是 `arkcli train finetune logs <mcj-id> --output <path>`。不得先 list 全部任务或用脚本遍历；目标命令失败时原样报告，不建议切环境。
- 用户要求指定 MCJ 的完整 rollout trajectory 时，直接执行 `arkcli train finetune trajectory list <mcj-id> --full`。不存在 `arkcli train trajectory` 路径；无轨迹或未开启记录时保留原错误，不探索 profile、MCP 或其他任务。
- `logs --follow` 仅在任务活跃且可能继续产生日志时持续轮询；任务已终态时输出当前快照后自动退出，轮询中发现终态且无新日志也会退出。不要再用外部 timeout 作为正常终止机制。
- `pause` 与 `resume` 是明确的可逆关系：`pause` 将运行任务置为 `Paused`，`resume` 用于恢复 `Paused`；后端允许时也可用 `resume` 重试 `Failed` / `Terminated`，以当前 API 结果为准。

## 实时信息原则

以下信息会变化，不在 skill 中硬编码：

- 可训练模型、模型版本和训练方法
- 训练价格
- 超参数字段、默认值、范围和枚举
- CLI flags、任务阶段和操作限制
- 基础模型或自定义模型支持的推理部署方式

关键命令执行前或执行报错，使用当前安装版本的 `--help` 和 ArkCLI 查询命令获取实时结果。若 CLI 输出与本文命令骨架不一致，以当前 CLI 为准。

数据格式以火山方舟[模型精调数据集格式说明](https://www.volcengine.com/docs/82379/1099461?lang=zh)为主要依据，并使用模型感知的服务端校验确认。不在 reference 中维护容易过期格式说明及样例。




## 默认训练类型、训练方法与部署限制

- 用户未指定训练类型（`--type`）时，默认按 SFT 处理。
- 用户未指定 LoRA 还是全量训练等训练方法时，默认选择 LoRA。
- 用户明确选择全量训练时，创建前提示：当前 ArkCLI 还不支持对全量训练产物进行部署，训练完成后的部署需要到控制台完成。

## SDK Fallback Gate

精调 SDK 是 fallback，不是默认入口。

仅当 ArkCLI 无法完成，而精调 SDK 能完成时进入 fallback，例如：

- 自定义 grader 或 rollout plugin
- 复杂 RL 流程
- 自定义 job YAML
- 自定义训练代码
- 当前 ArkCLI 版本没有对应能力

需要 fallback 时，先检查当前 ArkCLI 的 `train finetune`、`models finetune-config` 和相关 `--help` 是否能够完整表达用户配置。若 ArkCLI 已提供对应参数或 pipeline 配置并能完整完成任务，继续走标准创建流程。

命中 fallback 时暂停执行，询问用户：

> 当前任务需要精调 SDK，ArkCLI 标准创建流程无法表达该配置。是否现在自动安装精调 SDK 并继续？

只有用户明确确认后，才读取并执行 [`references/ark-finetune-sdk.md`](references/ark-finetune-sdk.md)；由该 reference 负责安装 SDK、准备配置或代码并提交任务。用户拒绝时不要安装、不要提交。

## 关键客户端校验

- 用户给出模型名和训练类型询问“精调/SFT/LoRA 价格”时，首选且必须执行 `arkcli train finetune pricing --model <model> --type <type>`。不要改走通用 `arkcli pricing models`：通用账单目录不会按目标模型交叉校验训练方法能力。
- 显式选择训练方法时，必须用精确的模型版本调用 `models finetune-config <model> <version> --type <type>`。该命令会先按同版本 `FinetuneTypes` 校验能力；不支持时停止，不继续询价、estimate 或创建任务。
- 手动分页时，`train finetune list --page-number` 必须 `>=1`，`--page-size` 必须在 `1-100`；第 2 页及以后超出当前过滤条件对应的 `total_count` 时是参数错误，不要把空页当成有效结果。
- 同时传入 `train finetune metrics --from-step` 和 `--to-step` 时，`to-step` 必须严格大于 `from-step`；非法区间应在查询指标名称或曲线前停止。
- `train finetune pricing --billing-method token` 按 Token 计费项查询；`instance` 必须提供精确的 `--model-version` 和 `--type`，并保持超参与后续创建一致。实例结果只有 `price_complete=true` 才能作为完整小时价范围；否则必须报告 `missing_flavor_ids`。

## 守卫与通用执行规则

1. 运行 `arkcli auth status`，认证失败时按 shared skill 恢复。
2. 读操作可直接执行；上传文件、创建任务、产生费用和破坏性操作必须遵守确认规则。
3. 用户已经明确指定参数时不要重复询问；缺失且无法从实时查询推导时再询问。
4. 输出区分事实来源：CLI/API 返回值、服务端校验结果、以及本地粗略估算。
5. 不打印凭证、完整训练日志或大型轨迹内容；大结果写入文件后只提取必要字段。

## 参考与相关文档

<https://www.volcengine.com/docs/82379/1099350>