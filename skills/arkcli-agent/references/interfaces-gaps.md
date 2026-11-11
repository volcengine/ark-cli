# 接口链路与 Gap

## 本期接入边界

- BYOK 的模型字段和每次请求凭证已接入，参数及边界见 [byok.md](byok.md)。火山已提供 Dream create/list/get/cancel/wait，参数、单次创建和异步状态见 [dream.md](dream.md)；BytePlus 暂不支持 Dream。
- Todo Write 复用 `--tool`，不属于基础 toolset，也不需要 Models。启用前读 [Todo 配置](todo-write.md)，包含权限前提、仅开启 Todo 的完整条目及保留其他工具的更新方式。
- 国内搜索提供方切换不新增 CLI flag，也不擅自把已有 web_search 改成其他工具；以回读配置和实际工具结果验证。未知工具结果与事件类型应保留原始 JSON，不因不能 pretty 展示而丢弃。
- MA 开通继续走已有服务开通检查和确认；不新增未经确认的开通 Action。trace、原始 kernel event、main/subagent header、出网白名单和 Nacos/TLS 的内部改动不扩展成用户命令；PE tool 的公开契约未确认。

## 接口链路

- Agent: `CreateAgent` / `GetAgent` / `ListAgents` / `UpdateAgent` / `DeleteAgent` / `ListAgentVersions`
- Skill: public SkillHub V1 `GET /v1/skills`（CLI 兼容输出名 `ListMarketSkills`）+ data-plane `POST /api/v3/skills` for `CreateSkill` + TOP `Get/List/DeleteSkill`、`Create/List/Get/DeleteSkillVersion`
- Env: `CreateEnvironment` / `GetEnvironment` / `ListEnvironments` / `UpdateEnvironment` / `DeleteEnvironment`
- Session: `CreateSession` / `GetSession` / `ListSessions` / `UpdateSession` / `DeleteSession`
- Session data-plane: `GET/POST /api/v3/sessions/:session_id/resources`, `GET/POST /api/v3/sessions/:session_id/events`, `GET /api/v3/sessions/:session_id/events/stream`, `GET /api/v3/sessions/:session_id/threads`, `GET /api/v3/sessions/:session_id/threads/:thread_id`
- Files data-plane: `GET/POST /api/v3/files`, `GET/DELETE /api/v3/files/:file_id`
- Memory: `ListMemoryStores` / `CreateMemoryStore` / `ListMemories` / `CreateMemory` 等
- Vault: `ListVaults` / `CreateVault` / `ListOAuthProviders` / `CreateVaultOAuthFlow` / `ListCredentials` / `CreateCredential` 等

## 当前已对齐 / 已有可接受替代

- Agent / Env / Session / Memory / Vault / Credential 主体 CRUD 已有命令面。
- Memory 条目创建使用 `--path`、`--content`，更新支持修改 Path / Content；创建和更新均不支持 tags。Memory 更新已移除 `--tags`，通过 `--file` / stdin 传入的顶层 Tags（大小写不敏感）会被过滤，不会保存或修改标签。
- Environment 自定义脚本已支持：`--setup-script` 写入 `Config.SetupScript`，支持 `@file`。
- Session 创建、`+new session`、`+iterate` 已支持 `AgentWithOverrides` / Environment overrides；对应快捷参数是 `--agent-overrides` 和 `--environment-overrides`，底层仍直联 OpenTOP，不走 BFF。
- Credential ENV 与 OAuth 换签已支持：ENV 通过 `Auth.Type=environment_variable`、`SecretName`、`SecretValue`、`Networking`，OAuth 换签通过 `Auth.Refresh`；敏感值支持 `@file`。
- Skill 搜索走无需鉴权的 SkillHub V1 `GET /v1/skills`，只映射 `query`、`pageNumber`、`pageSize`、`sourceType` 和 `keywords`；custom skill 创建 zip 走数据面 `POST /api/v3/skills`，避免 TOP 大小限制；custom Skill 的版本更新、删除、查询和下载走 OpenTOP Skill/SkillVersion actions。
- Files API 已有 `list/get/upload/wait/delete`；`session resources add --path` 可自动 upload -> wait active -> mount。
- Event send 已支持 `--events` 数组、text/tool confirmation，以及 `--image` / `--document` 多模态便捷参数：file ID 可直接发送，本地路径或 `@file` 会自动上传 Files API、等待 active 后发送。参数互斥与事件格式见 [events-chat.md](events-chat.md)。
- Session events/list/stream、threads/list/get 走数据面直联，不依赖 ArkBFF。
- Session 主体列表已对齐 `ListSessionsForTop`：`--agent-id` 发送 `AgentIds`，`--page/--limit` 发送 `PageNumber/PageSize`，`--page-all` 使用页码连续拉取。
- `+tail` 已有人类可读 pretty 输出；`+new session` 无参数时是 PRD 会话选择器入口，可继续已有 session 或选择 agent/env 起新 session；`+new session <agent-id> --environment-id <env-id>` 固定创建新 session 后 one-shot/stdin/TTY REPL，支持 `/allow`、`/deny`、`/interrupt`。`+chat <prompt>` 保留为 Responses API 快速对话。继续已有 session 也可走 `agent session events send/stream` 或 `+tail`；`+new session`、`+iterate`、`+tail` 的数据面等待由 stream + list 补偿 channel 负责。
- `+debug` 已聚合 session/events/resources/threads；`+export` 已导出可见诊断包。
- `+new-agent --fork` 已支持复制已有 Agent，并默认用 `copy-<source-name>` 命名。
- `+iterate` 已支持 GetAgent 当前版本 -> 可选 UpdateAgent -> CreateSession -> `--message` one-shot / TTY REPL；`--environment-id/--env-id` 是可选参数，未传时真实执行会用 `ListEnvironments` 自动选择当前项目最近创建的 Environment；没有可用环境才要求用户创建或显式传入。
- 火山新建 Agent 不传 `--tool` 时发送 `agent_toolset_20260916`，含 bash/read/write/edit/glob/grep/web_fetch/web_search；传 `--tool` 时按完整数组全量覆盖。旧版本不自动升级；BytePlus 默认仍为 `agent_toolset_20260701`。

## Channel（火山）

Channel 将飞书机器人消息路由到 Managed Agent Session，不是 CLI 自己监听飞书，也不是普通 Session 的别名。本节新命令仅在火山产品提供。

| 命令 | 用途 |
| --- | --- |
| `agent channel create --file ./channel.json` | 创建绑定；也支持 `--name`、`--provider`、`--app-id`、`--policy`、`--session-config`、`--enabled` |
| `agent channel get <id>` / `list` | 查询配置及 Validation；list 支持 `--name/--app-id/--agent-id/--provider/--enabled/--connection-status` |
| `agent channel update <id> --file ./patch.json` | 更新；`--enabled=false` 明确禁用；Provider 不可修改 |
| `agent channel delete <id> --yes` | 用户确认后删除绑定；不会删除历史 Session |
| `agent channel validate --file ./credentials.json` | 校验未保存的 Provider + FeishuApp；带 `--id <id>` 则校验并更新已保存 Channel 的验证状态 |
| `agent channel sessions list <id>` | 关联 Session 列表，支持 `--chat-type`、`--status`（重复）、`--keyword` |
| `agent channel participants list <id> --session-id <session-id>` | 参与者列表，可加 `--keyword`；Session 列表的 Participants 只是摘要，不能据此声称已列全 |
| `agent channel app register --app-name <name>` | 发起新应用注册，已有应用用 `--app-id`，两者不同时传 |
| `agent channel app status <registration-id>` | 单次查询注册状态；严格按返回 PollInterval 间隔查询，在 ExpireIn 内等待用户授权；不是自动授权 |
| `agent channel app config <id>` | 获取飞书应用回调/事件配置；默认脱敏；需要完整配置时按用户授权加 `--output ./private-channel.json` |

所有命令支持显式 `--file` TOP JSON/YAML 对象；不隐式读取 stdin。列表使用 `--limit/--page`，`--page-all` 沿 NextPage 合并 `Result` 数组，并保留 Summary/ProjectionSyncedAt/ProjectionDelayed。ProjectionDelayed=true 表示投影延迟，不是会话不存在。

创建请求示例（必须替换真实资源 ID，密钥不写在对话或命令参数中）：

```json
{
  "Name": "arkcli-feishu-channel",
  "Provider": "feishu",
  "Enabled": true,
  "FeishuApp": {"AppId": "<app-id>"},
  "Policy": {
    "Identity": {"Group": "bot", "P2P": "bot"},
    "Group": {"Trigger": "mention", "ReplyMode": "thread"},
    "Context": {"Enabled": true},
    "Display": {"StreamFinalAnswer": true, "ShowToolCalls": false}
  },
  "SessionConfig": {
    "Agent": {"Id": "<agent-id>", "VersionMode": "latest"},
    "EnvironmentId": "<env-id>",
    "MemoryIds": [],
    "VaultIds": []
  }
}
```

```bash
ARKCLI_NO_UPDATE_NOTIFIER=1 ARKCLI_CALLER_TYPE=ai_agent ARKCLI_CALLER_NAME=codex ARKCLI_SKILL_NAME=arkcli-agent arkcli agent channel create --file ./channel.json --app-secret-env FEISHU_APP_SECRET --dry-run --format json
```

`FEISHU_APP_SECRET` 由用户在安全环境预置，不输出它。用户确认实际接入飞书后才去掉 dry-run。SessionConfig.Agent.VersionMode 为 latest 时省略 Version；pinned 时使用已核实的正整数 Version。MemoryIds/VaultIds 空数组表示不绑定；SessionConfig 是整体替换，更新应提交完整对象。Policy 中显式提供的子对象也是完整替换，不把局部字段当深层合并。

update/delete/已保存的 validate/app config 接受 `--config-version`。省略时 CLI 仅 GetChannel 一次补当前 ConfigVersion；冲突保留错误，不自动刷新后重试。dry-run 会展示依赖查询和占位版本，不实际调用。

凭证 Valid 不表示机器人已连通：分别检查 Validation.Credential 和 Validation.Connection，再让用户从飞书发一条测试消息并回查关联 Session。Enabled=true 也不等于验证通过。失败时保留错误码和当前版本，不反复重建机器人。

注册返回 pending/succeeded/denied/expired/failed；仅 succeeded 表示授权完成。AppSecret、EncryptKey、VerificationToken 默认脱敏，成功凭证需要用户授权后用 status/config 的 `--output` 保存至新的 0600 文件，不能覆盖现有文件或把文件正文贴到对话。注册凭证有有效期，不承诺过期后可重复获取。

用户打开授权页后，按返回的 PollInterval 主动轮询 `app status`，无需等用户再次说“完成了”；以 ExpireIn 为截止时间，遇到 succeeded/denied/expired/failed 停止。pending 时不重复注册；过期或拒绝后先向用户说明，不自动重新申请授权。

### 已创建但连接失败：手动配置与复验

CLI 在 Connection=failed 或获取应用配置时增加顶层 `Guidance`，包含下一步命令；后端 `Result` 和错误码保持原样。退出码 0 只说明本次 API 调用成功，不代表飞书消息链路已接通。保留已有 Channel ID，不重新创建应用或 Channel。

1. 沿用创建时的 profile、环境，执行 `agent channel app config <id> --output ./channel-config.private.json`。文件必须是新路径，完整配置仅保存在本地私密文件中。
2. 打开对应应用的“事件与回调”，先核对安全配置，再配置地址。字段对应如下：

| 飞书页面配置 | 导出文件字段 |
| --- | --- |
| Encrypt Key | `Result.FeishuConfig.EncryptKey` |
| Verification Token | `Result.FeishuConfig.VerificationToken` |
| 事件订阅：发送到开发者服务器 | `Result.FeishuConfig.EventRequestUrl` |
| 回调配置：发送到开发者服务器 | `Result.FeishuConfig.CallbackRequestUrl` |

3. 按页面可编辑项完成配置；不把默认脱敏输出中的 `[REDACTED]` 当真实值。宿主要求人工填写敏感字段时，把这一步交给用户；不要把文件正文、EncryptKey 或 VerificationToken 发到对话中。按页面要求保存、发布。
4. 执行 `agent channel validate --id <id>`，分别查看 Credential 和 Connection。仍失败时保留错误码、RequestId 和当前版本，不能声称已连通。最后从飞书发送测试消息，执行 `agent channel sessions list <id>`，核实对应 Session 和回复。

自动配置报 `app_config_permission_denied`、缺少 `application:application:patch` 时，不要求用户反复搜索一个可能未对其开放的同名权限；可尝试上述手动配置。页面提示“Challenge code 没有返回”表示地址校验未通过，需核对安全参数、地址及服务端响应；仅凭该提示不能断定是密钥错误，也不要通过重复创建 Channel 解决。

不实现未公开的 ResetMapping、Retry 或内部文件投递策略命令；不调用飞书内部 API 绕过服务端权限。

## 当前 Gap

- PRD 中 `arkcli +new-agent "..."` 的 CLI 内置 LLM draft + confirm 未实现；当前由调用 arkcli 的 AI agent 负责自然语言理解、模型/skill/tools 选择，再调用结构化 `+new-agent` 或 `agent agent create`。
- PRD 中 `+outcome` shortcut 未实现。
- `+new session` 选择器已覆盖 P0 交互链路，但 token 数、近 7 天 session 次数等 PRD 展示字段依赖后端返回；当前只展示接口可可靠取得的 id/name/title/status/time/version。
- `+iterate` 尚未实现 PRD 的 TTY environment/resource 选择器和富 diff；当前省略 environment 时自动选择最近创建项，`--diff` 输出结构化请求预览。
- Session resources 原生 get/update/delete 未完全暴露；CLI get 由 list 派生，update/delete unsupported。
- `session resources add` 只封装了 file / local path 体验；`github_repository`、复杂 `memory_store` 等资源只能在 `session create --resource` 或底层 payload 中手写，缺少友好 typed flags 和 add 链路。
- `+export` 中 workspace tarball / memory snapshot 暂无可用读取接口，只能在 manifest 中标 unsupported。
- `arkcli agent +mcp-login` 只对后端 provider 列表中 `CredentialType=mcp_oauth` 的 URL 可靠；static bearer provider 走手动 credential create。Notion / Lark Base 等 provider 仍受后端 metadata discovery 可用性限制。
- PRD 示例中的 inline/local skill 目录形态未做成直接 `--skill '{type:inline,...}'`；当前推荐路径是本地 zip 用 `agent skill create --zip` 或 `agent agent create --skill-zip` 上传成 custom skill。
- `agent agent list` 默认拉单页；需要遍历全部候选时使用全局 `--page-all`，并按数据量设置 `--page-limit`。模糊复制/查找若仍返回多个候选，需要让用户确认。
- Env 创建对线上后端需要显式 `Config.Networking`；PRD 中只传 `{Type: cloud}` 的简写不够用，skill 文档已按线上行为改为 `{Type: cloud, Networking: {Type: unrestricted}}`。
- 数据面 `/api/v3/sessions/:id/events/stream` 支持重复的 `event_deltas=agent.message&event_deltas=agent.thinking` 查询参数；`events stream`、`+tail`、`events send --stream`（兼容 `--wait`）、`+new session` 和 `+iterate` 默认启用，`--no-event-deltas` 可回退到完整事件。CLI 会渲染 `event_start/event_delta` 并在最终完整事件到达时避免重复展示，断线补偿仍通过不带该查询参数的 `events list` 完成。
- 前端的 `StreamManagedAgentSessionEvents` 仍是 ArkBFF 的 POST 流式 preview 协议。CLI 按“不直连 ArkBFF”的约束直联数据面，因此只对齐已确认的数据面 Event Deltas 语义，不依赖或宣称复用 BFF preview 的请求封装。
- `[MA]开通时支持赠送额度包` 在需求单当前状态为已终止，现有 OpenTOP 仅有人工确认后的 `OpenChargeItems`，没有可确认的 gift quota 字段/Action；CLI 保持手动确认开通，不自动伪造赠送额度。
- “企业自定义 Skill”若指当前账号/项目下的 custom Skill，CLI 已通过 TOP `ListSkills` 选择并以 `Type=custom` 挂载；若要求额外企业 Skill 空间/组织 scope，当前公开 `ListSkills` 请求契约没有可确认的 enterprise/space 字段，不能硬接。现有 `--skill-space-id` 不代表后端已支持该过滤字段。
