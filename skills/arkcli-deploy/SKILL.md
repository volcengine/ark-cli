---
name: arkcli-deploy
version: 1.4.0
description: "arkcli +deploy：创建推理接入点（Endpoint）的统一首选入口 —— **用户说『创建/新建/create 一个 endpoint/接入点』或『部署/上线/deploy 某模型』，只要意图是新建一个接入点，一律优先走这里，不要走 arkcli-infer-endpoint 的 create**。当用户需要把模型部署成在线推理接入点时使用。注意：要对**已有** Endpoint 做获取/列表/启停/更新等全生命周期管理，才走 arkcli-infer-endpoint；本 skill 只负责一键创建（创建外的增删改查不在此）。创建成功后会自动把多语言调用示例渲染到 ./ark-examples/<ep-id>/。反触发：TTS/ASR/语音模型不能 +deploy，只能转 models search 说明广场可搜但 arkcli 不支持 Endpoint 创建。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli +deploy --help"
---

# arkcli +deploy

**前置：** 先用 Read 读 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md) 获取共享认证/配置/写操作守卫规则。

**新增 flag `--set-default <modality>`**: 部署成功后自动把新 endpoint 设为 active profile 该 modality (`text` / `image` / `video`) 的默认资源。仅在 `--dry-run` 之外 + 用户明确传 modality 时生效；失败仅 stderr warn，不阻断部署主流程。详见 [`../arkcli-shared/references/profile-defaults.md`](../arkcli-shared/references/profile-defaults.md)。

**写操作 + 计费**：`+deploy` 创建在线推理 Endpoint 是真实写操作，会产生计费资源。**任何执行前必须先 `--dry-run` 或与用户显式确认**，即使用户语气紧急。

**模型开通是独立计费写动作，非交互环境一律不自动开通**：若目标基础模型尚未开通，`+deploy` 会触发"开通模型"（账号级计费写）。**在 agent / CI / 管道这类非 TTY 环境，开通被硬拒——`--yes` 也不放行**（`--yes` 只在真人交互终端里用于跳过 `[y/N]`）。命中时 CLI 返回 `model_activation_required` + console 链接，你**必须结束本轮、把"开通（计费）"这件事连同链接交还给真人**，由真人在交互终端确认或在网页 console 开通。**严禁自己补 `--yes` / `echo Y` / 设 `ARKCLI_ALLOW_HEADLESS_ACTIVATION` 替用户开通**——你打印一句"请确认"然后同一轮自己加 `--yes` 跑掉，等于没问。`ARKCLI_ALLOW_HEADLESS_ACTIVATION=1` 只留给真·无人值守自动化（CI 流水线），不是 agent 该设的。`--dry-run` 预演**永不开通**。

**实名前置（开通类硬闸门）**：`+deploy` 会触发开通模型，**第一步**先 `arkcli auth status` 读 `volc_sso.identity.verified`——`false` 即停并把实名页 `https://console.volcengine.com/user/authentication/detail/` 贴给用户、暂停等待（详见 [`../arkcli-auth/references/realname-gate.md`](../arkcli-auth/references/realname-gate.md)）。**不要**先试 `+deploy`、撞到 model-id 无效 / 模型未开通报错再回头查实名——那些报错会把你带偏。

**语音模型硬边界**：TTS / ASR / 配音 / 朗读 / 播客 / 音色设计 / 实时语音交互，或模型名命中 `doubao-seed-tts-*`、`doubao-seed-asr-*`、`seedasr-*` 时，**不要执行 `+deploy`**。这些模型在 arkcli 当前只支持 [`models search`](../arkcli-models/SKILL.md) 做广场发现；不能创建 Endpoint，也不能通过开通模型绕过。

## 示例代码：`+deploy` 创建后自动渲染 + 独立命令 `+code-example`

两条路径都可用，都走 OpenTOP `OpenGetSampleCode`：

- **`+deploy` 创建成功后自动渲染**：把该接入点的多语言调用示例写到 `./ark-examples/<ep-id>/`（用 ep-id 当作可直接调用的 model id，示例直接调你刚建的接入点），并在 stderr 打一行落盘摘要。best-effort —— 取不到示例只软提示，不影响 Endpoint 创建/启动。
- **独立命令 `arkcli +code-example`**：按基础模型名/版本单独生成示例，转 [`../arkcli-code-example/SKILL.md`](../arkcli-code-example/SKILL.md)；注意它按 model-version 提供，部分版本后端无 group 会返回 not found，缺失时再降级到 `ark-examples/` 静态示例或方舟控制台示例代码页。

## 子命令穷举（只有这一个）

| 调用 | 说明 |
|------|------|
| `arkcli +deploy --name <ep-name> --model <model-id> [...]` | 创建 Endpoint；不加 `--dry-run` 即真实创建 |

> ⚠️ **没有** `arkcli deploy ...` / `arkcli endpoint create` / `arkcli +deploy create` 等子命令。整个能力就是一个 `+deploy` 命令加 flag。

## 反幻觉清单

- `--name`、`--model` 必填
- JSON 类 flag（`--rate-limit` / `--moderation` / `--intelligent-router` / `--tags` 等）字段名一律 **PascalCase**：`Rpm`、`Tpm`、`Strategy`、`Mode`，不是 `rpm`/`tpm`
- 独立 `+code-example` 的 flag 是 **`--model`（基础模型名/带版本 id）+ `--lang`**，不是 `--endpoint-id`：`arkcli +code-example --model <id> --lang python`（细节见 [`../arkcli-code-example/SKILL.md`](../arkcli-code-example/SKILL.md)）
- `+deploy` 创建成功后会**自动**把示例渲染到 `./ark-examples/<ep-id>/`（按 ep-id）；想按基础模型名另出一份则跑 `+code-example`
- 模型未开通时 `+deploy` 的开通在**非 TTY 下被硬拒、`--yes` 也不放行**；**禁止自己补 `--yes` / `echo Y` / 设 `ARKCLI_ALLOW_HEADLESS_ACTIVATION`**，必须把开通（计费）交还真人在终端 / console 处理
- 语音模型（TTS / ASR / 播客 / 音色 / 实时语音交互）广场可搜不等于可部署；命中这类模型时停在 `arkcli models search <keyword>`，不要给 `+deploy` 命令

## 路由判断

- 用户已有模型 ID + 想正式部署 → `arkcli +deploy --name <ep> --model <id> --dry-run`，预演无误后去掉 `--dry-run` 真实创建
- 用户要部署 / 接入语音模型，或模型名看起来是 `*-tts-*` / `*-asr-*` / `seedasr-*` → 转 [`arkcli-models`](../arkcli-models/SKILL.md) 说明"只支持广场检索，不支持 Endpoint 创建"
- 用户传入自定义模型 ID（`cm-xxxxx`）时，真实创建前会先查是否已有引用该自定义模型且状态为 `Running` 的 Endpoint；若有则直接复用并输出已有 `endpoint-id`，不会再创建第二个计费资源。`--dry-run` 仍按创建预演执行
- 用户语气紧急要求"立刻创建" → **不要跳过确认**，仍然先 `--dry-run` 并复述 `model/name/region`
- 已通过 `arkcli infer endpoint create` 拿到 `Id` → **不要**再 `+deploy` 创建第二个，转 `arkcli-infer-endpoint`；要调用示例转 `arkcli-code-example`（按模型名生成）

## 反触发（路由到别处，附完整命令避免下游幻觉）

| 用户意图 | 路由到 | 完整示范命令 |
|---------|--------|------------|
| 只想试模型效果 / 一次性生成 | `arkcli-chat` / `arkcli-gen` | `arkcli +chat --model <id> '...'` 或 `arkcli +gen --model <id> '...'` |
| 要某模型的调用示例 | `arkcli-code-example` | `arkcli +code-example --model <model-id> --lang python`（按模型名/版本生成；缺失版本降级到静态示例或控制台示例页） |
| 模型 ID 未定 | `arkcli-models` | `arkcli models search <keyword>` 或 `arkcli models list` |
| 语音模型部署 / TTS 接入点 / ASR Endpoint | `arkcli-models` | `arkcli models search <keyword>`（只做广场发现；当前不支持 Endpoint 创建） |
| 401 / 鉴权失败 | `arkcli-auth` | `arkcli auth status`，必要时 `arkcli auth login` |
| profile / region / project 不符预期 | `arkcli-config` | `arkcli profile show --format json` (旧 `arkcli config show` 已 deprecated) |
| 脚本化 / CI / 需要精细控制每个参数、跳过护栏 | `arkcli-infer-endpoint` | `arkcli infer endpoint create --model <id> --name <ep>` |

## 典型链路

1. **从模型选择到正式接入**：`arkcli auth status` → `arkcli models search/get` → `arkcli +deploy ... --dry-run` → `arkcli +deploy ...`（自定义模型若已有 Running Endpoint 会复用）
2. **从试用切换到正式接入**：`arkcli +chat` / `arkcli +gen` 验证效果 → `arkcli +deploy ... --dry-run` → 真实创建
3. **创建后做调用集成**：`+deploy` 已把示例自动写到 `./ark-examples/<ep-id>/`，直接用即可；想按基础模型名另出一份跑 `arkcli +code-example --model <model-id> --lang <lang>`（部分版本无示例时降级到 `ark-examples/` 静态示例或控制台示例页）

详细 flag、JSON 字段示例、错误码见 [`references/arkcli-deploy.md`](references/arkcli-deploy.md)。

## 参考

- [arkcli-chat](../arkcli-chat/SKILL.md) -- 快速对话试用，不创建 Endpoint
- [arkcli-gen](../arkcli-gen/SKILL.md) -- 图片/视频一步生成
- [arkcli-models](../arkcli-models/SKILL.md) -- 部署前确认模型 ID
- [arkcli-code-example](../arkcli-code-example/SKILL.md) -- 已有 endpoint 时生成调用代码
- [arkcli-shared](../arkcli-shared/SKILL.md) -- 认证与全局参数
