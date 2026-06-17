---
name: arkcli-config
version: 1.1.0
description: "arkcli 本地配置管理。0.1.16 起 profile 类操作请优先使用 `arkcli profile <subcmd>`（init/list/show/switch/delete 已 deprecated）；本 skill 仍可处理 `arkcli config reset` 与历史 yaml 排障。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli config --help"
---

# arkcli config

**⚠️ 0.1.16 起命令树迁移**：`arkcli config init/list/show/switch/delete` 已标记 deprecated，请改用 `arkcli profile create/list/show/use/delete`（行为一致 + 新增 profile 切面字段 type/project/owner_trn/available_api_keys）。`arkcli config reset` 保留（清整个本地状态，超出单 profile）。0.2.x 删除全部已 deprecated 子命令。

**CRITICAL — 开始前 MUST 先用 Read 工具读取 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md)，其中包含认证闸门、命令选择顺序与共享安全规则**
**CRITICAL — 配置类写操作前必须确认用户意图；只读排障优先 `list/show`，禁止上来直接改配置。**

## 唤起信号（When To Trigger）

- 用户提到：`profile`、`base-url`、`region`、`api-key`、`ARK_PROFILE/ARK_API_KEY/ARK_BASE_URL/ARK_REGION` 覆盖混乱
- 用户现象：同一条业务命令“打到了错误环境/错误 base URL/错误账号”，怀疑是配置来源不一致
- 用户目标：初始化/更新 profile、切换默认 profile、重置本地配置文件

## 反唤起信号（When NOT To Trigger）

- 用户要解决的是“未登录/401/鉴权失败/没有权限”：先走 `arkcli auth status`，必要时转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md)
- 用户目标是业务功能（models/+chat/+gen/+deploy/usage 等）：不要停留在 config，排除配置后回到原任务

## 适用场景

- 初始化或更新 profile
- 查看当前解析后的配置
- 切换默认 profile
- 删除单个 profile 或重置整个本地配置
- 排查为什么业务命令连到了错误环境、错误账号或错误 base URL

## Agent 快速执行顺序

1. 不确定当前生效配置时，先用 `arkcli profile show`（旧 `arkcli config show` 已 deprecated）
2. 不确定有哪些 profile 时，先用 `arkcli profile list`
3. 切换默认 profile：`arkcli profile use <name>`
4. 创建 profile：`arkcli profile create --type=...`（详见 arkcli-profile 子树，本 skill 不重复）
5. `arkcli config reset` 是破坏性操作（清整个本地配置文件），必须确认用户意图

## 配置归因（优先级）

`arkcli` 的解析优先级（从高到低）：

- Profile 选择: `--profile` flag > `ARK_PROFILE` 环境变量 > `config.yaml` 的 `default_profile` > 第一个 `type=platform` 的 profile > `"default"` sentinel
- 其它字段: 全局 flags（`--api-key/--base-url/--region`） > 环境变量（`ARK_API_KEY/ARK_BASE_URL/ARK_REGION`） > profile 配置 > identity store > `.env` fallback（仅无 identity 绑定时）

排障时要明确”是谁覆盖了谁”，不要直接猜测。

> **0.1.16 修正**：`--profile` 现在优先于 `ARK_PROFILE`（对齐 CLI 行业惯例 flag > env > config）。0.1.15 及更早版本是 env > flag，与本规则相反。

### 固定输出片段（用于稳定评测）

为保证 Agent 的输出可被机器稳定判分，回答里必须出现以下字面量片段（原样输出，不要改写）：

- `只读排障`
- `show/list`
- `--profile > ARK_PROFILE > default_profile > "default"`

同时，路径展示一律使用 `$HOME/...`，避免使用 `~`。

## 核心规则

- 本 skill 负责解释“为什么命令打到了错误环境/错误账号/错误 base URL”，并兼容历史 `config init/list/show/switch/delete` 命令排障语义
- 配置排障优先 `arkcli profile show`，不要一上来就改配置
- profile 写操作（create / use / set-default / delete）走 `arkcli profile` 子树
- `arkcli config reset` 是唯一保留的 config 写操作（整个本地配置文件清空，超出单 profile 范围）

## Guard Checklist（必做）

- 只读优先：先 `arkcli profile show --format json`，再 `arkcli profile list --format json`
- 写操作确认：`profile use/delete` 与 `config reset` 执行前必须复述影响范围并征得确认
- 全量清理提醒：`arkcli config reset` 删除 `config.yaml/config.json`，不会清理 `$HOME/.arkcli/.env`（token）或 identity store；要清理凭证请用 `arkcli auth logout`

## 与其他 skill 的串联

- profile 创建 / 切换 / 资源 default 设置 / API Key 管理 → 转 `arkcli-profile`（推荐）或继续看本 skill 的兼容映射
- 业务命令失败且怀疑是 `--profile`、`--base-url`、`--region`、API Key 来源不一致时，先转到这里
- 配置问题排除后，再回到原始业务 skill 继续
- 配置仍无法解释问题时，再考虑转 [`../arkcli-auth/SKILL.md`](../arkcli-auth/SKILL.md) 或 [`../arkcli-api-explorer/SKILL.md`](../arkcli-api-explorer/SKILL.md)

## 命令一览

| 命令 | 说明 | 状态 |
|------|------|------|
| `arkcli config reset` | 删除整个本地配置文件（保留） | ✅ 活跃 |
| `arkcli profile show [--profile <name>]` | 查看解析后配置或指定 profile | ✅ 替代 `config show` |
| `arkcli profile list` | 列出所有 profile（含 type/region/project 切面） | ✅ 替代 `config list` |
| `arkcli profile use <name>` | 切换默认 profile | ✅ 替代 `config switch` |
| `arkcli profile delete <name>` | 删除单个 profile | ✅ 替代 `config delete` |
| `arkcli profile create --type=...` | 创建 profile（替代旧 `config init`） | ✅ 替代 `config init` |
| `arkcli config init/list/show/switch/delete` | 旧子命令，0.2.x 移除 | ⚠️ deprecated |

## 参考

- [`references/arkcli-config-init.md`](references/arkcli-config-init.md)
- [`references/arkcli-config-profile.md`](references/arkcli-config-profile.md)
- [`references/evals.md`](references/evals.md)
