# 模型运行参数 metadata

创建、更新 Agent，或对 Session 做 override、upgrade 前，需要配置模型运行参数时，先查询目标模型的 metadata。`agent model list` 用于选择模型，不提供这些参数的完整可选值。

## 查询

工具模型的 `epa_tool_model.support` 不由此接口提供；使用 `agent model list --usage tool` 读取完整 ArkModels 并按版本筛选，缓存及参数说明见 [工具模型查询](agent.md#查询工具模型与缓存)。不要将此处 key 查询为空解释为不支持工具模型。

```bash
arkcli agent model config <foundation-model-name> --provider ark --format json
arkcli agent model config <foundation-model-name> --key thinking.support_value,thinking.default --key service_tier.support_value,service_tier.default --format json
```

| 输入 | 说明 |
| --- | --- |
| `<foundation-model-name>` | 必填，使用选中模型的 `agent model list` 结果中的 `name`，不是带版本的 `model`、模型 ID 或 `ep-*`。不要自行截断版本号推导名称。 |
| `--provider` | 默认 `ark`，与前端查询一致；不是账号或产品切换参数。其他 provider 只有在已确认目标模型身份时使用。 |
| `--key` | 可重复、可逗号分隔；不传时查询全部 metadata。key 原样交给服务端，不维护客户端白名单。 |

命令是当前 profile 下的只读控制面查询，调用 `ListManagedAgentModelMetaDatas`；不接受模型版本、分页或 `--dry-run`。返回服务端 `ResponseMetadata` 和 `Result` 数组，每项包含 `FoundationModelName/Provider/Key/Data`。`Data` 是字符串数组，接口不返回内部 Schema。

如果只有带版本模型标识，先从 `agent model list` 中匹配对应 `model`，读取该项 `name`；不要直接把 Endpoint ID 当作名称。查询失败应保留服务端错误，不自动切账号、产品或 provider 重试。

## 参数逐项说明

| 参数 | 含义 | 查询 key |
| --- | --- | --- |
| `speed` | 模型提供的速度模式；仅在目标模型支持时设置，不等同于所有模型都支持的通用加速开关。 | `speed.support_value`；当前契约未定义 `speed.default`。 |
| `thinking` | 思考模式，例如部分模型支持 `enabled`、`disabled` 或 `auto`。 | `thinking.support_value`、`thinking.default` |
| `reasoning_effort` | 推理投入程度，例如部分模型支持 `low`、`medium`、`high`。 | `reasoning_effort.support_value`、`reasoning_effort.default` |
| `service_tier` | 模型服务档位，例如部分模型支持 `default`、`fast` 或 `auto`。不是 CLI 套餐选择。 | `service_tier.support_value`、`service_tier.default` |

上述选值仅是示例，不是统一枚举。`speed` 的选值也应直接读取目标模型的返回值。`protocol` 可作为模型协议信息读取，不代表允许用户修改服务端管理的 `Protocol` 字段。

例如某次返回 `Key: "thinking.support_value", Data: ["enabled", "disabled"]` 与 `Key: "thinking.default", Data: ["enabled"]`，表示该目标模型声明这两个可选值，默认值为 `enabled`；不能推广到其他模型。

## 使用规则

- 按 `Result[].Key` 查找对应 `Data`；默认值通常为单元素数组，不能把 supported values 的第一个元素当默认值。
- 未返回、空数组或默认值与支持值矛盾时，说明实际返回情况，不猜测缺失枚举或默认值。
- 不存在的模型或 provider 也可能成功返回已知 key、但 `Data` 全为空；未知 key 可能返回 `Result: []`。因此 metadata 查询不能作为模型存在性校验。
- metadata 是配置指引，不新增客户端业务拦截。用户明确提供的参数仍按创建、override、upgrade 各自的契约交给接口校验；保存成功也不能反推该值是 metadata 声明支持的值。
- 不自动补写用户没要求的参数或默认值。尤其 upgrade 的省略字段按 upgrade 基线规则保留，不能用 metadata 默认值覆盖。
- 用户切换模型时重新查询新模型，不能沿用旧模型的参数范围。查询成功不代表模型已开通，也不证明推理调用一定可用。

参数写入位置和覆盖语义见 [Agent](agent.md)、[Session overrides](session-files.md)、[Session upgrade](session-upgrade.md)。
