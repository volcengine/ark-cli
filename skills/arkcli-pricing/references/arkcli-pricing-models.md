# pricing models

> **前置条件:** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

列出 ARK 基础模型(foundation model)的结算单价目录。`Price` 是后端按当前账号合同 / 活动 / 套餐计算后的**最终单价(含折扣)**;`OriginalPrice` 是公示原价。

## 命令

```bash
# 列全量基础模型(默认 PageSize=50, PageNumber=1)
arkcli pricing models

# 单基础模型查询
arkcli pricing models --model deepseek-v4-pro

# 单 custom model 查询(精调出来的 cm-*)
# CLI 会先调 GetCustomModel 反查 base,再走标准目录查询;
# 响应顶层多挂一个 resolved_custom_model 字段
arkcli pricing models --model cm-20260611124416-77wrq

# 按模态(后端 FoundationModelDomain 枚举)过滤
arkcli pricing models --modality LLM
arkcli pricing models --modality ComputerVision
arkcli pricing models --modality Audio
arkcli pricing models --modality Embedding

# 翻页
arkcli pricing models --page-size 20 --page-number 2

# 指定价格区域(默认走当前 profile region)
arkcli pricing models --region cn-beijing
```

## 参数

### 可选过滤

| 参数 | 类型 | 说明 |
|------|------|------|
| `--model` | string | 单基础模型名(如 `doubao-seed-1-6` / `deepseek-v4-pro`)或 custom model id(`cm-*`)。`cm-*` 由 CLI 自动反查 base,响应附带 `resolved_custom_model` 字段 |
| `--modality` | string | 后端 `FoundationModelDomain` 枚举,**不接受 alias**:`LLM` / `ComputerVision` / `Audio` / `Embedding` / `Router` |
| `--region` | string | 价格区域,默认沿用 profile 当前 region |
| `--page-number` | int | 页码,1-based,默认 1 |
| `--page-size` | int | 每页条数,默认 50 |

### --modality 用法

后端 `FoundationModelDomain` 字段**不区分** image / video / 3D,它们都在 `ComputerVision` 下。

| 用户语义 | 应传值 | 备注 |
|---|---|---|
| 文本 / 大语言模型 | `LLM` | |
| 图像生成 / 图片 | `ComputerVision` | 包含 image / video / 3D |
| 视频生成 | `ComputerVision` | 同上,需要客户端二次过滤 |
| 音频 / 语音 | `Audio` | TTS / ASR 等 |
| 向量 | `Embedding` | |
| 智能路由 | `Router` | |

⚠️ 不要传 `text` / `video` / `image` 这种自然语言别名,arkcli 会原样透传给后端,而后端只认 5 个枚举值——会得到空结果或报错。

## 返回值

JSON 结构(节选关键字段):

```json
{
  "page_number": 1,
  "page_size": 50,
  "total_count": 75,
  "items": [
    {
      "FoundationModelName": "deepseek-v4-pro",
      "DisplayName": "DeepSeek-V4-pro",
      "VendorName": "DeepSeek",
      "State": "Available",
      "ChargeItems": [
        { "Type": "InferencePrompt",     "Price": 0.012, "OriginalPrice": 0.012, "UnitCode": "千tokens" },
        { "Type": "InferenceCompletion", "Price": 0.024, "OriginalPrice": 0.024, "UnitCode": "千tokens" },
        { "Type": "ContextSessionHit",   "Price": 0.001, "UnitCode": "千tokens" },
        { "Type": "ContextSessionStorage","Price": 0.000017, "UnitCode": "千tokens/小时" },
        { "Type": "BatchInferencePrompt","Price": 0.006, "UnitCode": "千tokens" },
        { "Type": "BatchInferenceCompletion","Price": 0.012, "UnitCode": "千tokens" }
      ],
      "InferenceFreeUsage": { "Total": 500000, "Consumed": 0 },
      "ResourcePackItems": [
        { "Type": "FreeInference", "Total": 500000, "Consumed": 500000 }
      ],
      "SubServices": [{ "SubService": "base", "Status": "Available" }],
      "IsOverdue": false
    }
  ]
}
```

## 关键字段口径

| 字段 | 含义 |
|---|---|
| `Price` | **当前账号折后单价**(后端已结算合同/活动/套餐),Agent 直接展示这个值 |
| `OriginalPrice` | 公示原价。`Price < OriginalPrice` 即有折扣;差值/比例可用作"折扣力度"说明 |
| `UnitCode` | 计费单位,常见 `千tokens`,存储类是 `千tokens/小时` |
| `Type` | ChargeItem 类型(见下方表) — 通常向用户报推理价只展示 `InferencePrompt` + `InferenceCompletion` |
| `State` | 模型在当前账号下是否可用(`Available` / `Unavailable`) |
| `InferenceFreeUsage` | 免费推理额度,`Total` 总数,`Consumed` 已消耗 |
| `ResourcePackItems` | 资源包消耗,`FreeInference` type 的就是免费推理资源包 |
| `SubServices` | 子服务开通态(`base` / `context-cache` / `fast-infer`),关联 [arkcli-deploy](../../arkcli-deploy/SKILL.md) 的 `--moderation` 和 caching 行为 |
| `IsOverdue` | true 即欠费,影响推理调用成败 |
| `MultiChargeItems` | 阶梯计费(按 token 长度分档),少数模型有 |
| `resolved_custom_model` | **仅当 `--model` 传 cm-* 时出现**(顶层字段,不在 items 里),见下节 |

## Custom model(cm-*)反查

后端 `ListModelChargeItems` 不索引 custom 模型条目——精调出来的 cm-* 费率本来就嵌在 base 模型 ChargeItems 里(以 `Finetune*` 前缀的 type 表示)。CLI 替使用者做了反查糖衣:

**触发条件**:`--model` 传任何以 `cm-` 开头的字符串。

**执行链路**:
1. CLI 先调 `custom_model.get_custom_model { Id: "cm-xxx" }`
2. 从结果取 `FoundationModel.Name` / `FoundationModel.ModelVersion` / `CustomizationType`
3. 用 base name 走标准 `ListModelChargeItems`
4. 响应 `Items` 是 base 模型的全 ChargeItems(**不裁剪**),顶层多一个 `resolved_custom_model`

**响应增量字段**:

```json
{
  "resolved_custom_model": {
    "id": "cm-20260611124416-77wrq",
    "customization_type": "FinetuneSft",
    "foundation_model_name": "doubao-seed-1-6-flash",
    "foundation_model_version": "250615"
  },
  "page_number": 1,
  "page_size": 50,
  "total_count": 1,
  "items": [ /* base 模型全 ChargeItems,Agent 自行挑 Finetune* */ ]
}
```

**Agent 挑 ChargeItem 的对照(按 `customization_type`)**:

| customization_type | 调用费(询问 cm-* 推理多少钱时抽这两条) | 训练费(询问怎么训出来的多少钱时抽这条) |
|---|---|---|
| `FinetuneSft`(全量 SFT) | `FinetuneInferencePrompt` + `FinetuneInferenceCompletion` | `Finetune` |
| `FinetuneLoRA` / `DPOLoRA` / `GRPOLoRA` / `OPDLoRA` | `FinetuneInferencePrompt` + `FinetuneInferenceCompletion` | `LoraFinetune` |
| `DPO` / `GRPO` / `PPO` / `OPD`(全量 RL) | `FinetuneInferencePrompt` + `FinetuneInferenceCompletion` | `Finetune` |
| `Pretrain` | `FinetuneInferencePrompt` + `FinetuneInferenceCompletion` | `Finetune` |

> 视觉/视频精调出来的 cm-* 调用费分别走 `FinetuneT2ICompletion` / `FinetuneI2ICompletion` / `FinetuneT2VCompletion` / `FinetuneI2VCompletion` / `FinetuneTo3DCompletion` 等,具体看下方 ChargeItem.Type 全枚举对照。

**反查失败的错误**:
- `cm-xxx` 不存在 → `GetCustomModel "cm-xxx": ... The specified CustomModel cm-xxx is not found.`
- 无权限 → 透传 `AccessDenied`,转 [arkcli-auth](../../arkcli-auth/SKILL.md)

## ChargeItem.Type 全枚举对照

base 模型的 `ChargeItems` 数组**是全口径的**——下面所有 type 都可能出现在同一个 base 模型节点下,Agent 按用户问题语义抽对应那条。

### 推理类(直接调用 base 模型)

| Type | 含义 |
|---|---|
| `Inference` | 推理整体均价(部分模型给) |
| `InferencePrompt` | 推理输入价 |
| `InferenceCompletion` | 推理输出价 |
| `BatchInferencePrompt` | 批量推理输入价 |
| `BatchInferenceCompletion` | 批量推理输出价 |
| `BatchInferenceCacheHit` | 批量推理 prompt cache 命中价 |
| `ContextSessionHit` | Context Cache 命中价 |
| `ContextSessionStorage` | Context Cache 存储价(单位 `千tokens/小时`) |
| `FastInferencePrompt` / `FastInferenceCompletion` / `FastInferenceCacheHit` | 低延迟推理(子服务 `fast-infer` 开通后生效) |
| `VisionPrompt` / `BatchVisionPrompt` | 多模态视觉输入价 |

### 精调训练类(用 base 做精调时的训练费率)

| Type | 含义 |
|---|---|
| `Finetune` | 全量精调训练价 |
| `LoraFinetune` | LoRA 精调训练价 |

### 精调推理类(精调出来的 custom 模型调用时的推理价)

| Type | 含义 |
|---|---|
| `FinetuneInferencePrompt` | 精调模型推理输入价 |
| `FinetuneInferenceCompletion` | 精调模型推理输出价 |
| `FinetuneI2ICompletion` | 精调模型图生图(image-to-image)输出价 |
| `FinetuneI2VCompletion` | 精调模型图生视频输出价 |
| `FinetuneT2ICompletion` | 精调模型文生图输出价 |
| `FinetuneT2VCompletion` | 精调模型文生视频输出价 |
| `FinetuneTo3DCompletion` | 精调模型 3D 生成输出价 |

### 视觉/多模态生成类(base 模型做生成时的输出价)

| Type | 含义 |
|---|---|
| `T2ICompletion` / `T2ITokenCompletion` | 文生图 |
| `I2ICompletion` / `I2ITokenCompletion` | 图生图 |
| `T2VCompletion` / `I2VCompletion` | 文生视频 / 图生视频 |
| `V2VCompletion` / `V2V1080Completion` / `V2V4KCompletion` | 视频转视频(普通 / 1080p / 4K) |
| `NV2VCompletion` / `NV2V1080Completion` / `NV2V4KCompletion` | 静音视频转视频 |
| `FLF2VCompletion` | 首尾帧生成视频 |
| `To3DCompletion` / `BatchTo3DCompletion` | 3D 生成 |
| `ToVCompletion` / `ToVSilentCompletion` / `BatchToVCompletion` / `BatchToVSilentCompletion` | 通用视频生成(含静音) |
| `BatchV2VCompletion` / `BatchV2V1080Completion` / `BatchV2V4KCompletion` | 批量视频转视频 |

### 音频类

| Type | 含义 |
|---|---|
| `FastAudioPrompt` / `FastAudioCacheHit` | 音频低延迟推理 |

> **重要**: 一个用户问题往往只关心其中 1-2 条 type。常见映射:
> - "推理多少钱" → `InferencePrompt` + `InferenceCompletion`
> - "做精调多少钱" → `Finetune` 或 `LoraFinetune`(乘以 `tokens × epoch`)
> - "精调出来的模型调用多少钱" → `FinetuneInferencePrompt` + `FinetuneInferenceCompletion`
> - "图生图多少钱" → `T2ICompletion` / `I2ICompletion`(基础模型) 或 `FinetuneT2ICompletion` / `FinetuneI2ICompletion`(精调出来的)

## 常见用法

**给用户报某模型的输入/输出价格**

```bash
arkcli pricing models --model doubao-seed-1-6 \
  --transform 'items.0.ChargeItems.#(Type=="InferencePrompt").Price'
```

或不带 transform 拿全量,Agent 自己挑出 `InferencePrompt` / `InferenceCompletion` 两条。

**判断账号是否有折扣**

```bash
arkcli pricing models --model X
# 比对 Price 和 OriginalPrice,Price < OriginalPrice → 有折扣
```

**列某模态全部模型对比价格**

```bash
arkcli pricing models --modality LLM --page-size 100
# Agent 在结果里按 vendor / FoundationModelName 排序汇总
```

## 常见错误

| 错误 | 原因 | 处理 |
|------|------|---------|
| 空 items / total_count=0 | `--modality` 传了 `text` / `video` 等别名而不是后端枚举(没传 `--model` 时才会走到这条) | 翻译成 `LLM` / `ComputerVision` 等再传 |
| `foundation model "X" not found in pricing list` (exit=2) | `--model X` 名字在 pricing list 里查不到 | 检查拼写;或 `arkcli pricing models`(不带 `--model`)拉全量看支持的名字 |
| 模型 State=Unavailable | 当前账号还没开通 | 转 `arkcli +deploy`(自动开通)或 `arkcli models activate` |
| `IsOverdue=true` | 账号欠费 | 不影响价格查询,但会影响推理调用 |
| `GetCustomModel "cm-xxx": ... is not found` | cm- 前缀 id 不存在或不属于当前账号 | 检查 id 拼写;或先 `arkcli api custom_model.list_custom_models` 拉清单 |

## 参考

- [arkcli-pricing](../SKILL.md) — pricing skill 概览
- [arkcli-shared](../../arkcli-shared/SKILL.md) — 认证和全局参数
- [arkcli-usage](../../arkcli-usage/SKILL.md) — 看实际消耗(已经发生的 token / 请求)
