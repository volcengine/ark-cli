# arkcli models custommodel available-quantizations

查可用量化模式（quantize 前必跑）。

## Usage

```bash
arkcli models custommodel available-quantizations <id>
```

## Flags

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<id>` | 是 | string | 自定义模型 ID `cm-xxxxx` |

## Output

返回 JSON，包含：

- `quantizations`：可用量化方式列表（如 `["W8A8"]`，具体集合随 base model 而异）
- `supported_inference_types_by_quantization`：字典形式的前置部署方式预判，key 是量化方式，value 是该方式支持的部署/付费形态（如 `{"W8A8":["token","model_unit"]}`）

**注意**：
- 只读命令，可放心调用
- 不要从其他 base model 的可用列表"推断"
- 若返回空列表，说明该 base model 不支持量化——继续原模型部署即可
- `supported_inference_types_by_quantization` 来自 ArkModels 的 `QuantSupportedMethods` 元数据；若后端没有返回该字段，字典可能为空。创建后的最终可部署形态仍以 `custommodel get <new-id> --transform artifact_types` 为准
