# 查询和管理指定精调任务

本 reference 处理已知 job id 的详情、观察和生命周期操作。需要根据指标选择产物、导出 custom model 或继续部署时，改读 [`export-deploy.md`](export-deploy.md)。

## 命令速查

| 命令                                               | 何时用                     | 常用参数                                                |
| ------------------------------------------------ | ----------------------- | --------------------------------------------------- |
| `arkcli train finetune get <job-id>`             | 查完整详情，PRD-canonical 读入口 | `--transform`                                       |
| `arkcli train finetune status <job-id>`          | 查状态；语义上等同详情查询           | `--transform`                                       |
| `arkcli train finetune watch <job-id>`           | 持续监测进度及指标变化             | `--interval`、`--timeout`、`--quiet`、`--rich`         |
| `arkcli train finetune metrics <job-id>`         | 查指标曲线                   | `--metric`、`--from-step`、`--to-step`、`--output`     |
| `arkcli train finetune logs <job-id>`            | 查训练日志                   | `--tail`、`--since`、`--search`、`--follow`、`--output` |
| `arkcli train finetune trajectory list/get`      | 查 RL 轨迹                 | 仅在任务和 TLS 配置支持时使用                                   |
| `arkcli train finetune update <job-id>`          | 改名称或描述                  | `--name`、`--description`                            |
| `arkcli train finetune pause/resume/stop/delete` | 生命周期操作                  | 写/删/终止前必须确认；需要时才加 `--yes`                           |

## 先查询当前状态

任何写操作前先执行：

```bash
arkcli train finetune get <job-id>
```

确认 job id、当前阶段和目标操作是否匹配。具体阶段限制以当前命令 `--help` 和后端错误为准，不维护静态阶段矩阵。

## 详情与过程观察

按用户目标选择最小命令：

```bash
arkcli train finetune get <job-id>
arkcli train finetune watch <job-id>
arkcli train finetune metrics <job-id>
arkcli train finetune logs <job-id>
arkcli train finetune trajectory list <job-id>
```

执行前查看对应 `--help`。

- 一次性状态用 `get`，持续等待终态用 `watch`。
- 指标先查询可用 metric 名称，再按真实名称过滤。
- 日志、轨迹和指标可能很大；优先使用过滤、tail 或 `--output`，不要把完整内容灌入上下文。
- trajectory 仅在任务和当前 ArkCLI 能力支持时使用。
- 任务失败时返回阶段、错误原因和 CLI 提供的 hint，不要盲目重复提交。

## 任务元数据

更新名称或描述前，打印 job id、当前值和目标值并取得用户确认：

```bash
arkcli train finetune update <job-id> --help
arkcli train finetune update <job-id> --name <name>
```

只发送用户要求修改的字段。

## 暂停、恢复、停止和删除

先分别查看当前接口：

```bash
arkcli train finetune pause --help
arkcli train finetune resume --help
arkcli train finetune stop --help
arkcli train finetune delete --help
```

规则：

- `pause` 用于希望后续继续的任务。
- `resume` 仅在当前阶段允许时执行。
- `stop` 是不可逆终止，必须明确说明影响并获得确认。
- `delete` 删除任务记录，必须展示目标 job id、名称和当前阶段，并获得单独确认。**Phase 强校验**：仅终态（`Completed` / `Failed` / `Terminated`）可删，非终态在客户端直接拒绝并提示先执行 `train finetune stop`，避免用户输 y 后被 backend `phase_mismatch` 拒。
- Agent 环境要求 `--yes` 时，只能在用户确认后添加。
- 阶段不允许时，根据 CLI/backend hint 给出下一步，不通过反复重试绕过限制。

## 训练产物

用户只想查看产物时，可以执行：

```bash
arkcli train finetune artifacts list <job-id>
```

用户要选择最佳产物、导出或部署时，不在本文件继续，读取 [`export-deploy.md`](export-deploy.md)。
