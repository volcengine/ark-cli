# arkcli infer endpoint delete

Delete an infer endpoint (irreversible)

## Usage

```bash
arkcli infer endpoint delete <endpoint-id> [flags]
```

## Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `<endpoint-id>` | The ID of the endpoint to delete | Yes |

## Flags

| Flag | Type | Description | Required |
|------|------|-------------|----------|
| `--yes` | bool | Skip the interactive Y/N prompt (TTY only). **Non-interactive mode: this flag does NOT authorize deletion** — set `ARKCLI_ALLOW_HEADLESS_DELETE=1` instead. | No |
| `-h`, `--help` | | help for delete | No |

## 确认机制（高危，必读）

删除是不可逆操作。`infer endpoint delete` 用**三分支确认门**：

| 环境 | 行为 |
|------|------|
| 交互终端（TTY） | 弹强警告 + `[y/N]`，输入 `y` 才执行；`--yes` 跳过提问当确认 |
| 非交互（agent / CI / 管道）+ `ARKCLI_ALLOW_HEADLESS_DELETE=1` | 打审计行后放行（给真·无人值守自动化） |
| 非交互 + 未设 env | **hard-refuse**（`requires_headless_gate`）；`--yes` 在此被无视 |

设计意图：任何 agent 都能反射地加 `--yes`，但 agent 的 stdin 不是真 TTY。所以「真 TTY 或显式 env」是 agent 造不出来的硬门槛，防止 agent 自批 `--yes` 替用户静默删除。对齐 `infer endpoint create` 开通的 `ARKCLI_ALLOW_HEADLESS_ACTIVATION` 模式。

### Agent 调用须知

agent（Claude Code / OpenCode 等）通过 skill 调用 `infer endpoint delete` 时是非交互环境：

- **默认 hard-refuse**，即使带 `--yes` 也删不了。
- 要在 agent / CI 中删除，必须显式设 env：

  ```bash
  ARKCLI_ALLOW_HEADLESS_DELETE=1 arkcli infer endpoint delete <id> --yes
  ```

- 交互式真人终端直接 `arkcli infer endpoint delete <id>`，按提示确认即可。

### dry-run

`--dry-run` 是只读预演：先 `GetEndpoint` 拿详情展示，不调真删 API、不弹确认。火山 `delete_endpoint` 不支持服务端 dry-run（raw request 只有 `Id`），故用客户端模拟。

```bash
arkcli infer endpoint delete <id> --dry-run
```

### 删除前建议

删除前 endpoint 若为 `Running` 状态，部分场景需先 `stop` 再 `delete`；状态不允许时后端返回 `OperationDenied.EndpointStatus`，arkcli 翻译成「endpoint status does not allow this operation」并提示用 `infer endpoint get <id>` 检查状态。

## Global Flags

| Flag | Type | Description |
|------|------|-------------|
| `--api-key` | string | ARK API key override |
| `--base-url` | string | Custom API base URL |
| `--debug` | | Print request and response debug details to stderr |
| `--dry-run` | | Preview request metadata without executing when supported |
| `--format` | string | Output format: json (default "json") |
| `--page-all` | | Automatically fetch all pages when supported |
| `--page-delay` | int | Delay in milliseconds between pages (default 200) |
| `--page-limit` | int | Maximum pages to fetch with --page-all (default 10) |
| `--profile` | string | Active config profile |
| `--region` | string | ARK region override |
| `--transform` | string | Transform output with a GJSON-style path expression |
