# arkcli infer endpoint start

Start an infer endpoint endpoint

## Usage

```bash
arkcli infer endpoint start <endpoint-id> [flags]
```

## Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `<endpoint-id>` | The ID of the endpoint to start | Yes |

## Flags

| Flag | Type | Description | Required |
|------|------|-------------|----------|
| `-h`, `--help` | | help for start | No |

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