# arkcli infer endpoint get

Get infer endpoint endpoint details

## 调用异常：先取事实，再区分状态与可达性

已知精确 ID 时直接 `get <endpoint-id>`，不用先全量 list。保留完整 ID，不能把 `ep-m-` 改成 `ep-`，也不能凭时间前缀补随机后缀。NotFound 按主 Skill 的当前列表 + 历史用量分支核对；当前不可见不等于已删除。

1. 记录 `Status`、`StatusReason`、`ModelReference`、Project 与限流等实际返回字段；没有的字段标“未返回”，不补默认值。模型绑定有歧义时走 `resources resolve`，不能用 Custom Model 的基础模型 lineage 冒充实际绑定。
2. 复用本轮身份、region、base URL 与请求时间上下文，保留错误码、RequestId 和脱敏报文。不要要求用户粘贴完整 Authorization / API Key，也不为了排障自动切 Profile、刷新/轮换 Key。
3. `Running` 只证明控制面状态；`usage stats --endpoint <id> --start <日期> --end <日期>` 只证明该窗口的聚合用量，不提供单请求错误率或调用链。错误率/延迟诊断转 Doctor，近期无数据须说明聚合延迟。
4. 用户已授权真实试调用时，才沿该模型支持的 API/模态做一次最小复现；文本对话使用 `+chat --model <endpoint-id>`，不是 `--endpoint`。需要图片/视频输入时保留原输入条件，不用纯文本探针证明多模态成功。未实测时明确写“数据面未验证”，不宣称一切正常。

| 现象 | 需要核对的分支；不是仅凭错误码认定的根因 |
|---|---|
| 401 | 凭证类型、Bearer 格式、空白字符、有效期、身份匹配；仅展示脱敏信息，按宿主认证协议处理 |
| 403 | 当前 Key 的资源范围、IAM 权限、region 与实际 Project；account-wide 不等于具备所有权限，也不能先认定 Project 隔离；不建议直接扩大到全部权限 |
| 404 / NotFound | 原始 ID、版本、region、可见范围、base URL 与 API 路径；目录存在不证明该 Endpoint 可访问 |
| 500 | API 与模型模态是否匹配、原请求参数、服务错误及 RequestId；不能把所有 vision 模型当 embedding，也不凭 500 断言是用户客户端或网络 |
| ClosedEndpoint | `get` 核对状态；启动、重建是写操作，按用户意图与宿主授权执行，不为诊断自动修复 |

OpenAI 兼容 SDK 通常自行拼 API 路径，base URL 不重复带 `/chat/completions`；Platform / Agent Plan / Coding Plan 路径不可混用。具体以执行上下文和客户端契约为准，不靠换路径碰运气。

### 特惠区、保障包与删除的业务边界

- `ClosedEndpoint` 或费用相关错误也可能与账号欠费有关：先核对 Endpoint 状态和错误原文，再引导用户在火山引擎控制台费用中心确认现金余额、欠费与恢复状态。`usage balance` 是套餐/资源额度，不是账户现金余额，Billing 账单也不能证明实时可用余额；本轮无法查到时明确交由费用中心核验，不直接认定已欠费或自动充值。
- 从特惠区迁到保障包按新 Endpoint 迁移处理，不把更新名称/限流说成“原地升级保障包”。先确认目标计费产品和创建路径；新 EP 创建后单独验证数据面，再由用户确认切流与旧资源停用。
- 旧 EP 的 `context_id` 不能直接复用于新 EP；重新建立目标上下文，不能复用旧缓存成功结果证明新链路可用。
- 删除 Endpoint 不等于退订 TPM 保障包；包未退订可能继续计费。用户问扣费时转 Billing 核账与对应退订流程，不能承诺删 EP 即停止所有费用。
- `OperationDenied.EndpointStatus` 时先 `get` 核对，必要的 stop/delete 仍分别受宿主写操作守卫约束。

## Usage

```bash
arkcli infer endpoint get <endpoint-id> [flags]
```

## Arguments

| Argument | Description | Required |
|----------|-------------|----------|
| `<endpoint-id>` | The ID of the endpoint to get | Yes |

## Flags

| Flag | Type | Description | Required |
|------|------|-------------|----------|
| `-h`, `--help` | | help for get | No |

## Global Flags

| Flag | Type | Description |
|------|------|-------------|
| `--debug` | | Print request and response debug details to stderr |
| `--format` | string | Output format: json (default "json") |
| `--page-all` | | Automatically fetch all pages when supported |
| `--page-delay` | int | Delay in milliseconds between pages (default 200) |
| `--page-limit` | int | Maximum pages to fetch with --page-all (default 10) |
| `--profile` | string | Active config profile |
| `--transform` | string | Transform output with a GJSON-style path expression |
