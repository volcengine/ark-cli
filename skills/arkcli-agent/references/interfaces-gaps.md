# 接口链路与 Gap

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
- 默认 Tools 已对齐当前实现：不传 `--tool` 时发送 `agent_toolset_20260701`，含 bash/read/write/edit/glob/grep/web_fetch/web_search；传 `--tool` 时按完整数组全量覆盖。

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
