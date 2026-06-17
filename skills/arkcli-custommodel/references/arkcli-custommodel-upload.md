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

# dry-run 预览请求体（不实际提交）
arkcli models custommodel upload --name X --base-model Y --tos tos://b/p --dry-run
```

## Flags

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `--name` | 是 | string | 自定义模型名（账号内唯一）；须以英文字母开头，后续支持中文、字母、数字、下划线、短横线 |
| `--base-model` | 是 | string | 基座模型 ID，如 `doubao-seed-1-6-251015`；可先用 `arkcli models search` 找到 |
| `--tos` | 是 | string | TOS URI：`tos://<bucket>/<prefix>`，指向权重目录（bucket 需先在 TOS 控制台创建并上传权重）|
| `--quantization` | 否 | string | 上传时附带量化模式（也可上传后再单独 `quantize`） |
| `--description` | 否 | string | 描述，最多 300 字符 |
| `--dry-run` | 否 | bool | 预览请求体不提交 |

## Output

返回 JSON，含新建的 `cm-xxxxx` 和初始 status（通常 `preparation`）。

**注意**：
- **异步**任务：返回成功只代表请求已受理，不代表权重已就绪。后续必须 `custommodel get <id>` 轮询直到 `status=ready` 才能 deploy / quantize
- TOS bucket 必须与当前 ARK region 同区域，跨 region 会失败
- 若上传时已带 `--quantization`，仓库里会同时存在原版和量化版两个 cm-xxxxx
