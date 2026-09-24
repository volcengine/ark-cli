---
name: arkcli-resources
version: 1.2.2
description: "arkcli resources 实时控制面查询：列出当前/指定 profile 可见资源及其调用兼容性；把 Endpoint 解析为权威模型、模态与候选工作流。read-only，不写 profile.yaml。用户临时给出 ep-... 但未说明该走 Chat、Understand 还是 Gen 时优先使用。反触发：用户已观察到 Endpoint NotFound，并要判断 ID 是否不完整或仅存在于历史用量时，owning skill 必须是 arkcli-infer-endpoint。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli resources --help"
---

# arkcli resources

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、配置排查与命令选择顺序**
**CRITICAL — 用户已经看到 Endpoint NotFound，又要核对当前完整 ID / 历史用量时，必须先转读 [`../arkcli-infer-endpoint/SKILL.md`](../arkcli-infer-endpoint/SKILL.md)，由其按 `infer endpoint list → usage stats --endpoint` 证据链主导。不得只跑 `resources resolve/list` 就猜测 ID 后缀或宣称已删除。**

## 使用原则

- `arkcli resources` 域始终只读。即使用户在资源列表语境中提出创建意图，也必须先转读 [`../arkcli-deploy/SKILL.md`](../arkcli-deploy/SKILL.md)，不能由本 Skill 创建、修改或删除资源
- 普通“部署模型并获得 Endpoint”的产品意图继续走 `arkcli +deploy`。该在线工作流不支持 Client Preview，**不得执行 `arkcli +deploy --dry-run`**，也不得为了获得 Preview 而静默降级成底层 CRUD
- 转入 Deploy Skill 后，先只读核对模型、名称、Region、配置与计费影响并复述给用户；在收到本轮新的明确确认前，**不得在同一轮执行真实** `arkcli +deploy`。严禁 Agent 自行添加 `--yes`、`echo Y` 或设置 `ARKCLI_ALLOW_HEADLESS_ACTIVATION`
- 只有用户明确要求 raw CRUD、精确 CreateEndpoint 请求或 CI/脚本预览时，才转 [`../arkcli-infer-endpoint/SKILL.md`](../arkcli-infer-endpoint/SKILL.md)，使用叶子命令 `arkcli infer endpoint create ... --dry-run`；Preview 完成后仍需新的确认才能真实执行
- `arkcli resources list` 是 read-only 实时控制面查询，**每次都打上游**，没有本地缓存
- 列表是资源发现结果，可能受 project、创建者和模态过滤；列表为空或当前 default 未出现在 items 中，不等于指定 Endpoint 不存在或不可调用。
- 只读核对身份或默认资源、没有指定精确资源 ID 且列表为空时，报告已确认事实和信息缺口，结束核对；不要为补齐字段转到 `profile show/list/keys list`。空列表不是配置修复或 Key 同步的授权。
- `arkcli resources resolve <ep-id>` 先按 `endpoint_model_type` 判定真实绑定。Custom Model Endpoint 的 `model_id` / `custom_model_id` 必须保持 `cm-...` 身份，基础模型只通过 `base_model_*` 表达 lineage 与能力来源；不按 ID/模型名子串猜用途
- 派发逻辑跟 profile.Type 走：platform → `ListEndpoints`，agent-plan / coding-plan → 对应 plan API
- `agent-plan-team` 三模态使用团队席位 Key + 套餐模型；`coding-plan-team` 只有 text 使用团队席位 Key，image/video 虽可看到 platform Endpoint，但调用还需要后付费 API Key
- `resources list` 区分“账号可见”与“当前 profile 可调用”：读取 `invocable` 与 `required_overrides`，不要看到 ID 就断言当前凭证可用
- `invocable=true` 只说明 Profile/资源类型兼容，不证明 API Key 未失效、具有权限或额度。凭证归属与数据面真实调用结果要分开核验，元数据不能代替请求成功。
- 这个 skill 不负责改 default —— 用户要换 default 走 [`../arkcli-profile/SKILL.md`](../arkcli-profile/SKILL.md) 的 `profile set-default`
- `--profile X` 真切身份（P0-A 修正）：用 X 的 token / UserID 打控制面，不是 active=A 的身份打完再展示成 B 的资源

## 适用场景

- 用户问"当前 profile 下有哪些 endpoint / 模型可用"
- 用户跑 `profile set-default` 时报 `<id> 不在可用列表`，回这里看真实可用 ID
- 用户跑 `+chat / +gen --model` 报 `InvalidEndpointOrModel.NotFound`，回这里确认 ID 在 active profile 下可见
- 用户切了 profile，想知道新 profile 下的可用资源跟旧的有什么差异
- 用户只给出一个 Endpoint，不知道该走 `+chat`、`+understand` 还是 `+gen`
- 用户传了 Endpoint + API Key，但未给 Base URL，需要从 Endpoint 权威 region 派生

## 反唤起信号

- 用户要 **找模型** / **挑模型** / "哪个模型最强 / 性价比最高" → 转 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md)（带 enrich + 加权排序）
- 用户要 **创建 endpoint** → 转 [`../arkcli-deploy/SKILL.md`](../arkcli-deploy/SKILL.md)（`arkcli +deploy`）；本轮只做只读核对、复述和确认，不直接执行
- 用户明确要 **raw CreateEndpoint / CI / 精确请求预览** → 转 [`../arkcli-infer-endpoint/SKILL.md`](../arkcli-infer-endpoint/SKILL.md)，使用 `infer endpoint create --dry-run`
- 用户要 **管理 endpoint**（start / stop / get / update / list 详情）→ 转 [`../arkcli-infer-endpoint/SKILL.md`](../arkcli-infer-endpoint/SKILL.md)
- 用户已经拿到 Endpoint NotFound，想判断 ID 是否少尾部或历史上是否存在 → 转 [`../arkcli-infer-endpoint/SKILL.md`](../arkcli-infer-endpoint/SKILL.md)；不要停在 resources 单点解析

## resources vs models 的区别

| 维度 | `arkcli resources list` | `arkcli models ...` |
|------|------------------------|----------------------|
| Scope | 当前 profile 下可发现资源及兼容性 | 全平台基础模型 catalog |
| 输出 | endpoint ID（`ep-xxx`）或 plan 模型名 | foundation_model 全字段 + ArkModels enrich |
| 派发 | 按 profile.type 切 endpoint / plan / coding API | 通用 ListFoundationModel |
| 缓存 | 无 | 有 cache scope（profile/region/project） |
| 主要用途 | 发现候选、检查 Profile/资源类型兼容性 | 找模型、对比模型、确认 capability |

简言之：`resources list` 提供当前 Profile 下的资源发现与兼容性，`models` 提供平台模型目录；二者都不保证真实数据面调用成功。

## Agent 快速执行顺序

1. 用户给了 `ep-...` → `arkcli resources resolve <ep-id> --format json`，先读 `endpoint_model_type` 与 `model_id`，再读 `supported_workflows` / `generation_modality` / `requires_user_intent`
2. 不确定当前 profile → `arkcli auth whoami --format json`，读取当前身份及可用的 `profile.name` / `profile.type`；字段缺失时按当前编译产品诊断，不猜 type。`profile show/list` 可能同步并回写本地 Key 库存，不作为普通生成前的只读准入。
3. text 资源 → `arkcli resources list --modality text --format json`
4. image / video 资源 → `arkcli resources list --modality image --format json` / `--modality video`
5. 多 profile 对比 → 分别跑 `--profile A --modality text` 和 `--profile B --modality text`
6. 读取每项的 `invocable` / `required_overrides`；`is_default: true` 只表示默认偏好，不保证当前凭证可调用
7. 已明确的当前 default / 用户指定 EP 不在列表时，在同一身份下解析同一个 EP；核对 region、Running 状态、目标 API/工作流/模态及 warnings，再核对 Profile/Key 数据面兼容性。不得绕过明确的不可调用项或 required_overrides；不得仅因列表为空创建 EP，或静默切模型、Profile、default。

## 命令一览

| 命令 | 说明 |
|------|------|
| `arkcli resources list` | 列当前/指定 profile 下指定 modality 的可见资源、数据面与调用兼容性 |
| `arkcli resources resolve <endpoint-id>` | 权威解析 Endpoint 绑定模型、模态、候选工作流与 region |

## 输出形态

```json
{
  "profile": "platform_cn-beijing_default",
  "type": "platform",
  "modality": "text",
  "items": [
    {
      "id": "ep-20260424-aaaaa",
      "resource_kind": "endpoint",
      "data_plane": "platform",
      "credential_kind": "paygo",
      "invocable": true
    }
  ],
  "current_default": "ep-20260424-bbbbb",
  "item_count": 1
}
```

`is_default` 仅在 `items[].id == current_default` 时出现；`invocable=false` 时查看
`required_overrides`（例如 `["api_key"]`）。可见资源不等于当前凭证可调用。

`resources resolve` 的绑定身份契约：

- `endpoint_model_type=CustomModel`：`model_id == custom_model_id == cm-...`；`base_model_*` 仅描述训练 lineage 和能力元数据来源，不得拿 `base_model_id` 替代真实绑定。
- `endpoint_model_type=FoundationModel`：`model_id` 仍是基础模型 ID，既有语义不变。
- `resolution_warnings` 非空时保持 fail-soft，不得根据同时出现的 Custom/Foundation 引用自行猜绑定。

## 常见错误

- coding-plan 的 image/video 使用 platform Endpoint 池，调用前核对后付费 Key。列表为空时检查查询 scope/权限，已有精确目标则先解析；创建资源是另一个用户意图，转 Deploy Skill 处理，不由空列表自动触发。
- `coding-plan resources list: 缺 AccountID (请先 arkcli auth login)` → 仅 text 路径需要 AccountID; SSO 没登录或 token 解析时 claims.Sub 为空, 重新走 `arkcli auth login volc-sso`
- `ListEndpoints: NotLogin / Unauthorized` → 登录态/STS 过期 / `--profile X` 的 X 没在 identity store 里有 token；先 `auth login`
- `unsupported profile type "X" for resources list` → 按共享规则核对当前有效上下文，再转 Config/Profile Skill 排查；不得未经用户意图重建或修改 Profile。

## 参考

- [`../arkcli-profile/SKILL.md`](../arkcli-profile/SKILL.md) — 看完 resources 后要换 default 时进
- [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md) — 找模型 / 对比能力时进
- [`../arkcli-deploy/SKILL.md`](../arkcli-deploy/SKILL.md) — 没看到想要的 endpoint 时进创建链路
- [`../arkcli-shared/references/execution-context.md`](../arkcli-shared/references/execution-context.md) — Profile/凭证/资源矩阵与临时覆盖
