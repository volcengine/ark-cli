# config profile (deprecated 兼容映射)

> **0.1.16 起 profile 写操作迁移到 `arkcli profile` 子树**：`config list/show/switch/delete` 已 deprecated（仍可用，0.2.x 删除）。本文档保留旧命令与新命令的对照，方便 Agent 在排障老脚本时识别。

## 常规准入与 Profile 管理的边界

普通 CLI 只查当前身份/Profile 时，用 `arkcli auth status --format json`、`arkcli auth whoami --format json`；默认模型和路由看 `arkcli resources list --modality <text|image|video> --format json`。先遵守 shared 的宿主认证边界；禁止额外认证探测的宿主不能照搬这些认证命令。

`profile show/list/keys list` 会尝试同步远端 Key，可能回写本地库存、清空失效 Key 或改选默认 Key。它们的脱敏输出不代表无配置副作用。用户只要求 Chat/Gen 准入、或明确“不改配置/Key”时，不执行它们，也不以旧 `config show/list` 代替；缺失信息如实报告未知。

## 显式 Profile 管理用法

以下 show/list 只用于用户明确要求的 Profile 详情、清单或管理任务，并先说明上述同步影响；切换、删除、重置仍需明确授权。

```bash
# 列出所有 profile（含 type/region/project/owner_trn 切面）
arkcli profile list --format json

# 查看当前 / 指定 profile
arkcli profile show --format json
arkcli profile show --profile default --format json

# 切换默认 profile
arkcli profile use default --format json

# 删除单个 profile
arkcli profile delete default --format json

# 重置整个本地配置文件（保留，超出单 profile 范围）
arkcli config reset --format json

# 更新策略（只允许 update.mode）
arkcli config set update.mode automatic
arkcli config set update.mode disabled
```

公开模式只有 `automatic` 和 `disabled`。`disabled` 关闭静默安装，但保留隐式版本检查、更新提示以及显式 `arkcli update` / `arkcli update --check`。历史 YAML 中的 `notify` 仅作为兼容输入读取，不再是可设置选项。

对应 production gate 开启后，真正全新且安装版本等于 registry `latest` 的 stable 全局 npm 安装会先进入与 exact install 绑定的惰性 enrollment；第一次成功的、符合安全条件的人工业务命令会完成 grace 与 consent，重新校验 active 后即可调度。重装为当前 latest 时不改 mode，已有 automatic 会换发 exact receipt；非 latest 则持久化 disabled。

长期锁定版本仍建议先执行 `arkcli config set update.mode disabled`，再安装精确 npm 版本；非 latest 安装在成功确认 registry 后也会持久化 `disabled`。安装 `@latest` 本身不会将 `disabled` 改回 `automatic`；恢复时先安装 latest，再显式设置 automatic。配置位于 `$HOME/.arkcli/config.yaml`。

## 旧命令兼容映射

| 旧（deprecated） | 新 |
|------|------|
| `arkcli config list` | `arkcli profile list` |
| `arkcli config show` | `arkcli profile show` |
| `arkcli config switch <name>` | `arkcli profile use <name>` |
| `arkcli config delete <name>` | `arkcli profile delete <name>` |
| `arkcli config init ...` | `arkcli profile create --type=...`（行为更明确：必须指定 type） |
| `arkcli config reset` | 仍可用（清整个本地配置文件，无替代） |

> **为什么迁移**：profile 现在承载 `type / region / project / owner_trn / available_api_keys` 五个切面，命令路径 `arkcli profile <verb>` 比 `arkcli config <verb>` 更精确（config 还包括 reset、`update.mode` 和全局排障）。
