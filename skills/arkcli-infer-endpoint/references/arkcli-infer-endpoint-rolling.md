# arkcli infer endpoint rolling

替换**已有**推理接入点（Endpoint）绑定的模型版本，不新建 Endpoint。语义与 `infer endpoint update`（只改名称 / 描述 / 限流）正交：rolling 只换模型版本。不要把 `--model` 塞进 `update`。

命令组：`plan` / `create` / `get` / `list` / `cancel` / `rollback`。写操作（create / cancel / rollback）由 CLI 确认门保护；plan / get / list 只读。命令在全部产品可用，最终是否可切换由 Endpoint 的 `SupportRolling` 和 eligibility 决定。

## plan — 只读预检

```bash
arkcli infer endpoint rolling plan <endpoint-id> --model <model-id>
```

返回源/目标模型、`Supported`、`BlockingReasons` 和 `Warnings`。该命令不发起写请求，适合在 create 前先判断是否可切换。

## create — 发起滚动切换

```bash
arkcli infer endpoint rolling create <endpoint-id> --model <model-id> [flags]
```

将 `<endpoint-id>` 从当前模型版本滚动切换到 `--model` 指定的新版本。

| Flag | Type | Description | Required |
|------|------|-------------|----------|
| `--model` | string | 目标模型 ID（基础模型 名/版本，如 `doubao-seed-2-0-pro-260328`） | **Yes** |
| `--security-level` | string | 目标模型安全级别；需要按产品约束消歧时传入 | No |
| `--activate-target-model` | bool | 目标模型未开通时，在独立确认门后先开通再滚动（账号级计费写操作，rolling 的确认参数不授权该动作） | No |
| `--description` | string | 记录在滚动任务上的可选描述 | No |
| `--wait` | bool | 阻塞直到滚动进入终态（Succeeded / Reverted） | No |
| `--wait-timeout` | duration | `--wait` 的最长等待时长（Go duration，如 `15m`；`0` = 不限）。默认 `15m` | No |
| `--poll-interval` | duration | `--wait` 轮询间隔（Go duration，最小 `1s`）。默认 `5s` | No |
| `--yes` | bool | 宿主托管确认参数；Skill 生成的命令不带该参数，由宿主在执行前完成授权（仅授权本次滚动，不授权模型开通） | No |
| `--dry-run` | bool | 输出本地 `CreateEndpointRolling` Client Preview，零网络 | No |

**确认门**：交互终端会提示 `Switch endpoint ep-xxx from <当前模型/版本> to <目标模型/版本>? This starts a rolling update.`，输入 `y` 确认。Skill 只生成不带确认参数的命令，非交互场景由宿主在执行前统一授权；CLI 返回 `requires_confirmation` / `requires_headless_gate` 时停止并报告，不要自行补确认参数重试。该确认只授权本次滚动，**不**代表授权目标模型的开通 / 实名等前置。

**默认输出**：JSON `{Id, EndpointId, Status}`（`--wait` 时输出轮询到的终态 `RollingInfo`）。拿到的 `Id` 用于 `get` / `cancel` / `rollback`。

## get — 查询滚动任务详情

```bash
arkcli infer endpoint rolling get <rolling-id>
```

只读，输出 `RollingInfo`（Status、RollingIn / RollingOut 模型快照、RollingGray、轨迹等）。

## list — 列出滚动任务

```bash
arkcli infer endpoint rolling list [flags]
```

| Flag | Type | Description |
|------|------|-------------|
| `--endpoint-id` | string | 按 endpoint id 过滤 |
| `--status` | string | 按状态过滤：`Running` / `Reverting` / `Succeeded` / `Reverted`（大小写不敏感） |
| `--page-number` | int | 页码（>=1） |
| `--page-size` | int | 每页数量 |
| `--sort-by` | string | 排序字段（透传后端） |
| `--sort-order` | string | 排序方向：`Asc` / `Desc` |
| `--format` | string | 输出格式：`table` / `json`（默认 json） |

支持全局 `--page-all` 自动翻页。`--format table` 的列：`ID  ENDPOINT  STATUS  GRAY  ROLLING_IN  ROLLING_OUT  UPDATE_TIME`。

## cancel — 取消进行中的滚动

```bash
arkcli infer endpoint rolling cancel <rolling-id> [flags]
```

取消当前切换并**立即全量**切回原模型版本。高危：当前流量较大时原模型可能瞬时压力过高、触发限流，建议仅在有明确需求或非业务高峰期使用；需要逐步切回请用 `rollback`。写操作，确认参数宿主托管，Skill 生成的命令不带 `--yes`（仅授权本次 cancel）。

## rollback — 回滚滚动

```bash
arkcli infer endpoint rolling rollback <rolling-id> [flags]
```

按后端灰度策略**渐进**回滚到滚动开始前的模型版本（与 cancel 的立即全量切回不同）。写操作，确认参数宿主托管，Skill 生成的命令不带 `--yes`（仅授权本次 rollback）。

## PTU 与智能路由入口

Web 控制台中的其他 rolling 形态使用独立命令组，不混入普通 foundation-model rolling：

```bash
arkcli infer endpoint ptu rolling create <endpoint-id> --custom-model <cm-id> [flags]
arkcli infer endpoint ptu rolling list --endpoint-id <endpoint-id>
arkcli infer endpoint ptu rolling cancel <endpoint-id> [flags]

arkcli infer endpoint ptu seamless-rolling create <endpoint-id> --custom-model <cm-id> [--strategy <json|@file>] [--wait]
arkcli infer endpoint ptu seamless-rolling get <rolling-id>
arkcli infer endpoint ptu seamless-rolling list --endpoint-id <endpoint-id>
arkcli infer endpoint ptu seamless-rolling cancel <endpoint-id> [flags]

arkcli infer endpoint router rolling set <endpoint-id> --model <router-model-id> --router <json|@file> [flags]
arkcli infer endpoint router rolling get <endpoint-id>
```

PTU 命令只接受 `cm-` 自定义模型；Router 命令只接受智能路由 Endpoint，并通过 `TargetRouter` 发起普通 `CreateEndpointRolling`。写命令的确认参数同样由宿主托管。

## Global Flags

| Flag | Type | Description |
|------|------|-------------|
| `--debug` | | Print request and response debug details to stderr |
| `--format` | string | Output format: json (default "json") |
| `--page-all` | | Automatically fetch all pages when supported |
| `--page-delay` | int | Delay in milliseconds between pages (default 200) |
| `--page-limit` | int | Maximum pages to fetch with --page-all (default 10) |
| `--profile` | string | Active config profile |
| `--transform` | string | Transform output with a GJSON-style path expression |

## Examples

发起滚动并等待完成：

```bash
arkcli infer endpoint rolling create ep-20260428194457-p8b47 \
  --model doubao-seed-2-0-pro-260328 --wait
```

只发起、不等待（确认由交互终端或宿主 Hook 完成）：

```bash
arkcli infer endpoint rolling create ep-20260428194457-p8b47 \
  --model doubao-seed-2-0-pro-260328 --format json
```

本地预演（零网络，不发请求）：

```bash
arkcli infer endpoint rolling create ep-20260428194457-p8b47 \
  --model doubao-seed-2-0-pro-260328 --dry-run
```

按 endpoint 列出进行中的滚动（表格）：

```bash
arkcli infer endpoint rolling list --endpoint-id ep-20260428194457-p8b47 \
  --status Running --format table
```

取消 / 回滚（高危动作，确认由交互终端或宿主 Hook 完成）：

```bash
arkcli infer endpoint rolling cancel er-xxxx
arkcli infer endpoint rolling rollback er-xxxx
```
