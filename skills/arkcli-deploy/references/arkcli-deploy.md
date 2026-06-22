# +deploy 详细参考

> **前置：** 先读 [`../SKILL.md`](../SKILL.md)。本文件只补充上面没写的 flag 细节、JSON 示例、错误码。

## Agent 必读要点（不要跳过）

1. 子命令穷举：只有 `arkcli +deploy`。**不存在** `arkcli deploy ...` / `arkcli endpoint create` / `arkcli +deploy create`。
2. **写操作 + 计费**：默认就是真实创建，`--dry-run` 才是预演。
3. JSON 类 flag 字段名一律 **PascalCase**：`Rpm`、`Tpm`、`Strategy`、`Mode`，不是小写。
4. **`+code-example` 已迁到 OpenTOP，当前可用**：`arkcli +code-example --model <model-id> --lang python`（按基础模型名生成，不接受 `--endpoint-id`）；详见 [`../../arkcli-code-example/SKILL.md`](../../arkcli-code-example/SKILL.md)。
5. `--model cm-xxxxx` 真实执行时会先复用已有 Running Endpoint；只有找不到可复用 Endpoint 时才创建。
6. ~~创建成功后示例代码落到 `./ark-examples/<endpoint-id>/`~~ — 0.1.16 暂不生成示例文件。

## 命令模板

```bash
# 最简部署（不带 --dry-run = 真实创建）
arkcli +deploy --name my-endpoint --model doubao-seed-2-0-pro-260215

# 推荐：先 --dry-run 预演
arkcli +deploy --name my-endpoint --model doubao-seed-2-0-pro-260215 --dry-run

# 带速率限制（注意 PascalCase 字段名）
arkcli +deploy --name my-endpoint --model doubao-seed-2-0-pro-260215 \
  --rate-limit '{"Rpm": 60, "Tpm": 10000}'

# 带审核
arkcli +deploy --name my-endpoint --model doubao-seed-2-0-pro-260215 \
  --moderation '{"Strategy": "Basic"}'

# 带智能路由
arkcli +deploy --name my-endpoint --model doubao-seed-2-0-pro-260215 \
  --intelligent-router '{"Strategy": "Balanced", "Mode": "Automatic"}'

# 指定项目
arkcli +deploy --name my-endpoint --model doubao-seed-2-0-pro-260215 --project-name my-project

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
| `--project-name` | string | 项目名称 |
| `--rate-limit` | JSON string | 速率限制，如 `{"Rpm": 60, "Tpm": 10000, "Ipm": 100}` |
| `--moderation` | JSON string | 审核配置。Strategy: `Basic` / `Customized` / `Default` / `Skip` |
| `--dry-run` | bool | 试运行模式，不实际创建 |
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

创建成功后返回 Endpoint 信息及多语言调用示例代码（落到 `./ark-examples/<endpoint-id>/`）。

## 常见错误

| 错误 | 原因 | 处理 |
|------|------|------|
| 缺少必填参数 | 未提供 `--name` 或 `--model` | 必须指定 |
| JSON 格式错误 | `--rate-limit` 等 JSON 参数格式不对 | 单引号包裹，字段用 PascalCase |
| 模型不存在 | `--model` ID 无效 | `arkcli models search <keyword>` 找正确名字 |

## 参考

- [arkcli-deploy](../SKILL.md) -- skill 入口
- [arkcli-shared](../../arkcli-shared/SKILL.md) -- 共享认证与全局参数
