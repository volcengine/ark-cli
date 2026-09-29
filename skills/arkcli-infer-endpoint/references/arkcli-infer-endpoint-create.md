# arkcli infer endpoint create

创建一个推理接入点（Endpoint）。

## Usage

```bash
arkcli infer endpoint create [flags]
```

## Flags

| Flag | Type | Description | Required |
|------|------|-------------|----------|
| `--model` | string | 模型 ID（基础模型或自定义模型） | Yes |
| `--name` | string | Endpoint 名称 | Yes |
| `--billing-method` | string | 计费 / 推理方式；当前仅支持 `token`，不传时保持默认行为 | No |
| `--rpm` | int | 速率限制 RPM | No |
| `--tpm` | int | 速率限制 TPM | No |
| `--dry-run` | bool | 输出本地创建计划，不调用模型查询、开通或创建 API | No |
| `-h`, `--help` | | help for create | No |

Client Preview 的 `ModelReference` 与模型开通检查依赖在线信息，因此输出为
`fidelity=partial` 并列出 `unresolved`。它不是服务端 validation，也不证明模型
已开通或请求会被接受。


## 模型未指定时

`--model` 是必填参数。用户已经明确选择脚本化 / CI / 无护栏 raw CRUD，但没有给模型或只给品牌、系列、家族名时，不要凭记忆补全，也不要先用缺失参数试跑 `--dry-run`：

1. 读取 [`../SKILL.md` 的「raw CRUD 创建前的模型澄清」](../SKILL.md#raw-crud-创建前的模型澄清)。
2. 只执行一次有界的实时 `arkcli models search [<keyword>] --size 10 --format json`，从本轮结构化结果形成 0/1/N 候选。
3. 多候选优先用宿主的结构化选择能力让用户选择；选项只能来自本轮返回，禁止全量 `models list`、逐候选 `models get` 或凭记忆补版本。
4. 选定唯一完整模型 ID 后再运行本地 `--dry-run`。用户只要求预览时停止；真实创建继续遵守共享写操作守卫。

普通“创建 / 部署 Endpoint”意图仍应走 [`arkcli-deploy`](../../arkcli-deploy/SKILL.md)；本节不能把普通创建降级成无护栏 raw CRUD。

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

## 示例

```bash
# 创建 Endpoint
arkcli infer endpoint create \
  --model doubao-seed-2-0-pro-260215 \
  --name my-endpoint

# 显式要求 token 推理方式（当前与不传时的创建 payload 保持一致，但会先校验模型是否支持 token）
arkcli infer endpoint create \
  --model doubao-seed-2-0-pro-260215 \
  --name my-token-endpoint \
  --billing-method token


# 典型返回
{
  "Id": "ep-20260421180049-ngwkm"
}

# 只提取 Id，供后续命令复用
endpoint_id=$(arkcli infer endpoint create \
  --model doubao-seed-2-0-pro-260215 \
  --name my-endpoint \
  --transform Id)
```

## 返回值

创建成功后返回新建 Endpoint 的 `Id`。例如：

```json
{
  "Id": "ep-20260421180049-ngwkm"
}
```

这个 `Id` 是后续操作的主入口。

## 创建后的下一步

### 1. 查看详情

```bash
arkcli infer endpoint get "$endpoint_id"
```

### 2. 管理生命周期

```bash
arkcli infer endpoint stop "$endpoint_id"
arkcli infer endpoint start "$endpoint_id"
```

### 3. 生成调用示例

`arkcli +code-example` 已迁到 OpenTOP，当前可用。按**基础模型名**生成示例（不接受 `--endpoint-id`）：

```bash
arkcli +code-example --model doubao-seed-2-0-pro --language python
```

详见 [`../../arkcli-code-example/SKILL.md`](../../arkcli-code-example/SKILL.md)。

### 4. 不要再调用 `+deploy`

如果你已经通过 `infer endpoint create` 拿到了 `Id`，说明 Endpoint 已经创建完成。

此时不应该再执行：

```bash
arkcli +deploy ...
```

因为 `+deploy` 本身也是“创建 Endpoint”的工作流，它会创建新的资源，而不是复用刚创建出来的 `Id`。

## 注意事项

- `infer endpoint create` 偏向标准资源创建
- `+deploy` 偏向”创建 Endpoint + 模型开通/复用检查/profile 同步”的任务工作流
- 二者都会创建 Endpoint，不应串行重复执行
- 语音模型（TTS / ASR / 配音 / 播客 / 音色 / 实时语音交互，或 `doubao-seed-tts-*` / `doubao-seed-asr-*` / `seedasr-*`）当前不能用本命令创建 Endpoint；只能通过 `arkcli models search <keyword>` 做广场发现和选型说明
- `--billing-method` 目前只有 `token` 一个枚举值；传 `token` 时不会额外写入 CreateEndpoint 请求字段，只会在创建前校验对应模型是否支持 token 推理方式
- 基础模型的支持方式来自 `ArkModels.data[].Features.ShareService`；自定义模型优先看可部署版本 `EndpointSupportedMethods.ShareService`，必要时兜底到 `AvailableDeploymentTypes=Shared`

## 参考

- [arkcli-infer-endpoint](../SKILL.md) -- infer endpoint 能力概览
- [arkcli-infer-endpoint-get](arkcli-infer-endpoint-get.md) -- 查看 Endpoint 详情
- [arkcli-infer-endpoint-start](arkcli-infer-endpoint-start.md) -- 启动 Endpoint
- [arkcli-infer-endpoint-stop](arkcli-infer-endpoint-stop.md) -- 停止 Endpoint
- [arkcli-code-example](../../arkcli-code-example/SKILL.md) -- 生成多语言调用示例（已迁到 OpenTOP，当前可用）
