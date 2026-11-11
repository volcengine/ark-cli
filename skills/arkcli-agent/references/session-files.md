# Session 与 Files

`session resources list/get/add` 是数据面命令，支持显式 `--api-key` / `--base-url`。显式地址必须同时提供显式 Key，且地址应为 MA 数据面地址，不是 BYOK 模型的 BaseUrl；只提供 Key 时使用当前产品、Region、环境的 MA 标准地址，不继承套餐路由。这不改变 `session create/get` 等控制面命令的认证要求。

## Env / Session

### Session 项目一致性

创建 Session（包括 `session create`、`+new session`、`+iterate`）时，所选 Agent、Environment 及其他带项目归属的关联资源必须与 Session 处于同一 Project；override 不豁免该要求。

- 先通过 `auth status` 确认所选 Platform profile 的项目上下文，再用资源的 list/get 返回值核对实际 `ProjectName`。`account-wide` / “账号全部资源”只是查询范围，不是具体 Project，也不证明列出的资源属于同一项目。
- 若 Session 的目标 Project 尚未确定、资源归属缺失或存在冲突，先补查或向用户确认，不按名称、同账号、可见性或创建时间推断。对没有项目归属的资源，不虚构 `ProjectName`。
- 选择环境时先筛选同 Project 的候选，再按用户需求选择，并显式传 `--environment-id`（或入口支持的 `--env-id`）；不能依赖账号范围内“最新环境”的自动选择。没有同项目候选时，确认后在目标项目创建，不借用其他项目资源。
- 使用已确认的项目级 Platform profile 完成创建；没有对应 profile 时按 Profile skill 处理，不擅自修改全局默认，不伪造全局 `--project-name`。创建后用 `session get` 回读项目与资源绑定；预览成功不能代替项目归属核对。

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

Session 创建可以对已有 Agent / Environment 做一次性配置覆盖。override 内的 `System`、`Tools`、`McpServers`、`Skills` 和 Environment `Config` 是服务端定义的替换语义，非 nil 字段不会与基 Agent 自动合并；传数组时要传完整数组。`Multiagent` 按字段是否出现区分省略、`null` 和空数组，不能一概当作整对象替换；具体配置和语义见 [Multiagent](agent.md)。

```bash
arkcli agent session create \
  --agent-id agent-xxx \
  --environment-id env-xxx \
  --agent-overrides '{model: {id: doubao-seed-2-1-pro-260628, thinking: enabled, reasoning_effort: low, service_tier: auto}, Tools: [...]}' \
  --environment-overrides @./environment-overrides.yaml \
  --format json
```

`--agent-overrides` 映射为 `AgentWithOverrides`，`--environment-overrides` 映射为 `Environment`，缺少 override 对象中的 `Id` 时 CLI 会从 `--agent-id` / `--environment-id` 补齐。override 是服务端 one-of 变体，CLI 会在最终请求中移除对应的 `AgentId` / `AgentVersion` 或 `EnvironmentId`，避免同时传两个互斥变体。`arkcli +new session` 和 `arkcli +iterate` 也支持同名参数。需要传自定义 Agent 配置时不要同时使用 `--agent` 和 `--agent-overrides`。

`AgentWithOverrides.Model` 支持 `Id/Provider/Protocol/BaseUrl/Headers/TokenLimits/InputModalities` 及 `Speed/Thinking/ReasoningEffort/ServiceTier`；不再过滤用户显式提供的模型接入字段。类型、省略语义和请求凭证见 [BYOK](byok.md)。按目标模型能力选择参数，不猜默认值，也不把配置保存成功当成推理成功。

上述模型结构归一化同时适用于 `--agent-overrides`、`--agent-overrides @file` 以及通用 `session create --file` 内的 `AgentWithOverrides`。显式空值 `--agent-overrides=` / `--environment-overrides=` 是无效输入，CLI 返回 validation error，不创建 Session。

### 归档（火山）

`agent agent archive <agent-id>`、`agent env archive <env-id>`、`agent session archive <session-id>` 是不可逆归档，不是删除，也不自动中断任务。先 get 核对目标，再取得用户对归档的明确确认，才加 `--yes`。不要因为归档失败自动改用 delete 或 interrupt；没有恢复命令。

```bash
ARKCLI_NO_UPDATE_NOTIFIER=1 ARKCLI_CALLER_TYPE=ai_agent ARKCLI_CALLER_NAME=codex ARKCLI_SKILL_NAME=arkcli-agent arkcli agent session archive <session-id> --dry-run --format json
ARKCLI_NO_UPDATE_NOTIFIER=1 ARKCLI_CALLER_TYPE=ai_agent ARKCLI_CALLER_NAME=codex ARKCLI_SKILL_NAME=arkcli-agent arkcli agent session archive <session-id> --yes --format json
```

Agent 归档不产生新版本，保留历史配置；Environment 归档不删除数据。已有 Session 的行为以服务端为准，不承诺停止运行或释放费用。Session 归档要求满足后端状态前提，运行中拒绝时先汇报，不能擅自结束用户任务。

火山三类 `list` 默认排除归档资源，查询归档资源必须带 `--include-archived`。Agent / Environment 另支持 `--status active` 或 `--status terminated`，可重复；只查归档资源时同时传 `--include-archived --status terminated`。单独传 `--status terminated` 不会自动包含归档资源，空结果不能据此判断没有归档资源。这里 `terminated` 是后端归档筛选值，不写成 `archived`。Session `--status` 仍使用它原有的运行状态枚举，不套用 Agent/Env 枚举。

```bash
ARKCLI_NO_UPDATE_NOTIFIER=1 ARKCLI_CALLER_TYPE=ai_agent ARKCLI_CALLER_NAME=codex ARKCLI_SKILL_NAME=arkcli-agent arkcli agent env list --include-archived --page-all --format json
ARKCLI_NO_UPDATE_NOTIFIER=1 ARKCLI_CALLER_TYPE=ai_agent ARKCLI_CALLER_NAME=codex ARKCLI_SKILL_NAME=arkcli-agent arkcli agent agent list --include-archived --status terminated --page-all --format json
ARKCLI_NO_UPDATE_NOTIFIER=1 ARKCLI_CALLER_TYPE=ai_agent ARKCLI_CALLER_NAME=codex ARKCLI_SKILL_NAME=arkcli-agent arkcli agent env list --include-archived --status terminated --page-all --format json
```

看实际 `ArchiveTime` / `Status` 判断归档结果；旧响应缺字段不推断为已经归档。若服务端版本尚不支持 Status，不把空结果当作“没有归档资源”。`--page-limit` 是最多请求页数，达到上限后检查 NextPage。BytePlus 不新增上述 flags/命令，继续使用旧查询方式。

## Files 与 Session Resources

### 查看挂载资源

先 `agent session get <session-id>` 核对 Session，再 `agent session resources list <session-id>` 查看接口实际返回的挂载项。需要单项时用 `resources get <session-id>` 按该命令 help 中的资源 ID/file ID/mount path 筛选；文件 metadata 用 `agent file get <file-id>`。列出的挂载项不等于完整 Sandbox 目录树，file get 也不等于下载内容。没有公开读取接口时直接说明边界，不生成 `fs tree`、任意 Sandbox path 读取或内部接口命令。

Memory 挂载须使用返回的 store ID，再查询 `agent memory-store get` / `memories list` / `memories get --view full`。条目默认 basic 视图不包含正文；Path 更新后以响应中的 ID 和 Path 为准，不假设旧 ID 仍可用。不连接 OV 内部存储，也不据此绕过账号或资源权限。

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
## 工具版本

Session 的 AgentWithOverrides.Tools 是完整数组；不提供就沿用绑定 Agent 的配置，不自动升级旧工具集。显式选择 20260916 或提示词增强工具前，读取 [工具配置](agent.md#默认-agent-工具)，保留完整的其他工具、权限与 Models。
