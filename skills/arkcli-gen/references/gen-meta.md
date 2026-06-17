---
name: gen-meta
description: arkcli gen get / list / delete 三个标准 CRUD 命令的 reference, 与 +gen 工作流配套使用以管理已提交的异步生成任务（视频 / 3D / 未来扩展）。
---

# gen get / list / delete

这是 `+gen` 的"对账面"——`+gen` 负责提交生成任务并自动轮询等待；这三个命令负责事后查询任务状态、列出历史任务、删除已完成任务。

## 何时使用

- `gen get <task-id>` —— 已经有 `task_id`，想拿回任务的最新状态（含 `output_url` / `usage` / `error`）。**任务一旦 `succeeded`，`gen get` 会把产物自动下载到本地**（默认当前目录，文件名 `<task-id>.mp4`），跟 `+gen --wait` 的落地行为一致；`--save-to=""` 可关闭。
- `gen list` —— 列出当前账户下的异步任务，支持按 `status` / `model` / `service-tier` / `task-id` 过滤。
- `gen delete <task-id>` —— 删除一个已完成的任务记录。

## 模态覆盖

这三个命令背后的端点（SDK `arkruntime.ListContentGenerationTasks` / `GetContentGenerationTask` / `DeleteContentGenerationTask`）在服务端是 **modality-agnostic** 的：

- `+gen` 视频任务（`seedance-*` 模型族）创建的任务
- 3D 任务（`seed-3d` 模型族）创建的任务
- 任何通过 `arkcli api arkruntime.create_content_generation_task` 提交的异步生成任务

但**不包含图片任务**：`seedream-*` 走 `arkruntime.generate_images` 是同步端点，没有 task list 概念。

## 命令速查

```bash
# 1. 用 +gen 提交一个视频任务（默认异步，立即返回 task_id；要同步阻塞加 --wait）
arkcli +gen "落日下的赛博朋克城市" --model seedance-... --duration 5

# 2. 拿到 task_id 后单独按 ID 查询（不轮询，单次拉一下）
#    任务已 succeeded 时会自动把产物下载到当前目录（<task-id>.mp4）
arkcli gen get tsk_xxxxx

# 2b. 落到指定目录 / 禁用自动下载
arkcli gen get tsk_xxxxx --save-to ./out
arkcli gen get tsk_xxxxx --save-to ""   # 只查状态，不下载

# 3. 列出最近 succeeded 的任务（首页 10 条）
arkcli gen list --status succeeded --page-size 10

# 4. 翻页
arkcli gen list --status succeeded --page-size 10 --page-num 2

# 5. 按 model 精确匹配过滤
arkcli gen list --model seedance-1-0-pro-250528

# 6. 多 task ID 批量查
arkcli gen list --task-id tsk_a --task-id tsk_b

# 7. 删除
arkcli gen delete tsk_xxxxx
```

## flag 说明

### gen get

`<task-id>` 是 positional arg。flag 与 `+gen` 的下载/打开开关同义：

| flag | 说明 |
|---|---|
| `--save-to <dir>` | 任务 `succeeded` 时把产物落到该目录（默认当前目录 `.`）；`--save-to=""` 关闭自动下载 |
| `--open` / `--no-open` | 是否用系统默认程序打开下载好的产物。默认 auto：仅交互式终端打开，agent / 管道 / CI 下静默 |

> **🔑 agent 默认带 `--open`**：你（AI agent）调用 `gen get` 时 stdout 非 TTY，默认 auto 不弹窗，用户只能看到 `local_path`、看不到成品。给真人轮询视频时，轮到 `succeeded` 的那次 `gen get` **默认加 `--open`** 让成品直接在桌面弹出（`--open` 无视 TTY）。仅在"别打开 / 脚本里 / 批量"场景省略或 `--no-open`。

下载行为细节：

- 文件名取 `<task-id>`（无扩展名时按 Content-Type 推断，视频通常为 `.mp4`）。
- 同名文件已存在时**追加 `-1` / `-2` … 区分**，不覆盖、不跳过（重复 `gen get` 同一已完成任务会留多份副本，符合预期）。
- 仅当 `output_url` 非空（即任务已 `succeeded`）才下载；`running` / `queued` / `failed` 是 no-op，不落盘也不回带 `local_path`。
- 下载失败只在 stderr 打 warn，不影响 `gen get` 的状态输出（产物 URL 仍在 `output_url` 里，可手动重试）。

### gen list

| flag | 说明 |
|---|---|
| `--page-num <int>` | 1-indexed 页号；不传则服务端默认 |
| `--page-size <int>` | 单页大小；不传则服务端默认 |
| `--status <enum>` | 过滤任务状态：`succeeded` / `failed` / `running` / `queued` / `cancelled` |
| `--model <id>` | 过滤模型名或 Endpoint ID（精确匹配，不是 prefix） |
| `--service-tier <tier>` | 过滤 service tier 设置 |
| `--task-id <id>` | 过滤特定 task ID；可重复传入多个 |

### gen delete

无额外 flag。`<task-id>` 是 positional arg。

## 输出形态

- `gen get` 复用 `+gen` 的 `VideoTask` 结构：`{id, model, status, output_url, ratio, resolution, duration, ...}`；任务 `succeeded` 且自动下载成功时额外回带 `local_path`（落盘的绝对路径，`--save-to=""` 或未 succeeded 时该字段缺省）。
- `gen list` 返回 `{total: <int>, items: [<item>, ...]}`；每个 `item` 是 SDK 列表 item 的原始 JSON 形态，**注意以下 3 处与 `gen get` 的差异**：
  - 错误字段叫 `failure_reason`，不是 `error`
  - **不含 `resolution` / `ratio` / `duration`**（要拿这些字段需要再调一次 `gen get`）
  - `subdivisionlevel` / `fileformat` 这两个字段的 JSON key 是单词式（无下划线），SDK wire 真相
- `gen delete` 返回 `{id, deleted: true}`；服务端无响应体，这里是 arkcli 合成的。

## 常见错误

| 现象 | 原因 |
|---|---|
| `gen get` / `gen delete` 报 `missing task id` | positional arg 没传 |
| `gen list` 返回空 `items[]` | filter 太严，或当前账户下确实没任务 |
| `gen get` 拿不到 `output_url` | 任务还在 `running` / `queued`；看 `status` 字段 |
| 把图片任务返回的 ID 拿过来 `gen get` 报 not found | 图片走同步端点不是 task list，用法错位 |

## 与 raw API 的关系

| 子命令 | 等价的 raw API |
|---|---|
| `gen get <id>` | `arkcli api arkruntime.get_content_generation_task --params '{"id":"<id>"}'` |
| `gen list ...` | `arkcli api arkruntime.list_content_generation_tasks --params '{"page_num":1,"filter":{...}}'` |
| `gen delete <id>` | `arkcli api arkruntime.delete_content_generation_task --params '{"id":"<id>"}'` |

CLI 子命令是 raw API 的人类友好包装；命令更短、flag 更直观。raw API 仍然是兜底通道。

## 与 `+gen` 的协作

- `+gen` 自身轮询超时时会回带 `task_id` 和提示 —— 用 `gen get <task-id>` 接着追，比重新跑 `+gen` 划算（重新跑会再创建一个新任务并扣一次配额）。轮询到 `succeeded` 的那次 `gen get` 会顺手把产物下载到本地，不必再手动 curl `output_url`。
- 异步提交（不带 `--wait`）+ `gen get` 轮询是视频的主流用法：`arkcli +gen ... --modality video` 拿到 `task_id` → 反复 `gen get <id>` 直到 `succeeded` → 产物自动落地。带 `--wait` 则是同步阻塞版，两条路径的落地行为一致。
- `gen list` 是排查"我之前提交的那个任务到底在不在"的唯一入口；提交后 `+gen` 没拿到正常返回时建议先 `gen list --status running` 看看。
