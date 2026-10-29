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
               过滤 (Endpoint.ModelReference.FoundationModel.name
               → GetFoundationModel → task_types/filter_task_types;
               unknown 保留但标 modality_confidence=unknown 并在 stderr 告警)
  agent-plan
    text     → ListAgentPlanVirtualModelID (Edition=personal, BizInfo=PlanTier)
    image    → 同接口 + ModelCapability.OutputModalities 含 image
    video    → 同接口 + ModelCapability.OutputModalities 含 video
  coding-plan
    text     → ListCodingPlanVirtualModelID (Edition=personal, BizInfo=PlanTier)
    image    → 借道 platform ListEndpoints + modality filter (S10, commit f69be53)
    video    → 同 image
  agent-plan-team
    text     → ListAgentPlanVirtualModelID (Edition=enterprise, BizInfo=席位档位)
    image/video → 同接口 + OutputModalities 过滤；凭证语义是 team_seat
  coding-plan-team
    text     → ListCodingPlanVirtualModelID (Edition=enterprise, BizInfo=席位档位) + team_seat
    image/video → platform Endpoint；team_seat 不可调用，required_overrides=["api_key"]
```

## Endpoint 模态解析契约

`resources list` 对 platform Endpoint 使用控制面的结构化关系：

```text
Endpoint ID
  -> ModelReference.FoundationModel.name
  -> GetFoundationModel
  -> task_types / filter_task_types
  -> text | image | video | unknown
```

- Endpoint 名称、FoundationModel `name` 和 DisplayName 都是标识或展示字段，不是能力证据。
- 国内 `seedream` / `seedance` 前缀可以作为人类示例，但禁止作为代码或 Skill 的 image/video 路由规则。
- 无法取得模型元数据、任务类型未知或结果冲突时标记为 `unknown`。列表会保留该 Endpoint 并提示人工核对；设为模态默认值和真正发起生成时仍需通过权威解析或显式 `--modality`，不能静默猜测。

## 输出形态

```json
{
  "profile": "<name>",
  "type": "<platform | agent-plan | coding-plan>",
  "modality": "<text|image|video>",
  "items": [
    {
      "id": "<id>",
      "resource_kind": "<endpoint|model>",
      "data_plane": "<platform|agent_plan|coding_plan>",
      "credential_kind": "<paygo|agent_plan|team_seat>",
      "invocable": true,
      "is_default": true
    }
  ],
  "current_default": "<id-or-empty>",
  "item_count": <n>
}
```

`is_default` 仅表示 items[].id 跟 profile.yaml 的 default 匹配。它不保证当前凭证
可调用；`invocable=false` 时读取 `required_overrides`。

## resolve Endpoint

用户只给 Endpoint，或需要确认它该走 `+chat` / `+understand` / `+gen` 时：

```bash
arkcli resources resolve <endpoint-id> --format json
```

输出示意：

```json
{
  "resource_id": "ep-...",
  "resource_region": "cn-beijing",
  "resource_kind": "endpoint",
  "data_plane": "platform",
  "endpoint_model_type": "CustomModel",
  "model_id": "cm-...",
  "model_name": "<custom-model-name>",
  "custom_model_id": "cm-...",
  "base_model_id": "<foundation-model-id>",
  "base_model_name": "<foundation-model-name>",
  "base_model_version": "<foundation-model-version>",
  "input_modalities": ["text"],
  "output_modalities": ["text"],
  "supported_workflows": ["chat"],
  "generation_modality": "unknown",
  "requires_user_intent": false,
  "resolution_source": "endpoint_custom_model_and_foundation_metadata"
}
```

- 解析必须先读 `endpoint_model_type`，不能因为 Custom Model Endpoint 同时携带
  FoundationModel 引用就把后者当成真实绑定。
- Custom Model Endpoint 中，`model_id == custom_model_id` 是实际绑定身份；
  `base_model_*` 仅表达训练 lineage，并为 task/API/modality 提供权威能力元数据。
- Foundation Model Endpoint 继续使用原契约：`model_id` 是基础模型 ID，不输出
  `custom_model_id`。
- `resource_region` 是 Endpoint 权威 region，可用于 Endpoint + API Key 时派生 Base URL。
- `requires_user_intent=true` 表示同一资源可服务多个候选工作流，最终由用户任务选择。
- `resolution_warnings` 非空或模态为 `unknown` 时，保留已知绑定身份，但不要从
  Endpoint ID、Custom Model 名或基础模型名猜能力与工作流。

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

临时 API Key / Base URL / Endpoint 的执行决策见
[`../../arkcli-shared/references/execution-context.md`](../../arkcli-shared/references/execution-context.md)。
