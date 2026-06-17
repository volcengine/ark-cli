# arkcli models custommodel delete

删除（破坏性，不可逆）。

## Usage

```bash
arkcli models custommodel delete <id> [--yes] [--dry-run]
```

## Examples

```bash
# 预览删除请求，不实际删除
arkcli models custommodel delete cm-xxxxx --dry-run

# 交互式删除：会弹 [Y/N] 二次确认
arkcli models custommodel delete cm-xxxxx

# 非交互删除：跳过本地二次确认
arkcli models custommodel delete cm-xxxxx --yes
```

## Flags

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<id>` | 是 | string | 自定义模型 ID `cm-xxxxx` |
| `--yes` | 否 | bool | 跳过 [Y/N] 交互确认 |
| `--dry-run` | 否 | bool | 全局 flag：预览删除请求，不实际删除，也不弹二次确认 |

**注意**：
- 删除**不可逆**——同 ID 不会复用；已部署 endpoint 的引用会立即失效
- 删除前推荐先 `custommodel get <id> --transform id,name,active_endpoints` 列出引用情况，向用户复述
- `--yes` 的边界：只有用户已经明确确认删除目标 `cm-xxxxx`，并理解 `active_endpoints` 非空会破坏推理链路时，agent 才能把 `--yes` 加到命令里
- 用户没明说"删除"、只是询问能否删除或让你检查风险时，永远不要执行删除
