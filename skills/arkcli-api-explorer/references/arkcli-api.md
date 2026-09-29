# arkcli api（Raw API Explorer）

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证闸门、全局 flags、配置覆盖排查与安全规则。

`arkcli api` 用于调用仓库中已注册的 Action（operation），作为产品命令的兜底入口。它面向低频、专业或高风险场景，不应替代 `arkcli <domain> <verb>` 或 `arkcli +<shortcut>`。

## 唤起信号 / 反唤起信号

适合进入 `arkcli api` 的信号：

- 用户明确说要“验证某个 Action/operation/registry”
- 用户目标是“排障/回归验证/契约核对”，且产品命令确实没有覆盖
- 用户遇到 `unknown action`，需要确认是否已注册

不适合进入 `arkcli api` 的信号：

- 用户目标是 `models/+chat/+gen/+deploy/usage` 等已覆盖场景
- 用户只是未登录/鉴权失败/环境配置混乱（优先回到认证与配置排障）

## 没有产品命令封装的数据面能力（直接走 raw API）

下面这些数据面能力当前**没有专属产品命令**，需要直接通过 `arkcli api` 调用对应的 `arkruntime.*` Action：

| 能力 | Action | 说明 |
|------|--------|------|
| Embedding（文本向量） | `arkruntime.create_embeddings` | 数据面 OpenAI 兼容 `/embeddings`，传 `model` + `input` 即可 |

```bash
arkcli api arkruntime.create_embeddings --params '{"model":"doubao-embedding-large-text-250515","input":"火山方舟"}'
```

需要新增其它能力的产品命令时，从这里降级到 raw API 是合法兜底，不要把 raw API 当默认入口。

## 命令

```bash
# 列出所有已注册 Action（无参等价于 --list）
arkcli api --list
arkcli api

# 调用某个已注册 Action
arkcli api <registered-action> --params '{...}'

# 只在客户端预览 descriptor 和最终 payload，不请求后端
arkcli api <registered-action> --params '{...}' --dry-run
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<registered-action>` | 否 | string | Action 名（例如 `model.list_foundation_models`）。不提供时默认进入 list 模式 |
| `--list` | 否 | bool | 列出全部已注册 Action |
| `--params` | 否 | string | JSON 请求体（字符串形式）。省略时默认 `{}` |
| `--dry-run` | 否 | bool | 仅 invoke 模式可用；输出 `preview.v1` Client Preview，不请求后端 |

## 返回值

- list 模式：返回已注册 Action 列表（按名称排序）。
- invoke 模式：返回该 Action 的原始响应对象（按契约定义）。
- invoke + `--dry-run`：返回 `mode=client_preview` 和 `steps[0]` 中的
  `protocol`、`target`、脱敏后的 `payload`；不创建 transport，不验证服务端。
- 输出遵循当前二进制帮助中的全局 `--format` 与 `--transform`（GJSON path 表达式）；机器处理优先用完整 JSON，不把历史“仅 json”当成输出能力限制。

`arkcli api --list --dry-run` 会报错，因为 list 本身已经是纯本地枚举，
没有远端请求可供 Preview。

## 如何确定 `--params` 的结构

禁止猜参数。推荐按下面顺序定位契约：

1. `arkcli api --list` 找到 action 名（例如 `model.list_foundation_models`）
2. 使用当前版本的 reference/契约；有匹配版本的源码时，可在 `internal/apis/<domain>/...` 搜索 operation 定义与 req/resp。已安装 CLI 的宿主不一定有源码，不要求用户下载整个仓库才能调用。
3. 以已核实的请求字段（源码中为 `json:"..."` tag）构造 `--params` JSON；拿不到 schema 时先补证据，不靠字段大小写轮番试错。registry 枚举只证明 Action 已注册，不证明猜测的 payload 有效。

例如 `model.list_foundation_models` 的请求体来自 `ListFoundationModelsRequest`，包含 `PageSize`、`PageNumber`、`SortBy`、`SortOrder`、`Filter` 等字段。

## 风险与守卫

### JSON 与 shell 变量

静态 JSON 用外层单引号、内部双引号：`--params '{"PageSize":100}'`。动态值用 JSON 编码器构造，不把用户文本拼进转义字符串；字符串用 `--arg`，已校验的数字/布尔/数组才用 `--argjson`：

```bash
# 只构造参数，不执行远端请求；字段仍必须按目标 Action 的真实 schema 选择
page_size=100
params=$(jq -cn --argjson size "$page_size" '{PageSize:$size}')
# 字符串示例：引号、反斜杠和换行由 jq 编码，不使用 eval
name_params=$(jq -cn --arg name "$resource_name" '{Name:$name}')
# 调用时整体引用变量：arkcli api <已核实Action> --params "$params"
```

`jq` 失败就停止，不发送空串或部分 JSON。构造成功只证明 JSON 语法正确，不证明请求字段或资源权限正确。响应先完整捕获，再按实际 envelope 与字段大小写解析；不得用 `head` 截断后拼回“合法 JSON”。

- 默认只读优先：能用查询验证就先查询验证。
- 写/删/可能产生费用的 Action：执行前必须让用户明确确认意图。
- 写/删/可能产生费用的 Action：先运行同一条命令的 `--dry-run` Client
  Preview 核对 payload；Preview 后仍必须让用户明确确认，不能自动去掉 flag。
- `--params '{"DryRun":true}'` 中的 `DryRun` 是后端请求字段，不等于 CLI
  `--dry-run`。未加 CLI flag 时命令仍会创建 transport 并发出真实请求。
- `--debug` 仅用于排障：会输出请求/响应调试信息到 stderr，容易产生噪声。
- 输出降噪：优先用 `--transform` 抽取稳定字段（便于脚本与后续步骤消费）。

## 常见用法模式

```bash
# 1) 最小调用（先跑通再逐步加字段）
arkcli api model.list_foundation_models --params '{}'

# 2) 先做本地 Client Preview
arkcli api model.list_foundation_models \
  --params '{"PageSize":10,"PageNumber":1}' \
  --dry-run \
  --transform 'steps.0.payload'

# 3) 加分页
arkcli api model.list_foundation_models --params '{"PageSize":10,"PageNumber":1}'

# 4) 只提取关键字段（示例 path，按实际输出结构调整）
arkcli api model.list_foundation_models --params '{"PageSize":10,"PageNumber":1}' --transform 'Result.Items.#.Name'
```

## 常见错误

| 错误/现象 | 原因 | 处理方式 |
|----------|------|----------|
| `unknown action "..."` | Action 未注册/拼写错误 | 先 `arkcli api --list`；再检查 `internal/apis/` 是否已注册该 operation |
| `invalid --params JSON: ...` | JSON 不合法 | 确保 `--params` 是合法 JSON（双引号、无多余逗号）；必要时先用工具校验 JSON |

## 注意事项

- `--params` 是字符串 JSON：命令行引号和转义容易出错，优先从最小 `{}` 开始逐步加字段。
- `--transform` 建议用于提取稳定字段，减少 stdout 噪声，便于脚本集成。
- 写操作/删除操作：必须在执行前让用户确认意图；能只读验证就先只读验证。

## 最小评估（建议每次改动后跑）

```bash
# 1) list 能工作（无参/--list 都行）
arkcli api --list --transform '0.name'

# 2) 正常调用能工作（以已知 action 为例）
arkcli api model.list_foundation_models --params '{"PageSize":1,"PageNumber":1}' --transform 'Result.Items.#.Name'

# 3) Client Preview 不请求后端，输出统一 plan
arkcli api model.list_foundation_models \
  --params '{"PageSize":1,"PageNumber":1}' \
  --dry-run \
  --transform 'steps.0.payload'
```
