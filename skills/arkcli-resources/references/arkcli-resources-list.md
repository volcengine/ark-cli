# resources list 详细参考

> **前置**：先读 [`../SKILL.md`](../SKILL.md)。

## Flag 一览

| 参数 | 必填 | 默认 | 说明 |
|------|------|------|------|
| `--profile` | 否 | active | 显式 target profile；用 P0-A 修正后的 `RebuildForProfile` 切身份打控制面 |
| `--modality` | 否 | `text` | `text` / `image` / `video` |

## 派发逻辑

```
profile.Type:
  platform   → ListEndpoints (PageAll) 列所有 active endpoint, 按 modality
               过滤 (text=非 ContentGeneration; image=seedream-*; video=seedance-*;
               custom model 走 unknown 放行避免错杀, 见 internal/service/
               inferendpoint/endpoint_modality.go)
  agent-plan
    text     → ListAgentPlanLatestModel
    image    → AgentPlanImageModels  (硬编 console snapshot)
    video    → AgentPlanVideoModels  (硬编 console snapshot)
  coding-plan
    text     → ListArkCodeLatestModel (必传 AccountID, 从 SSO 派生)
    image    → 借道 platform ListEndpoints + modality filter (S10, commit f69be53)
    video    → 同 image
```

## 输出形态

```json
{
  "profile": "<name>",
  "type": "<platform | agent-plan | coding-plan>",
  "modality": "<text|image|video>",
  "items": [
    {"id": "<id>"},
    {"id": "<id>", "is_default": true}
  ],
  "current_default": "<id-or-empty>",
  "item_count": <n>
}
```

`is_default` 仅在 items[].id 跟 profile.yaml 的 `Resources.<modality>.Default` 匹配时出现。

## 跟旧 list 形态的差异（0.1.16 前）

0.1.16 把原 `profile.AvailableTextModels` / `AvailableImageModels` / `AvailableVideoModels` 静态列表全部移除（S2），全部走 `resources list` 实时拉。这意味着：

- 老脚本里读 `profile.yaml.available_text_models` 字段的逻辑全部失效
- Agent 想知道"可用模型清单"必须跑 `resources list`，不能假设 profile.yaml 里有
- profile.yaml 现在只持 default（`Resources.<modality>.Default`），不持 available list

## 跟 `arkcli api ListEndpoints` 的区别

| 维度 | `arkcli resources list` | `arkcli api ListEndpoints` |
|------|------------------------|----------------------------|
| 入口 | 产品命令（shortcuts） | Raw API explorer |
| 派发 | 按 profile.Type 分流 endpoint / plan / coding | 永远 ListEndpoints |
| 输出 | 极简 `[{"id": ...}, ...]` | 完整 Endpoint 全字段 |
| identity scope | `RebuildForProfile` 切 target 身份 | 用 active profile |

Agent 优先 `arkcli resources list`，只在需要 Endpoint 全字段（status / quota / created_at）时回退 `arkcli api ListEndpoints --params ...`。
