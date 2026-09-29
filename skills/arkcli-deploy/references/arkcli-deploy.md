# +deploy 详细参考

> **前置：** 先读 [`../SKILL.md`](../SKILL.md)。本文件只补充上面没写的 flag 细节、JSON 示例、错误码。

## Agent 必读要点（不要跳过）

1. 子命令穷举：只有 `arkcli +deploy`。**不存在** `arkcli deploy ...` / `arkcli endpoint create` / `arkcli +deploy create`。
2. **写操作 + 计费**：执行即真实创建；该工作流不支持 `--dry-run`。
3. JSON 类 flag 字段名一律 **PascalCase**：`Rpm`、`Tpm`、`Strategy`、`Mode`，不是小写。
4. **`+code-example` 已迁到 OpenTOP，当前可用**：`arkcli +code-example --model <model-id> --language python`（按基础模型名生成，不接受 `--endpoint-id`）；详见 [`../../arkcli-code-example/SKILL.md`](../../arkcli-code-example/SKILL.md)。
5. `--model cm-xxxxx` 真实执行时会先复用已有 Running Endpoint；只有找不到可复用 Endpoint 时才创建。
6. 新建成功后会 best-effort 生成示例到 `./ark-examples/<endpoint-id>/`；示例获取或写入失败只在 stderr 提示，不回滚 Endpoint。必须确认实际文件存在，不能仅凭创建成功承诺示例已生成。复用已有 Endpoint 不走新建后的示例生成步骤。

## 创建前的业务核对

先完成模型消歧，再复用本轮精确版本的模型详情；模型下线或版本不存在时停止，
不能用开通动作、改 prompt 或反复创建来修复无效版本。自定义模型仍保持真实 `cm-...`
身份，不把基础模型 lineage 当成部署目标。

- `--tags` 是 `[{"Key":"env","Value":"prod"}]` 对象数组，不是 `{"env":"prod"}`。
- JSON flags 用单引号包住完整 JSON；`--view summary` 等单值 flag 必须有值。
- 本地语法正确不代表服务端业务校验会通过；`+deploy` 不支持离线 preview，
  不得用不存在的 `--dry-run` 代替真实的只读核对。
- 基准中 moderation 的业务经验保留为 Volc 选型提示：普通文本优先 `Default`，
  视觉模型需核对 `Default` / `Basic`，图像/视频生成未明确要求时优先省略走服务端默认。
  这不是所有版本的封闭允许值表；CLI DTO 还定义了 `Customized` 等策略。
  用户要求关闭审核时，不把口语机械映射为 `Skip` 或静默换成其他策略，应说明目标模型
  的支持边界并核对实际错误。模型类别与策略兼容性以该目标的权威信息为准。
- 限流读取精确模型详情的 `rate_limit` 及当前账号配额；区别 RPM/TPM、并发、
  创建任务速率，不能把图像/视频模型的任务限流当成文本 RPM。模型目录上限不等于
  当前账户剩余配额。用户说“不限流”时不自行写 `Rpm:-1`；
  返回的负值/缺失字段不能直接当作可提交值。核对可支持的值，保留用户明确目标，
  无法满足时说明约束，不通过省略 flag 冒充已设置无限额度。

## 命令模板

```bash
# 最简部署（真实创建）
arkcli +deploy --name my-endpoint --model doubao-seed-2-0-pro-260215

# 执行前复述 model/name/region 并取得明确确认

# 带速率限制（注意 PascalCase 字段名）
arkcli +deploy --name my-endpoint --model doubao-seed-2-0-pro-260215 \
  --rate-limit '{"Rpm": 60, "Tpm": 10000}'


# 带审核
arkcli +deploy --name my-endpoint --model doubao-seed-2-0-pro-260215 \
  --moderation '{"Strategy": "Basic"}'

# 带智能路由
arkcli +deploy --name my-endpoint --model doubao-seed-2-0-pro-260215 \
  --intelligent-router '{"Strategy": "Balanced", "Mode": "Automatic"}'

# 使用另一套已存在的项目上下文
arkcli --profile my-project-profile +deploy --name my-endpoint --model doubao-seed-2-0-pro-260215

# 自定义模型：若已有 Running Endpoint 会直接复用，否则创建
arkcli +deploy --name my-custom-endpoint --model cm-xxxxx
```

## 参数

### 必填

| 参数 | 类型 | 说明 |
|------|------|------|
| `--name` | string | 接入点名称 |
| `--model` | string | 模型 ID（基础模型或自定义模型） |

### 常用可选

| 参数 | 类型 | 说明 |
|------|------|------|
| `--description` | string | 接入点描述 |
| `--rate-limit` | JSON string | 速率限制，如 `{"Rpm": 60, "Tpm": 10000, "Ipm": 100}` |
| `--moderation` | JSON string | 审核配置。Strategy: `Basic` / `Customized` / `Default` / `Skip` |
| `--view` | string | 创建后查看 Endpoint 详情 |









### 高级配置

| 参数 | 类型 | 说明 |
|------|------|------|
| `--model-unit-id` | string | 模型单元 ID |
| `--batch-only` | bool | 仅批量推理模式 |
| `--data-delivery` | bool | 数据交付模式 |
| `--domain` | string | 接入点 Domain（不指定时自动根据模型推断） |
| `--dedicated-to-ptu` | bool | 专用于 PTU |
| `--dedicated-to-dynamic-ptu` | bool | 专用于动态 PTU |
| `--content-generation` | JSON string | 内容生成配置 |
| `--inference-foundry` | JSON string | 推理工厂配置 |
| `--tags` | JSON string | 标签，如 `[{"Key": "env", "Value": "prod"}]` |
| `--allow-data-collected` | bool | 允许数据收集 |
| `--service-info` | JSON string | 服务信息 |
| `--need-watermark` | bool | 需要水印 |
| `--watermark-info` | JSON string | 水印信息 |
| `--intelligent-router` | JSON string | 智能路由。Strategy: `Balanced` / `CostFirst` / `EffectFirst`。Mode: `Automatic` / `CandidateSet` / `Ordered` |
| `--is-intelligent` | bool | 是否启用智能路由 |
| `--fallback` | JSON string | 降级配置 |
| `--limit-coefficient` | JSON string | 限制系数 |
| `--deployment-type` | string | 部署类型 |
| `--is-agentic` | bool | 是否为 Agentic |
| `--agentic-strategy` | JSON string | Agentic 策略。Mode: `Auto` / `Custom` |
| `--is-aicc` | bool | 是否为 AICC |
| `--specify-region` | string | 指定区域 |

## 返回值

stdout 返回 `endpoint_id`、`model_type`、`status`、`reused`；示例代码不是 stdout 字段。
按 `reused` 区分“新建”与“复用”，不要把复用已有资源说成创建了新资源。
`status=created` 只证明创建流程返回成功，不证明已完成一次数据面推理；需要实时状态时只读查询
`infer endpoint get`。只有实际执行了用户要求的推理验证并取得有效输出，才能说“调用验证通过”。
示例落盘情况检查 stderr 与实际文件；示例缺失、默认资源设置失败等后置问题应单独报告，
不得为了修复它们再次运行 `+deploy`。

## 常见错误

| 错误 | 原因 | 处理 |
|------|------|------|
| 缺少必填参数 | 未提供 `--name` 或 `--model` | 必须指定 |
| JSON 格式错误 | `--rate-limit` 等 JSON 参数格式不对 | 单引号包裹，字段用 PascalCase |
| 模型不存在 | `--model` ID 无效 | `arkcli models search <keyword>` 找正确名字 |



## 参考

- [arkcli-deploy](../SKILL.md) -- skill 入口
- [arkcli-shared](../../arkcli-shared/SKILL.md) -- 共享认证与全局参数
