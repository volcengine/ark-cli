# auth login

## 推荐顺序

交互式（推荐）：

```bash
arkcli auth login
```

明确指定方式时：

```bash
arkcli auth login volc-sso                                      # 火山 SSO (浏览器)
arkcli auth login --no-browser                                  # 火山 SSO 无浏览器 (TTY 交互 / 非 TTY 出 Phase 1)
arkcli auth login --no-browser --code <授权码>                  # 无浏览器 Phase 2 (agent/沙箱喂码完成登录)
```

> **0.1.16 暂关 AK/SK 登录通道**：`--access-key/--secret-key` flag 已注释、交互菜单也移除了 "AK/SK" 选项。CI / 自动化 / agent / 沙箱场景请改用 `arkcli auth login --no-browser`（见下方「无浏览器两段式」）。**`--no-browser` / `--code` 由 `auth login` 注册并可被 `volc-sso` 继承；也支持 `auth login volc-sso --no-browser`。** AK/SK 通道恢复将在后续版本同步更新本文档。

## 无浏览器两段式跨进程 flow（agent / 沙箱 / CI 必读）

**根因**：老 `--no-browser` 是单进程交互式 —— 打印 URL 后当场 `阻塞读 stdin` 等你粘授权码。agent / 沙箱里每条命令是独立短命进程，stdin 不是真人终端（常是 `/dev/null`），`读 stdin 立刻 EOF`，进程在拿到授权码前就崩（`读取授权码失败: EOF`）。解法是把单进程拆成**两段跨进程命令**，中间用一次性落盘 session 接力 PKCE（对齐 lark-cli `auth login --no-wait` / `--device-code` split-flow）。

> ⚠️ **两段必须在同一运行环境执行**：Phase 1 把 `code_verifier`+`state` 落到 `~/.arkcli/.sso-pending.json`，Phase 2 靠读这个文件接力。**两条命令必须共享同一 `HOME` / 同一持久化卷**（同一容器 / 同一开发机）。若无状态调度把两步派到不同 worker、不同容器或不同 `HOME`，Phase 2 找不到 pending 会报「没有待完成的…会话」。

**CLI 按 stdin 是否 TTY 自动分流，agent 无需判断**：

```text
arkcli auth login --no-browser [--code X]
  ├─ 有 --code X ──────────► Phase 2: 读盘换 token + 激活（须与 Phase 1 同一 HOME）
  └─ 无 --code
       ├─ 非 TTY(agent/沙箱) ► Phase 1: 出 URL + 落盘 PKCE/state + 立即退出（不读 stdin）
       └─ 是 TTY(真人) ──────► 一条命令交互式（打印 URL → 当场粘贴），行为同旧版
```

> `--code` **必须与 `--no-browser` 一起用**（它只是无浏览器 Phase 2 的喂码入口）；单独写 `auth login --code X` 会报 `--code 仅在 --no-browser 模式下有效`。

**agent 执行步骤**：

1. **Phase 1** —— 跑 `arkcli auth login --no-browser`。stdout 是结构化 JSON：
   ```json
   {"stage":"authorize_pending","method":"sso_no_browser",
    "authorize_url":"https://signin.volcengine.com/authorize/oauth/authorize?...",
    "next_command":"arkcli auth login --no-browser --code <授权码>","expires_in_sec":600}
   ```
   把 `authorize_url` **原样**转发给用户（不要改写/编解码），请他在任意设备浏览器完成 SSO，复制页面显示的 **base64 授权码**回来。命令已退出（exit 0），不会挂住。
2. **Phase 2** —— 拿到授权码后跑 `arkcli auth login --no-browser --code <授权码>`（`--code` 必须连 `--no-browser`，且与 Phase 1 同一 `HOME`）。CLI 读 Phase 1 落盘的 `~/.arkcli/.sso-pending.json`（含 PKCE `code_verifier` + `state`）换 token、激活身份、删盘。成功后输出 `{"auth_method":"sso_no_browser","path":"..."}`。

**出错恢复（先看 pending 还在不在，再决定重跑 Phase 1 还是重试 Phase 2）**：

| 报错 | pending 状态 | 怎么办 |
|------|-------------|--------|
| 会话过期（超 `expires_in_sec`，TTL 默认 10min） | 已被清 | **重跑 Phase 1** 拿新 URL |
| `没有待完成的…会话` / pending 不存在（含 Phase 1、2 不在同一 HOME） | 无 | **重跑 Phase 1**（并确保两段同一 HOME） |
| `base64 解码失败` / `state 不匹配(CSRF)` / token 交换失败 | **仍在盘** | **TTL 内直接重试 Phase 2**：纠正后重跑 `--code`——贴对当前 `authorize_url` 的码；token 交换瞬时失败可让用户从**同一 authorize_url** 重新授权拿新码（PKCE/state 仍有效，**无需**重跑 Phase 1） |

> `state 不匹配(CSRF)` 最常见原因是粘了**上一轮 / 别处** authorize_url 的码，不是攻击；pending 未删，粘当前这轮的码重试 Phase 2 即可。真疑似被注入时再手动重跑 Phase 1 换全新会话。

**其它**：
- 凭证可移植仍可用：有浏览器的机器登录后把 `~/.arkcli/` 拷进沙箱复用，与两段式互补。

## 说明

- `volc-sso` 适合火山方舟交互式开发
- `--no-browser` 是真正的 cross-device 流（对齐 volcengine-cli `ve login --remote`）：
  - CLI 不监听任何端口
  - 用户在 **任意设备**（手机/远程电脑/平板）浏览器打开 authorize URL
  - 完成登录后，浏览器页面显示一段 **base64 字符串**（不是 URL，不是 6 位数字）
  - 把这段 base64 串粘贴回 CLI 提示符 `授权码:` 处即可
- 登录成功后凭证写入 identity store，并绑定或切换到对应 tenant 的 profile（同 tenant 重登更新当前 profile 的 identity_key；跨 tenant 登录创建以 tenant 命名的新 profile 并设为 default）
- `arkcli auth apikey` 只用于选择普通 API Key；`arkcli auth apikey create` 只用于创建、验证并保存普通 API Key 到当前身份，非交互环境收到 `requires_confirmation` 后必须走共享 Skill 的宿主确认流程。已有 Profile 要切换默认 Key 时仍用 `profile keys refresh/use`。这些命令都不能修复 Agent Plan 个人版专属 Key 或团队版 Running 席位 Key，普通池里有 Key 不代表 Plan profile 可用。
- SSO 首次登录会按 profile 类型从不同来源拉 Key，并把选中的 Project 写入对应 profile：`platform` / `coding-plan` 个人版走普通池，`agent-plan` 个人版走 `Scene=RealAgentPlanPersonal`，团队版走 `GetSeatInfo.Result.ApiKey`。
- 真人 TTY 下，普通池为空会询问是否创建一把全资源 Key；Project 使用当前具体值，账号全部资源/空值回退 `default`，名称与控制台一致为 `api-key-YYYYMMDDHHmmss`。确认后只写一次，再最多读 3 次状态；拒绝或非 TTY 时不创建。
- Agent Plan 个人版若专属记录存在但状态非 Active（如 `Restricted`），TTY 会说明旧 Key 将立即失效并询问是否轮转。确认后只轮转一次，再最多读 3 次状态；第一轮 Active 即结束。无专属记录、仍有 Active 记录但明文不可读、拒绝或非 TTY 时不轮转。
- 如果所选 profile 的 API Key 获取报错（例如 Agent Plan 轮转后检查 3 次仍为 `Restricted`），TTY 不再让该局部故障回滚整个 SSO：已选 region/project 保持不变，失败类型从本轮候选中移除，然后重新展示消费场景选择。失败类型不会在后续 backfill 中再次调用 fetcher；用户改选 Coding Plan/platform 时各自仍使用自己的 Key 池，不发生跨池降级。
- 临时选择另一套已存在的上下文使用 `--profile <name>`；永久切换使用 `arkcli profile use <name>`
- 需要重选 Project 时使用 `arkcli profile project [<name>]`。根命令不再提供 `--project-name`，`ARK_PROJECT_NAME` 也不参与运行时解析

## volcengine-cli 登录态借用 (v3, 1.0.4 起, 仅火山 SSO 交互式登录)

**触发条件** (三个都满足才走):
1. 用户跑 `arkcli auth login` 交互式菜单里选**浏览器 SSO** (case 0) 或**无浏览器 SSO** (case 1); 显式 `arkcli auth login volc-sso --no-browser` 不走 (它跳过交互菜单直接进 OAuth)
2. 本机 `PATH` 上有 `ve` 可执行 (`exec.LookPath`), `ve -v` 版本 >= `1.0.45`, `ve --help` 输出含 `ecs` 关键字 (火山官方 CLI 指纹)
3. `ve` 已 `ve login` 且当前登录态可用 (SDK `clicreds.NewCliCredentials("", "").Get()` 返 STS 且 `sts.GetCallerIdentity()` 返 `IdentityType="User"`)

**接管动作**:
1. 复用 ve 的 STS + IAM 身份 (AccountID / UserID / Trn / UserName / IsRoot 来自 `GetCallerIdentity()` 服务端权威值)
2. 通过 `sso.ActivateIdentity` (source=ve 分支) 落盘 arkcli identity: `identity_store/volc-<AccountID>/metadata.json` (`source="ve"` + 身份事实) + `sts.json` (STS 三元组 + expires_at_ms) + `user_id.json`; **不落 `token.json`** (arkcli 侧不再持 IDToken/refresh_token/ClientID, 这些字段由 volcengine-go-sdk 内部管)
3. 触发 `BuildFirstProfileSet` 派生 platform / agent-plan / coding-plan profile (跟原生 SSO 一样)
4. `.env` 只落 STS 三元组 + expires_at + user_id, **不落** `VOLCENGINE_ID_TOKEN` / `VOLCENGINE_REFRESH_TOKEN` / `VOLCENGINE_OAUTH_CLIENT_ID`

**降级条件** (任一不满足即降级 arkcli 原生 SSO OAuth 流, **不阻断登录**):
- 未装 ve / 版本 <1.0.45 / 不是火山官方 CLI (指纹校验失败)
- ve 未登录 / login cache 缺 refresh_token / refresh_token 过期 (SDK 抛 `CliConsoleLoginRefreshTokenExpired` / `CliConfigLoginSessionMissing` 等 8 个白名单 code 之一)
- ve 登录态是 SSO Identity Center (`IdentityType != "User"`, 语义跟 arkcli IAM 用户对不上)
- 借用途中 `sts.GetCallerIdentity()` 网络失败 / SDK IO 异常等环境异常

**跟原生 SSO 的差异对照**:

| 维度 | 原生 SSO | v3 ve handoff |
|------|---------|--------------|
| 浏览器授权 | 需要 (每次登录) | **不需要** (借用 ve 已有 refresh_token 状态) |
| identity_store 落盘 | `token.json` + `sts.json` + `user_id.json` + `metadata.json.source="arkcli"` | 无 `token.json`; `sts.json` + `user_id.json` + `metadata.json.source="ve"` |
| `.env` 里 IDToken/RT/ClientID | 有 | **没有** |
| STS refresh | arkcli 用 `token.json` 的 refresh_token 走 OIDC endpoint 换 | volcengine-go-sdk 内部持 refresh_token 自动 refresh (arkcli 侧只调 `clicreds.Get()`) |
| `auth whoami.auth_method` | `"sso"` | `"sts"` (合法登录态, `logged_in=true`) |
| 身份切换语义 | SSO clean-slate | 同上 (source 字段不影响 identity_key 判定) |

**Agent 使用要点**:
- 用户跑 `arkcli auth login` 后 `auth whoami` 显示 `auth_method="sts"` 时**不要**误以为是"未登录 / init-volc / AK/SK 态", 那可能是 v3 ve handoff 正常结果 — 看 `identity_store/<key>/metadata.json.source` 值区分 (`"ve"` = 借用 ve, `"arkcli"` = 原生 SSO)
- 引导用户重登时不再需要主动说"先跑 ve login" — arkcli 会自动检测 + 降级
- `arkcli auth logout` 只清 arkcli 自己的 identity_store, **不动 ve 的登录态** — 用户如果想彻底登出, 还需要额外跑 `ve logout`

## --no-browser 跟浏览器流的 OAuth2 差异（agent 排障用）

| 维度 | `volc-sso`（浏览器） | `--no-browser`（cross-device） |
|------|-----|-----|
| client_id | `trn:signin:::devtools/same-device` | `trn:signin:::devtools/cross-device` |
| redirect_uri | `http://127.0.0.1:<port>/oauth/callback` | `https://signin.volcengine.com/authorize/oauth/authorize` |
| CLI 本地端口 | 需要（callback server） | 不需要 |
| 用户粘贴内容 | — (自动 callback) | base64 串 |
| refresh token | 必须用 same-device client_id refresh | 必须用 cross-device client_id refresh |

`.env` / identity store 都持久化了 `client_id`（`VOLCENGINE_OAUTH_CLIENT_ID` / `IdentityToken.client_id`），后续 `TryRefreshVolcToken` 会用对的 client_id 续期。**用户/agent 不应手动改这个字段**，否则 refresh 会被 server 拒（RFC 6749 §6）。
