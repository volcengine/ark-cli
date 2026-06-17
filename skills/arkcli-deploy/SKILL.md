---
name: arkcli-deploy
version: 1.2.0
description: "arkcli +deploy：一键创建 Endpoint 的快捷方式。当用户需要把模型部署成在线推理接入点时使用。注意：要对**已有** Endpoint 做获取/列表/启停/更新等全生命周期管理，走 arkcli-infer-endpoint；本 skill 只负责一键创建（创建外的增删改查不在此）。本命令此前会同时生成多语言示例代码，该内嵌子能力暂时下线（依赖的 NodeBFF 接口待迁移），但 Endpoint 创建/启动主流程不受影响。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli +deploy --help"
---

# arkcli +deploy

**前置：** 先用 Read 读 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md) 获取共享认证/配置/写操作守卫规则。

**新增 flag `--set-default <modality>`**: 部署成功后自动把新 endpoint 设为 active profile 该 modality (`text` / `image` / `video`) 的默认资源。仅在 `--dry-run` 之外 + 用户明确传 modality 时生效；失败仅 stderr warn，不阻断部署主流程。详见 [`../arkcli-shared/references/profile-defaults.md`](../arkcli-shared/references/profile-defaults.md)。

**写操作 + 计费**：`+deploy` 创建在线推理 Endpoint 是真实写操作，会产生计费资源。**任何执行前必须先 `--dry-run` 或与用户显式确认**，即使用户语气紧急。

**实名前置（开通类硬闸门）**：`+deploy` 会触发开通模型，**第一步**先 `arkcli auth status` 读 `volc_sso.identity.verified`——`false` 即停并把实名页 `https://console.volcengine.com/user/authentication/detail/` 贴给用户、暂停等待（详见 [`../arkcli-auth/references/realname-gate.md`](../arkcli-auth/references/realname-gate.md)）。**不要**先试 `+deploy`、撞到 model-id 无效 / 模型未开通报错再回头查实名——那些报错会把你带偏。

## 示例代码：独立命令可用 / `+deploy` 内嵌渲染暂下线

区分两条路径，别混为一谈：

- **`+deploy` 创建成功后的内嵌自动渲染** —— **仍暂下线**：依赖的 `GetExampleCodeItems` NodeBFF 接口已停，`+deploy` 不再自动生成 `./ark-examples/<endpoint-id>/` 调用示例，会打印"示例代码暂时不可用"提示。Endpoint 本身正常创建并启动，不受影响。
- **独立命令 `arkcli +code-example`** —— **可用**：已迁到 OpenTOP `OpenGetSampleCode`，需要示例时单独跑它（见下）。

需要给用户调用代码时：
- 独立命令 `arkcli +code-example` **可用**（已迁到 OpenTOP `OpenGetSampleCode`），转 [`../arkcli-code-example/SKILL.md`](../arkcli-code-example/SKILL.md)；注意它按 model-version 提供，部分版本后端无 group 会返回 not found，缺失时再降级到 `ark-examples/` 静态示例或方舟控制台示例代码页
- `+deploy` 自身**不再自动**把示例写进 `./ark-examples/<endpoint-id>/`，需要示例就单独跑上面的命令

## 子命令穷举（只有这一个）

| 调用 | 说明 |
|------|------|
| `arkcli +deploy --name <ep-name> --model <model-id> [...]` | 创建 Endpoint；不加 `--dry-run` 即真实创建 |

> ⚠️ **没有** `arkcli deploy ...` / `arkcli endpoint create` / `arkcli +deploy create` 等子命令。整个能力就是一个 `+deploy` 命令加 flag。

## 反幻觉清单

- `--name`、`--model` 必填
- JSON 类 flag（`--rate-limit` / `--moderation` / `--intelligent-router` / `--tags` 等）字段名一律 **PascalCase**：`Rpm`、`Tpm`、`Strategy`、`Mode`，不是 `rpm`/`tpm`
- 独立 `+code-example` 的 flag 是 **`--model`（基础模型名/带版本 id）+ `--lang`**，不是 `--endpoint-id`：`arkcli +code-example --model <id> --lang python`（细节见 [`../arkcli-code-example/SKILL.md`](../arkcli-code-example/SKILL.md)）
- `+deploy` 创建成功后**不再自动**把示例写到 `./ark-examples/<endpoint-id>/`（内嵌渲染段仍下线）；要示例单独跑 `+code-example`

## 路由判断

- 用户已有模型 ID + 想正式部署 → `arkcli +deploy --name <ep> --model <id> --dry-run`，预演无误后去掉 `--dry-run` 真实创建
- 用户传入自定义模型 ID（`cm-xxxxx`）时，真实创建前会先查是否已有引用该自定义模型且状态为 `Running` 的 Endpoint；若有则直接复用并输出已有 `endpoint-id`，不会再创建第二个计费资源。`--dry-run` 仍按创建预演执行
- 用户语气紧急要求"立刻创建" → **不要跳过确认**，仍然先 `--dry-run` 并复述 `model/name/region`
- 已通过 `arkcli infer endpoint create` 拿到 `Id` → **不要**再 `+deploy` 创建第二个，转 `arkcli-infer-endpoint`；要调用示例转 `arkcli-code-example`（按模型名生成）

## 反触发（路由到别处，附完整命令避免下游幻觉）

| 用户意图 | 路由到 | 完整示范命令 |
|---------|--------|------------|
| 只想试模型效果 / 一次性生成 | `arkcli-chat` / `arkcli-gen` | `arkcli +chat --model <id> '...'` 或 `arkcli +gen --model <id> '...'` |
| 要某模型的调用示例 | `arkcli-code-example` | `arkcli +code-example --model <model-id> --lang python`（按模型名/版本生成；缺失版本降级到静态示例或控制台示例页） |
| 模型 ID 未定 | `arkcli-models` | `arkcli models search <keyword>` 或 `arkcli models list` |
| 401 / 鉴权失败 | `arkcli-auth` | `arkcli auth status`，必要时 `arkcli auth login` |
| profile / region / project 不符预期 | `arkcli-config` | `arkcli profile show --format json` (旧 `arkcli config show` 已 deprecated) |

## 典型链路

1. **从模型选择到正式接入**：`arkcli auth status` → `arkcli models search/get` → `arkcli +deploy ... --dry-run` → `arkcli +deploy ...`（自定义模型若已有 Running Endpoint 会复用）
2. **从试用切换到正式接入**：`arkcli +chat` / `arkcli +gen` 验证效果 → `arkcli +deploy ... --dry-run` → 真实创建
3. **创建后做调用集成**：记录返回的 `endpoint-id`；要多语言示例跑 `arkcli +code-example --model <model-id> --lang <lang>`（按模型名生成，非 endpoint-id；部分版本无示例时降级到 `ark-examples/` 静态示例或控制台示例页）

详细 flag、JSON 字段示例、错误码见 [`references/arkcli-deploy.md`](references/arkcli-deploy.md)。

## 参考

- [arkcli-chat](../arkcli-chat/SKILL.md) -- 快速对话试用，不创建 Endpoint
- [arkcli-gen](../arkcli-gen/SKILL.md) -- 图片/视频一步生成
- [arkcli-models](../arkcli-models/SKILL.md) -- 部署前确认模型 ID
- [arkcli-code-example](../arkcli-code-example/SKILL.md) -- 已有 endpoint 时生成调用代码
- [arkcli-shared](../arkcli-shared/SKILL.md) -- 认证与全局参数
