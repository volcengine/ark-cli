# arkcli models custommodel quantize

量化（异步）。

## Usage

```bash
arkcli models custommodel quantize <id> --quantization <mode> [flags]
```

## Examples

```bash
# 标准用法（必须先跑 available-quantizations 确认 mode 合法，并查看该 mode 预期支持的部署形态）
arkcli models custommodel quantize cm-xxxxx --quantization int8

# 同时附描述
arkcli models custommodel quantize cm-xxxxx \
  --quantization int8 \
  --description "int8 quantized for low-latency inference"
```

## Flags

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<id>` | 是 | string | 待量化的源模型 ID（必须 `status=ready`） |
| `--quantization` | 是 | string | 量化模式；**必须从 `available-quantizations <id>` 返回的集合里选** |
| `--description` | 否 | string | 量化结果模型的描述，最多 300 字符 |

## Output

返回 JSON，含新生成的 `cm-yyyyy`（与源不同）和初始 status。

**注意**：
- **异步**任务：返回成功只代表受理，必须 `custommodel get <new-id>` 轮询新 ID 至 `ready`
- 量化结果是**独立的新自定义模型**——源模型保留不变，账号下会同时存在两个 cm-xxxxx
- 量化模式选择对推理性能/精度影响较大，建议生产前两边各部署一个 endpoint 做对照
- 创建前如需判断某个量化方式是否支持 token / 模型单元部署，先看 `available-quantizations` 返回的 `supported_inference_types_by_quantization`

## 资源关系

量化链路里有三类资源，ID 不可混用：

| 资源 | ID 形态 | 说明 |
|------|---------|------|
| 源自定义模型 | `cm-xxxxx` | `quantize` 的输入，保持不变 |
| 量化结果模型 | 新的 `cm-yyyyy` | `quantize` 返回的新模型，需要继续 `get` 轮询到 `ready` |
| 推理 endpoint | `ep-xxxxx` | 用 ready 的源模型或量化模型经过 `+deploy` 创建，`+chat` / `+gen` 应使用 endpoint 或可调用模型 ID |

创建前可用 `available-quantizations` 返回的 `supported_inference_types_by_quantization` 预判某个量化方式支持的部署形态；创建后用 `custommodel get <new-id> --transform artifact_types` 返回的 `supported_inference_types` 做最终确认。
