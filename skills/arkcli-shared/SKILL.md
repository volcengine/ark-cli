---
name: arkcli-shared
version: 2.0.1
description: "arkcli 共享执行协议：首次配置入口、业务命令执行前的认证闸门、命令路由与选择顺序、输出/安全/二次确认规则。深度细节（身份解析、AK-SK 边界、API Key 恢复、实名闸门、profile 默认与漂移、全局 flags、故障分流）按需在 references/ 加载。当用户第一次使用 arkcli、遇到未登录/鉴权失败、需要判断该走产品命令还是 raw api、或任何 arkcli-* skill 需要公共上下文时触发。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli --help"
---

# arkcli 共享规则

本 skill 是 `arkcli` 的统一执行协议**入口**。所有 `arkcli-*` skill 在执行前都应先读取本文件。

> **分层约定**：本文正文只放"几乎每个任务都命中"的规则。稀有路径 / 查表类细节下沉到 reference，命中对应场景时再读：
>
> | 场景 | 读哪个 |
> |---|---|
> | "我的 / my xxx" 自指资源过滤 | [`../arkcli-auth/references/identity-resolution.md`](../arkcli-auth/references/identity-resolution.md) |
> | AK/SK 态能调什么 / 数据面 API Key 报错恢复 | [`../arkcli-auth/references/auth-modes.md`](../arkcli-auth/references/auth-modes.md) |
> | 开通 / 部署 / 精调前的实名检查 | [`../arkcli-auth/references/realname-gate.md`](../arkcli-auth/references/realname-gate.md) |
> | `+chat`/`+gen`/`+deploy` 的默认资源与漂移 nudge | [`references/profile-defaults.md`](references/profile-defaults.md) |
> | 全局 flags 速查 | [`references/global-flags.md`](references/global-flags.md) |
> | 报错不知归类 / 故障分流 | [`references/troubleshooting.md`](references/troubleshooting.md) |

## 配置与首次使用

- 首次使用或怀疑配置归因（`--profile` / `ARK_PROFILE` / 全局 flags / `.env` 谁覆盖谁）不对时，先看 [`../arkcli-config/SKILL.md`](../arkcli-config/SKILL.md)
- profile 类写操作（create / use / set-default / keys / models / delete / rename）走 [`../arkcli-profile/SKILL.md`](../arkcli-profile/SKILL.md)
- 看可用资源（endpoint / plan 模型 ID）走 [`../arkcli-resources/SKILL.md`](../arkcli-resources/SKILL.md)
- 首次远端调用前，先看 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
- `arkcli` 是产品 CLI，不是 OpenAPI Action 浏览器；不要从 Action 名反推命令设计
- 安装: `npm i @volcengine/ark-cli -g`（公开版）

## 统一 CLI 与 Profile

`arkcli` 是唯一的二进制：

```bash
arkcli profile create --type platform --set-default          # 新建火山 profile（旧 config init/switch 已 deprecated）
arkcli profile use <name>                                    # 切换默认 profile
```


切换 profile 会联动切换登录身份、API Key、控制面路由等全部上下文。详细命令树看 [`../arkcli-profile/SKILL.md`](../arkcli-profile/SKILL.md)。判断当前租户：`arkcli profile show` 看 `tenant` 字段。

## 命令路由与执行顺序

优先按**用户目标**判断，而不是按命令名思考：

| 用户目标 | 路径 | 关键点 |
|---|---|---|
| 试用模型 / 快速验证效果 | `auth` → `models`（可选）→ [`+chat`](../arkcli-chat/SKILL.md) / [`+gen`](../arkcli-gen/SKILL.md) | **不需要** Endpoint |
| 专项多模态理解（转写/抽取/字幕/框目标…） | `auth` → [`+understand`](../arkcli-understand/SKILL.md) | 有明确产出形态时走 understand，不是 chat |
| 正式接入 / 稳定调用 | `auth` → `models` → [`infer endpoint list`](../arkcli-infer-endpoint/SKILL.md) → 没有就 [`+deploy`](../arkcli-deploy/SKILL.md) | 核心资源是 Endpoint，不是 +chat/+gen |
| 排查存量调用 / 看消耗 | `auth` → [`usage`](../arkcli-usage/SKILL.md) | — |
| 本地 AI Agent 集成 | [`+connect`](../arkcli-connect/SKILL.md) | — |

**易混动词路由**（避免选错 skill）：

- **列 / 绑 / 分 / 轮换席位 APIKey** → [`arkcli-plans`](../arkcli-plans/SKILL.md)；**用 / 消耗 / 还剩多少额度** → [`arkcli-usage`](../arkcli-usage/SKILL.md)
- profile **写操作**（create/use/set-default/keys）→ [`arkcli-profile`](../arkcli-profile/SKILL.md)；配置**排障 / 老 yaml** → [`arkcli-config`](../arkcli-config/SKILL.md)
- 开放式带图**对话** → `+chat`；**生成**图/视频 → `+gen`；有产出形态的**理解** → `+understand`

**命令选择顺序**（始终按此）：

1. 先用产品命令 `arkcli <domain> <verb>` 或 `arkcli +<workflow>`
2. 有对应 reference 文档先读 reference 再执行
3. 产品命令确实不覆盖，最后才走 `arkcli api`（**不要**把 Raw API Explorer 当默认入口）
4. 读操作优先直接执行；写 / 删除 / 切换默认配置前必须确认用户意图

## 认证闸门

除 `arkcli auth ...`、`arkcli profile list/show`、`arkcli +connect list` 外，默认认为业务命令**需要先过认证检查**。不要跳过认证检查就连续重试一串业务命令。

1. 先运行 `arkcli auth status`；已登录就继续目标命令
2. 未登录 / 凭证失效：按当前 profile 的 `tenant` 选登录命令
   - **火山**（`tenant=volc` 或未设置）：直接 Bash 执行 `arkcli auth login volc-sso`
   - 执行前一句话告知用户"检测到未登录，正在启动 SSO 登录，请在浏览器完成授权"；Bash 调用设 `timeout=600000`（10 分钟）；成功后立即回到原始任务，不要停在 auth 结果
   - SSO 同时覆盖控制面 BFF 和数据面，所以默认走 SSO；**AK/SK 登录通道 0.1.16 暂关**。CI / agent / 沙箱（非 TTY）走 `arkcli auth login --no-browser` **两段式**：Phase 1 跑它拿 `authorize_pending` JSON 里的 `authorize_url` 转发给用户，待其浏览器授权后回粘 base64 授权码，再跑 `arkcli auth login --no-browser --code <授权码>` 完成（细节见 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)）。**不要**在非 TTY 直接指望它阻塞等粘贴（旧版会 `EOF` 崩）
   - 启动 SSO 失败（无浏览器 / `open` 失败 / 端口占用 / 超时）→ 不原地重试，把 stderr 原样贴回用户、请其手动在终端登录后回来
   - 登录流程细节、whoami、apikey 选择见 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)

**追加闸门 / 错误恢复**（命中才读，避免常驻）：

- 命中**开通 / 部署 / 精调 / 激活模型**（`+deploy` / `infer endpoint create` / `train finetune create` / `models activate`）→ **第一步**先做实名检查，见 [`../arkcli-auth/references/realname-gate.md`](../arkcli-auth/references/realname-gate.md)
- 当前是 **AK/SK 态**、要判断能调什么，或**数据面 API Key** 报 401/403/AccessDenied 要恢复 → [`../arkcli-auth/references/auth-modes.md`](../arkcli-auth/references/auth-modes.md)
- 用户说"**我的 / my** xxx"要按身份过滤资源 → [`../arkcli-auth/references/identity-resolution.md`](../arkcli-auth/references/identity-resolution.md)

## 输出规则

- 当前全局 `--format` 只支持 `json`
- `stdout` 只放结构化结果；解释 / 调试 / 错误都走 `stderr`

## 安全规则

- 禁止输出完整 AK/SK、token、secret
- 写入、删除、切换配置前需要确认用户意图
- 只有已注册 Action 才能通过 `arkcli api` 调用
- 涉及创建 Endpoint、修改 profile、清理凭证等操作时，优先建议先用只读命令或 `--dry-run` 验证

## 二次确认错误处理（human-in-the-loop）

高危操作（删除资源、变更凭证、产生费用等）需要二次确认。CLI 自动检测环境：交互式终端显示 Y/N 提示；非交互式（Agent 调用）返回 `ExitValidation` 错误、`type="requires_confirmation"`。

当 arkcli 返回 `ExitValidation` 且 `type="requires_confirmation"` 时：

1. **不要直接报错给用户** —— 这是正常的二次确认流程
2. **调用 AskUser** —— 提示内容可从 CLI 错误的 `hint` 字段提取，或通用提示"即将执行高危操作，确认继续吗？"
3. 用户确认 → 给原命令加 `--yes` flag 重试
4. 用户取消 → 返回"操作已取消"

当前会触发二次确认的命令：`plans personal rotate-apikey`、`plans team rotate-apikey`、`models activate <model-name>`、`profile delete <profile-name>`、`profile project [<project-name>]`、`config delete <profile-name>`。后续新增高危命令遵循同一约定。

## Agent 禁止行为

- 不要把 `arkcli api` 当默认入口
- 不要在未检查认证状态前连续重试业务命令
- 不要把中间步骤当最终结果；登录、查模型、切 profile 完成后应回到用户原始任务
- 不要在业务 skill 里重复共享规则；共享规则统一以本 skill（及其 references）为准
- 不要把"试用模型"误写成"需要先创建 Endpoint"（试用走 `+chat`/`+gen`）；也不要把"正式接入"误导到一次性试用命令（正式接入走 `+deploy`）
- 不要为单个业务命令加 `--project` / `--project-name`；Project Name 只有全局 `--project-name` 一个入口，持久化优先走 `arkcli profile create --project <name>`，不要写进 `.env` / 老 `config.json`

## 参考

- [arkcli-auth](../arkcli-auth/SKILL.md) — 认证状态检查、登录与退出登录；其 `references/` 放身份解析 / AK-SK 边界 / API Key 恢复 / 实名闸门
- [arkcli-config](../arkcli-config/SKILL.md) — profile、base-url、region 配置排障
- [arkcli-api-explorer](../arkcli-api-explorer/SKILL.md) — 产品命令未覆盖时的 raw API 兜底入口
- [references/profile-defaults.md](references/profile-defaults.md) — profile 默认资源、漂移检测、跨模态
- [references/global-flags.md](references/global-flags.md) — 常用全局 flags 速查
- [references/troubleshooting.md](references/troubleshooting.md) — 故障分流与能力边界
