# arkcli infer endpoint list

List infer endpoint endpoints

## Usage

```bash
arkcli infer endpoint list [flags]
```

## Flags

| Flag | Type | Description | Required |
|------|------|-------------|----------|
| `--format` | string | Output format: table \| json | No |
| `--model` | string | Filter by model ID (custom model) | No |
| `--status` | string | Filter by endpoint status | No |
| `--mine` | bool | 仅列出当前 SSO 子账号创建的接入点（服务端 `sys:ark:createdBy` Tag 过滤） | No |
| `--page-number` | int | Page number (>=1) | No |
| `--page-size` | int | Page size | No |
| `-h`, `--help` | | help for list | No |

## "我的接入点" 语义（覆盖 shared default）

本命令**已内建 `--mine`**，按 [`../../arkcli-auth/references/identity-resolution.md`](../../arkcli-auth/references/identity-resolution.md) 的"决策顺序"必须**优先**走它，
不要再退回 `whoami + jq` 客户端过滤。

```bash
arkcli infer endpoint list --mine --page-all --page-size 100 --page-delay 800 --format json
```

行为：

- `--mine` 在 shortcut 层把当前 `cfg.UserID` / `cfg.UserName` 拼成 `IAMUser/<UserID>/<UserName>`，
  作为 `TagFilters.{Key: "sys:ark:createdBy", Values: [...]}` 下发到 `ListEndpoints`。
- 与 `--status` / `--model` / `--page-*` 可自由叠加（例如 `--mine --status Running`）。
- 仅在 **SSO 子账号**登录态下可用。其它态会**直接报错**，不静默退化：
  - 非 SSO（AK/SK / APIKey）：`--mine requires an SSO sub-user login`，按 profile tenant 引导登录（火山：`arkcli auth login volc-sso`）
  - SSO root：`--mine is not supported for root logins`，引导改用 sub-user 登录
  - SSO 但 UserID/UserName claim 缺失：引导重登刷新身份

## Global Flags

| Flag | Type | Description |
|------|------|-------------|
| `--api-key` | string | ARK API key override |
| `--base-url` | string | Custom API base URL |
| `--debug` | | Print request and response debug details to stderr |
| `--dry-run` | | Preview request metadata without executing when supported |
| `--page-all` | | Automatically fetch all pages when supported |
| `--page-delay` | int | Delay in milliseconds between pages (default 200) |
| `--page-limit` | int | Maximum pages to fetch with --page-all (default 10) |
| `--profile` | string | Active config profile |
| `--region` | string | ARK region override |
| `--transform` | string | Transform output with a GJSON-style path expression |