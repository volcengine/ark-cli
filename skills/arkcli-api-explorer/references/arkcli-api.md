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
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<registered-action>` | 否 | string | Action 名（例如 `model.list_foundation_models`）。不提供时默认进入 list 模式 |
| `--list` | 否 | bool | 列出全部已注册 Action |
| `--params` | 否 | string | JSON 请求体（字符串形式）。省略时默认 `{}` |

## 返回值

- list 模式：返回已注册 Action 列表（按名称排序）。
- invoke 模式：返回该 Action 的原始响应对象（按契约定义）。
- 输出遵循全局 `--format`（当前仅 `json`）与 `--transform`（GJSON path 表达式）。

## 如何确定 `--params` 的结构

禁止猜参数。推荐按下面顺序定位契约：

1. `arkcli api --list` 找到 action 名（例如 `model.list_foundation_models`）
2. 在代码中搜索 action 名，定位到对应的 operation 定义与 req/resp：
   - `internal/apis/<domain>/...`
3. 以 req 结构体的 `json:"..."` tag 为准构造 `--params` JSON。

例如 `model.list_foundation_models` 的请求体来自 `ListFoundationModelsRequest`，包含 `PageSize`、`PageNumber`、`SortBy`、`SortOrder`、`Filter` 等字段。

## 风险与守卫

- 默认只读优先：能用查询验证就先查询验证。
- 写/删/可能产生费用的 Action：执行前必须让用户明确确认意图。
- `--debug` 仅用于排障：会输出请求/响应调试信息到 stderr，容易产生噪声。
- 输出降噪：优先用 `--transform` 抽取稳定字段（便于脚本与后续步骤消费）。

## 常见用法模式

```bash
# 1) 最小调用（先跑通再逐步加字段）
arkcli api model.list_foundation_models --params '{}'

# 2) 加分页
arkcli api model.list_foundation_models --params '{"PageSize":10,"PageNumber":1}'

# 3) 只提取关键字段（示例 path，按实际输出结构调整）
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
```
