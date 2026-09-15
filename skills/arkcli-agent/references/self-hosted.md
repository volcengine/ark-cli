# Self-hosted Environment 与 Work Queue

本页描述当前已实现的 API，不把 PRD 中的拟议 Worker Credential、lease_id、complete 或 Worker 列表当作现网契约。

## 创建与绑定

```bash
arkcli agent env create \
  --name enterprise-worker-env \
  --runtime-type self_hosted \
  --dry-run --format json
```

`--runtime-type` 仅在 env create 提供，映射为 `Config.Type`，优先于 `--config`/`--file` 中的 Type。也可继续使用 `--config '{Type: self_hosted}'`。不要附带 cloud 专用的 Networking、Packages、Env、SetupScript；CLI 不主动丢弃这些显式配置，由服务端判定是否合法。环境更新沿用 `env update`，不要把它当作 cloud/self_hosted 类型切换入口。

创建前确认最终参数；创建后从结果取得 Environment ID 并用 `env get` 回读。Session 沿用 `agent session create --agent-id <agent-id> --environment-id <env-id>`。环境创建不代表 Worker 在线，Session 创建不代表任务已执行；必须有企业 Worker 消费队列并回传工具结果。

## 查询和排障

```bash
arkcli agent env work stats <env-id> --format json
arkcli agent env work list <env-id> --limit 100 --page-all --format json
arkcli agent env work get <env-id> <work-id> --format json
```

- 上述命令走 `/api/v3/environments/{environment_id}/work` 数据面，复用当前产品/profile 的 API Key、Region、项目，不绕到 ArkBFF。显式 `--base-url` 必须与显式 API Key 配对；不要把 Key 写入文件、日志或报告。
- list 返回 `data` 和可选 `next_page`。用 `--page <next_page>` 继续；它是不透明 cursor，不是数字页码。`--page-all` 默认最多 10 页、每页 100 条，支持 `--page-limit` 和 `--page-delay`；达到页数上限后保留 `next_page`，不能声称结果完整。
- 当前 list 没有服务端 state/session_id/worker_id 过滤参数，不要编造 flags。按 Session 定位可用返回的 `data.id`；本期后端 Work ID 等于 Session ID，但优先使用实际返回值。
- 状态为 `queued → starting → active → stopping → stopped`。`stopped` 只表示 Work 结束，不等于 Session 成功；结合 `session get`、events 或 `+debug` 核实业务结果。
- `depth` 是未被 Poll 领取的 queued 数；`pending` 是已 Poll 但未 Ack 的 queued 数，不是 running 数。`workers_polling` 是最近 30 秒内具名 Worker 的 Poll 活动计数，不等于全部活跃任务的 Worker 数；0 不足以断定所有 Worker 离线。`oldest_queued_at` 覆盖 depth 与 pending。
- get/list/stats 是只读命令，不支持 `--dry-run`。查询失败时报告服务端错误，不转向其他产品或账号。

## 停止任务

先 get 核对目标和状态；只预览时：

```bash
arkcli agent env work stop <env-id> <work-id> --dry-run --format json
```

只有用户确认目标与停止影响后才执行 `agent env work stop <env-id> <work-id> --yes`。默认请求 `force=false`（优雅停止）；用户明确要求强制停止才增加 `--force`。不传 `--yes` 时真实执行返回 `requires_confirmation` 且不发请求。Preview 是零网络计划，不需要 `--yes`，也不代替执行授权或在线校验。

当前后端 stop 只接受 `force`，没有 `mode/reason` 参数。成功返回原始 Work 状态，不等候远端进程退出：优雅停止通常先进入 stopping；force 立即标记 stopped，但不能证明企业 Sandbox 已被杀死，也不能据此声称旧 Tool Result 一定被拒绝。之后单独 get 核对 Work 状态，并结合企业 Worker 日志核实清理结果；不要自动升级为强制停止。

## 本期边界

arkcli 提供环境/Session 管理与 Work 运维，不在本机常驻 Poll 或执行 Agent 下发的 bash/文件工具。Worker 接入使用前端已提供的官方 Go/Python/Java SDK；当前实现使用 API Key，不虚构 Environment Key 管理命令，也不把 Vault Credential 当作 Worker Credential。部署执行工具的 Worker 是独立操作，需要用户指定隔离环境并确认权限、工作目录与凭证边界。
