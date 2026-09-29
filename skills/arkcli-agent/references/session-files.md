# Session 与 Files

## Env / Session

```bash
arkcli agent env create --name arkcli-<domain>-env-<timestamp> --config '{Type: cloud, Networking: {Type: unrestricted}}' --format json
arkcli agent session create --agent-id <agent-id> --environment-id <env-id> --title arkcli-<domain>-session-<timestamp> --format json
arkcli agent session get <session-id> --format json
```

Session 标准 CRUD 走 OpenTOP；Session resources / events / threads 走数据面直联。
线上创建 cloud env 时需要显式带 `Networking: {Type: unrestricted}`；只传 `{Type: cloud}` 会被后端校验为缺少 `config.networking`。

读取 Session 时保留 stdout、stderr 和退出码，先检查业务错误，再按实际 envelope 取字段：OpenTOP 通常为 `Result.Id` / `Result.AgentId`，其他入口可能包在 `data`；以完整 JSON 为准，不能把顶层 `.Id == null` 当作 Session 不存在。字段未回显、成功空结果、权限失败、鉴权失败和解析错误分别报告；已有完整响应只修正本地取值路径，不为取错层级重新查询，也不合并 stderr 后当 JSON 解析。

### Session TOS 资源

创建 Session 时可以直接绑定一个 TOS 目录。CLI 会把它编排为 `Resources` 数组中的一项，字段与前端及 OpenTOP 契约一致：`Type=tos`、`TosBucket=<bucket>`、`TosKey=<prefix>/`。

只有用户明确要求绑定 TOS 且提供了地址时才传 `--tos-path`。用户未提到 TOS 时不要默认绑定；用户要求绑定但没有给出地址时，先询问完整的 `tos://<bucket>/<prefix>/`，不要猜 Bucket 或 Prefix。

```bash
arkcli agent session create \
  --agent-id agent-xxx \
  --environment-id env-xxx \
  --tos-path tos://my-bucket/analysis/ \
  --format json
```

`--tos-path` 也接受不带协议的 `my-bucket/analysis/`。它表示目录而不是单个对象，因此 prefix 必须非空并以 `/` 结尾；`tos://bucket/`、`bucket`、`bucket/path` 会在 CLI 层校验失败。已有的 `--resource '[...]'` 仍可使用；同时传两者时 TOS 资源会追加到原始资源数组中。

`arkcli +new session` 和 `arkcli +iterate` 创建 Session 时同样支持 `--tos-path`。该参数只绑定已有 TOS 目录，不负责创建 Bucket、上传对象或申请 TOS 权限；调用前需确保账号已开通 TOS，并对 Bucket/Prefix 有读取权限。

### 其他创建时资源

复杂资源在 `session create --resource` 中声明，不等同于仅支持 file 的 `session resources add`。例如绑定用户已选定的 memory store：

```bash
arkcli agent session create --agent-id <agent-id> --environment-id <env-id> \
  --resource '[{Type: memory_store, MemoryStoreId: ms-xxx, Name: user-memory}]' --format json
```

规范字段使用 PascalCase，类型值仍是 `memory_store` 等 snake_case。当前 CLI 也会把 `memory_store_id` 等已注册别名归一化为 `MemoryStoreId`，不能断言 snake_case 必然失败。遇到 required 错误，先检查资源对象层级、真实 ID 和本地 Preview 的最终 payload；不要反复猜字段、虚构 ID 或擅自创建替代资源。

### Environment 自定义脚本

Environment 的初始化脚本位于 `Config.SetupScript`。可以直接传文本，也可以用 `@` 读取本地脚本文件：

```bash
arkcli agent env create \
  --name arkcli-analysis-env \
  --config '{Type: cloud, Networking: {Type: unrestricted}}' \
  --setup-script @./bootstrap.sh \
  --format json
```

`--setup-script` 会覆盖 `Config.SetupScript`；复杂的 `Config` 仍可用 `--config @./env.yaml` 或 `--file` 透传。CLI 会把 `setup_script`、`networking` 等常见 lower/snake case 写法归一化成 OpenTOP 的字段。

### Session overrides

override 模型运行参数前，按[模型参数 metadata](model-config.md)查询目标模型；切换模型需重新查询，不自动注入默认值。

Session 创建可以对已有 Agent / Environment 做一次性配置覆盖。override 内的 `System`、`Tools`、`McpServers`、`Skills`、`Multiagent` 和 Environment `Config` 是服务端定义的替换语义，非 nil 字段不会与基 Agent 自动合并；传数组时要传完整数组。

```bash
arkcli agent session create \
  --agent-id agent-xxx \
  --environment-id env-xxx \
  --agent-overrides '{model: {id: doubao-seed-2-1-pro-260628, thinking: enabled, reasoning_effort: low, service_tier: auto}, Tools: [...]}' \
  --environment-overrides @./environment-overrides.yaml \
  --format json
```

`--agent-overrides` 映射为 `AgentWithOverrides`，`--environment-overrides` 映射为 `Environment`，缺少 override 对象中的 `Id` 时 CLI 会从 `--agent-id` / `--environment-id` 补齐。override 是服务端 one-of 变体，CLI 会在最终请求中移除对应的 `AgentId` / `AgentVersion` 或 `EnvironmentId`，避免同时传两个互斥变体。`arkcli +new session` 和 `arkcli +iterate` 也支持同名参数。需要传自定义 Agent 配置时不要同时使用 `--agent` 和 `--agent-overrides`。

`AgentWithOverrides.Model` 只支持 `Id`、`Speed`、`Thinking`、`ReasoningEffort`、`ServiceTier`。`Id` 可以为当前 Session 切换模型；CLI 会从最终请求中剔除 `Provider`、`Protocol`、`BaseUrl`、`Headers` 及其他未定义的 Model 字段，不会因为这些字段在输入中出现而本地报错，也不要依赖被剔除字段生效。模型参数必须以**切换后的模型** metadata 为准：`Thinking`、`ReasoningEffort` 与 `ServiceTier` 的支持值会随模型变化。Managed Agent 的 `ServiceTier` 产品枚举为 `auto`、`default`、`fast`，没有 `priority`；不要猜测或把其他业务中的 priority 概念映射到这里。`auto` 在服务端回读时可能被解析为实际生效的 `default`，这不等同于字段丢失。拿不到目标模型 metadata 时，省略不确定的运行参数并让服务端使用模型默认值。

上述归一化和 Model 字段过滤同时适用于 `--agent-overrides`、`--agent-overrides @file` 以及通用 `session create --file` 内的 `AgentWithOverrides`。显式空值 `--agent-overrides=` / `--environment-overrides=` 是无效输入，CLI 返回 validation error，不创建 Session。

### Environment 状态筛选

当前 `arkcli agent env list` **没有注册 `--status`**，不要生成下面这种命令：

```bash
arkcli agent env list --status active
```

`ListEnvironments` 后端请求也没有 `Status` 过滤字段；返回的 Environment 对象只有可选的 `ArchiveTime`。需要按状态找环境时，先拉全量，再由调用 arkcli 的 AI Agent 本地筛选：

```bash
arkcli agent env list --page-all --format json
```

- `ArchiveTime` 非空：`archived`
- `ArchiveTime` 为空或字段缺失：`active`
- 前端领域类型虽然预留了 `updating`，但当前返回契约没有可可靠推导它的字段；不要声称支持 `updating` 筛选。
- 不要只过滤默认第一页；状态筛选必须配合 `--page-all`，并关注全局 `--page-limit` 是否导致结果仍被截断。
- env `--page-all` 默认使用 `Limit=100`，并把每页响应的 `NextPage` 作为下一次请求的 `Page`；`--page-limit` 限制的是最多请求页数，不是结果条数。

## Files 与 Session Resources

- 本地文件上传到 Files API：`arkcli agent file upload --path ./data.csv --purpose user_data --wait-active`。
- URL/TOS 注册：`arkcli agent file upload --url https://... --purpose user_data` 或 `--url tos://bucket/path/file --tos '{bucket: b, prefix: arkfiles/}'`。
- `agent file upload/delete` 支持命令级 `--dry-run`：只输出零网络 `preview.v1`，不会上传、轮询或删除。删除预览不要求 `--yes`，真实删除仍需要明确确认并传 `--yes`。
- 视频/特殊文件预处理可传 `--preprocess-configs`；有效期可传 `--expire-at <unix-seconds>`。
- 上传后等待状态：`arkcli agent file wait <file-id>`。
- 查询已有文件：`arkcli agent file list --purpose user_data --limit 20`；全量遍历用 `arkcli --page-all agent file list`，CLI 默认 `limit=100` 并沿 `has_more + last_id -> after` 拉取。
- 用户给本地路径并要挂到已有 session 时，优先一步完成：

```bash
arkcli agent session resources add <session-id> --path ./data.csv --mount-path data.csv
```

这会先上传 file，再等待 active，最后 add session resource。
`session resources add --path` 支持透传上传相关参数：`--purpose`、`--tos`、`--preprocess-configs`、`--expire-at`、`--wait-timeout`、`--wait-interval`。

边界：

- `session resources add` 当前只支持 `type=file`、`file_id`、`mount_path`，不支持 PRD 中更复杂的 `github_repository` / `memory_store` 参数。
- 后端会自动在 `mount_path` 前添加 `/mnt/session/uploads/`。CLI 在 add 前会向 stderr 提示这一点，但不会改写用户传入的路径；例如传 `reports/data.csv`，最终路径为 `/mnt/session/uploads/reports/data.csv`。避免重复传入 `uploads/` 或完整受管前缀，并且不要传包含 `..` 的越界路径。
- `resources get` 是 CLI 派生能力：调用 list 后按 `resource_id` / `file_id` / `mount_path` 本地筛选。
- 线上数据面当前没有原生 resources update/delete，CLI 保持 fail-fast unsupported。
- `session resources add` 后端会复制源文件到 session 受管 uploads 路径，因此 resources list 看到的 file_id 可能不同于 add 时传入的源 file_id。
- 不要调用 ArkBFF/NodeBFF 获取文件；CLI 只直连 `/api/v3/files` 数据面。

## 调试与导出

- 已知 Session ID 的异常定位用 `arkcli +debug <session-id> --format json`，结合 Session、近期 events、resources 和 threads 分清失败、等待确认与仍在运行；聚合结果中的 warnings 表示部分查询失败，不等于对应数据为空。
- 用户需要诊断包时用 `arkcli +export <session-id> --output ./session-diagnostics.tar.gz`。它导出 CLI 可见的诊断数据，不是完整工作区、memory store 或环境快照；以 archive 的 manifest/notes、事件数量和 warnings 说明范围与限制。
- 导出文件位于当前 CLI 宿主；交付前检查可能含有的消息、代码或业务数据。远端沙箱路径不能当成用户电脑路径，不自动上传或公开诊断包。

## 示例

```bash
arkcli agent file upload --path ./sales.csv --purpose user_data --wait-active --format json
arkcli agent session resources add sess-xxx --path ./sales.csv --mount-path sales.csv --format json
```
