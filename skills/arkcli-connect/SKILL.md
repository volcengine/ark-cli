---
name: arkcli-connect
version: 1.1.0
description: "arkcli +connect：将 arkcli 内嵌的 AI skills 安装到本机检测到的所有 AI Agent 中，支持安装、列出已支持 agent、卸载。当用户需要将 arkcli 能力同步到 Claude Code 等本地 agent 时使用。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli +connect --help"
---

# arkcli +connect

**前置：** 先用 Read 读 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md) 获取共享安全规则。

把 arkcli 内嵌的 skills 同步到本机 AI Agent（Claude Code、Cursor、Gemini CLI 等）的 skills 目录。**纯本地文件操作，不需要认证。**

## 子命令穷举（只有这三个）

| 调用 | 说明 |
|------|------|
| `arkcli +connect` | 默认行为：安装到所有检测到的 agent |
| `arkcli +connect list` | 只读：列出支持的 agent 与检测状态 |
| `arkcli +connect uninstall` | 破坏性：从所有 agent 卸载 |

> ⚠️ **没有** `+connect install` / `+connect setup` / `+connect sync` / `+connect remove` 等子命令；安装就是默认行为，不要凭直觉补 install。

## 路由判断

- 用户想把 arkcli skills 装进本地 agent → 跑 `arkcli +connect`，建议先 `arkcli +connect list` 预检
- 用户只想知道支持哪些 agent → **只**跑 `arkcli +connect list`，**不要**顺手装
- 用户想清理已装 skills → `arkcli +connect uninstall`，**先要求用户明确确认**（破坏性，会按 `ark-` / `arkcli-` 前缀 `rm -rf` agent 目录下所有同前缀 skill）

## 反触发（应该路由到别处）

- 401 / 鉴权失败 / "auth login 报错" → 走 `arkcli-auth`，**与 +connect 无关**（+connect 不需要认证）
- profile / base-url / region 配置问题 → 走 `arkcli-config`
- 想生成代码示例、调用模型 → 走 `arkcli-code-example` / `arkcli-chat`

## 关键事实（写在 SKILL.md 内，避免 Agent 幻觉）

- 安装走 **purge-then-install**：每个 agent 上先删掉 skills 目录下**所有** `ark-` / `arkcli-` 前缀的子目录（清掉历史遗留 skill，如已废弃的 `ark-experience`、`arkcli-experience-video` 等），再把当前内嵌 skill 写入
- **默认还会移除两个抢路由的第三方生成 skill**：`byted-ark-seedance-skill` / `byted-ark-seedream-skill`（按**精确名**匹配，不误伤其他 `byted-*`）。原因：它们的 description 带"推荐优先"+逐字触发词，会在"生图/生视频"意图上压过 `arkcli +gen`，且自身要独立 `ark-` API Key，赢了路由反而失败。移除后生成类意图统一落 `arkcli +gen`（用登录态 profile 凭证，跟视频一样能成）。日志显式打印 `removed N conflicting (...)`，非静默。要保留它们：`arkcli +connect --keep-conflicting`
- `uninstall` 同样按 `ark-` / `arkcli-` 前缀清理，不依赖"当前 binary 里有的 skill 名"——历史遗留同样能清掉
- 多个 agent 共享同一个 skills 目录时，按路径**自动去重**，只 purge / 装一次
- **Codex 特例**：Codex 同时扫 `~/.codex/skills`(CODEX_HOME) 和共享的 `~/.agents/skills`。若两处都装 arkcli，Codex 会把每个 skill 列**两遍**。因此 `+connect` 把 Codex 的 skill 装进共享的 `~/.agents/skills`(与 cline/warp 去重成一份)、靠 `~/.codex` 目录存在来检出 Codex；并在每次 `+connect` 时**清掉 legacy `~/.codex/skills` 里的 arkcli**(只清 arkcli/seed，gocli 等其它 skill 不动)。这样 Codex 里每个 arkcli skill 只出现一次
- `list` 只扫描本地文件系统，不写入、不联网、不需要认证
- 支持的 agent 列表硬编码在二进制里，新增 agent 需要重新编译
- `npm install` arkcli 时 postinstall **会自动跑 `+connect`**：在能打开 `/dev/tty` 的交互式终端上直接把内嵌 skills 装到检测到的所有 agent，无需用户确认；CI / 非交互终端 / 拿不到 `/dev/tty` 一律静默跳过；想完全静音可设 `ARKCLI_SKIP_POSTINSTALL=1`

详细行为、错误码、输出示例见 [`references/arkcli-connect.md`](references/arkcli-connect.md)。
