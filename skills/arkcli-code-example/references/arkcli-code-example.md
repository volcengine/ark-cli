# +code-example

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

为指定基础模型生成多语言 SDK 调用示例代码，同时写入本地文件。

数据源是火山方舟 OpenTOP `OpenGetSampleCode`：按 `(model, version)` + `scenario` 拉取一批"代码块"，arkcli 自动按 `(语言 × IO 任务类型)` 去重、配好安装命令后渲染落盘。

## 什么时候用

- 已经明确模型名（或合并 ID），想直接拿多语言调用代码
- 想把示例保存到本地目录，作为后续接入模板

不要在这些场景使用本命令：

- 只是想快速试模型效果：转 `arkcli +chat` 或 `arkcli +gen`
- 还不知道模型名：先 `arkcli models search/get`
- 需要先创建 Endpoint：先 `arkcli +deploy` 或 `arkcli infer endpoint create`

## 用法

```bash
# 基础用法（传基础模型名）
arkcli +code-example --model doubao-seedream-5-0

# 指定模型版本（两种写法等价）
arkcli +code-example --model doubao-seedream-5-0 --version 260128
arkcli +code-example --model doubao-seedream-5-0-260128

# 只看 Python 示例
arkcli +code-example --model doubao-seedream-5-0 --lang python

# 只看 curl 示例（shell 是别名，等价）
arkcli +code-example --model doubao-seedream-5-0 --lang curl

# 指定输出目录
arkcli +code-example --model doubao-seedream-5-0 --output-dir ./my-examples

# 需要结构化结果时，显式要求 JSON
arkcli +code-example --model doubao-seedream-5-0 --format json
```

## 参数

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `--model` | string | ✓ | 基础模型名（`doubao-seedream-5-0`）或合并 ID（`doubao-seedream-5-0-260128`） |
| `--version` | string | | 模型版本，如 `260128`；若 `--model` 已带版本可省略 |
| `--lang` | string | | 语言过滤：`python` / `go` / `java` / `node` / `curl`（`shell` 是 `curl` 的别名），默认全部输出 |
| `--output-dir` | string | | 本地输出目录，默认 `./ark-examples/<model>` |
| `--format json` | flag | | 输出结构化结果，适合 Agent 或脚本继续消费 |

> **0.1.17 变更**：旧版的 `--endpoint-id` 已移除 —— 新数据源 `OpenGetSampleCode` 按基础模型取码，不再支持按已有接入点 ID 取码。如需把代码指向某个 endpoint，自行把生成代码里的 `model="..."` 改成你的 `ep-xxx` 即可。

## 推荐执行顺序

1. 不确定认证状态时，先执行 `arkcli auth status`
2. 不确定模型名或版本时，先执行 `arkcli models search/get`
3. 需要结构化解析时，追加 `--format json`
4. 需要控制本地写入位置时，追加 `--output-dir`

## 输出

1. **本地文件**：写入 `--output-dir` 目录；未显式指定时，默认写到 `./ark-examples/<model>/`。文件按"IO 任务类型"分子目录（对齐后端返回的 `IOType` 维度），同类型下再按语言分文件。例如图像模型：

```text
ark-examples/<model>/
├── text_to_image_generate_single_image/
│   ├── python.py
│   ├── go/main.go
│   ├── java/Main.java
│   ├── node.ts
│   └── curl.sh
├── image_to_image_single_image_to_single_image/
│   └── ...
└── ...
```

文本/对话类模型的子目录名会是对应的对话类 IO 类型。带安装步骤的语言（如 Python 的 `volcengine-python-sdk[ark]`、Go 的 `go get ...`），安装命令会作为文件头注释一并写入。

2. **结构化结果**：追加 `--format json` 后，得到按 `(语言 × IO 类型)` 展开的结构化条目（含 `language` / `io_type` / `call_method` / `install` / `code` 等字段），适合后续流程解析。

3. **人类可读代码块**：默认按后端给定顺序打印各示例代码块，便于直接复制。

## 常见链路

### 1. 先查模型，再生成示例

```bash
arkcli models search doubao-seedream
arkcli +code-example --model doubao-seedream-5-0
```

### 2. 只要某一种语言，并让下游脚本消费 JSON

```bash
arkcli +code-example \
  --model doubao-seedream-5-0 \
  --lang python \
  --format json
```

## 注意事项

- 代码中的 `$ARK_API_KEY` 会自动替换为当前 profile 的 `api_key`（见 `arkcli profile show`）
- 需要先通过 `arkcli auth` 登录才能调用接口
- 模型名称可通过 `arkcli models search <keyword>` 查找
- 本命令会写入本地文件；如果不希望写到当前目录，请显式指定 `--output-dir`
- 示例只覆盖 API Key 鉴权方式（`Authorization: Bearer $ARK_API_KEY`）

## 常见问题

| 现象 | 常见原因 | 处理方式 |
|------|----------|----------|
| 鉴权失败 / 未登录 | 本地没有可用凭证 | 先执行 `arkcli auth status`，必要时转 `arkcli auth login` |
| 不知道该填哪个模型名 | 只有模糊关键词，没有确定模型 ID | 先执行 `arkcli models search <keyword>` |
| 想让代码指向现有 endpoint | `--endpoint-id` 已移除 | 生成后把代码里的 `model="..."` 手动改成你的 `ep-xxx` |
| 下游脚本难以解析输出 | 直接消费了默认的人类可读输出 | 追加 `--format json` |

## 参考

- [arkcli-code-example](../SKILL.md) — skill 概览
- [arkcli-models](../../arkcli-models/SKILL.md) — 查模型名与版本
- [arkcli-deploy](../../arkcli-deploy/SKILL.md) — 创建 Endpoint 并进入正式接入
- [arkcli-infer-endpoint](../../arkcli-infer-endpoint/SKILL.md) — 创建或管理已有 Endpoint
- [arkcli-shared](../../arkcli-shared/SKILL.md) — 认证和全局参数
