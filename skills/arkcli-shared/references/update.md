# 版本检查与显式升级

本 reference 只负责回答“当前版本是否最新”和执行用户明确要求的升级。它不访问业务账号，不需要先跑 `auth status`。

`update.mode` 的读取、写入、默认值和 automatic production gate 属于 [`../../arkcli-config/SKILL.md`](../../arkcli-config/SKILL.md)。不要因发现新版本或仅看到 postinstall enrollment 就推断 active automatic consent；新安装 pending、registry latest reconciliation、exact receipt 换发与版本锁定流程以 config Skill 为准。

本 reference 的所有 Agent 命令都必须携带下方 `ARKCLI_CALLER_TYPE=ai_agent` 等 caller metadata。它不仅用于归因，也保证 AI Skill 不消耗人工首次宽限、不激活 exact consent、不调度 automatic。不要删除这些变量后用 Agent 模拟人工命令。

## 只检查

默认先读零网络的本地缓存：

```bash
ARKCLI_NO_UPDATE_NOTIFIER=1 \
ARKCLI_CALLER_TYPE=ai_agent \
ARKCLI_CALLER_NAME=<agent-id> \
ARKCLI_SKILL_NAME=arkcli-shared \
arkcli --format json update --check
```

按 `status` 判断，不要只看兼容字段 `update_available`：

| `status` | 含义 | Agent 行为 |
|---|---|---|
| `unknown` | 没有当前发行渠道的有效缓存 | 明确说“未知”；需要实时结果时再用 `--refresh` |
| `up_to_date` | 缓存或 registry 已确认当前版本不旧于 latest | 报告已是最新，并说明 `source` |
| `update_available` | `latest` 严格新于 `current` | 报告版本差异；不要自动升级 |

`source` 是 `none`、`cache` 或 `registry`，`checked_at` 是缓存或查询时间。用户要求实时结果时再刷新：

```bash
ARKCLI_NO_UPDATE_NOTIFIER=1 \
ARKCLI_CALLER_TYPE=ai_agent \
ARKCLI_CALLER_NAME=<agent-id> \
ARKCLI_SKILL_NAME=arkcli-shared \
arkcli --format json update --check --refresh
```

`--refresh` 会访问当前二进制内置的 npm registry。查询失败时原样报告错误；不要把旧缓存包装成实时结果，也不要把 `status=unknown` 解读成“无更新”或“已是最新”。

## 显式升级

只有用户在本轮明确要求升级时才进入写流程。先检查并展示 `current -> latest` 和发行渠道，再取得本轮确认；确认后才可在非交互命令中添加 `--yes`：

```bash
ARKCLI_NO_UPDATE_NOTIFIER=1 \
ARKCLI_CALLER_TYPE=ai_agent \
ARKCLI_CALLER_NAME=<agent-id> \
ARKCLI_SKILL_NAME=arkcli-shared \
arkcli update --yes
```

显式 `arkcli update` 和 `arkcli update --check` 在所有更新模式下都可用；`disabled` 只关闭静默安装。`disabled` 仍允许隐式检查和版本提示。

用户要求长期锁定版本时，不要只执行 npm 降级。先引导用户持久设置 `arkcli config set update.mode disabled`，再安装精确版本。重装不得沿用旧 exact receipt：安装版本等于 registry 当前 `latest` 时不改 mode，已有 automatic 会换发 receipt；非 latest 持久化 disabled。安装 `@latest` 不会把 disabled 改回 automatic，恢复时先安装 latest，再显式设置 automatic。

`update` 会启动外部 npm，属于 `opaque_external_execution`，不支持 `--dry-run`。不要伪造 Client Preview，也不要用 raw API 代替；用版本差异说明和明确确认完成安全闭环。

只有 npm 成功、同一 npm global root 下的 package 版本等于 `latest`，且 npm 管理的目标 executable 的 `--version` 也等于 `latest`，才能报告升级成功。命令失败时不得声称升级完成。

## 失败与平台分流

- `npm` 缺失：转述 CLI 给出的手动命令，不自行安装 Node/npm。
- npm prefix、Node 或 NVM 不一致：停止，提示用户切回安装当前 arkcli 的 Node 环境；不要用 PATH 中另一个 npm 强行继续。
- 当前 binary 不在 npm global package 内：说明该副本不受影响。即使 npm package 验证成功，也不要声称当前副本已经变更。
- Windows 的显式更新只表示“已安排”；只有新一次 `arkcli --version` 显示目标版本，或 `update-apply.log` 记录 `installed and verified`，才能报告完成。
- Windows 更新失败时保留旧安装、日志和 `.arkcli-update-*` 恢复目录，不把失败解释成“部分成功”。
- macOS/Linux 的显式更新仍是手工 npm 升级语义；不要把 automatic 原子 cutover 描述成显式命令已采用的路径。

不要用 cron、系统服务或 Agent 定时任务模拟自动更新，也不要因普通更新提示中断当前业务命令或改变其退出码。
