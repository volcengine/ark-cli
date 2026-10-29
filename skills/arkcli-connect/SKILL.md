---
name: arkcli-connect
version: 1.3.0
description: "arkcli +connect：将 arkcli 内嵌的 AI skills 安装到本机检测到的所有 AI Agent 中，支持安装、列出已支持 agent、卸载。当用户需要将 arkcli 能力同步到 Claude Code 等本地 agent 时使用。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli +connect --help"
---

# arkcli +connect

**前置：** 先用 Read 读 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md) 获取共享安全规则。

把 arkcli 内嵌的 skills 同步到本机 AI Agent（Claude Code、Cursor、Gemini CLI 等）的 skills 目录。**纯本地文件操作，不需要认证。**

## 调用形态（子命令仍只有这三个）

| 调用 | 说明 |
|------|------|
| `arkcli +connect` | 默认行为：安装到所有检测到的 agent |
| `arkcli +connect --path <skills-dir>` | 安装到指定的本地 skills 目录（项目级 / 自定义路径），不扫描 agent、不改全局目录 |
| `arkcli +connect list` | 只读：列出支持的 agent 与检测状态 |
| `arkcli +connect uninstall` | 从所有 agent 删除 ownership manifest 记录且未被用户修改的 ArkCLI skills |
| `arkcli +connect uninstall --path <skills-dir>` | 只从指定目录删除 ownership manifest 记录且未被用户修改的 ArkCLI skills |
| `arkcli +connect uninstall --purge-prefix` | 显式破坏性清理：删除所有检测目录中的 `ark-` / `arkcli-` 前缀目录或软链 |
| `arkcli +connect uninstall --path <skills-dir> --purge-prefix` | 只在指定目录执行显式全前缀清理 |

> ⚠️ **没有** `+connect install` / `+connect setup` / `+connect sync` / `+connect remove` 等子命令；安装就是默认行为，不要凭直觉补 install。`--path` 是 flag，不是子命令。

## 路由判断

- 用户想把 arkcli skills 装进本地 agent → 跑 `arkcli +connect`，建议先 `arkcli +connect list` 预检
- 用户想把 arkcli skills 装进某个 repo/project 的本地 skills 目录 → 跑 `arkcli +connect --path <skills-dir>`，例如 `arkcli +connect --path .claude/skills`；`--path` 接收具体 skills 目录，不是项目根目录
- 用户只想知道支持哪些 agent → **只**跑 `arkcli +connect list`，**不要**顺手装
- 用户想清理 ArkCLI 自己安装的 skills → `arkcli +connect uninstall`，仍需先确认 agent 范围；默认只删 manifest 记录且摘要未变化的 managed skills
- 用户想清理 repo/project 本地 skills 目录里的 managed skills → `arkcli +connect uninstall --path <skills-dir>`
- 用户明确要求连历史未知条目一起全清 → 解释风险并在确认后使用 `--purge-prefix`；它会删除用户手工维护的同前缀目录或软链

## 反触发（应该路由到别处）

- 401 / 鉴权失败 / "auth login 报错" → 走 `arkcli-auth`，**与 +connect 无关**（+connect 不需要认证）
- profile / base-url / region 配置问题 → 走 `arkcli-config`
- 想生成代码示例、调用模型 → 走 `arkcli-code-example` / `arkcli-chat`

## 关键事实（写在 SKILL.md 内，避免 Agent 幻觉）

- 安装走 **authoritative catalog update**：`.arkcli-managed-skills.json` 记录 ArkCLI 上次写入的精确名称、产品和目录摘要；每次安装都会用当前 CLI catalog 覆盖所有同名 Skill，并删除 manifest 中已退出当前 catalog 的旧 Skill
- 当前 catalog 的精确官方名称具有最高优先级：无论现有同名目录是否被 manifest 记录、摘要是否变化、是否由用户手改，都会被当前版本覆盖；需要保留自定义内容时必须复制到不同名称
- 不在当前 catalog、也未被旧 manifest 管理的其他目录或软链不会进入安装事务，即使名称带 `ark-` / `arkcli-` 前缀也默认保留
- 目录摘要仍用于 provenance 和默认卸载保护，不再阻断安装升级；安装在 staging/backup 中完成并保持事务性，失败时回滚，不产生半套新旧 Skills
- 历史 BytePlus candidate 的 `arkcli-managed-agent` 若存在于旧 manifest，会作为退出 catalog 的旧 Skill 删除；canonical `arkcli-agent` 作为当前官方名称安装。这是统一 catalog 规则的结果，不再需要产品特判
- `--path <skills-dir>` 是隔离安装：相对路径按当前 `PWD` 解析，绝对路径原样使用；只在该目录收敛当前 catalog 和旧 manifest，并保留其他不同名条目。它**不**扫描 agent、不清 legacy 私有目录、不删除 `byted-ark-*` conflicting skills、不 patch Claude/opencode 等 harness 配置
- **默认还会移除两个抢路由的第三方生成 skill**：`byted-ark-seedance-skill` / `byted-ark-seedream-skill`（按**精确名**匹配，不误伤其他 `byted-*`）。原因：它们的 description 带"推荐优先"+逐字触发词，会在"生图/生视频"意图上压过 `arkcli +gen`，且自身要独立 `ark-` API Key，赢了路由反而失败。移除后生成类意图统一落 `arkcli +gen`（用登录态 profile 凭证，跟视频一样能成）。日志显式打印 `removed N conflicting (...)`，非静默。要保留它们：`arkcli +connect --keep-conflicting`
- `uninstall` 默认只删除 ownership manifest 中记录且摘要未变化的 managed skills；只有显式 `--purge-prefix` 才恢复旧式全前缀删除
- 多个 agent 共享同一个 skills 目录时，按路径**自动去重**，只更新一次 ownership manifest
- **多路径扫描特例**：Codex 同时扫 `~/.codex/skills`(CODEX_HOME) 和共享的 `~/.agents/skills`；Pi 同时扫 `~/.pi/agent/skills` 和共享的 `~/.agents/skills`。若两处都装 arkcli，会把每个 skill 列**两遍**或报同名冲突。因此 `+connect` 把 Codex / Pi 的 skill 都装进共享的 `~/.agents/skills`(与 cline/warp 去重成一份)，靠 `~/.codex` / `~/.pi/agent` 目录存在来检出对应 agent；legacy 私有目录只删除带旧 manifest 的 owned Skills，或与当前官方树逐字节一致的迁移副本，未知/修改过的同前缀条目保留
- `list` 只扫描本地文件系统，不写入、不联网、不需要认证
- 支持的 agent 列表硬编码在二进制里，新增 agent 需要重新编译
- `npm install` arkcli 时 postinstall **会自动跑 `+connect`**：在能打开 `/dev/tty` 的交互式终端上按 authoritative catalog 规则更新检测到的所有 agent，无需用户确认；成功时只输出唯一安装目标数、对应 Agent 名称和安全卸载命令 `arkcli +connect uninstall`，不会逐目录展开进度明细。**下载完成后无论是否检测到 agent，都会输出当前编译产品的登录 SOP**；产品文案按 locale 输出。它不会执行 `--purge-prefix`。CI / 非交互终端 / 拿不到 `/dev/tty` 一律静默跳过；想完全静音可设 `ARKCLI_SKIP_POSTINSTALL=1`

详细行为、错误码、输出示例见 [`references/arkcli-connect.md`](references/arkcli-connect.md)。
