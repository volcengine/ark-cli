---
name: arkcli-code-example
version: 1.2.0
description: "arkcli +code-example：为指定基础模型生成多语言（Python / Go / Java / Node / curl）调用示例代码并写入本地文件。数据源是火山方舟 OpenTOP OpenGetSampleCode。当用户需要拿某个基础模型的 SDK / curl 调用示例、保存为本地接入模板时使用。"
metadata:
  requires:
    bins: ["arkcli"]
  cliHelp: "arkcli +code-example --help"
---

# arkcli +code-example

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

为指定基础模型生成多语言 SDK 调用示例代码，同时写入本地文件。数据源是火山方舟 OpenTOP `OpenGetSampleCode`。

## 什么时候用

- 用户说"生成示例代码 / SDK 示例 / curl 命令 / 怎么调用这个模型"，且已经知道模型名
- 想把某个基础模型的多语言调用模板保存到本地

**反路由**（不要用本命令）：

- 只是想试模型效果 → [`../arkcli-chat/SKILL.md`](../arkcli-chat/SKILL.md)（文本/多模态对话）或 [`../arkcli-gen/SKILL.md`](../arkcli-gen/SKILL.md)（出图/视频）
- 还不知道模型名 → 先 [`../arkcli-models/SKILL.md`](../arkcli-models/SKILL.md) `models search/get`
- 需要先创建接入点 → [`../arkcli-deploy/SKILL.md`](../arkcli-deploy/SKILL.md)

## 快速开始

```bash
# 拿某基础模型的全部语言示例（写入 ./ark-examples/<model>/）
arkcli +code-example --model doubao-seedream-5-0

# 指定版本 + 只看 Python + 结构化输出
arkcli +code-example --model doubao-seedream-5-0 --version 260128 --lang python --format json
```

- `--model` 必填，可传基础模型名或合并 ID（`doubao-seedream-5-0-260128`）
- `--version` 可选；`--model` 已带版本时可省略
- `--lang` 可选：`python` / `go` / `java` / `node` / `curl`（`shell` 是 `curl` 别名），默认全部
- `--output-dir` 可选，默认 `./ark-examples/<model>`
- `--format json` 输出结构化条目（含 `language` / `io_type` / `install` / `code`）

> **0.1.17 变更**：旧版 `--endpoint-id` 已移除——新数据源按基础模型取码，不再支持按已有接入点 ID 取码。如需指向某 endpoint，生成后把代码里的 `model="..."` 手动改成你的 `ep-xxx`。

## 命令一览

| 命令 | 说明 |
|------|------|
| `arkcli +code-example --model <name> [--version <ver>] [--lang <lang>] [--output-dir <dir>] [--format json]` | 生成多语言示例代码并写入本地 |

## 输出形态

- **本地文件**：按"IO 任务类型"分子目录（对齐后端 `IOType` 维度），同类型下按语言分文件；带安装步骤的语言把安装命令写进文件头注释。
- **结构化 JSON**：`--format json` 时给出按 `(语言 × IO 类型)` 展开的条目。
- **人类可读**：默认按后端顺序打印各代码块。

详细参数、目录结构示例、常见链路与 FAQ 见 [`references/arkcli-code-example.md`](references/arkcli-code-example.md)。

## 参考

- [references/arkcli-code-example.md](references/arkcli-code-example.md) — 完整命令手册
- [arkcli-models](../arkcli-models/SKILL.md) / [arkcli-deploy](../arkcli-deploy/SKILL.md) / [arkcli-infer-endpoint](../arkcli-infer-endpoint/SKILL.md) / [arkcli-shared](../arkcli-shared/SKILL.md)
