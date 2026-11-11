# MCP 与 Vault

## 查询可挂 MCP

后端已注册、可被 Agent 使用的 MCP provider 通过下面命令查：

```bash
arkcli agent vault oauth-provider list --limit 100 --format json
```

- 返回结果里的 MCP server 信息可以作为创建 / 更新 Agent 时 `--mcp-server` 的来源；不要凭空猜 URL。
- 如果用户想“给 Agent 加 GitHub / Notion / 飞书等 MCP”，先查 provider 列表，按 provider 的名称、URL、credential 类型筛选。
- 挂到 Agent 时通常还要补一条 `mcp_toolset` 工具配置，`mcp_server_name` 要和 `--mcp-server` 里的 `name` / `Name` 对齐。
- 如果 provider 需要 OAuth，先按下面的 OAuth MCP 登录链路拿 credential；如果是 static bearer，走 `agent vault credentials create`。

## ENV Credential 与换签

Credential 的 `Auth.Type` 支持 `environment_variable`。CLI 提供了和前端表单等价的快捷参数；敏感值支持 `@file`，不要写进 shell 历史：

```bash
arkcli agent vault credentials create <vault-id> \
  --display-name github-token \
  --secret-name GITHUB_TOKEN \
  --secret-value @./github-token.txt \
  --networking-type limited \
  --allowed-host api.github.com \
  --format json
```

这会构造 `Auth: {Type: environment_variable, SecretName, SecretValue, Networking}`。`--networking-type` 只能是 `limited` 或 `unrestricted`；limited 可以重复传 `--allowed-host`，unrestricted 不应同时传 host。

MCP OAuth 的托管换签字段也可直接传入：

```bash
arkcli agent vault credentials create <vault-id> \
  --auth-type mcp_oauth \
  --mcp-server-url https://example.com/mcp \
  --access-token @./access-token.txt \
  --client-id <client-id> \
  --client-secret @./client-secret.txt \
  --refresh-token @./refresh-token.txt \
  --token-endpoint https://example.com/oauth/token \
  --scope read \
  --format json
```

这会写入 `Auth.Refresh`，其中 client secret 位于 `TokenEndpointAuth.ClientSecret`，由后端负责托管换签。也可以继续使用 `--auth @./credential.json` 透传完整对象；快捷参数会覆盖对象中同名字段。

## ENV OAuth（仅火山版）

ENV OAuth 将 OAuth access token 作为 `SecretName` 指定的环境变量注入沙箱；它不是 MCP OAuth，不需要 `McpServerUrl`，不要使用 `+mcp-login`。

### 用户需要浏览器授权

先通过 `agent vault oauth-provider list` 查询，选择 `CredentialType=env_oauth` 的 provider，使用返回的真实名称，不要用 MCP provider 名称替代。确认目标 Vault、权限 scope 和出网范围后发起：

```bash
arkcli agent vault env-oauth-flow create \
  --vault-id <vault-id> \
  --display-name skill-oauth \
  --provider-name <registered-env-oauth-provider-name> \
  --secret-name SKILL_ACCESS_TOKEN \
  --networking-type limited \
  --allowed-host api.example.com \
  --redirect-url <backend-allowed-business-success-url> \
  --format json
```

- `--secret-name` 是沙箱环境变量名。`--networking-type limited` 配合真实目标域名的 `--allowed-host`；`unrestricted` 不传 host，不要为绕过校验扩大出网范围。
- `--redirect-url` 是**授权完成后的业务成功页**，须符合后端允许列表；不是上游 OAuth 应用登记的 callback URI，不要自行拼本地 callback。上游 callback 由后端配置。
- 预置模式传 `--provider-name`。自定义模式不传 provider-name，改为同时传 `--authorize-endpoint`、`--token-endpoint`、`--client-id`、`--client-secret @./client-secret.txt`；两种模式互斥。`--scope` 可选，`--source` 可标记调用来源。
- 支持 `--file <JSON/YAML路径>` 提供完整 TOP CamelCase 请求，flags 覆盖同名字段；不接受 `--file -`。支持叶子 `--dry-run`，只生成脱敏客户端计划，不能验证 provider、URL 白名单或权限。
- 返回 `Result.RedirectUrl` 和 `Result.OAuthFlowId` **只表示流程创建成功，不是凭证已创建**。请用户打开返回的授权链接；CLI 不自动启动浏览器或等待。用户完成后用 `agent vault credentials list <vault-id>` 确认新凭证，不因等待而重复创建流程。不打印 token 或 client secret。
- 授权被拒绝、超时或后端报错时保留原错误并停止，不偷偷改用手工 token 模式。

### 用户已有合法 token，需要直接保存

```bash
arkcli agent vault credentials create <vault-id> \
  --display-name skill-oauth \
  --auth-type env_oauth \
  --secret-name SKILL_ACCESS_TOKEN \
  --access-token @./access-token.txt \
  --refresh-token @./refresh-token.txt \
  --client-id <client-id> \
  --client-secret @./client-secret.txt \
  --token-endpoint https://example.com/oauth/token \
  --networking-type limited \
  --allowed-host api.example.com \
  --format json
```

`--provider-name` 可选；`--access-token` 写入 `Auth.AccessToken`，不是 `SecretValue`。刷新配置写入 `Auth.Refresh`，仅在用户确有对应配置时传入；没有 refresh token 不要伪造。更新使用 `credentials update <vault-id> <credential-id>`，支持相同参数。`--auth`/`--file` 也可提供完整 Auth；混用快捷字段时保留显式 `Auth.Type=env_oauth`。建议始终明确传 `--auth-type env_oauth`，避免和普通 ENV 混淆。

## OAuth MCP 登录

凭证必须与目标 provider 的服务语义匹配；不能为了让创建通过，把某服务的 Key 绑定到
无关 provider，也不能写入 metadata/tags。不要逐个 curl/web_fetch 探测 provider URL，
连接与授权结果以对应创建/登录命令的真实返回为准。OAuth 是外部授权流，不用聊天选择
代替；metadata discovery 或授权失败时保留错误并停止，不擅自改用 token 绕过。
多个符合用户目标的 provider 按共享协议消歧，使用本轮名称、URL、CredentialType；
不能从记忆补出未注册的 provider。

不要凭空猜 MCP OAuth URL。正确顺序：

1. 查后端 provider：

```bash
arkcli agent vault oauth-provider list --limit 100 --format json
```

2. 只对 `CredentialType=mcp_oauth` 的 provider 使用 `arkcli agent +mcp-login`。当前实测可用示例是 GitHub MCP：

```bash
arkcli agent +mcp-login \
  --vault-id vlt-xxx \
  --name arkcli-github-mcp-oauth-<timestamp> \
  --mcp-server-url https://api.githubcopilot.com/mcp/ \
  --format json
```

3. `arkcli agent +mcp-login` 会本地启动 callback server，调用 `CreateVaultOAuthFlow`，打印/打开授权 URL，等待后端 callback 创建 credential，成功后输出 `credential_id`。
4. 不方便自动开浏览器时加 `--no-open`，手动复制授权 URL。
5. 完成后用 `agent vault credentials list <vault-id>` 查看已创建的 credential；需要查看非敏感字段时再使用 `get`。

注意：

- `CredentialType=static_bearer` 的 provider 不适合 `arkcli agent +mcp-login`，应走 `agent vault credentials create` 写 bearer auth。
- 不要把用户 token 写进 agent metadata / tags。
- `agent vault credentials get` 可能回显敏感 auth 字段，除非用户明确要求排障，否则优先用 `list`。
- 实测 `https://mcp.notion.so/v1`、`https://mcp.larksuite.com/base` 在当前后端 metadata discovery 中失败；不要作为默认示例。

## Agent 挂 MCP

- 只有用户明确给 MCP server URL/name/credential，或上下文明确要求时才写 `--mcp-server`；如果用户只说要某类 MCP，先用 `agent vault oauth-provider list` 查可用 provider。
- 挂 MCP 通常需要同时写 `McpServers` 和 `Tools` 中的 `mcp_toolset`。
- 注意：`--tool` 是完整 `Tools` 数组替换。需要保留默认 bash/read/write/edit/glob/grep/web_fetch/web_search 时，必须把完整默认 `agent_toolset_20260701` 和 `mcp_toolset` 放在同一个数组里一起传，不要只传单个 `mcp_toolset`，也不要只写 `{type: agent_toolset_20260701}` 这种不完整占位。
- 不确定默认 tools 结构时，参考 `references/agent.md#默认-agent-工具`；先用同一条 `agent agent create ... --dry-run` 检查 `preview.v1` 中的完整默认 `Configs` 与追加的 `mcp_toolset`，再经用户确认后真实创建。
- 下例的 `tools-with-default-agent-toolset-and-github-mcp.json` 不是内置文件，表示调用方自行准备的完整 Tools 数组文件。

```bash
arkcli agent agent create \
  --name arkcli-mcp-agent \
  --model <items[].model from agent model list> \
  --mcp-server '{type: url, name: github, url: https://api.githubcopilot.com/mcp/}' \
  --tool @tools-with-default-agent-toolset-and-github-mcp.json
```

- `MCPConnectionFailed` 多数是 URL、协议、鉴权、`tools/list` 初始化失败，不等于 Agent 创建失败。

## 示例

```bash
arkcli agent vault oauth-provider list --limit 100 --format json
arkcli agent vault create --display-name arkcli-mcp-login-vault-20260706 --format json
arkcli agent +mcp-login --vault-id vlt-xxx --name arkcli-github-mcp-oauth-20260706 --mcp-server-url https://api.githubcopilot.com/mcp/ --format json
```
