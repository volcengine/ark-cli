# API Key：查询、导出与注入

先读 [共享协议](../../arkcli-shared/SKILL.md)。这里只处理用户明确要求的 Key 任务，不是所有业务命令的前置步骤；宿主对凭证与授权的限制始终优先。

## 1. 元数据查询不改变当前 Key

- “有哪些 Key / 谁的 Key / 哪把 Key 有用量”先用 `arkcli api apikey.list --params '{"PageSize":100}' --page-all --format json`，按实际 envelope 的 `Items` 读取 Id、名称、掩码、状态与 Project 等字段；检查分页完整性。
- 账户 Key 列表不是 Profile 本地库存。不要用会选 Key 并写凭证的 `auth apikey` 代替 list，也不要为只读查询自动执行 `profile keys refresh/use`。
- `Key` 的掩码不能用作推理凭证，不能去掉星号或拼后缀“还原”。用户给名称/后缀时只在真实结果中匹配；0 个说明未找到，多个需要消歧，不取第一条。
- 普通 Key 池、个人 Agent Plan 专属池、团队席位 Key 不是同一个池；普通列表有 Key 不证明 Plan Key 已就绪。

## 2. 用户明确需要凭证时

先确认精确目标和用途：当前 Profile 的 Key、某个账户 Key，还是注入已授权的客户端。不要从“列 Key”自动扩张成全量明文导出。

- 当前 Profile 已存的 Key 有产品入口 `arkcli profile apikey get [--profile <name>] [--plain]`；它导出持久 Profile 凭证，不代表本次 `--api-key`/环境变量 override 的值。只在明确授权的导出/注入任务中使用。
- 账户中指定 Key：list 定位真实数字 `Id`，再使用已注册 `apikey.get_raw`，请求字段为 `{"Id":123}`（示意 ID 不可实用）。返回的 `Result.ApiKey` 才是明文；不能把 `Id`、`KeyID`、`SID` 或名称混用。
- 当前 registry 提供 `apikey.get_raw` 和 `apikey.get_sid`，不要发明 `apikey.get` 作为替代。`get_sid` 是持有 ApiKey 后反查 SID/Project/状态，不是凭 Id 取明文；两者输入和返回值不同。先核对当前 `api --list`，不能用相似 Action 名碰撞。
- `get_raw` 属于敏感凭证访问，个人 Plan 可能触发协议/Access 闸门；明确拒绝或状态不可确认时停止，不改走文件、另一 Action 或另一身份绕过。
- 遵守共享“不输出完整凭证”的交付边界：明文只进入用户授权的配置/安全目标，不进入聊天、debug、临时报告、命令展示或版本库；反馈只写目标、名称/后缀和完成状态。宿主不允许导出时如实说明。
- 明确授权批量导出时仍需锁定目标清单、权限与安全落点，记录逐项成功/失败；默认串行，不自动轮换、创建、禁用或改默认 Key。

## 3. 下游注入

优先让 ArkCLI/Helper 按选定 Profile 自行消费凭证，不手工转抄。确需外部客户端时，在同一受控进程内取值并写入已授权位置，禁止 `set -x`、echo Key 或将秘密展开成可见命令；失败只报告非敏感摘要。

导出/注入与轮换完全不同：查询失败不意味着需要重置 Key，读取明文不意味着旧 Key 失效，配置写入成功也不证明下游已加载或可调用。需按目标客户端核验，不能默默更换费用主体。

### 获准后、同一受控进程内的解析模板

仅当宿主允许导出、用户已指定 Key 与安全落点，且不能通过 Helper 直接交接时采用。下面只描述取值步骤，**不是业务调用前置检查**：

```bash
set +x
# key_id 来自已完整读取的 list；是数字 Id，不是掩码、名称或 SID
key_params=$(jq -ecn --argjson id "$key_id" 'select(($id|type)=="number" and $id>0 and ($id|floor)==$id) | {Id:$id}') || exit 1
# 实际调用前补齐共享协议要求的归因环境；不加 --debug
key_response=$(arkcli api apikey.get_raw --params "$key_params" --format json) || exit 1
resolved_key=$(printf '%s' "$key_response" | jq -er '.Result.ApiKey | select(type=="string" and length>0 and (contains("*")|not))') || exit 1
# 只在此进程内交给已经获准的客户端/配置写入器；不 echo、不列进 argv 或报告
# 完成交接后立即清理进程变量
unset resolved_key key_response key_params
```

模板中的 handoff 步骤必须使用目标客户端真实接口，不能在清理变量后声称已注入。超出 JSON 工具可精确表达范围的数字 Id 使用支持精确整数的编码器，不能舍入后查询另一条记录。禁止用 `tr -d '"'` 代替 JSON 解码；非零退出、空值或掩码值均为失败，不暴力换 Id。

批量导出只遍历用户批准的真实 Id 清单、默认串行；每项记录 Id/成功或失败/安全落点，不能把所有明文汇总进对话。交付前核对清单完整性、未改变默认 Key、秘密未进入 Trace，失败项如实报告而不轮换重试。
