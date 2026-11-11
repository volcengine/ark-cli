# arkcli models custommodel upload

从 TOS 导入（异步）。

> **前置条件：必须先开通火山引擎 TOS（对象存储）服务，并把模型权重上传到一个 bucket，才能拿到 `tos://<bucket>/<prefix>` URI 给 `--tos` 用。**
> - TOS 控制台：<https://console.volcengine.com/tos>
> - TOS 上传文档（命令行/SDK/控制台多种方式）：<https://www.volcengine.com/docs/6349/74820?lang=zh>
>
> Agent 处理流程分两种：
> - 用户已经给出 `tos://<bucket>/<prefix>`：继续确认 `--name`、`--base-model`、`--tos`，写操作确认后再执行 `upload`
> - 用户还没有 TOS URI：**不要**直接帮他跑 `upload`。先给出上面两条链接，让用户开通 TOS、创建 bucket、上传权重，拿到 URI 后回来继续
>
> bucket region 必须与当前 ARK region 一致。

## Usage

```bash
arkcli models custommodel upload --name <name> --base-model <foundation-model-id> --tos tos://<bucket>/<prefix> [flags]
```

## Examples

```bash
# 最简：必填三项
arkcli models custommodel upload \
  --name my-finetune-v1 \
  --base-model doubao-seed-1-6-251015 \
  --tos tos://my-bucket/finetune/v1/

# 带描述与上传时量化
arkcli models custommodel upload \
  --name my-finetune-v1-int8 \
  --base-model doubao-seed-1-6-251015 \
  --tos tos://my-bucket/finetune/v1/ \
  --quantization int8 \
  --description "lora-sft on customer support corpus, int8 quantized"

```

`upload` 会启动依赖 TOS 的异步导入，无法在不读取外部状态的条件下给出可靠计划，
因此不注册 `--dry-run`。复述并确认 `name/base-model/tos/quantization` 后再真实执行。

## 上传前检查

- CLI 分页读取 TOS 目录，检查目录顶层的 `config.json`、至少一个 `*.safetensors`，以及 `tokenizer.json` / `tokenizer_config.json` 之一（与控制台实际校验规则一致）。嵌套目录、相邻前缀不算；仅有 `tokenizer.model` 不满足当前规则。
- 目录无法读取、文件缺失或分页异常时停止，不调用 `UploadModel`。
- 查询当前账号的 `Foundry` 功能资格；仅已启用时调用 `CheckCustomModel`。`Supported=false` 会在输出 `message` 中展示原因，仍按指定基础模型上传，不会误判为链接无效。检查请求失败或响应不完整时停止，修复后重试。
- 检查不自动修改 `--base-model` / `--quantization`，也不代表异步导入已完成。不要用真实上传作为只读检查。

## Flags

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--name` | 是 | string | 自定义模型名（账号内唯一）；须以英文字母开头，后续支持中文、字母、数字、下划线、短横线 |
| `--base-model` | 是 | string | 基座模型 ID，如 `doubao-seed-1-6-251015`；可先用 `arkcli models search` 找到 |
| `--tos` | 是 | string | TOS URI：`tos://<bucket>/<prefix>`，指向权重目录（bucket 需先在 TOS 控制台创建并上传权重）|
| `--quantization` | 否 | string | 上传时附带量化模式（也可上传后再单独 `quantize`） |
| `--description` | 否 | string | 描述，最多 300 字符 |

## Output

返回 JSON，含新建的 `cm-xxxxx` 和初始 status（通常 `preparation`）。

**注意**：
- **异步**任务：返回成功只代表请求已受理，不代表权重已就绪。后续必须 `custommodel get <id>` 轮询直到 `status=ready` 才能 deploy / quantize
- TOS bucket 必须与当前 ARK region 同区域，跨 region 会失败
- 若上传时已带 `--quantization`，仓库里会同时存在原版和量化版两个 cm-xxxxx
