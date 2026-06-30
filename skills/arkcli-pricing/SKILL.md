---
name: arkcli-pricing
version: 1.0.0
description: "查询火山引擎 ARK 基础模型结算单价（含当前账号折扣）以及 AgentPlan / CodingPlan 套餐订阅价格。Price 字段就是后端按账号合同 / 活动 / 套餐折后的最终单价，OriginalPrice 是公示原价。当用户问模型多少钱、定价、单价、价格、Agent Plan 多少钱、Coding Plan 多少钱、套餐价格、折扣价、按 token 收费、不同模态价格对比、模型免费额度时使用。反触发：TTS/ASR/语音模型费用不支持查询，不要用 Audio pricing，只能转 models search 说明广场发现边界。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli pricing --help"
---

# arkcli pricing

**CRITICAL — 开始前 MUST 用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md),其中包含认证闸门、配置排查与共享安全规则。**

## 适用场景

- 查询基础模型(foundation model)的结算单价（输入 / 输出 / 批量推理 / 上下文缓存等多种 ChargeItem)
- 查看免费额度(`InferenceFreeUsage`、`ResourcePackItems`)和欠费状态(`IsOverdue`)
- 查看子服务开通态(`SubServices`:`base` / `context-cache` / `fast-infer`)
- 按模态过滤价格目录
- 查询 **AgentPlan / CodingPlan 套餐订阅价格**(个人版 + 企业版,共 12 档)

## 业务定位

- **两类定价分开管**:**模型按量计费**(pricing models)和 **套餐订阅价**(pricing plans)是两个完全独立的命令。决策见下方"模型 vs 套餐"段
- 仅覆盖 **foundation 模型(基础模型)** 的目录式定价。但 base 模型的 ChargeItems **是全口径**——既包含基础模型推理价,也包含用它做精调时的训练费率(`Finetune` / `LoraFinetune`),以及精调出来的 custom model 调用时的推理价(`FinetuneInference*` / `FinetuneI2I*` 等)。所以查 base 模型一次就能拿到所有相关费率。
- **后端 API** 不返回 custom 模型(精调出来的 `cm-*`)单独条目——费率本来就嵌在 base 模型 ChargeItems 数组里。但 **arkcli 已经做了反查糖衣**:`pricing models --model cm-xxx` 会自动调 GetCustomModel 拿到 base name 后再查价,响应顶层多挂一个 `resolved_custom_model` 字段(`customization_type` + `foundation_model_name` / `foundation_model_version`),Agent 据此挑对应 ChargeItem.Type 即可,不需要手动做两段式
- 不覆盖 Endpoint 维度的实际消耗 → 用 [arkcli-usage](../arkcli-usage/SKILL.md)
- **不含限速 / Token 上限 / context window**:这些是 `arkcli-models` 的范畴,转 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md)。pricing 只管"钱"
- **不覆盖语音模型费用**：TTS / ASR / 配音 / 朗读 / 播客 / 音色 / 实时语音交互，或 `doubao-seed-tts-*` / `doubao-seed-asr-*` / `seedasr-*` 等广场语音模型，当前在 arkcli 只支持 `models search` 发现；不要用 `pricing models --model` 或 `--modality Audio` 回答其价格
- **Price = 含当前账号折扣的最终单价**(后端已按合同/活动/套餐计算),不需要再做客户端折扣计算
- **OriginalPrice = 公示原价**;`Price < OriginalPrice` 即说明账号有折扣

## 命令

| 命令 | 说明 |
|------|------|
| [`pricing models`](references/arkcli-pricing-models.md) | 列基础模型结算单价目录(按量计费、按 token 收费) |
| [`pricing plans`](references/arkcli-pricing-plans.md) | 列订阅套餐价格(AgentPlan / CodingPlan 个人版 + 企业版) |

## 模型 vs 套餐 — 用哪个命令

**用户问的是"按 token 算的单价"还是"包月套餐价"?这是关键判别点。**

| 用户措辞 | 该用 | 理由 |
|---|---|---|
| "DeepSeek-V4 多少钱" / "doubao-seed 多少钱" / "GPT-X 价格" | `pricing models` | 模型名 → 按 token 计费 |
| "图生图多少钱一张" | `pricing models` | 按调用计费,在 ChargeItems 里 |
| "音频识别每分钟多少" / "TTS 多少钱" / "语音模型费用" | `arkcli-models` | 当前 arkcli 仅支持语音模型广场发现，不支持费用查询 |
| "做精调多少钱" / "精调一个模型多少 token 钱" | `pricing models` | 精调费率在 base 模型 ChargeItems 内嵌(看 `Finetune` / `LoraFinetune` type) |
| "我的 cm-xxx 调用多少钱" / "这个精调出来的模型多少钱" | `pricing models --model cm-xxx` | CLI 自动反查 base + 透出 `resolved_custom_model.customization_type`,按下表抽 ChargeItem |
| "Agent Plan 多少钱" / "Agent Plan 个人版 small 多少钱" | `pricing plans` | 套餐名 → 包月订阅 |
| "Coding Plan 多少钱" / "lite / pro 套餐价" | `pricing plans` | 同上 |
| "Agent Plan 企业版价格" / "团队版多少钱" | `pricing plans --edition enterprise` | 同上 |

混淆兜底:用户说"X 多少钱"且 X 是 model name → models;X 是 plan name → plans;两者都不像 → 先 `pricing plans` 列出 12 个 plan 看是否命中,再回 models。

## --modality 取值映射(Agent 侧负责翻译)

后端 `FoundationModelDomain` 字段是**5 个粗粒度枚举**,arkcli 直接透传不做翻译。当用户用自然语言说"video / 图像 / 文本"时,你(Agent)需要先翻译再传:

| 用户可能说 | 后端真实枚举 | 注意 |
|---|---|---|
| 文本 / text / LLM / 语言模型 | `LLM` | |
| 图像 / 图片 / image / 视觉 / vision | `ComputerVision` | |
| 视频 / video | `ComputerVision` | ⚠️ 后端不区分 image/video/3D,**返回包含 image 和 3D 模型** |
| 3D / hyper3d | `ComputerVision` | 同上 |
| 音频 / audio / 语音 / voice / TTS / ASR | 不查询 | `Audio` 是后端枚举，但当前 arkcli skill 不用它回答广场语音模型费用；只说明不支持 |
| 向量 / embedding / 嵌入 | `Embedding` | |
| 路由 / router | `Router` | |

**重要**:用户说"video 多少钱"时,你应该:
1. 调 `arkcli pricing models --modality ComputerVision`(不能直接 `--modality video`)
2. 在响应中筛出 video 类模型(看 `DisplayName` 或 `FoundationModelTag`),或显式告诉用户"后端按视觉大类返回,包含 image / video / 3D"

## ChargeItem.Type 关键枚举对照(响应解析必备)

每个 base 模型的 `ChargeItems` 是**全口径**——可能包含基础推理、精调训练、精调推理、各类生成任务的费率混在一起。Agent 按用户问题只挑相关 1-2 条。

| 用户问题语义 | 抽哪条 Type |
|---|---|
| 推理多少钱 | `InferencePrompt` + `InferenceCompletion` |
| 批量推理多少钱 | `BatchInferencePrompt` + `BatchInferenceCompletion` |
| 上下文缓存多少钱 | `ContextSessionHit` + `ContextSessionStorage` |
| 做精调多少钱 | `Finetune`(全量) 或 `LoraFinetune`(LoRA) |
| 精调出来的模型调用多少钱 | `FinetuneInferencePrompt` + `FinetuneInferenceCompletion` |
| 图生图多少钱(base 模型) | `T2ICompletion` / `I2ICompletion` |
| 图生图多少钱(精调模型) | `FinetuneT2ICompletion` / `FinetuneI2ICompletion` |
| 视频生成 | `T2VCompletion` / `I2VCompletion` / `V2VCompletion`(及 1080/4K 变体) |
| 3D 生成 | `To3DCompletion` / `FinetuneTo3DCompletion` |
| 低延迟推理 | `FastInferencePrompt` / `FastInferenceCompletion` |

完整 Type 枚举见 [`references/arkcli-pricing-models.md`](references/arkcli-pricing-models.md) 的「ChargeItem.Type 全枚举对照」段落。

## CustomizationType → ChargeItem.Type 映射

`pricing models --model cm-*` 的响应顶层会带 `resolved_custom_model.customization_type`,据此挑费率 — 完整对照表见 [`references/arkcli-pricing-models.md`](references/arkcli-pricing-models.md) 的「Custom model(cm-*)反查」段落。

## 快速决策

- 用户问"X 模型多少钱":`arkcli pricing models --model X`,从 `ChargeItems` 抽 `InferencePrompt` / `InferenceCompletion`
- 用户问"我的 cm-xxx 多少钱":`arkcli pricing models --model cm-xxx`,看响应 `resolved_custom_model.customization_type` → 按上面映射表抽 `Finetune*` 那条;`resolved_custom_model.foundation_model_name` 顺带告知用户它的 base 是什么
- 用户问"所有 LLM 价格":`arkcli pricing models --modality LLM`
- 用户问"TTS / ASR / 语音模型多少钱":不执行 pricing；转 `arkcli models search <keyword>` 只说明可发现与当前不支持费用查询
- 用户问"Agent Plan / Coding Plan 多少钱":`arkcli pricing plans` 列全 12 档,或带 `--plan agent-plan-personal-small` 单查
- 用户问"我有什么免费额度":`pricing models`,看 `InferenceFreeUsage` 和 `ResourcePackItems[Type=FreeInference]`
- 用户问"我账号有折扣吗":对比 `Price` vs `OriginalPrice`,差值即折扣(两个命令都支持)
- 用户问"我的实际消耗":这不是 pricing 的事,转 [arkcli-usage](../arkcli-usage/SKILL.md)
- **企业版套餐没权限**:`pricing plans` 会把那 6 行 enterprise 的 `error` 字段填成 `AccessDenied: ...`,其余 6 行 personal 正常返。这是预期行为,不是 bug

## 常见降级

- 鉴权错误:转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
- 用户在问消耗(已发生的用量)而非单价:转 [`../arkcli-usage/SKILL.md`](../arkcli-usage/SKILL.md)
- 用户在问语音模型费用 / TTS 定价 / ASR 价格:转 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md) 只说明广场可搜和当前 arkcli 不支持费用查询

## 参考

- [arkcli-usage](../arkcli-usage/SKILL.md) — 实际 token / 请求消耗(含历史范围)
- [arkcli-deploy](../arkcli-deploy/SKILL.md) — 创建 Endpoint(部署前自动检测开通)
- [arkcli-shared](../arkcli-shared/SKILL.md) — 认证和全局参数
