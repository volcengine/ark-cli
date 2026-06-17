---
name: arkcli-resources
version: 1.0.0
description: "arkcli resources 实时控制面查询：按 profile.type 派发，列出当前/指定 profile 下可用的 endpoint / plan 模型 ID 列表。read-only，不写 profile.yaml。change default 走 `arkcli profile set-default` 或 `+deploy --set-default`。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli resources --help"
---

# arkcli resources

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、配置排查与命令选择顺序**

## 使用原则

- `arkcli resources list` 是 read-only 实时控制面查询，**每次都打上游**，没有本地缓存
- 派发逻辑跟 profile.Type 走：platform → `ListEndpoints`，agent-plan / coding-plan → 对应 plan API
- **团队版 `agent-plan-team` / `coding-plan-team` 行为等同对应个人版**：agent-plan-team ≈ agent-plan（自带生文/生图/生视频预置模型）；coding-plan-team ≈ coding-plan（自带生文模型，`--modality image|video` 借道 platform endpoint 池）。资源派发 / base_url / 模型清单跟个人版完全一致，只是凭证来自团队席位
- 这个 skill 不负责改 default —— 用户要换 default 走 [`../arkcli-profile/SKILL.md`](../arkcli-profile/SKILL.md) 的 `profile set-default`
- `--profile X` 真切身份（P0-A 修正）：用 X 的 token / UserID 打控制面，不是 active=A 的身份打完再展示成 B 的资源

## 适用场景

- 用户问"当前 profile 下有哪些 endpoint / 模型可用"
- 用户跑 `profile set-default` 时报 `<id> 不在可用列表`，回这里看真实可用 ID
- 用户跑 `+chat / +gen --model` 报 `InvalidEndpointOrModel.NotFound`，回这里确认 ID 在 active profile 下可见
- 用户切了 profile，想知道新 profile 下的可用资源跟旧的有什么差异

## 反唤起信号

- 用户要 **找模型** / **挑模型** / "哪个模型最强 / 性价比最高" → 转 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md)（带 enrich + 加权排序）
- 用户要 **创建 endpoint** → 转 [`../arkcli-deploy/SKILL.md`](../arkcli-deploy/SKILL.md)（`arkcli +deploy`）
- 用户要 **管理 endpoint**（start / stop / get / update / list 详情）→ 转 [`../arkcli-infer-endpoint/SKILL.md`](../arkcli-infer-endpoint/SKILL.md)

## resources vs models 的区别

| 维度 | `arkcli resources list` | `arkcli models ...` |
|------|------------------------|----------------------|
| Scope | 当前 profile 下"我能用什么" | 全平台基础模型 catalog |
| 输出 | endpoint ID（`ep-xxx`）或 plan 模型名 | foundation_model 全字段 + ArkModels enrich |
| 派发 | 按 profile.type 切 endpoint / plan / coding API | 通用 ListFoundationModel |
| 缓存 | 无 | 有 cache scope（profile/region/project） |
| 主要用途 | 设 default、验 `--model <id>` 是否 active | 找模型、对比模型、确认 capability |

简言之：`resources list` 回答 **"我（当前 profile）能用什么"**，`models` 回答 **"平台上有什么"**。

## Agent 快速执行顺序

1. 不确定当前 profile → `arkcli profile show --format json`（看 `type`）
2. text 资源 → `arkcli resources list --modality text --format json`
3. image / video 资源 → `arkcli resources list --modality image --format json` / `--modality video`
4. 多 profile 对比 → 分别跑 `--profile A --modality text` 和 `--profile B --modality text`
5. 输出里 `is_default: true` 标的是当前 profile 的 default，用户切换 default 走 `arkcli profile set-default`

## 命令一览

| 命令 | 说明 |
|------|------|
| `arkcli resources list` | 列当前/指定 profile 下指定 modality 的可用资源 ID（实时） |

## 输出形态

```json
{
  "profile": "platform_cn-beijing_default",
  "type": "platform",
  "modality": "text",
  "items": [
    {"id": "ep-20260424-aaaaa"},
    {"id": "ep-20260424-bbbbb", "is_default": true},
    {"id": "ep-20260424-ccccc"}
  ],
  "current_default": "ep-20260424-bbbbb",
  "item_count": 3
}
```

`is_default` 仅在 `items[].id == current_default` 时出现；如果 default 是空字符串则 items 全无 `is_default`。

## 常见错误

- coding-plan profile 下 `resources list --modality image|video` 不再 fail-fast (S10): 会借道 platform 控制面 ListEndpoints, 列出同账号已 `+deploy` 的 endpoint id, 用户拿来当 `+gen --model <ep-id>` 或 `profile set-default --modality image <ep-id>`. 列表为空 → 用户在 platform 上还没 deploy, 先 `arkcli +deploy <model>`
- `coding-plan resources list: 缺 AccountID (请先 arkcli auth login)` → 仅 text 路径需要 AccountID; SSO 没登录或 token 解析时 claims.Sub 为空, 重新走 `arkcli auth login volc-sso`
- `ListEndpoints: NotLogin / Unauthorized` → 登录态/STS 过期 / `--profile X` 的 X 没在 identity store 里有 token；先 `auth login`
- `unsupported profile type "X" for resources list` → profile.yaml 被手改成不认识的 type；用 `profile show` 看 `type` 字段，需要 `profile create` 重建

## 参考

- [`../arkcli-profile/SKILL.md`](../arkcli-profile/SKILL.md) — 看完 resources 后要换 default 时进
- [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md) — 找模型 / 对比能力时进
- [`../arkcli-deploy/SKILL.md`](../arkcli-deploy/SKILL.md) — 没看到想要的 endpoint 时进创建链路
