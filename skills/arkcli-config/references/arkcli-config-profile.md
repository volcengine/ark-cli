# config profile (deprecated 兼容映射)

> **0.1.16 起 profile 写操作迁移到 `arkcli profile` 子树**：`config list/show/switch/delete` 已 deprecated（仍可用，0.2.x 删除）。本文档保留旧命令与新命令的对照，方便 Agent 在排障老脚本时识别。

## 推荐用法（新）

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
```

## 旧命令兼容映射

| 旧（deprecated） | 新 |
|------|------|
| `arkcli config list` | `arkcli profile list` |
| `arkcli config show` | `arkcli profile show` |
| `arkcli config switch <name>` | `arkcli profile use <name>` |
| `arkcli config delete <name>` | `arkcli profile delete <name>` |
| `arkcli config init ...` | `arkcli profile create --type=...`（行为更明确：必须指定 type） |
| `arkcli config reset` | 仍可用（清整个本地配置文件，无替代） |

> **为什么迁移**：profile 现在承载 `type / region / project / owner_trn / available_api_keys` 五个切面，命令路径 `arkcli profile <verb>` 比 `arkcli config <verb>` 更精确（config 还包括 reset / 全局排障）。
