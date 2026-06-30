---
name: arkcli-auth
version: 1.3.0
description: "arkcli 认证管理：交互式登录、Volc SSO 登录、查看状态、退出登录、生成 ARK API Key (apikey)、以及云开发机/CI 用 `arkcli init-volc` 从 VOLC_INIT_* 环境变量无交互引导 platform profile。0.1.16 起 SSO 登录走 Gate 1+2 自动绑定 Profile 切面 (type/region/project/owner_trn)；AK/SK login 通道暂关。当用户需要初始化凭证、排查鉴权问题、切换认证方式、生成或重选 ARK API Key、或在已注入凭证的环境无交互引导时使用。反触发：用户问 TTS/ASR/语音模型能力、接入或调用时，不要引导 `auth apikey`，只转 models search 说明 arkcli 当前仅支持广场发现。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli auth --help"
---

# arkcli auth

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、配置排查与共享安全规则**
**CRITICAL — 用户目标是其他业务命令时，必须先判断是不是被认证阻塞，再决定是否进入本 skill。**

**⚠️ 0.1.16 变化总览（必读）**：
1. SSO 登录引入 **Gate 1+2**：浏览器流后比对 SSO trn 与 `is_default profile.OwnerTrn`，4-case 分别走 `BuildFirstProfile` (新建) / `GUIDE_SKIP` (复用) / 提示切 default / 提示新建。详见 `docs/volc-sso.md`。
2. **AK/SK 登录通道暂关**：`auth login --access-key / --secret-key` 已注释,promptui 也移除"AK/SK"选项；SSO（火山）+ `arkcli auth login --no-browser`（**根命令 flag, 不是 volc-sso 子命令 flag**）是唯一登录入口。
3. **auth status / auth whoami 输出新增 profile 切面字段**：`active_profile.{name,type,region,project,owner_trn}` 和 `profiles_summary[...]`；顶层 `auth_method/logged_in/volc_sso/ark_api_key` 等老字段全部保留（向后兼容）。
4. **Profile 管理迁移到 `arkcli profile`**：`config init/list/show/switch/delete` 已 deprecated，详见 [`../arkcli-config/SKILL.md`](../arkcli-config/SKILL.md)。
5. **0.1.17 首登动态选 project**：`BuildFirstProfile` 的 project 步骤改为经 IAM `ListProjects` 拉当前身份名下真实 active project 列表交互选（拉取失败/无权限回退兜底 `default`，不阻断登录）。登录后想换 project 不必重登：`arkcli profile project [<name>]`（拉同一列表重选，把 platform profile 重派生到新 project，个人版 plan profile 保留），详见 [`../arkcli-profile/SKILL.md`](../arkcli-profile/SKILL.md)。

## 适用场景

- 第一次登录 `arkcli`
- 切换到 Volc SSO
<!-- AK/SK 通道在 0.1.16 暂关 (见上方说明); 恢复后取消下条注释 -->
<!-- - 使用 AK/SK 做非交互登录 -->
- 登录后重新获取或切换 ARK API Key
- 查看当前凭证状态
- **回答"我是谁 / 我的 IAM ID 是多少 / 我属于哪个账号"——用 `arkcli auth whoami`**
- 清理本地登录状态
- 其他业务 skill 因未登录、凭证过期、身份不匹配而被阻塞
- **云开发机 / CI 已注入 `VOLC_INIT_*` 凭证，无交互引导** —— 用 `arkcli init-volc`（不是 SSO）

## 无交互引导（init-volc）

云开发机 / CI 等**已经把火山凭证注入成 `VOLC_INIT_*` 环境变量**的场景，用 `arkcli init-volc` 一条命令、零交互地落一个火山 platform profile 并设为 default，让后续 arkcli / OpenCode 调用开箱即用（数据面用 API Key，控制面用 STS）。

- 触发词："云开发机引导 / 无交互初始化 / 已注入 VOLC_INIT 怎么让 arkcli ready / CI 里跳过 SSO"
- 跟 `auth login` 的区别：`init-volc` **不登录、不交互、不联网**，纯消费环境变量；本地终端用户首次引导仍走 `auth login`（SSO）
- 详细环境变量契约、落地行为、输出见 [`references/arkcli-auth-init-volc.md`](references/arkcli-auth-init-volc.md)

## Agent 快速执行顺序

1. 业务命令开始前如果不确定认证状态，先执行 `arkcli auth status`
2. 需要识别当前用户身份（"我创建的 / 我的 xxx" 语义）时，用 `arkcli auth whoami`，不要去 `~/.arkcli/.env` 里手动解 JWT
3. **未登录或凭证失效时，火山方舟场景直接通过 Bash 执行 `arkcli auth login volc-sso`**（不要只是"提示用户去跑"）。SSO 是 0.1.16 唯一可用登录通道，覆盖控制面 BFF + 数据面绝大多数能力
   - 执行前用一句话告知用户："检测到未登录，我现在为你启动 SSO 登录，请在弹出的浏览器中完成授权"
   - Bash 调用必须设 `timeout=600000`（10 分钟）
   - 启动失败（浏览器没装、`open` 失败、端口被占用、超时等）→ 不原地重试，把 stderr 贴回给用户，请用户手动在终端跑对应租户的登录命令（火山：`arkcli auth login volc-sso`）
   - 命令成功后立即回到用户原始任务，不要停在 auth 结果
4. **agent / 沙箱 / CI 无浏览器登录走两段式 `--no-browser`**（AK/SK 通道 0.1.16 暂关，提到 AK/SK 时告知并引导走这里）：agent 终端通常非 TTY，`--no-browser` 现为**两段式**，不再阻塞读 stdin（旧版在沙箱必报 `读取授权码失败: EOF` —— 进程在拿到授权码前就被 EOF 打断）：
   - **Phase 1**：跑 `arkcli auth login --no-browser`。它打印授权 URL 并以 JSON 输出 `{"stage":"authorize_pending","authorize_url":"...","next_command":"..."}` 后**立即退出**（不傻等）。把 `authorize_url` 原样转发给用户，请他在**任意设备**浏览器完成 SSO，复制页面显示的 **base64 授权码**回来。
   - **Phase 2**：拿到授权码后跑 `arkcli auth login --no-browser --code <授权码>` 完成登录（读 Phase 1 落盘的 PKCE/state 换 token）。`--code` **必须连 `--no-browser`**（单独写会报 `--code 仅在 --no-browser 模式下有效`）。
   - **两段必须同一运行环境**：Phase 1 落盘 `~/.arkcli/.sso-pending.json`、Phase 2 读它接力，**两条命令须共享同一 `HOME` / 同一持久化卷**（同一容器 / 开发机）；跨容器或跨 HOME 派发会让 Phase 2 报「没有待完成的…会话」。期间不要动 `~/.arkcli/`。
   - **flag 位置**：`--no-browser` / `--code` 都挂在 `auth login` **根命令**上，不是 `volc-sso` 子命令；`auth login volc-sso --no-browser` 会报 `unknown flag`。
   - **出错恢复**：`会话过期(TTL 10min)` / `没有待完成的会话` → **重跑 Phase 1**（确保同一 HOME）；`base64 解码失败` / `state 不匹配(CSRF)` / token 交换瞬时失败 → pending 仍在盘，**TTL 内纠正后直接重试 Phase 2**（贴对当前 authorize_url 的码，必要时从同一 URL 重新授权拿新码），**不必**重跑 Phase 1。
   - 真人终端（TTY）下 `arkcli auth login --no-browser` 仍是一条命令交互式（打印 URL 后当场粘贴），行为不变。
5. 用户明确要求挑选登录方式时，因 stdin 交互 agent 无法替用户选，让用户自己跑 `arkcli auth login`
6. 只有在用户明确要求清理本地凭证时，才执行 `arkcli auth logout`

## 核心规则

- `auth status` 是默认入口；不要上来就 `login`
- 登录成功后，应回到用户原始目标继续执行，而不是停在 auth 结果本身
- `auth logout` 是破坏性操作，必须由用户明确提出
- `auth status` 会对敏感字段做掩码，可直接用于排障，并会展示当前生效的 `project_name`
- `auth login` 成功后会输出 `auth_method`；凭证存储位置是实现细节，不再回显路径
- SSO 登录（`arkcli auth login volc-sso`）与 `auth apikey` 都会在选中 API Key 之后写入凭证存储；0.1.16 final clean-slate 模型: 整 arkcli 同一时间只 active 一个 identity, 新 SSO 跟旧 sub 不一致时清空所有 profile (含跨 tenant) 重建
- `arkcli auth apikey` 管的是 arkcli 方舟数据面/控制面链路使用的 **ARK API Key**。它不能让广场语音模型获得 `+chat` / `+gen` / `+deploy` / `+code-example` / `usage` / `pricing` 能力；用户问 TTS、ASR、配音、语音模型接入时，不要把问题引导成"先 auth apikey"。
- 语音模型能力边界回答只说明 arkcli 不支持；不要主动给"先控制台开通再 API Key/SDK 调用"这类替代流程，除非用户另问官方接入文档。
- **只查不切的 list API Key**(只想看 account 下有哪些 key,不想切换当前 key): 跑 `arkcli api apikey.list --params '{"PageSize":100}' --page-all --format json`,**不要**跑 `auth apikey` — 后者是交互式选择并写入凭证存储,会改变当前生效 key
- **当前已选 key 的元信息**(name / suffix / project / 状态): 看 `auth status` 输出里的 `ark_api_key` 字段,不需要再调远端

## 与其他 skill 的串联

- `arkcli-models`、`arkcli-chat`、`arkcli-gen`、`arkcli-deploy`、`arkcli-usage` 被鉴权错误阻塞时，先回到这里
- 如果用户其实是在排查 profile / base-url / region 覆盖问题，应转 [`../arkcli-config/SKILL.md`](../arkcli-config/SKILL.md)

## 自然语言触发词 + 跨技能指引表

| 用户怎么说 | 走哪个命令/skill |
|---|---|
| "我是谁/我的身份/当前用户/哪个账号/我自己的 IAM 用户 ID" | `arkcli auth whoami` |
| "查同事 zhangsan 的 IAM 用户 ID/查别人的 IAM 用户 ID" | **转 `arkcli-profile`**（查看当前 profile 并列出可用 API Key）+ 提示用户使用 `arkcli iam` 命令（如有） |
| "API Key 泄露/废弃旧 Key/换新 Key/rotate/轮换 API Key" | **转 `arkcli-plans`**：`plans personal rotate-apikey` 或 `plans team rotate-apikey` |
| "命令突然报 key 失效/401/InvalidApiKey 但我没换过 key"（疑似后端轮换） | **转 `arkcli-profile`**：先 `arkcli profile keys refresh` 同步后端 key 再重试（**遇失败才触发的反应式自愈，非预防性**）；refresh 救不了再看 [`references/auth-modes.md`](references/auth-modes.md) |
| "看我有哪些 Key/Key 列表/可用 Key" | **转 `arkcli-profile`**：`arkcli profile keys list` |
| "切换默认 Key/用另一个 Key" | **转 `arkcli-profile`**：`arkcli profile keys use <key>` |
| "AK/SK 登录/access key/secret key" | **告知通道暂关**：当前版本 AK/SK 登录通道暂时关闭，请使用 SSO 登录，运行 `arkcli auth login`

## 命令一览

| 命令 | 说明 |
|------|------|
| `arkcli auth status` | 查看当前认证状态（凭证健康度） |
| `arkcli auth whoami` | 查看当前认证身份（用户名 / IAM 用户 ID / 账号 ID 等），脚本与 Skill 用 |
| `arkcli auth login` | 交互式选择登录方式（火山 SSO 浏览器 / 火山 SSO 无浏览器） |
| `arkcli auth login volc-sso` | 浏览器 SSO 登录 |
| `arkcli auth login --no-browser` | 无浏览器 SSO（cross-device）。TTY：一条命令交互式粘贴；非 TTY（agent/沙箱）：Phase 1，打印 URL + `authorize_pending` JSON 后退出 |
| `arkcli auth login --no-browser --code <授权码>` | 无浏览器 SSO **Phase 2**：把 base64 授权码喂回完成登录（agent/沙箱两段式的第二步） |
<!-- AK/SK 通道 0.1.16 暂关; 恢复后取消下行注释 -->
<!-- \| `arkcli auth login --access-key <ak> --secret-key <sk>` \| 非交互 AK/SK 登录 \| -->
| `arkcli auth apikey` | 获取并配置 ARK API Key（按 active profile 的 tenant 写入对应 identity store） |
| `arkcli auth logout` | 删除本地凭证 |

## 参考

- [`references/arkcli-auth-login.md`](references/arkcli-auth-login.md)
- [`references/arkcli-auth-status.md`](references/arkcli-auth-status.md)
- [`references/arkcli-auth-whoami.md`](references/arkcli-auth-whoami.md)
