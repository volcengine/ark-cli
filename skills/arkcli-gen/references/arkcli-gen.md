# +gen

> **前置条件：** 先阅读 [`../arkcli-shared/SKILL.md`](../../arkcli-shared/SKILL.md) 了解认证、全局参数和安全规则。

图片/视频生成的执行层文档（`+gen` 全参数）。**这是三步工作流的第 3 步**；完整工作流（① `resources list` 查 profile 模型 → ② `models get` 查 supported_params → ③ `+gen` 生成）见 [`../SKILL.md`](../SKILL.md)。

> **⚠️ 视频任务默认异步**：提交即返回 `task_id`（`status: queued`），用 `arkcli gen get <task_id>` 轮询；要同步阻塞加 `--wait`。图片任务同步返回。

## 命令

> **⚠️ `--model` 必须是 `<name>-<primary_version>` 完整形式（如 `doubao-seedream-5-0-260128`、`doubao-seedance-1-5-pro-251215`）或 Endpoint ID（`ep-xxx`）。直接传族名会 404 `InvalidEndpointOrModel.NotFound`。**
>
> `primary_version` 格式不固定——常见 6 位日期，也可能是 8 位日期、带限定前缀、短数字、甚至空串（此时 full ID 就是族名本身）。约半数模型非 6 位，详见 [`../../arkcli-models/SKILL.md`](../../arkcli-models/SKILL.md) 链路 0。**不要自行正则判断"是否像完整 ID"**。
>
> **调用前补全（必须）：** 若不确定 `--model` 是否已是完整形式，先执行：
>
> ```bash
> VER=$(arkcli models get <name> --transform 'primary_version' | tr -d '"')
> # --transform 输出带 JSON 双引号，必须 tr -d 剥掉；否则 $VER 会是 "260128" 导致拼接错误
> MODEL=${VER:+<name>-$VER}; MODEL=${MODEL:-<name>}   # VER 空串时退回 <name>
> arkcli +gen --model "$MODEL" "<prompt>"
> ```
>
> 若刚 `models search/list` 过，可直接复用返回里的 `primary_version` 字段，无需再调 `get`。

```bash
# 1) 文生图 (T2I) — seedream 模型
arkcli +gen --model doubao-seedream-5-0-260128 "简约商务笔记本电脑，深蓝色，白色背景"

# 2) 图生图 / image-edit (I2I) — 用 --input 艾特参考图
arkcli +gen --model doubao-seedream-5-0-260128 \
  --input @ref.jpg \
  "保留构图，把背景换成黄昏的海滩"

# 多张参考图（多张同时传 --input）
arkcli +gen --model doubao-seedream-5-0-260128 \
  --input @style1.jpg --input @style2.jpg \
  "融合两张图的风格，主体是一只柴犬"

# 一次出多张候选（image-count > 1 自动启用 sequential）
arkcli +gen --model doubao-seedream-5-0-260128 --image-count 4 "未来城市天际线 4 个风格变体"

# 3) 文生视频 (T2V)
arkcli +gen --model doubao-seedance-1-5-pro-251215 "一只柴犬在樱花树下奔跑，慢镜头"

# 4) 图生视频 / 首帧到视频 (I2V) — 第 1 张图默认作为首帧
arkcli +gen --model doubao-seedance-1-5-pro-251215 \
  --input @first.jpg \
  "镜头缓慢拉远，主体保持不动"

# 显式指定 first / last 帧
arkcli +gen --model doubao-seedance-1-5-pro-251215 \
  --input first:@start.jpg --input last:@end.jpg \
  "从这两帧之间生成补间动画"

# 5) 参考视频 (R2V) — 把视频当作参考素材
arkcli +gen --model doubao-seedance-2-0-r2v-260128 \
  --input ref:@reference.mp4 \
  "保持参考视频的运动轨迹，替换主角为机器人"

# 6) 参考音频 — 让视频节奏与音频同步
arkcli +gen --model doubao-seedance-2-0-260128 \
  --input ref:@beat.mp3 \
  "随节拍切换的城市夜景蒙太奇"

# 进阶 flag — 图片任务
arkcli +gen --model doubao-seedream-5-0-260128 \
  --guidance-scale 7.5 --optimize-prompt --output-format png \
  "产品摄影风格"

# 进阶 flag — 视频任务
arkcli +gen --model doubao-seedance-2-0-260128 \
  --frames 96 --camera-fixed --return-last-frame --draft \
  "城市夜景航拍"

# 输出完整 JSON
arkcli +gen --model doubao-seedance-1-5-pro-251215 --format json "产品广告视频"

# 自定义推理接入点 (endpoint ID) — 必须显式 --modality
arkcli +gen --model ep-20260416234150-zsd4v --modality image "简约商务笔记本"
arkcli +gen --model ep-20260416234150-zsd4v --modality video "一只柴犬奔跑"
```

## 参数

| 参数 | 必填 | 类型 | 说明 |
|------|------|------|------|
| `<prompt>` | 是 | positional | 生成提示词（位置参数，放在命令最后） |
| `--model` | 否 | string | **完整版本化模型 ID**（如 `doubao-seedream-5-0-260128`、`doubao-seedance-1-5-pro-251215`）或推理接入点 ID（`ep-xxx`）。仅传族名会直接 404 `InvalidEndpointOrModel.NotFound`。**0.1.16+ 可省略**：缺省时按 `--modality` fallback 到 active profile 的 `Resources.<modality>.Default`，未设时报 hint 引导 `arkcli profile set-default --modality <m> <id>` |
| `--modality` | 见说明 | string | 生成模态：`image` 或 `video`。使用 `seedream-*` / `seedance-*` 模型时可自动推断；使用 endpoint ID 时**必填** |
| `--input` | 否 | string（可重复） | 参考素材引用，按出现顺序进入 content[]。本地文件 `@<path>` / 远程 `https://...` `tos://...` 都可。可选 role 前缀 — 简写：`first:` `last:` `ref:` `none:`（none 显式忽略，简写 wire 上不传 role 让服务端按位置推断）；SDK 显式：`first_frame:` `last_frame:` `reference_image:` `reference_video:` `reference_audio:`（这些会真正写到 wire `content[].role` 字段）。**图片任务**：折叠为 image union；**视频任务**：第 1 张图默认首帧，其它图为参考图，视频→ref_video，音频→ref_audio |
| `--name` | 否 | string | 任务名覆盖 |
| `--version` | 否 | string | 模型版本覆盖 |
| `--project-name` | 否 | string | 项目名称（全局 flag） |
| `--size` | 否 | string | 图片输出尺寸，如 `1920x1920`；像素数过小时会被后端拒绝 |
| `--image-count` | 否 | int | 图片任务输出张数；`>1` 时自动转为 `sequential_image_generation=auto + max_images=N` |
| `--n` | 否 | int | `--image-count` 的别名（图片任务）；两者同传时 `--n` 优先 |
| `--ratio` | 否 | string | 输出宽高比覆盖，如 `16:9` / `9:16` / `1:1`（视频任务） |
| `--resolution` | 否 | string | 输出分辨率，如 `480p`、`720p`、`1080p`（视频任务） |
| `--duration` | 否 | int | 视频时长（秒） |
| `--frames` | 否 | int | 视频帧数（在支持的模型上覆盖 duration） |
| `--seed` | 否 | int | 随机种子，相同 seed 可复现结果 |
| `--watermark` | 否 | bool | 是否添加水印 |
| `--generate-audio` | 否 | bool | 是否同步生成音频（视频任务） |
| `--guidance-scale` | 否 | float | 图片任务的 classifier-free guidance scale，如 `7.5` |
| `--optimize-prompt` | 否 | bool | 图片任务：启用服务端 prompt 优化 |
| `--output-format` | 否 | string | 图片输出格式：`jpeg` 或 `png` |
| `--response-format` | 否 | string | 图片响应格式：`url`（默认）或 `b64_json`（直接返回 base64 编码图片，不存 URL） |
| `--prompt-thinking` | 否 | string | 图片任务：prompt 优化思考模式 `auto` / `enabled` / `disabled`（仅在 `--optimize-prompt=true` 时生效） |
| `--prompt-mode` | 否 | string | 图片任务：prompt 优化执行模式 `standard` / `fast` |
| `--sequential` | 否 | string | 图片任务：序列图模式 `auto` / `disabled`。留空时延续 `--image-count > 1` → `auto` 的自动行为；显式传 `disabled` 强制单图，即使 `--image-count > 1` |
| `--stream` | 否 | bool | 图片任务：流式 NDJSON 输出，详见 [`image-stream.md`](image-stream.md) |
| `--camera-fixed` | 否 | bool | 视频任务：固定虚拟镜头 |
| `--return-last-frame` | 否 | bool | 视频任务：返回最后一帧 URL（用于续接生成） |
| `--draft` | 否 | bool | 视频任务草稿模式：更快 / 更便宜 / 质量更低 |
| `--priority` | 否 | int | 视频任务调度优先级 0-9，越高越优先。**受 supported_params 约束**——Step 2 先查模型是否支持及范围（实测 seedance-2.0 / 2.0-fast 支持 `[0,9]`，1.5-pro 不支持）；模型不支持时传了会被校验拒 |
| `--service-tier` | 否 | string | 视频任务：服务等级 |
| `--safety-id` | 否 | string | 视频任务：调用方传入的安全标识 |
| `--execution-expires-after` | 否 | int | 视频任务服务端 TTL（秒） |
| `--callback-url` | 否 | string | 视频任务：服务端在生命周期事件（created/running/succeeded/failed）上 POST 通知到此 URL |
| `--wait` | 否 | bool | **视频任务**：阻塞到任务完成再返回。**默认 false**——提交即返回 `task_id` 异步轮询（2.0 起默认行为，旧版默认同步等待） |
| `--extra-body` | 否 | string（JSON 对象） | 视频任务 forward-compat 通道：传 JSON 对象字符串，里面的 key 会 merge 到 top-level 请求 body，让你不升级 arkcli 也能透传服务端新增字段。例：`--extra-body '{"new_field":"value"}'` |
| `--tools` | 否 | string（可重复） | 工具开关，目前支持 `web_search` |
| `--save-to` | 否 | string | 保存生成产物的本地目录，默认 `.`（当前目录）；传 `--save-to=""` 显式关闭自动下载。下载失败不阻塞主流程 |

## 返回值

**默认输出**（精简）：

```json
{
  "kind": "image",
  "model": "doubao-seedream-5-0-260128",
  "status": "succeeded",
  "output_url": "https://...",
  "output_urls": ["https://..."],
  "local_path": "/work/ark-gen.jpeg",
  "local_paths": ["/work/ark-gen.jpeg"]
}
```

- `output_url` / `output_urls`：TOS 预签名 URL，**24 小时后失效**，不要长期引用
- `local_path` / `local_paths`：由 `--save-to` 触发自动下载后落盘的本地绝对路径；是**持久产物**，优先用它而不是 URL。关闭自动下载（`--save-to=""`）时这两个字段不存在
- `task_id`：视频任务才有。**视频默认异步**——提交即返回此 id + `status: queued`；用 `arkcli gen get <task_id>`（或 `arkcli api arkruntime.get_content_generation_task --params '{"id":"<task_id>"}'`）轮询到 `succeeded` 再取 `output_url`。要同步阻塞用 `+gen --wait`

**`--format json`**：输出完整的任务对象（不含 `local_path`）。

视频任务的完整对象除了 `id / status / output_url / ratio / resolution / duration / frames / generate_audio` 等常规字段外，还会回显服务端 echo 出的额外字段（按需出现，由模型与服务端决定）：

| 字段 | 说明 |
|---|---|
| `usage` | `{prompt_tokens, completion_tokens, total_tokens}` token 用量遥测 |
| `revised_prompt` | 服务端最终下发给模型的 prompt（经优化 / 安全过滤后） |
| `subdivisionlevel` | 任务细分等级（注意 JSON key 是单词式无下划线） |
| `fileformat` | 输出文件格式（同样单词式无下划线） |
| `safety_identifier` | 调用方传入的安全标识符回显 |
| `tools` | 任务实际用到的工具类型列表，例如 `["web_search"]` |

服务端不返回时这些字段不出现（omitempty），不影响 `output_url` / `local_path` 等核心字段的提取。

## 常见错误

| 错误 | 原因 | 处理方式 |
|------|------|---------|
| `model is required` | 未指定 `--model` | 必须指定模型名 |
| `Error code: 404 - InvalidEndpointOrModel.NotFound` | `--model` 传的是模型族名（如 `doubao-seedream-5-0`、`doubao-seedance-1-5-pro`），该族未注册族名别名 | 用 `arkcli models get <name> --transform 'primary_version'` 拿版本号，拼成 `<name>-<primary_version>` 再传 |
| 缺少 prompt | 未提供位置参数 | prompt 是必填的位置参数 |
| `image generation failed: InvalidParameter` | 图片尺寸像素数过小等参数错误；常见于 `1024x1024` 这类尺寸 | 改用 `1920x1920` 及以上尺寸，必要时加 `--debug` 看底层错误 |
| 视频迟迟没结果 | 视频默认**异步**，返回的是 `task_id` + `status: queued`（**非失败**） | 用 `arkcli gen get <task_id>` 轮询到 `succeeded`；**不要**重提 `+gen`（会建新任务）。要同步等可用 `+gen --wait` |
| `warn: auto-download failed: ...`（stderr） | 下载 `output_url` 失败（网络 / 403 等） | 不阻塞主流程；`output_url` 仍返回，用户可手动 `curl -o file.ext "<output_url>"` 抢救（注意 24h 时效） |

## 注意事项

- prompt 是位置参数，放在命令最后，建议用引号包裹
- `--input` 是数据驱动的多模态入口：图扩展名走图通道，视频扩展名走视频通道，音频扩展名走音频通道；其他扩展名在视频任务里**会被静默丢弃**（content-generation 没有 input_file slot），在图片任务里也不会被识别为参考图
- 视频任务的多 `--input` 顺序是有意义的：第一张图默认作为 first frame；如果要表达"这张是参考素材，不是首帧"，用 `ref:` 前缀
- 图片生成建议从 `1920x1920` 起步，`1024x1024` 这类尺寸当前可能直接被后端拒绝
- 图片生成通常几秒完成，视频生成可能需要数十秒到数分钟
- **预签名 URL 24h 失效**：`output_url` 只能在短时间内 HTTP `GET` 下载（注意不是 `HEAD`，签名绑定 method）。需要持久保存就依赖 `local_path` 或及时下载
- 默认文件名：有 `task_id`（视频）时取 task_id，否则用 `ark-gen`；Content-Type 推扩展名（`.jpeg` / `.mp4` 等）；同名自动 `-1`/`-2` 后缀
- 如需底层排查或自己掌控提交 + 轮询节奏：直接走 raw API：`arkcli api arkruntime.create_content_generation_task --params '{"model":"<id>","content":[{"type":"text","text":"<prompt>"},{"type":"image_url","image_url":{"url":"https://..."}}]}'` 拿 `id`，然后用 `arkruntime.get_content_generation_task --params '{"id":"<id>"}'` 轮询

## 参考

- [arkcli-gen](../SKILL.md) -- gen skill 概览
- [arkcli-shared](../../arkcli-shared/SKILL.md) -- 认证和全局参数
