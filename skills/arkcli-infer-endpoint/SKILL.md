---
name: arkcli-infer-endpoint
version: 1.3.6
description: "arkcli 推理接入点管理与显式 raw CRUD 创建能力。**正向触发：用户明确要求脚本化 / CI / 无护栏 / 原始 raw CRUD 创建 Endpoint 时，必须使用本 skill 的 `arkcli infer endpoint create` 路径，不能切到 `arkcli-deploy`**；模型缺失或只有品牌/家族名时仍留在本 skill，先执行实时有界候选与 0/1/N 澄清，模型未唯一确定前禁止 preview 或创建。本 skill 也负责对**已有** Endpoint 做获取、列表、启动、停止、删除、更新，以及在 NotFound 后用当前完整列表 + 历史 usage 判断 ID 是否不完整或仅有历史记录。普通的『创建/新建/部署一个 endpoint』仍走带产品护栏的 arkcli-deploy（`+deploy`），只有用户明确选择上述原始路径才例外。优先使用产品命令 `arkcli infer endpoint ...`，而不是 Raw API。TTS/ASR/语音模型连 raw create 也不要引导，只能转 models search 说明 arkcli 不支持 Endpoint 创建。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli infer endpoint --help"
---

# arkcli infer endpoint

**CRITICAL — 显式脚本化 / CI / 无护栏 / 原始 raw CRUD 创建是本 skill 的正向触发，不得因为出现“创建 Endpoint”就转到 `arkcli-deploy`。**

**CRITICAL — 模型候选复用本轮完整结果；截断、空结果的有界恢复统一复用 Deploy 规则。认证/权限失败不重试、不切身份、不修改配置或把凭证写进命令行；其他失败保留原始错误，不盲换命令。**

**CRITICAL — 开始前 MUST 先读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)。**

## 使用原则

- 推理接入点相关需求优先使用 `arkcli infer endpoint ...`
- 这些命令虽然是标准 CLI 类型，但实现入口仍然来自 `shortcuts/inferendpoint/`
- 只有产品命令无法覆盖时，才回退到 [`../arkcli-api-explorer/SKILL.md`](../arkcli-api-explorer/SKILL.md)
- `infer endpoint create` 成功后会返回 Endpoint `Id`
- 这个 `Id` 应作为后续 `get / start / stop` 的输入，也可以传给 [`../arkcli-code-example/SKILL.md`](../arkcli-code-example/SKILL.md) 生成带真实 `endpoint-id` 的调用示例
- 如果已经通过 `infer endpoint create` 拿到 `Id`，不要再调用 `+deploy` 试图"二次部署"；`+deploy` 本身就是创建 Endpoint 的工作流
- `infer endpoint create --billing-method` 当前只支持 `token`；它是可选项，显式传 `token` 时会先校验模型是否支持 token 推理方式，创建请求本身保持默认行为
- **语音模型不是 Endpoint create 目标**：TTS / ASR / 配音 / 播客 / 音色 / 实时语音交互，或模型名命中 `doubao-seed-tts-*`、`doubao-seed-asr-*`、`seedasr-*` 时，不要用 `infer endpoint create` 或 Raw CRUD 绕过；只转 [`arkcli-models`](../arkcli-models/SKILL.md) 说明"广场可搜，但当前 arkcli 不支持 Endpoint 创建"。

## 「我的接入点」语义

用户说**"我的推理接入点 / 我创建的 / 我有多少个 / 列出我的"** → 必须加 `--mine --page-all`：

```bash
arkcli infer endpoint list --mine --page-all --page-size 100 --format json
```

服务端按 `sys:ark:createdBy` tag 过滤，只返回当前 SSO sub-user 创建的 endpoint。
需要 SSO 子账号身份；root 账号 / AK-SK 报错不算空结果，认证处理服从当前 shared 宿主协议，不自动重登。
详细行为见 [`references/arkcli-infer-endpoint-list.md`](references/arkcli-infer-endpoint-list.md)。

### 本人查询为空的三级恢复

1. `--mine` 查询成功、完整且 `.TotalCount == 0` / `.Items == []` 才进入下一层；鉴权、权限、网络、参数错误及分页不完整均不能当空。
2. 向用户说明接下来查“当前用户 API Key 用过的 EP”，不冒充“本人创建”。按 [usage stats](../arkcli-usage/references/arkcli-usage-stats.md) 使用 `arkcli usage stats --start <29天前日期> --end <今天日期> --mine --mine-by=apikey --by endpoint --format json`；原请求有时间范围就保留。对完整结果里的非空 `ModelEndpoint` 去重，逐个 `infer endpoint get <真实ID>`，保留原 `--status` / `--model` 条件（get 结果本地核对）。不能给 endpoint list 发明 `--mine-by`。历史 ID 当前 NotFound 时单列“历史调用过、当前不可见”，不称已删除；其他失败原样报告。
3. 第二层也成功但没有非空历史 ID 时，询问用户是否查看账号全量；已明确授权该范围时可继续。执行 `arkcli infer endpoint list --page-all --page-size 100 --format json` 并保留原筛选条件，结果明确标“账号全量 fallback（已取消 --mine，不能保证本人创建）”。拒绝扩大范围则交付本人查询为空，不自动全账号。


## Endpoint NotFound 的只读排障

任何调用异常先读 [get reference 的排障分支](references/arkcli-infer-endpoint-get.md)，区分控制面状态、历史用量与真实调用证据；盘点/导出先读 [list reference](references/arkcli-infer-endpoint-list.md) 的分页、限流、时区与交付要求。已知精确 ID 时先直接 get，不为所有问题先列全账号。

用户已经观察到 `InvalidEndpointOrModel.NotFound`，或 `get` 查不到某个 Endpoint 时，不要直接宣称「已删除」，也不要猜测尾部字符。按两个独立事实源收敛：

**禁止仅按 ID 长度、形状、当前列表里的其他 ID 或常见格式，声称用户输入「缺少后缀 / 缺少随机字符 / 不是完整 ID」。** 只有当前列表返回了以用户输入为前缀的唯一真实 ID，才能把该完整 ID 作为证据复述；否则只能报告未找到。

1. 查当前身份与 project 可见的完整列表，同时核对精确 ID 和用户所给前缀。列表可能很大，必须在 shell 内只保留匹配项，避免把无关 Endpoint 全量灌入 Agent 上下文：

   ```bash
   arkcli infer endpoint list --page-all --page-size 100 --format json \
     | jq --arg id '<endpoint-id>' '[.Items[]? | select(.Id == $id or (.Id | startswith($id)))]'
   ```

   `<endpoint-id>` 必须替换为用户提供的原值；不得省略 `--page-all` 后只拿第一页下结论。
2. 读取 [`../arkcli-usage/references/arkcli-usage-stats.md`](../arkcli-usage/references/arkcli-usage-stats.md)，再用 `arkcli usage stats --start <YYYY-MM-DD> --end <YYYY-MM-DD> --endpoint <endpoint-id> --format json` 核对历史用量。用户给了时间范围就照用；没给时默认查最近 30 天，必须把占位符换成实际日期。

结论必须与证据强度对齐：

- 当前列表找到唯一的完整 ID：复述真实 ID，说明原输入是否少尾部；不要凭格式自己补齐。
- 当前列表无记录、但历史用量有记录：只能说「曾经有调用，当前不可见」，不能单凭这两步区分已删除、换身份或换 project。
- 两处都没有证据：报告「未找到」，建议核对原始 ID、profile 和 region；禁止为了让任务继续而创建 Endpoint、切 profile 或走 Raw API。

## `infer endpoint create` 与 `+deploy` 的边界

`infer endpoint create` 是**原始 CRUD**，适合脚本化 / CI / 需要无护栏可预测行为的场景。注意它的 flag 集其实是 `+deploy` 的**子集**（`+deploy` 才是参数超集），差别在"有没有工作流护栏"，不是"谁参数更多"。它**不包含** `+deploy` 的以下护栏：

| 护栏 | `+deploy` | `infer endpoint create` |
|------|-----------|------------------------|
| 实名前置校验（[realname-gate.md](../arkcli-auth/references/realname-gate.md)） | ✅ | ✅（foundation model 路径同样触发 `EnsureModelAvailable`，含实名拦截） |
| 自定义模型复用检查（避免重复计费） | ✅ | ❌ |
| 部署后 profile 默认资源同步（`--set-default`） | ✅ | ❌ |

**路由判定**：只要用户意图是"创建 / 新建 / create 一个 endpoint / 接入点"或"把模型部署 / 上线"（无论用词是 create 还是 deploy），一律路由到 `+deploy`（[arkcli-deploy](../arkcli-deploy/SKILL.md)）；**只有**用户明确需要绕过工作流护栏、脚本化 / CI、或要无护栏的可预测 raw CRUD 行为时，才使用 `infer endpoint create`。

**例外**：如果目标是语音模型（TTS / ASR / 播客 / 音色 / 实时语音交互），不要因为用户说了"create endpoint"就进入本 skill；语音模型在 arkcli 当前只支持 `models search` 发现，不支持 Endpoint 创建。

## raw CRUD 创建前的模型澄清

本节只在用户已经明确选择脚本化 / CI / 无护栏 raw CRUD 路径、但 `--model` 仍缺失或只有品牌、系列、家族名时生效。普通创建意图仍按上一节路由到 [`arkcli-deploy`](../arkcli-deploy/SKILL.md)，不能因为模型未定就改走 raw create。

候选生成必须复用 [`arkcli-deploy` 的「创建意图中的模型澄清」](../arkcli-deploy/SKILL.md#创建意图中的模型澄清)契约：初次执行有界的实时 `arkcli models search [<keyword>] --size 10 --format json`，只从本轮结构化结果读取真实字段。过滤、完整模型 ID、创建时间排序、模态覆盖、0/1/N 处理及截断/空结果恢复均与 `+deploy` 一致，不能另立更严格的一次查询即停止策略。只有恢复仍不完整时才停止；禁止全量盲查、逐候选循环 get 或凭记忆补版本。

- **0 个候选**：如实说明本轮没有查到，请用户补充用途、模态或关键词。
- **1 个候选**：复述本轮返回的完整模型 ID，请用户确认。
- **多个候选**：优先使用当前宿主提供的结构化选择能力，把本轮实时返回的精简候选直接列成选项；通用 Skill 不写死宿主工具名，如果宿主会自动补自由输入项，不要再添加“其他 / 手动指定”重复兜底。宿主没有结构化选择能力时，才退化为精简列表并要求用户明确回复一个完整 ID。

选定唯一模型后，仍保持在 `arkcli-infer-endpoint` 的 raw CRUD 路径：先生成 `arkcli infer endpoint create ... --dry-run` 的本地 `preview.v1`；用户只要求预览时到此停止，要求真实创建时继续遵守共享写操作守卫。模型未唯一确定前，禁止执行 `--dry-run`、真实 `infer endpoint create`、`+deploy` 或 Raw API 创建。

## 命令一览

| 命令 | 说明 |
|------|------|
| `arkcli infer endpoint create` | 创建推理接入点（原始 CRUD，无部署工作流护栏）；**新建接入点的默认意图请用 `+deploy`**，此命令仅限脚本化 / 无护栏 raw CRUD 场景 |
| `arkcli infer endpoint get <endpoint-id>` | 获取推理接入点详情 |
| `arkcli infer endpoint list [--mine]` | 列出推理接入点；用户说**"我的 / 我自己的 / 我创建的 / 我有多少"**时必须加 `--mine`（SSO sub-user 过滤） |
| `arkcli infer endpoint start <endpoint-id>` | 启动推理接入点 |
| `arkcli infer endpoint stop <endpoint-id>` | 停止推理接入点 |
| `arkcli infer endpoint delete <endpoint-id>` | 删除推理接入点（**不可逆**，需二次确认；非交互环境需 `ARKCLI_ALLOW_HEADLESS_DELETE=1`，`--yes` 不授权） |
| `arkcli infer endpoint update <endpoint-id>` | 更新推理接入点（名称 / 描述 / 限流） |

## 参考

- [`references/arkcli-infer-endpoint-create.md`](references/arkcli-infer-endpoint-create.md)
- [`references/arkcli-infer-endpoint-get.md`](references/arkcli-infer-endpoint-get.md)
- [`references/arkcli-infer-endpoint-list.md`](references/arkcli-infer-endpoint-list.md)
- [`references/arkcli-infer-endpoint-start.md`](references/arkcli-infer-endpoint-start.md)
- [`references/arkcli-infer-endpoint-stop.md`](references/arkcli-infer-endpoint-stop.md)
- [`references/arkcli-infer-endpoint-delete.md`](references/arkcli-infer-endpoint-delete.md)
- [`references/arkcli-infer-endpoint-update.md`](references/arkcli-infer-endpoint-update.md)
- [`../arkcli-code-example/SKILL.md`](../arkcli-code-example/SKILL.md)
