---
name: arkcli-api-explorer
version: 1.1.0
description: "arkcli Raw API Explorer：调用已注册的 Action 作为产品命令的兜底能力。当现有 `arkcli <domain> <verb>` 无法覆盖需求，或需要验证底层 Action 契约时使用。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli api --help"
---

# arkcli api（Raw API Explorer）

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证、配置覆盖排查与命令选择顺序**
**CRITICAL — 只有在现有产品命令和业务 skill 确实不覆盖时，才允许进入本 skill；禁止把 `arkcli api` 当默认入口。**
**CRITICAL — 任何 `arkcli api ...` 调用前，MUST 先用 Read 工具读取 [`references/arkcli-api.md`](references/arkcli-api.md)，禁止盲目猜参数与输出结构。**

## 使用原则

- 先产品命令：`arkcli <domain> <verb>` 或 `arkcli +<shortcut>`
- 再 skill / reference：确认是否已有稳定入口与正确参数
- 最后才 `arkcli api`：仅用于低频、专业或高风险的底层 Action 验证与兜底

不要因为某个 Action 存在，就反推出新的顶层 skill 或 `cmd/` 命令；只有升级为稳定产品能力时，才考虑补 `shortcuts/`。

## 适用场景

- 现有产品命令确实无法覆盖需求（且该需求不值得做成标准命令或 `+shortcut`）
- 需要验证某个已注册 Action 的输入输出契约（例如排障、回归验证、确认 transport 可达）
- 需要确认 registry 中是否已存在某个 Action（用于开发或排查 “注册缺失”）

## 唤起信号（When To Trigger）

- 用户明确提到：`action` / `operation` / `registry` / “已注册 Action” / “契约验证”
- 用户给出：`arkcli api ... --params ...` 并需要你补全/排错
- 用户遇到：`unknown action` / 需要确认某个 Action 是否存在
- 开发场景：在 `internal/apis/<domain>/` 新增 operation 后，要快速验证能否 `--list` 与可调用

## 反唤起信号（When NOT To Trigger）

- 用户目标是：对话（`+chat`）、生成（`+gen`）、部署（`+deploy`）、用量（`usage`）、查模型（`models`）等已有稳定产品路径
- 用户只是鉴权失败/未登录/环境 profile 混乱（应先走 `auth/config`，不要把问题导向 `api`）
- 用户只是想“找一个命令怎么用”（优先 `arkcli <domain> --help` + 对应 skill/reference）

## Agent 快速执行顺序

1. 先判断用户目标是否已有产品入口：`arkcli <domain> --help`，并优先转对应业务 skill
2. 如果只是认证/配置问题，不要误判为需要 `api`：
   - 认证失败先转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
   - profile/base-url/region 覆盖混乱先转 [`../arkcli-config/SKILL.md`](../arkcli-config/SKILL.md)
3. 枚举已注册 Action（无参等价于 list）：
   - `arkcli api --list`
   - `arkcli api`
4. 定位契约与必填字段（禁止猜 JSON）：
   - 在代码中查 `internal/apis/<domain>/` 对应的 req/resp 结构体
5. 只读优先、写操作必须二次确认，然后再执行：
   - `arkcli api <registered-action> --params '{...}'`
6. 输出尽量稳定、减少噪声：
   - 优先用全局 `--transform '<gjson path>'` 提取关键字段
   - 只有排障需要时才开 `--debug`（会输出请求/响应调试信息到 stderr）

## Guard Checklist（必须执行）

| 检查点 | 目的 | 做法 |
|--------|------|------|
| 产品命令覆盖判断 | 防止误用 `api` | 先 `arkcli <domain> --help` 并对照对应 skill |
| 认证闸门 | 防止把鉴权问题误判为缺能力 | 先 `arkcli auth status`；失败转 `arkcli-auth` |
| 契约事实源 | 防止猜参数 | 必须从 `internal/apis/<domain>/` 的 req/resp tag 生成 `--params` |
| 风险确认 | 防止写操作误触发 | 涉及创建/删除/修改/影响费用前，要求用户明确确认 |
| 噪声控制 | 防止整坨输出污染上下游 | 默认引导 `--transform` 提取关键字段；`--debug` 仅排障打开 |

## 示例

```bash
arkcli api --list
arkcli api model.list_foundation_models --params '{"PageSize":10,"PageNumber":1}'

# 只提取 items 里的 name（示例 path，按实际输出结构调整）
arkcli api model.list_foundation_models --params '{"PageSize":10,"PageNumber":1}' --transform 'Result.Items.#.Name'
```

## 常见错误与处理

| 现象 | 常见原因 | 处理方式 |
|------|----------|----------|
| `unknown action "..."` | Action 未注册或拼写错误 | 先 `arkcli api --list`；再在 `internal/apis/` 中确认是否已注册 |
| `invalid --params JSON: ...` | JSON 不合法（引号/转义/单引号包裹不当） | 确保 `--params` 是合法 JSON；必要时把 JSON 放到文件再用 shell 展开传入 |
| 输出字段和预期不一致 | 直接猜测了契约 | 先读 [`references/arkcli-api.md`](references/arkcli-api.md) 并查看 `internal/apis/<domain>/` 的 req/resp |

## 开发约束

- 如果 Action 未注册：补 `internal/apis/<domain>/` 的 operation 注册与 req/resp 契约
- 不要直接因为 Action 缺失就新增 `cmd/<action>`；`api` 是兜底入口
- 只有当它升级为稳定产品能力时，才在 `shortcuts/` 中补业务命令（并配套 skill/reference/test）

## 参考

- [arkcli-shared](../arkcli-shared/SKILL.md) -- 认证、配置覆盖排查与共享安全规则
- [arkcli-auth](../arkcli-auth/SKILL.md) -- 认证失败时的回退入口
- [arkcli-config](../arkcli-config/SKILL.md) -- profile / base-url / region 排障入口
- [`references/arkcli-api.md`](references/arkcli-api.md) -- api explorer 的命令语义、参数与排错
- [`references/evals.md`](references/evals.md) -- 最小评估用例（唤起/反唤起/排错）
- [`../../docs/developer-guide.md`](../../docs/developer-guide.md) -- 开发者指南（API Explorer 适用场景与开发流程）
