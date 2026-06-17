---
name: arkcli-train-finetune
description: 使用 ArkCLI 创建、查询和管理模型精调训练任务，并从训练指标选择最佳 step、导出训练产物为 custom model、衔接模型仓库与推理部署。适用于选择可训练模型与训练方法、查询价格和超参数、校验训练文件、预览并创建任务、列出任务、观察或操作指定任务，以及根据效果指标导出和部署精调产物。本 skill 不负责数据集管理。
---

# ArkCLI 精调训练

先读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，遵循认证、输出、安全和二次确认规则。

## 能力边界

- 创建精调任务：读取 [`references/create.md`](references/create.md)
- 列出或筛选任务：读取 [`references/list.md`](references/list.md)
- 查询、观察或操作一个指定任务：读取 [`references/manage.md`](references/manage.md)
- 根据指标选择 step、导出产物并部署：读取 [`references/export-deploy.md`](references/export-deploy.md)
- 不管理数据集，可以使用用户提供的本地文件、TOS URL。
- 训练产物的指标分析和 artifact export 由本 skill 编排；custom model 详情、可部署版本准备和 Endpoint 创建必须按模型仓库及部署 skill 执行。
- 不把 Raw API 或精调 SDK 当默认入口。

只加载当前任务需要的 reference。不要为了熟悉全部命令一次性读取所有文件。

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

当需要fallback时，先检查当前 ArkCLI 的 `train finetune`、`models finetune-config` 和相关 `--help` 是否能够完整表达用户配置。若 ArkCLI 已提供对应参数或 pipeline 配置并能完整完成任务，继续走标准创建流程。

命中 fallback 时暂停执行，询问用户：

> 当前任务需要精调 SDK，ArkCLI 标准创建流程无法表达该配置。是否现在自动安装精调 SDK 并继续？

只有用户明确确认后，才读取并执行精调 SDK Skill。当前配套入口为 [`references/ark-finetune-sdk.md`](references/ark-finetune-sdk.md)；由该 skill 负责安装 SDK、准备配置或代码并提交任务。用户拒绝时不要安装、不要提交。

## 通用执行规则

1. 先运行 `arkcli auth status`，认证失败时按 shared skill 恢复。
2. 读操作可直接执行；上传文件、创建任务、产生费用和破坏性操作必须遵守确认规则。
3. 用户已经明确指定参数时不要重复询问；缺失且无法从实时查询推导时再询问。
4. 输出区分事实来源：CLI/API 返回值、服务端校验结果、以及本地粗略估算。
5. 不打印凭证、完整训练日志或大型轨迹内容；大结果写入文件后只提取必要字段。

## 相关文档

<https://www.volcengine.com/docs/82379/1099350>
