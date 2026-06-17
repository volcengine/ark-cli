# arkcli models custommodel update

改名 / 改描述。

## Usage

```bash
arkcli models custommodel update <id> [flags]
```

## Examples

```bash
# 改名
arkcli models custommodel update cm-xxxxx --name new-display-name

# 改描述
arkcli models custommodel update cm-xxxxx --description "updated for v2 release"

# 同时改两个
arkcli models custommodel update cm-xxxxx --name v2 --description "..."
```

## Flags

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<id>` | 是 | string | 自定义模型 ID `cm-xxxxx` |
| `--name` | 否* | string | 新名（账号内唯一）；须以英文字母开头，后续支持中文、字母、数字、下划线、短横线 |
| `--description` | 否* | string | 新描述，最多 300 字符 |

*`--name` 与 `--description` 至少要传一个。

**注意**：本命令**只改展示属性**——不能改 base model / 训练类型 / 量化模式 / 状态。改 `--name` 不影响已部署 endpoint 的引用关系（endpoint 通过 ID 引用，与名无关）。

## Output

成功时 `result` 只返回 `id` 和本次实际修改的字段：

```json
{
  "id": "cm-xxxxx",
  "name": "new-display-name"
}
```

同时修改名称和描述时才会同时返回 `name` / `description`；不会返回未修改字段或完整模型详情。
