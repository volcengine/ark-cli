---
name: image-stream
description: arkcli +gen --stream 图片流式输出。每行一个 image_generation.* 事件 JSON (NDJSON), 适合做增量预览 / 可视化进度 / 结合 jq 提取每帧 URL。
---

# +gen --stream（图片流式）

`+gen` 默认是同步等待整个生成完成；`--stream` 让 CLI 切换到 NDJSON 输出，把服务端 SSE 流式事件原样吐到 stdout，每行一个事件。适合：

- 多图生成场景 (`--image-count >1`)，想要看到第一张就立刻处理
- 接 jq / 脚本流水线，做增量预览或并行下载
- 调试服务端流式行为

## 适用范围

**仅图片任务**。视频任务走的是异步 task + polling 模型（`+gen` 自身轮询），没有 SSE 通道；传 `--stream` + 视频模型会被 `+gen` 前置拦截：

```
--stream only applies to image generation tasks
hint: remove --stream (video tasks use async polling, not SSE),
      or pass --modality image with an image-capable model / endpoint
```

## 事件协议

底层走 SDK `arkruntime.Client.GenerateImagesStreaming`，三类事件：

| 事件类型 | 何时发 | 携带字段 |
|---|---|---|
| `image_generation.partial_succeeded` | 每张图片生成完成 | `url` / `b64_json` / `size` / `image_index` / `model` / `created` |
| `image_generation.partial_failed` | 单张图片生成失败（不一定终止流式） | `error.code` / `error.message` |
| `image_generation.completed` | 流式终结 | 可选 `usage{generated_images, output_tokens, total_tokens}` / `tools[]` |

CLI 输出每行一个 `ImageStreamDelta` JSON（service-layer 视图），关键字段：

```json
{
  "event_type": "image_generation.partial_succeeded",
  "model": "doubao-seedream-5-0-260128",
  "created": 1700000000,
  "image_index": 0,
  "url": "https://...",
  "size": "1024x1024"
}
```

注意：
- `partial_failed` 的 `error.code/message` 在 ImageStreamDelta 里是 `error_code` / `error_message` 两个 flat 字段
- `completed` 事件的 `usage` / `tools` 在 ImageStreamDelta 里改名为 `final_usage` / `final_tools`
- `b64_json`（base64 编码图片）字段只有当 ResponseFormat 为 `b64_json` 时才出现（默认 `url`）

## 命令速查

```bash
# 1. 流式生成单张图（看第一帧就能下手处理）
arkcli +gen "neon kitten" --model doubao-seedream-5-0-260128 --stream

# 2. 流式生成 3 张图，按顺序拿 URL
arkcli +gen "5 different futuristic cars" \
  --model doubao-seedream-5-0-260128 \
  --image-count 3 --stream \
  | jq -r 'select(.event_type=="image_generation.partial_succeeded") | .url'

# 3. 流式 + 解 base64（要先把 b64_json 切到 stream 上）
arkcli +gen "blue laptop" \
  --model doubao-seedream-5-0-260128 --stream \
  | jq -r 'select(.b64_json != "") | .b64_json' \
  | base64 -d > out.png

# 4. 监控错误事件 + 终结 usage
arkcli +gen "..." --model ... --stream \
  | jq 'select(.event_type=="image_generation.partial_failed" or .event_type=="image_generation.completed")'
```

## 与 `--save-to` 的关系

`--save-to` 是同步路径上的"自动下载"功能；`--stream` **不会触发自动下载**。

理由：流式的目的是让用户增量处理输出，CLI 不应该并行落盘干扰。需要同时落盘 + 流式监控时，自己 pipe 流式输出到下载脚本即可（见上面例 3）。

## 与 raw API 的关系

`+gen --stream` 等价于：

```bash
arkcli api arkruntime.generate_images_stream \
  --params '{"model":"...","prompt":"...","stream":true,"response_format":"url"}'
```

但 raw API 接受不到 `arkcli +gen` 那一层的 ParseAssetRefs / collapseImageAssets / modality 路由能力。普通使用走 `+gen --stream` 即可，raw API 只在调试 SDK 字段或测试服务端行为时使用。

## 不支持流式的场景

| 场景 | 为什么 | 替代方案 |
|---|---|---|
| 视频生成 (`seedance-*` / `--modality video`) | 视频走异步 task + poll 模型，无 SSE 通道 | 不传 `--stream`；用 `+gen` 默认轮询 |
| 同步等待累计结果 | `--stream` 模式下 stdout 是 NDJSON 流，不输出累计 GenerateOutput | 不传 `--stream`；走默认同步路径 |

## 常见错误

| 现象 | 原因 |
|---|---|
| `--stream only applies to image generation tasks` | 模型是视频族 (`seedance-*`)，或 `--modality video` 显式指定 |
| 流式输出第一行就是 `partial_failed` 后再无输出 | 上游真正失败了；看 `error_code` 字段；尤其常见于 `--size 1024x1024` 这种小尺寸 |
| 看不到 `completed` 事件 | `--image-count 1` 时服务端可能不发 completed，单张就是终结；多张才会发 |
